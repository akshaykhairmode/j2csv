package converter

import (
	"bytes"
	"io"
	"regexp"

	"github.com/rs/zerolog"
)

type chanReader struct {
	c      chan []byte     //We will receive data on this channel after stripping comments
	excess *bytes.Buffer   //We will store excess bytes here and write them when space is available
	logger *zerolog.Logger //console logger
}

var (
	singleLineComment = regexp.MustCompile(`//(.*)`)
	multiLineComment  = regexp.MustCompile(`/\*[\s\S]*?\*/`)
)

var stopByte = []byte("}")

// New take an reader and returns another reader. Send 0 to create default size buffer. The new reader will receive data after removal of single line and multiline comments.
func New(inp io.Reader, sizeInBytes int, logger *zerolog.Logger) io.Reader {
	cw := &chanReader{
		c:      make(chan []byte, 50),
		excess: bytes.NewBuffer(nil),
		logger: logger,
	}

	go cw.startParsingInput(inp, sizeInBytes)

	return cw
}

func (cw *chanReader) Read(buf []byte) (int, error) {

	var retn int

	//first copy excess bytes from previous operation
	n, err := cw.excess.Read(buf)
	//Ignore EOF here as we would get false EOF in middle when our regex parsing is slow.
	if err != nil && err != io.EOF {
		cw.logger.Fatal().Err(err).Msg("error while writing to buffer")
	}

	//We filled the inp buffer
	//so we return here as we dont have any capacity available to read into.
	if n >= len(buf) {
		return n, nil
	}

	retn += n //increase total read bytes count

	//space remaining in buf, so lets get some data from our channel
	//Blocks till channel is open, so if parsing is slow
	//this read method will block till we get some data.
	data, isChanOpen := <-cw.c
	if !isChanOpen && len(data) <= 0 {
		return retn, io.EOF
	}

	//for the available capacity in buf, we will write data we received from the channel.
	n = copy(buf[n:], data)
	//its possible we may not be able to write everything to buffer
	//so the excess data which we could not write to buf will be written to excess buffer.
	cw.excess.Write(data[n:])

	retn += n //increase total read bytes count
	return retn, nil
}

func (cw *chanReader) startParsingInput(inp io.Reader, sizeInBytes int) {
	if sizeInBytes <= 0 {
		sizeInBytes = 1 << 12 //default bytes = 4kb
	}
	//Create a buffer to write data.
	//The data we read from the input file will be written in this buffer.
	buf := bytes.NewBuffer(make([]byte, 0, sizeInBytes))
	for {
		err := cw.readFromInp(inp, buf, sizeInBytes) //Write data to our buffer.
		if err == io.EOF {                           //If we reach end of file break the loop.
			break
		}

		//runRegex will remove the single line and multi line comments and return the new bytes.
		finalBytes := runRegex(buf.Bytes())
		//once we get the filtered data, we push the data to the channel.
		cw.c <- finalBytes
		//We reset the buffer now so we can reuse the allocations on next read.
		buf.Reset()
	}

	//repeat steps here for remaining bytes.
	finalBytes := runRegex(buf.Bytes())
	cw.c <- finalBytes
	buf.Reset()
	//Close the channel here so that the read method can return EOF.
	close(cw.c)
}

func runRegex(b []byte) []byte {
	b = singleLineComment.ReplaceAll(b, nil)
	b = multiLineComment.ReplaceAll(b, nil)
	return b
}

func (cw *chanReader) readFromInp(inp io.Reader, buf *bytes.Buffer, sizeInBytes int) error {

	b := make([]byte, sizeInBytes)

	n, err := inp.Read(b) //Read file into b.
	if err == io.EOF {    //return if we reach EOF.
		return err
	} else if err != nil {
		cw.logger.Fatal().Err(err).Msg("converter : error while reading input")
	}

	b = b[:n]    //Reslice as its possible we have got partial read.
	buf.Write(b) //Write to buffer.

	//Our saftey check. We need this check as its possible that the end of buffer may be a partial comment match.
	//For example, lets say we have a comment [//This is a single line comment]
	//Its possible that Read method read it partially. [//This is a sing]
	//since we have read it till here our regex match will remove the string till sing.
	//Because of this next read will start from [le line comment] which is not a valid json and our json decoder will fail.
	//To avoid this, We will again read till a closing braces which signifies object closing.
	//NOTE :: This logic will not work in case the comments itself has json strings or there are nested json objects.
	//Have some idea or better logic? create a pull request or comment.
	for {
		sb := make([]byte, 1)
		n, err := inp.Read(sb)
		if err == io.EOF {
			return err
		}

		buf.Write(sb[:n])

		if bytes.Equal(sb, stopByte) {
			break //break if we find closing bracket
		}
	}

	return nil
}

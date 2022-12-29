package converter

import (
	"bufio"
	"bytes"
	"io"
	"regexp"

	"github.com/rs/zerolog"
)

type chanWriter struct {
	c      chan []byte
	excess *bytes.Buffer
	logger *zerolog.Logger
}

// New take an reader and returns another reader. Send 0 to create default size buffer. The new reader will receive data after removal of single line and multiline comments.
func New(inp *bufio.Reader, sizeInBytes int, logger *zerolog.Logger) io.Reader {
	cw := &chanWriter{
		c:      make(chan []byte, 50),
		excess: bytes.NewBuffer(nil),
		logger: logger,
	}

	go cw.startParsingInput(inp, sizeInBytes)

	return cw
}

func (cw *chanWriter) Read(buf []byte) (int, error) {
	var retn int

	//first copy excess bytes from previous operation
	n, err := cw.excess.Read(buf)
	if err != nil && err != io.EOF { //Ignore EOF here as we could write data later to it also
		cw.logger.Fatal().Err(err).Msg("error while writing to buffer")
	}

	if n >= len(buf) { //We filled the inp buffer, so we return as we dont have any capacity available.
		return n, nil
	}

	retn += n //incr

	//space remaining in buf, so lets get some data from our channel
	data, isChanOpen := <-cw.c //Blocks till channel is open
	if !isChanOpen && len(data) <= 0 {
		return retn, io.EOF
	}

	//for the remaining space
	n = copy(buf[n:], data)
	cw.excess.Write(data[n:])

	retn += n //incr

	return retn, nil
}

func (cw *chanWriter) startParsingInput(inp *bufio.Reader, sizeInBytes int) {
	if sizeInBytes <= 0 {
		sizeInBytes = 1 << 12 //4kb
	}
	buf := bytes.NewBuffer(make([]byte, 0, sizeInBytes))
	for {
		err := cw.readFromInp(inp, buf, sizeInBytes)
		if err == io.EOF {
			break
		}

		finalBytes := cw.runRegex(buf.Bytes())
		cw.c <- finalBytes
		buf.Reset()
	}

	finalBytes := cw.runRegex(buf.Bytes())
	cw.c <- finalBytes
	buf.Reset()
	close(cw.c)
}

func (cw *chanWriter) runRegex(b []byte) []byte {
	singleLineComment := regexp.MustCompile(`//(.*)`)
	b = singleLineComment.ReplaceAll(b, nil)

	multiLineComment := regexp.MustCompile(`/\*[\s\S]*?\*/`)
	b = multiLineComment.ReplaceAll(b, nil)

	return b
}

func (cw *chanWriter) readFromInp(inp *bufio.Reader, buf *bytes.Buffer, sizeInBytes int) error {

	b := make([]byte, sizeInBytes)

	n, err := inp.Read(b)
	if err == io.EOF {
		return err
	} else if err != nil {
		cw.logger.Fatal().Err(err).Msg("converter : error while reading input")
	}

	b = b[:n]
	buf.Write(b)

	for {
		singleByte, err := inp.ReadByte()
		if err == io.EOF {
			return err
		}
		buf.WriteByte(singleByte)

		if singleByte == '}' {
			break //break if we find closing bracket
		}
	}

	return nil
}

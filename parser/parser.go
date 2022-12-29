package parser

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"sort"

	"github.com/rs/zerolog"
)

type parser struct {
	headers    []string            //headers will be stored here.
	out        *csv.Writer         //Our output file will be csv
	decoder    *json.Decoder       //This is the json decoder we will use.
	utsHeaders map[string]struct{} //The columns which needs conversion from UNIX to string.
	logger     zerolog.Logger      //We will use the console logger of zerolog.
	pool       *pool               //To reduce some load on the GC.
}

func (p *parser) EnablePool() *parser {
	p.pool.enabled = true
	return p
}

func NewParser(out *csv.Writer, decoder *json.Decoder, logger *zerolog.Logger) *parser {

	return &parser{
		out:        out,
		decoder:    decoder,
		utsHeaders: map[string]struct{}{},
		logger:     *logger,
		pool:       &pool{},
	}
}

func (p *parser) ProcessArray(uts string) {

	p.startToken()

	p.setHeadersAndWriteFirstRow(uts, true)

	p.parseArrayElements()

	p.endToken()

}

func (p *parser) ProcessObjects(uts string) {

	p.setHeadersAndWriteFirstRow(uts, false)

	p.parseArrayElements()

}

func (p *parser) writeRow(row map[string]any, isFirstRow bool) {

	csvRow := p.pool.GetStringSlice() //get string slice from pool.

	for _, header := range p.headers { //We will loop on every header and get the value for that header. Since we are looping on headers we will skip extra elements which could be there in later objects

		value := row[header]

		if isFirstRow { //If its the first row no parsing is required as we are writing the headers.
			csvRow = append(csvRow, fmt.Sprintf("%v", value))
			continue
		}

		csvRow = append(csvRow, p.parseRowValue(header, value)) //get the proper value after conversion.
	}

	p.out.Write(csvRow)           //Write to our csv writer.
	p.pool.PutStringSlice(csvRow) //put the slice back in pool.
}

func (p *parser) parseArrayElements() {

	for p.decoder.More() {

		object := p.pool.GetMapStringAny()

		if err := p.decoder.Decode(&object); err != nil {
			p.logger.Fatal().Msgf("error while parseArrayElements decoding object : %v", err)
		}

		p.writeRow(object, false)
		p.pool.PutMapStringAny(object)
	}

	p.out.Flush()
}

func (p *parser) getHeaderAndFirstRow() ([]string, map[string]any) {
	if p.decoder.More() { //Check if we have an object
		object := map[string]any{}
		if err := p.decoder.Decode(&object); err != nil { //Decode the object into map.
			p.logger.Fatal().Msgf("error while decoding array object : %v", err)
		}

		headers := []string{}
		for key := range object {
			headers = append(headers, key)
		}

		sort.Strings(headers) //Sort headers or we will get random order every run because maps & json being unordered.

		return headers, object
	}

	p.logger.Fatal().Msgf("empty object") //If we dont get first object the the file would not have one and could be an empty array.

	return nil, nil
}

func (p *parser) setHeadersAndWriteFirstRow(uts string, isArray bool) {

	headerMap := map[string]any{}

	headers, row := p.getHeaderAndFirstRow()

	p.headers = headers
	for _, header := range headers {
		headerMap[header] = header //We are using map because we want to write this as first row itself and our writeRow method only takes map. Also this helps with fast lookups.
	}

	p.pool.SetPools(len(headers)) //set pool as we now know the header size
	p.setUTS(uts, headerMap)      //set uts so that later we can use this to convert the unix timestamp to string.
	p.writeRow(headerMap, true)   //Write the headers to csv file.
	p.writeRow(row, false)        //Write our first row after headers.
}

func (p *parser) endToken() {
	token := p.token()
	p.logger.Debug().Msgf("End Token : %v", token)
}

func (p *parser) startToken() {
	token := p.token()
	p.logger.Debug().Msgf("Start Token : %v", token)
}

func (p *parser) token() json.Token {
	token, err := p.decoder.Token()
	if err != nil {
		p.logger.Fatal().Msgf("error while reading token : %v", err)
	}

	return token
}

package parser

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"sort"

	"github.com/rs/zerolog"
)

type parser struct {
	headers    []string
	out        *csv.Writer
	decoder    *json.Decoder
	utsHeaders map[string]struct{}
	logger     zerolog.Logger
	pool       *pool
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

	csvRow := p.pool.GetStringSlice()

	for _, header := range p.headers {

		value := row[header]

		if isFirstRow {
			csvRow = append(csvRow, fmt.Sprintf("%v", value))
			continue
		}

		csvRow = append(csvRow, p.parseRowValue(header, value))
	}

	p.out.Write(csvRow)
	p.pool.PutStringSlice(csvRow)
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
	if p.decoder.More() {
		object := map[string]any{}
		if err := p.decoder.Decode(&object); err != nil {
			p.logger.Fatal().Msgf("error while decoding array object : %v", err)
		}

		headers := []string{}
		for key := range object {
			headers = append(headers, key)
		}

		sort.Strings(headers)

		return headers, object
	}

	p.logger.Fatal().Msgf("empty object")

	return nil, nil
}

func (p *parser) setHeadersAndWriteFirstRow(uts string, isArray bool) {

	headerMap := map[string]any{}

	headers, row := p.getHeaderAndFirstRow()

	p.headers = headers
	for _, header := range headers {
		headerMap[header] = header
	}

	p.pool.SetPools(len(headers)) //set pool as we now know the header size
	p.setUTS(uts, headerMap)
	p.writeRow(headerMap, true)
	p.writeRow(row, false)
}

func (p *parser) endToken() {
	token, err := p.decoder.Token()
	if err != nil {
		p.logger.Fatal().Msgf("error while decoding json : %v", err)
	}
	p.logger.Debug().Msgf("End Token : %v", token)
}

func (p *parser) startToken() json.Token {

	token, err := p.decoder.Token()
	if err != nil {
		p.logger.Fatal().Msgf("error while reading token : %v", err)
	}

	p.logger.Debug().Msgf("Start Token : %v", token)

	return token
}

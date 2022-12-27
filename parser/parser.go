package parser

import (
	"bufio"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"sort"

	"github.com/rs/zerolog"
)

type parser struct {
	headers    []string
	out        *csv.Writer
	inp        *bufio.Scanner
	decoder    *json.Decoder
	utsHeaders map[string]struct{}
	logger     zerolog.Logger
	pool       *pool
}

func (p *parser) EnablePool() *parser {
	p.pool.enabled = true
	return p
}

func NewParser(inp *bufio.Scanner, out *csv.Writer, decoder *json.Decoder, logger *zerolog.Logger) *parser {

	return &parser{
		out:        out,
		inp:        inp,
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

	p.parseDistinctElements()

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

func (p *parser) parseDistinctElements() {

	for p.inp.Scan() {

		txt := p.inp.Text()

		object := p.pool.GetMapStringAny()

		if err := json.Unmarshal([]byte(txt), &object); err != nil {
			p.logger.Warn().Err(err).Msgf("error while decode, invalid JSON")
			continue
		}

		p.writeRow(object, false)
		p.pool.PutMapStringAny(object)
	}

	p.out.Flush()
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

func (p *parser) getHeaderAndFirstRowFromObject() ([]string, map[string]any) {

	object := map[string]any{}

	//find first successful json object
	for p.inp.Scan() {

		txt := p.inp.Text()

		object = map[string]any{}
		if err := json.Unmarshal([]byte(txt), &object); err != nil {
			p.logger.Warn().Err(err).Msgf("error while decoding json")
			continue
		}

		break //we found the first valid json
	}

	if len(object) <= 0 {
		p.logger.Fatal().Msg("could not find first valid json object")
	}

	headers := []string{}
	for key := range object {
		headers = append(headers, key)
	}

	sort.Strings(headers)

	return headers, object
}

func (p *parser) getHeaderAndFirstRowFromArray() ([]string, map[string]any) {
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
	var headers []string
	var row map[string]any

	if isArray {
		headers, row = p.getHeaderAndFirstRowFromArray()
	} else {
		headers, row = p.getHeaderAndFirstRowFromObject()
	}

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

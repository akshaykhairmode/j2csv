package parser

import (
	"bufio"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"sort"
	"sync"

	"github.com/rs/zerolog"
)

type parser struct {
	headers    []string
	out        *csv.Writer
	inp        *bufio.Reader
	decoder    *json.Decoder
	utsHeaders map[string]struct{}
	logger     zerolog.Logger
	objPool    *sync.Pool
	strPool    *sync.Pool
}

func NewParser(inp *bufio.Reader, out *csv.Writer, logger *zerolog.Logger) *parser {

	decoder := json.NewDecoder(inp)

	return &parser{
		out:        out,
		inp:        inp,
		decoder:    decoder,
		utsHeaders: map[string]struct{}{},
		logger:     *logger,
	}
}

func (p *parser) Process(uts string) {

	p.setHeadersAndWriteFirstRow(uts)

	p.parseArrayElements()

	p.endToken()
}

func (p *parser) writeRow(row map[string]any, isFirstRow bool) {

	csvRow := p.strPool.Get().([]string)

	for _, header := range p.headers {

		value := row[header]

		if isFirstRow {
			csvRow = append(csvRow, fmt.Sprintf("%v", value))
			continue
		}

		csvRow = append(csvRow, p.parseRowValue(header, value))
	}

	p.out.Write(csvRow)

	//clear map and put it back in pool
	for k := range row {
		delete(row, k)
	}
	p.objPool.Put(row)

	//empty slice and put it back
	csvRow = csvRow[:0]
	p.strPool.Put(csvRow)
}

func (p *parser) parseArrayElements() {
	for p.decoder.More() {
		object := p.objPool.Get().(map[string]any)
		if err := p.decoder.Decode(&object); err != nil {
			p.logger.Fatal().Msgf("error while praseArray decoding object : %v", err)
		}

		p.writeRow(object, false)
	}

	p.out.Flush()
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

func (p *parser) setHeadersAndWriteFirstRow(uts string) {

	token := p.getJSONInputFormat()

	headerMap := map[string]any{}

	switch token {
	case json.Delim('['):
		headers, row := p.getHeaderAndFirstRowFromArray()
		p.headers = headers
		for _, header := range headers {
			headerMap[header] = header
		}
		p.setPools()
		p.setUTS(uts, headerMap)
		p.writeRow(headerMap, true)
		p.writeRow(row, false)
	default:
		p.logger.Fatal().Msgf("Invalid JSON, Only supports array of objects %s", `[{"key1":"value1","key2":"value2"}]`)
	}
}

func (p *parser) endToken() {
	endToken, err := p.decoder.Token()
	if err != nil {
		p.logger.Fatal().Msgf("error while decoding json : %v", err)
	}
	p.logger.Debug().Msgf("End Token : %v", endToken)
}

func (p *parser) getJSONInputFormat() json.Token {

	token, err := p.decoder.Token()
	if err != nil {
		p.logger.Fatal().Msgf("error while reading token : %v", err)
	}

	return token
}

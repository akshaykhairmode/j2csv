package parser

import (
	"encoding/json"
	"fmt"
	"io"
	"sort"
	"strings"
	"time"

	"github.com/rs/zerolog"
)

type parser struct {
	headers    []string
	out        io.WriteCloser
	inp        io.ReadCloser
	decoder    *json.Decoder
	utsHeaders map[string]struct{}
	logger     zerolog.Logger
}

func NewParser(inp io.ReadCloser, out io.WriteCloser, logger *zerolog.Logger) *parser {

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

func (p *parser) writeRow(row map[string]any) {
	for index, header := range p.headers {
		value := row[header]
		objectPool.Put(row)

		switch v := value.(type) {
		case float64:
			if _, ok := p.utsHeaders[header]; ok {
				t := time.Unix(int64(v), 0)
				p.out.Write([]byte(t.String()))
			} else {
				p.out.Write([]byte(fmt.Sprintf("%d", int64(v))))
			}
		default:
			p.out.Write([]byte(fmt.Sprintf("%s", v)))
		}

		if index < len(p.headers)-1 {
			p.out.Write([]byte(","))
		}

	}

	p.out.Write([]byte("\n"))
}

func (p *parser) parseArrayElements() {
	for p.decoder.More() {
		object := objectPool.Get().(map[string]any)
		if err := p.decoder.Decode(&object); err != nil {
			p.logger.Fatal().Msgf("error while praseArray decoding object : %v", err)
		}

		p.writeRow(object)
	}
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
		p.writeRow(headerMap)
		p.setUTS(uts, headerMap)
		p.writeRow(row)
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

func (p *parser) setUTS(uts string, headerMap map[string]any) {

	fields := strings.Split(uts, ",")

	if len(fields) <= 0 {
		return
	}

	for _, field := range fields {
		if _, ok := headerMap[field]; !ok {
			p.logger.Fatal().Msgf("Passed header %v does not match with file headers : %v", field, p.headers)
		}
		p.utsHeaders[field] = struct{}{}
	}
}

func (p *parser) getJSONInputFormat() json.Token {

	token, err := p.decoder.Token()
	if err != nil {
		p.logger.Fatal().Msgf("error while reading token : %v", err)
	}

	return token
}

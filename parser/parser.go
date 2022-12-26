package parser

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"io"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/rs/zerolog"
)

type parser struct {
	headers    []string
	out        *csv.Writer
	inp        io.ReadCloser
	decoder    *json.Decoder
	utsHeaders map[string]struct{}
	logger     zerolog.Logger
	objPool    *sync.Pool
	strPool    *sync.Pool
}

func NewParser(inp io.ReadCloser, out *csv.Writer, logger *zerolog.Logger) *parser {

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

		if isFirstRow || len(p.utsHeaders) <= 0 {
			csvRow = append(csvRow, fmt.Sprintf("%v", value))
			continue
		}

		switch v := value.(type) {
		case float64:
			if _, ok := p.utsHeaders[header]; ok {
				t := time.Unix(int64(v), 0)
				csvRow = append(csvRow, t.String())
				continue
			}
			csvRow = append(csvRow, fmt.Sprintf("%d", int64(v)))
		case string:
			if _, ok := p.utsHeaders[header]; ok {
				val, err := strconv.ParseInt(v, 10, 64)
				if err != nil {
					p.logger.Debug().Str("str", v).Msg("could not convert the string to int64")
					csvRow = append(csvRow, v)
					continue
				}
				t := time.Unix(val, 0)
				csvRow = append(csvRow, t.String())
			}

		default:
			csvRow = append(csvRow, fmt.Sprintf("%s", v))
		}
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

func (p *parser) setPools() {

	p.objPool = &sync.Pool{
		New: func() any {
			return make(map[string]any, len(p.headers))
		},
	}

	p.strPool = &sync.Pool{
		New: func() any {
			return make([]string, 0, len(p.headers))
		},
	}

}

func (p *parser) setUTS(uts string, headerMap map[string]any) {

	trimmed := strings.TrimSpace(uts)
	if trimmed == "" {
		return
	}

	fields := strings.Split(trimmed, ",")

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

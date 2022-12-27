package parser

import (
	"fmt"
	"strconv"
	"strings"
	"time"
)

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

func (p *parser) parseRowValue(header string, value any) string {

	_, isUTSColumn := p.utsHeaders[header]

	switch v := value.(type) {
	case float64:
		if isUTSColumn {
			t := time.Unix(int64(v), 0)
			return t.String()
		}
		return strconv.FormatInt(int64(v), 10)
	case string:
		if isUTSColumn {
			val, err := strconv.ParseInt(v, 10, 64)
			if err != nil {
				p.logger.Debug().Str("str", v).Msg("could not convert the string to int64")
				return v
			}
			text, err := time.Unix(val, 0).Local().MarshalText()
			if err == nil {
				return string(text)
			}
			return v
		}
	case bool:
		if v {
			return "true"
		}
		return "false"
	case nil:
		return ""
	default:
		return fmt.Sprintf("%v", v)
	}

	return fmt.Sprintf("%v", value)

}

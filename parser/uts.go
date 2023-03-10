package parser

import (
	"encoding/json"
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

	_, isUTSColumn := p.utsHeaders[header] //check if the column exist in
	switch v := value.(type) {
	case float64: //json decodes numbers as float64.
		if isUTSColumn {
			t := time.Unix(int64(v), 0)
			return t.String()
		}
		return strconv.FormatFloat(v, 'f', -1, 64)
	case string:
		if isUTSColumn {
			val, err := strconv.ParseInt(v, 10, 64) //first convert to int
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
	case map[string]any: //If its nested JSON, marshal it and return the string
		nested, err := json.Marshal(v)
		if err != nil {
			p.logger.Debug().Err(err).Msg("error while marshaling nested JSON")
			return fmt.Sprintf("%v", v)
		}
		return string(nested)
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

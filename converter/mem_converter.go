package converter

import (
	"bytes"
	"io"

	"github.com/rs/zerolog"
)

func ConvertInMemory(r io.Reader, logger *zerolog.Logger) *bytes.Buffer {

	buf, err := io.ReadAll(r)
	if err != nil {
		logger.Fatal().Err(err).Msg("could no read file in memory")
	}

	return bytes.NewBuffer(runRegex(buf))

}

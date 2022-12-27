package parser

import (
	"bufio"
	"encoding/csv"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/rs/zerolog"
)

func GetInputReader(inFile string, logger *zerolog.Logger) *bufio.Reader {
	if inFile == "" {
		flag.PrintDefaults()
		logger.Fatal().Msgf("Input file path cannot be empty")
	}

	fh, err := os.Open(inFile)
	if err != nil {
		logger.Fatal().Msgf("error while opening input file : %v", err)
	}

	logger.Info().Msgf("Reading input from path : %s", inFile)

	return bufio.NewReader(fh)
}

func GetOutWriter(inFile, outFile string, logger *zerolog.Logger) (*csv.Writer, string) {

	if outFile == "" {
		fileName := filepath.Base(inFile)
		ext := filepath.Ext(fileName)
		ts := time.Now().Unix()
		fname := fileName[0 : len(fileName)-len(ext)]
		outFile = fmt.Sprintf("j2csv-%s-%d.%s", fname, ts, "csv")
	}

	fh, err := os.OpenFile(outFile, os.O_CREATE|os.O_RDWR, 0644)
	if err != nil {
		logger.Fatal().Msgf("error while creating output file : %v", err)
	}

	logger.Info().Msgf("Output File Path >> %s", outFile)

	return csv.NewWriter(fh), outFile
}

package file

import (
	"archive/zip"
	"bufio"
	"encoding/csv"
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"time"

	"github.com/rs/zerolog"
)

type Close func()

func GetInputReader(inFile string, isStdin bool, logger *zerolog.Logger) (io.Reader, Close) {

	c := (func() {})

	if isStdin {
		return os.Stdin, c
	}

	if inFile == "" {
		flag.PrintDefaults()
		logger.Fatal().Msgf("Input file path cannot be empty")
	}

	var fh io.ReadCloser
	var err error

	if isZip(inFile) {
		fh = getZipReader(inFile, logger)
	} else {
		fh, err = os.Open(inFile)
		if err != nil {
			logger.Fatal().Err(err).Msg("error while opening input file")
		}
	}
	c = func() { closeFile(fh, logger) }

	logger.Info().Msgf("Reading input from path : %s", inFile)

	return bufio.NewReader(fh), c
}

func GetOutWriter(inFile, outFile string, isZip bool, logger *zerolog.Logger) (*csv.Writer, string, Close) {

	if outFile == "" {
		fileName := filepath.Base(inFile)
		ext := filepath.Ext(fileName)
		ts := time.Now().Unix()
		fname := fileName[0 : len(fileName)-len(ext)]
		outFile = fmt.Sprintf("j2csv-%s-%d.%s", fname, ts, "csv")
	}

	fh, err := os.OpenFile(outFile, os.O_CREATE|os.O_RDWR, 0644)
	if err != nil {
		logger.Fatal().Err(err).Msg("error while creating output file")
	}

	c := Close(func() { closeFile(fh, logger) })

	return csv.NewWriter(fh), outFile, c
}

func getZipReader(inFile string, logger *zerolog.Logger) io.ReadCloser {
	zr, err := zip.OpenReader(inFile)
	if err != nil {
		logger.Fatal().Err(err).Msg("error while opening input zip file")
	}

	if len(zr.File) > 1 {
		logger.Fatal().Msg("Only 1 file allowed in zip")
	}

	fh, err := zr.File[0].Open()
	if err != nil {
		logger.Fatal().Msgf("error while opening file insize zip : %v", err)
	}

	return fh
}

func closeFile(fh io.Closer, logger *zerolog.Logger) {
	logger.Debug().Msg("closing file")
	if err := fh.Close(); err != nil {
		logger.Error().Err(err).Msg("error while closing file")
	}
}

func isZip(name string) bool {
	ext := filepath.Ext(name)
	log.Println(ext)
	return ext == ".zip"
}

func ZipFile(fpath string, logger *zerolog.Logger) (string, error) {

	fname := filepath.Base(fpath)
	ext := filepath.Ext(fname)
	zipName := fpath[0:len(fpath)-len(ext)] + ".zip"

	fh, err := os.Open(fpath)
	if err != nil {
		logger.Warn().Err(err).Msg("zip : error while opening the out file")
		return zipName, err
	}
	defer fh.Close()

	zh, err := os.OpenFile(zipName, os.O_CREATE|os.O_RDWR, 0644)
	if err != nil {
		logger.Warn().Err(err).Msg("zip : error while creating zip file")
		return zipName, err
	}
	defer zh.Close()

	wr := zip.NewWriter(zh)
	defer wr.Close()

	fwr, err := wr.Create(fname)
	if err != nil {
		logger.Warn().Err(err).Msg("zip : error while creating file inside zip")
		return zipName, err
	}

	n, err := io.Copy(fwr, fh)
	if err != nil {
		logger.Warn().Err(err).Msg("zip : could not copy to zip")
		return zipName, err
	}

	logger.Debug().Msgf("Copied : %d bytes to zip", n)

	return zipName, nil
}

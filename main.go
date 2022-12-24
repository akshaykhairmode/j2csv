package main

import (
	"flag"
	"io"
	"os"
	"time"

	"j2csv/logger"
	"j2csv/parser"

	"github.com/rs/zerolog"
)

type flags struct {
	inFile  string //the file to read for the json input
	outFile string //the output file path
	uts     string //unix to string
	verbose bool   //enables debug logs
	help    bool   //prints command help
}

func main() {

	startTime := time.Now()
	fg := parseFlags()
	logWriter := logger.GetLogger(fg.verbose)
	fg.printAll(logWriter)

	input := parser.GetInputReader(fg.inFile, logWriter)
	output, outFilePath := parser.GetOutWriter(fg.inFile, fg.outFile, logWriter)
	defer closeFiles(logWriter, input, output)

	logWriter = logger.SetFatalHook(logWriter, outFilePath)

	parser.NewParser(input, output, logWriter).Process(fg.uts)

	logWriter.Info().Msgf("Done!!, Time took : %v", time.Since(startTime))

}

func closeFiles(logger *zerolog.Logger, fhs ...io.Closer) {
	for _, fh := range fhs {
		if err := fh.Close(); err != nil {
			logger.Debug().Msgf("error while closing file : %v", err)
		}
	}
}

func (f flags) printAll(logger *zerolog.Logger) {
	flag.VisitAll(func(f *flag.Flag) {
		logger.Debug().Msgf("Flag %s , Value : %s", f.Name, f.Value)
	})
}

func parseFlags() flags {
	fg := flags{}
	flag.StringVar(&fg.inFile, "f", "", "--f /home/input.txt (Required)")
	flag.StringVar(&fg.outFile, "o", "", "--f /home/output.txt")
	flag.StringVar(&fg.uts, "uts", "", "used to convert timestamp to string, usage --uts createdAt,updatedAt")
	flag.BoolVar(&fg.verbose, "v", false, "Enables verbose logging")
	flag.BoolVar(&fg.help, "h", false, "Prints command help")
	flag.Parse()

	if fg.help {
		flag.PrintDefaults()
		os.Exit(0)
	}

	return fg
}

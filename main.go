package main

import (
	"flag"
	"fmt"
	"os"
	"runtime"
	"time"

	"github.com/akshaykhairmode/j2csv/logger"
	"github.com/akshaykhairmode/j2csv/parser"

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

	logWriter = logger.SetFatalHook(logWriter, outFilePath)

	parser.NewParser(input, output, logWriter).Process(fg.uts)

	logWriter.Info().Msgf("Done!!, Time took : %v", time.Since(startTime))

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

// https://gist.github.com/j33ty/79e8b736141be19687f565ea4c6f4226
func PrintMemUsage() {
	var m runtime.MemStats
	runtime.ReadMemStats(&m)
	// For info on each, see: https://golang.org/pkg/runtime/#MemStats
	fmt.Printf("Alloc = %v MiB", bToMb(m.Alloc))
	fmt.Printf("\tTotalAlloc = %v MiB", bToMb(m.TotalAlloc))
	fmt.Printf("\tSys = %v MiB", bToMb(m.Sys))
	fmt.Printf("\tNumGC = %v\n", m.NumGC)
}

func bToMb(b uint64) uint64 {
	return b / 1024 / 1024
}

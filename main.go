package main

import (
	"bufio"
	"encoding/csv"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"os"
	"runtime"
	"time"

	"github.com/akshaykhairmode/j2csv/converter"
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
	stats   bool   //prints memory allocs/gc etc
	isArray bool   //if input is array of objects
}

func main() {

	startTime := time.Now()
	fg := parseFlags()
	logWriter := logger.GetLogger(fg.verbose) //get a console logger
	fg.printAll(logWriter)

	input := parser.GetInputReader(fg.inFile, logWriter) //get a buffered reader from the input file.
	output, outFilePath := parser.GetOutWriter(fg.inFile, fg.outFile, logWriter)

	logWriter = logger.SetFatalHook(logWriter, outFilePath) //If fatal log level is called, delete the output file.

	PrintMemUsage(fg.stats)
	if fg.isArray {
		processArray(output, input, logWriter, fg.uts)
	} else {
		processObjects(output, input, logWriter, fg.uts)
	}
	PrintMemUsage(fg.stats)

	logWriter.Info().Msgf("Done!!, Time took : %v", time.Since(startTime))

}

func processArray(output *csv.Writer, input io.Reader, logWriter *zerolog.Logger, uts string) {
	decoder := json.NewDecoder(input)
	p := parser.NewParser(output, decoder, logWriter).EnablePool()
	p.ProcessArray(uts)
}

func processObjects(output *csv.Writer, input *bufio.Reader, logWriter *zerolog.Logger, uts string) {
	newInput := converter.New(input, 0, logWriter)
	decoder := json.NewDecoder(newInput)
	p := parser.NewParser(output, decoder, logWriter).EnablePool()
	p.ProcessObjects(uts)
}

func (f flags) printAll(logger *zerolog.Logger) {
	flag.VisitAll(func(f *flag.Flag) {
		logger.Debug().Msgf("Flag %s , Value : %s", f.Name, f.Value)
	})
}

func parseFlags() flags {
	fg := flags{}
	flag.BoolVar(&fg.stats, "stats", false, "prints the allocations at start and at end")
	flag.StringVar(&fg.inFile, "f", "", "usage --f /home/input.txt (Required)")
	flag.StringVar(&fg.outFile, "o", "", "usage --o /home/output.txt")
	flag.StringVar(&fg.uts, "uts", "", "used to convert timestamp to string, usage --uts createdAt,updatedAt")
	flag.BoolVar(&fg.verbose, "v", false, "Enables verbose logging")
	flag.BoolVar(&fg.help, "h", false, "Prints command help")
	flag.BoolVar(&fg.isArray, "a", false, "use this option if its an array of objects")
	flag.Parse()

	if fg.help {
		flag.PrintDefaults()
		os.Exit(0)
	}

	return fg
}

// https://gist.github.com/j33ty/79e8b736141be19687f565ea4c6f4226
func PrintMemUsage(print bool) {

	if !print {
		return
	}

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

package main

import (
	"encoding/csv"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"os"
	"runtime"
	"time"

	"github.com/akshaykhairmode/j2csv/converter"
	"github.com/akshaykhairmode/j2csv/file"
	"github.com/akshaykhairmode/j2csv/logger"
	"github.com/akshaykhairmode/j2csv/parser"

	"github.com/rs/zerolog"
)

type flags struct {
	inFile  string //the file to read for the json input
	outFile string //the output file path
	uts     string //unix to string
	empty   string //fill empty columns with passed value
	verbose bool   //enables debug logs
	help    bool   //prints command help
	stats   bool   //prints memory allocs/gc etc
	force   bool   //will load the whole input file in memory
	stdIn   bool   //get data from stdin
	zip     bool   //create output in zip file
	isArray bool   //if input is array of objects
}

const (
	colorReset = "\033[0m"
	colorGreen = "\033[32m"
)

var fg flags

func main() {

	startTime := time.Now()
	parseFlags()
	logWriter := logger.GetLogger(fg.verbose) //get a console logger
	fg.printAll(logWriter)

	input, closeInput := file.GetInputReader(fg.inFile, fg.stdIn, logWriter) //get a buffered reader from the input file.
	defer closeInput()

	output, outFilePath, closeOutput := file.GetOutWriter(fg.inFile, fg.outFile, fg.zip, logWriter)

	logWriter = logger.SetFatalHook(logWriter, outFilePath, closeInput, closeOutput) //If fatal log level is called, delete the output file.

	PrintMemUsage(fg.stats)
	if fg.isArray {
		processArray(output, input, logWriter, fg)
	} else {
		processObjects(output, input, logWriter, fg)
	}
	PrintMemUsage(fg.stats)

	closeOutput()
	processZip(outFilePath, fg.zip, logWriter)

	logWriter.Info().Msgf("Done!!, Time took : %v", time.Since(startTime))

}

func processZip(outFilePath string, isZip bool, logWriter *zerolog.Logger) {

	format := func(s string) string {
		return fmt.Sprintf("Output File ====> %v%s%v", colorGreen, s, colorReset)
	}

	if !isZip {
		logWriter.Info().Msg(format(outFilePath))
		return
	}

	zipPath, err := file.ZipFile(outFilePath, logWriter)
	if err != nil {
		logWriter.Error().Msg("could not create zip file")
		os.Remove(zipPath)
		return
	}

	os.Remove(outFilePath)
	logWriter.Info().Msgf(format(zipPath))

}

func processArray(output *csv.Writer, input io.Reader, logWriter *zerolog.Logger, fg flags) {
	decoder := json.NewDecoder(input)
	p := parser.NewParser(output, decoder, logWriter).EnablePool().SetDefault(fg.empty)
	p.ProcessArray(fg.uts)
}

func processObjects(output *csv.Writer, input io.Reader, logWriter *zerolog.Logger, fg flags) {

	var newInput io.Reader

	if fg.force {
		newInput = converter.ConvertInMemory(input, logWriter)
	} else {
		newInput = converter.New(input, 0, logWriter) //converter is the package name we are using.
	}

	decoder := json.NewDecoder(newInput)
	p := parser.NewParser(output, decoder, logWriter).EnablePool().SetDefault(fg.empty)
	p.ProcessObjects(fg.uts)
}

func (f flags) printAll(logger *zerolog.Logger) {
	flag.VisitAll(func(f *flag.Flag) {
		logger.Debug().Msgf("Flag %s , Value : %s", f.Name, f.Value)
	})
}

func parseFlags() {
	flag.BoolVar(&fg.stats, "stats", false, "prints the allocations at start and at end")
	flag.StringVar(&fg.inFile, "f", "", "usage --f /home/input.txt (Required)")
	flag.StringVar(&fg.outFile, "o", "", "usage --o /home/output.txt")
	flag.StringVar(&fg.uts, "uts", "", "used to convert timestamp to string, usage --uts createdAt,updatedAt")
	flag.BoolVar(&fg.verbose, "v", false, "Enables verbose logging")
	flag.BoolVar(&fg.help, "h", false, "Prints command help")
	flag.BoolVar(&fg.isArray, "a", false, "use this option if its an array of objects")
	flag.BoolVar(&fg.force, "force", false, "force load input file in memory, use this if conversion is failing.")
	flag.BoolVar(&fg.stdIn, "i", false, "get input data from standard input")
	flag.BoolVar(&fg.zip, "z", false, "output file to be .zip")
	flag.StringVar(&fg.empty, "e", "", "usage --e NA, will put NA in columns where value does not exist")
	flag.Parse()

	if fg.help {
		flag.PrintDefaults()
		os.Exit(0)
	}
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

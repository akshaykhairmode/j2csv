package logger

import (
	"os"

	"github.com/akshaykhairmode/j2csv/parser"
	"github.com/rs/zerolog"
)

func GetLogger(verbose bool) *zerolog.Logger {

	zerolog.SetGlobalLevel(zerolog.InfoLevel)
	logger := zerolog.New(zerolog.ConsoleWriter{Out: os.Stdout}).With().Timestamp().Logger()

	if verbose {
		zerolog.SetGlobalLevel(zerolog.DebugLevel)
	}

	return &logger

}

func SetFatalHook(logger *zerolog.Logger, outFile string, cleanup ...parser.Close) *zerolog.Logger {

	logger.Debug().Str("outFile", outFile).Msg("Setting up hook")

	l := logger.Hook(FatalHook{
		OutFile: outFile,
		Logger:  logger,
		cleanup: cleanup,
	})

	return &l

}

type FatalHook struct {
	cleanup []parser.Close
	OutFile string
	Logger  *zerolog.Logger
}

func (h FatalHook) Run(e *zerolog.Event, level zerolog.Level, msg string) {
	if level == zerolog.FatalLevel {

		for _, f := range h.cleanup {
			f()
		}

		err := os.Remove(h.OutFile)
		if os.IsNotExist(err) {
			return
		}
		if err != nil {
			h.Logger.Debug().Err(err).Msg("error while removing out file")
		}
	}
}

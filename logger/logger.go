package logger

import (
	"context"
	"io"
	"os"
	"sync"
	"time"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/pkgerrors"
)

type LogFields struct {
	Key   string
	Value string
}

var once sync.Once

var log zerolog.Logger

func Logger(logLevel string) zerolog.Logger {
	once.Do(func() {
		zerolog.ErrorStackMarshaler = pkgerrors.MarshalStack
		zerolog.TimeFieldFormat = time.RFC3339Nano

		level, err := zerolog.ParseLevel(logLevel)
		if err != nil {
			panic(err)
		}

		//logLevel, err := strconv.Atoi(os.Getenv("LOG_LEVEL"))
		//if err != nil {
		//	logLevel = int(zerolog.DebugLevel) // default to Debug
		//}

		var output io.Writer = zerolog.ConsoleWriter{
			Out:        os.Stdout,
			TimeFormat: time.RFC3339,
		}

		//if os.Getenv("APP_ENV") != "development" {
		//	output = os.Stdout
		//}

		log = zerolog.New(output).
			Level(level).
			With().
			Timestamp().
			Caller().
			Logger()

		zerolog.DefaultContextLogger = &log
	})

	return log
}

func Log(ctx context.Context, fields ...LogFields) *zerolog.Logger {
	ctxLogger := zerolog.Ctx(ctx)
	tempContext := ctxLogger.With()
	for _, field := range fields {
		tempContext = tempContext.Str(field.Key, field.Value)
	}
	tempLogger := tempContext.Logger()
	return &tempLogger
}

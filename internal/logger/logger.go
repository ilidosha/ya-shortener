package logger

import (
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"os"
)

func SetupLog(debug bool) {
	if debug {
		// В дебаге выводим плоские строки
		cw := zerolog.ConsoleWriter{Out: os.Stdout}
		log.Logger = zerolog.New(cw).With().Timestamp().Logger()
		zerolog.SetGlobalLevel(zerolog.DebugLevel)
		return
	}

	// Иначе логируем в stdout в json'е
	log.Logger = zerolog.New(os.Stdout).With().Logger()
	zerolog.SetGlobalLevel(zerolog.InfoLevel)

	zerolog.MessageFieldName = "m"
	zerolog.ErrorFieldName = "e"
	zerolog.LevelFieldName = "l"
	zerolog.TimestampFieldName = "t"
}

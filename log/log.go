package log

import (
	"github.com/sirupsen/logrus"
	"os"
)

var (
	log = logrus.New()
)

type LogConfig struct {
	Level     string
	Out       string
	File      string
	Formatter string
}

func SetLogConf(logConfig *LogConfig) (*logrus.Logger, error) {
	if logConfig.Out != "file" {
		log.SetOutput(os.Stdout)
	} else {
		f, err := os.OpenFile(logConfig.File, os.O_WRONLY|os.O_APPEND|os.O_CREATE, 0644)
		if err != nil {
			return nil, err
		}
		log.SetOutput(f)
	}

	if logConfig.Formatter == "json" {
		log.SetFormatter(&logrus.JSONFormatter{})
	}

	switch logConfig.Level {
	case "trace":
		log.SetLevel(logrus.TraceLevel)
	case "debug":
		log.SetLevel(logrus.DebugLevel)
	case "info":
		log.SetLevel(logrus.InfoLevel)
	case "warn":
		log.SetLevel(logrus.WarnLevel)
	case "error":
		log.SetLevel(logrus.ErrorLevel)
	case "fatal":
		log.SetLevel(logrus.FatalLevel)
	case "panic":
		log.SetLevel(logrus.PanicLevel)
	default:
		log.SetLevel(logrus.DebugLevel)
	}

	return log, nil
}

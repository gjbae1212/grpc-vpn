package internal

import (
	"io"
	"os"
	"path/filepath"

	"github.com/fatih/color"
	"github.com/sirupsen/logrus"
)

type Logger struct {
	*logrus.Logger
}

// NewLogger is to return logrus logger.
func NewLogger(filePath string) (*Logger, error) {
	logger := logrus.New()
	logger.SetFormatter(&logrus.TextFormatter{
		ForceColors:   true,
		DisableColors: false,
		FullTimestamp: true,
	})

	if filePath == "" {
		logger.SetOutput(os.Stdout)
	} else {
		// make log directory
		logDir := filepath.Dir(filePath)
		if _, err := os.Stat(logDir); os.IsNotExist(err) {
			if err := os.MkdirAll(logDir, os.ModePerm); err != nil {
				return nil, err
			}
		}

		// make log file
		f, err := os.OpenFile(filePath, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0600)
		if err != nil {
			return nil, err
		}
		logger.SetOutput(io.MultiWriter(os.Stdout, f))
	}

	return &Logger{logger}, nil
}

// PanicWithMessage prints red message to stdout or file and then it raises a panic.
func (l *Logger) PanicWithError(err error) {
	l.Panicln(color.RedString("%s", err.Error()))
}

package logging

import (
	"io"
	stdlog "log"
	"os"
	"path"
	"strings"

	log "github.com/sirupsen/logrus"
)

// LogWriter represents writer object
var LogWriter io.WriteCloser

func openLogFile(filepath string) (io.WriteCloser, error) {
	f, err := os.OpenFile(filepath, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0600)
	if err != nil {
		return nil, err
	}
	return f, nil
}

func setLogOutput(w io.Writer) {
	log.SetOutput(w)
	stdlog.SetOutput(log.StandardLogger().Writer())
}

// Init initializes the logger
func Init(logdir string, logfile string, loglevel string) error {
	// Close the previously opened log file
	if LogWriter != nil {
		err := LogWriter.Close()
		if err != nil {
			return err
		}
		LogWriter = nil
	}

	level, err := log.ParseLevel(strings.ToLower(loglevel))
	if err != nil {
		setLogOutput(os.Stderr)
		log.WithError(err).Debug("Failed to parse log level")
		return err
	}
	log.SetLevel(level)
	log.SetFormatter(&log.TextFormatter{FullTimestamp: true, TimestampFormat: "2006-01-02 15:04:05.000000"})

	if strings.ToLower(logfile) == "stderr" || logfile == "-" {
		setLogOutput(os.Stderr)
	} else if strings.ToLower(logfile) == "stdout" {
		setLogOutput(os.Stdout)
	} else {
		logFilePath := path.Join(logdir, logfile)
		logFile, err := openLogFile(logFilePath)
		if err != nil {
			setLogOutput(os.Stderr)
			log.WithError(err).Debugf("Failed to open log file %s", logFilePath)
			return err
		}
		setLogOutput(logFile)
		LogWriter = logFile
	}
	return nil
}

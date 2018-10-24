package logging

import (
        "io"
        stdlog "log"
        "os"
        "path"
        "strings"

        log "github.com/sirupsen/logrus"
)

var LogWriter io.WriteCloser

func OpenLogFile(filepath string) (io.WriteCloser, error) {
        f, err := os.OpenFile(filepath, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
        if err != nil {
                return nil, err
        }
        return f, nil
}

func SetLogOutput(w io.Writer) {
        log.SetOutput(w)
        stdlog.SetOutput(log.StandardLogger().Writer())
}

func Init(logdir string, logfile string, loglevel string) error {
        // Close the previously opened log file
        if LogWriter != nil {
                LogWriter.Close()
                LogWriter = nil
        }

        level, err := log.ParseLevel(strings.ToLower(loglevel))
        if err != nil {
                SetLogOutput(os.Stderr)
                log.WithError(err).Debug("Failed to parse log level")
                return err
        }
        log.SetLevel(level)
        log.SetFormatter(&log.TextFormatter{FullTimestamp: true, TimestampFormat: "2006-01-02 15:04:05.000000"})

        if strings.ToLower(logfile) == "stderr" || logfile == "-" {
                SetLogOutput(os.Stderr)
        } else if strings.ToLower(logfile) == "stdout" {
                SetLogOutput(os.Stdout)
        } else {
                logFilePath := path.Join(logdir, logfile)
                logFile, err := OpenLogFile(logFilePath)
                if err != nil {
                        SetLogOutput(os.Stderr)
                        log.WithError(err).Debug("Failed to open log file %s", logFilePath)
                        return err
                }
                SetLogOutput(logFile)
                LogWriter = logFile
        }
        return nil
}

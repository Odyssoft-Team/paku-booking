package shared

import (
	"log"
	"os"
)

type Logger interface {
	Infof(format string, args ...any)
	Errorf(format string, args ...any)
	Debugf(format string, args ...any)
}

type StdLogger struct {
	info  *log.Logger
	err   *log.Logger
	debug *log.Logger
}

func NewStdLogger() *StdLogger {
	return &StdLogger{
		info:  log.New(os.Stdout, "INFO  ", log.LstdFlags|log.Lmicroseconds),
		err:   log.New(os.Stderr, "ERROR ", log.LstdFlags|log.Lmicroseconds),
		debug: log.New(os.Stdout, "DEBUG ", log.LstdFlags|log.Lmicroseconds),
	}
}

func (l *StdLogger) Infof(format string, args ...any)  { l.info.Printf(format, args...) }
func (l *StdLogger) Errorf(format string, args ...any) { l.err.Printf(format, args...) }
func (l *StdLogger) Debugf(format string, args ...any) { l.debug.Printf(format, args...) }

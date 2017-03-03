package log

import (
	"os"

	golog "log"
)

type Logger struct {
	*golog.Logger
	verbose bool
}

func NewLogger(verbose bool) *Logger {
	l := &Logger{
		verbose: verbose,
		Logger:  golog.New(os.Stdout, "", golog.Ldate|golog.Ltime),
	}
	return l
}

func (l *Logger) Fatal(v ...interface{}) {
	l.Logger.Fatal(v...)
}

func (l *Logger) Fatalf(format string, v ...interface{}) {
	l.Logger.Fatal(format, v)
}

func (l *Logger) Panic(v ...interface{}) {
	l.Logger.Panic(v...)
}

func (l *Logger) Print(v ...interface{}) {
	if l.verbose {
		l.Logger.Print(v...)
	}
}

func (l *Logger) Printf(format string, v ...interface{}) {
	if l.verbose {
		l.Logger.Printf(format, v...)
	}
}

func (l *Logger) Write(p []byte) (int, error) {
	if l.verbose {
		l.Logger.Printf("%s", p)
	}
	return len(p), nil
}

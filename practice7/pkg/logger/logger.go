package logger

import (
	"fmt"
	"log"
	"os"
)

type Interface interface {
	Info(message interface{}, args ...interface{})
	Error(message interface{}, args ...interface{})
}

type Logger struct {
	infoLog  *log.Logger
	errorLog *log.Logger
}

func New(level string) *Logger {
	return &Logger{
		infoLog:  log.New(os.Stdout, "INFO: ", log.Ldate|log.Ltime|log.Lshortfile),
		errorLog: log.New(os.Stderr, "ERROR: ", log.Ldate|log.Ltime|log.Lshortfile),
	}
}

func (l *Logger) Info(message interface{}, args ...interface{}) {
	l.infoLog.Printf(fmt.Sprintf("%v", message), args...)
}

func (l *Logger) Error(message interface{}, args ...interface{}) {
	l.errorLog.Printf(fmt.Sprintf("%v", message), args...)
}

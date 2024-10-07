package grpc

import "log"

type Logger struct {
	Open bool
}

func (l *Logger) Printf(format string, v ...any) {
	if !l.Open {
		return
	}
	log.Printf(format, v...)
}

func (l *Logger) Println(v ...any) {
	if !l.Open {
		return
	}
	log.Println(v...)
}

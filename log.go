package twig

import (
	"io"
	"log"
)

func newLog(w io.Writer, name string) *log.Logger {
	return log.New(w, name, log.LstdFlags|log.Lshortfile)
}

func newEventLog(w io.Writer, name string) *EventLogger {
	return &EventLogger{
		Logger: log.New(w, name, log.LstdFlags|log.Lshortfile),
	}
}

type EventLogger struct {
	Logger
}

func (el *EventLogger) On(eg EventRegister) {
	eg.On("logger", el)
}

func (el *EventLogger) OnEvent(topic string, ev *Event) {
	el.Println(ev.Body)
}

type Logger interface {
	Print(i ...interface{})
	Println(i ...interface{})
	Printf(format string, args ...interface{})
	Fatal(i ...interface{})
	Fatalln(i ...interface{})
	Fatalf(format string, args ...interface{})
	Panic(i ...interface{})
	Panicln(i ...interface{})
	Panicf(format string, args ...interface{})
}

package twig

import (
	"fmt"
	"io"
	"log"
	"os"
)

func newLog(w io.Writer, name string) *log.Logger {
	return log.New(w, name, log.LstdFlags|log.Lshortfile)
}

func newStdEventLog() *StdEventLogger {
	return &StdEventLogger{
		Logger: log.New(os.Stdout, "twig-", log.LstdFlags|log.Lshortfile),
	}
}

type StdEventLogger struct {
	*log.Logger
	twig *Twig
}

func (el *StdEventLogger) On(eg EventRegister) {
	eg.On("logger", el)
}

func (el *StdEventLogger) OnEvent(topic string, ev *Event) {
	el.Println(ev.Body)
}

func (el *StdEventLogger) Attach(t *Twig) {
	el.twig = t
	prefix := fmt.Sprintf("twig@%s-%s-", t.Name(), t.ID())
	el.SetPrefix(prefix)
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

	Attacher
	EventAttacher
}

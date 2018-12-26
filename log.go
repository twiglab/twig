package twig

import (
	"io"
	"log"
	"os"
)

func NewLog(w io.Writer) Logger {
	return log.New(w, "twig-log-", log.LstdFlags|log.Llongfile)
}

func NewStdLog(name string) Logger {
	return log.New(os.Stdout, name, log.LstdFlags|log.Llongfile)
}

func NewErrLog(name string) Logger {
	return log.New(os.Stderr, name, log.LstdFlags|log.Llongfile)
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

package twig

import (
	"io"
	"log"
)

func newLog(w io.Writer, name string) *log.Logger {
	return log.New(w, name, log.LstdFlags|log.Lshortfile)
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

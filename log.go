package twig

import (
	"io"
	"log"
)

// NewLog 创建Logger
func NewLog(w io.Writer, name string) *log.Logger {
	return log.New(w, name, log.LstdFlags|log.Lshortfile)
}

// Logger Twig的Logger接口，用于日志输出
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

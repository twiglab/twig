package twig

import (
	"os"
	"os/signal"
)

type Resetter interface {
	Reset()
}

type SignalFunc func(os.Signal)

func OsExit() SignalFunc {
	return func(_ os.Signal) {
		os.Exit(0)
	}
}

func Signal(f SignalFunc, sig ...os.Signal) {
	ch := make(chan os.Signal)
	signal.Notify(ch, sig...)

	for s := range ch {
		f(s)
	}
}

package twig

import (
	"os"
	"os/signal"
)

type Resetter interface {
	Reset()
}

// 信号处理函数
// 返回true 退出
// 返回false 等待处理下一个信号
type SignalFunc func(os.Signal) bool

func Exit() SignalFunc {
	return func(_ os.Signal) bool {
		return true
	}
}

// 监听系统信号
func Signal(f SignalFunc, sig ...os.Signal) {
	ch := make(chan os.Signal)
	defer close(ch)

	signal.Notify(ch, sig...)

	for s := range ch {
		if f(s) {
			break
		}
	}
}

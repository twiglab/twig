package twig

import (
	"os"
	"os/signal"
)

// Reloader 描述一个可以被重新加载的对象
// 在某些信号发生时候， 可以对Relaoder对象进行Reload操作，用于重新加载
type Reloader interface {
	Reload() error
}

// Reload
// bug: 无法处理所有信号(TODO)
/*
func Reload(r Reloader, sig ...os.Signal) {
	return func(s os.Signal) bool {
		for _, g := range sig {
			if s == g {
				if err := r.Reload(); err != nil {
					// log
				}
				return false
			}
		}
	}
}
*/

// 信号处理函数
// 返回true 退出
// 返回false 等待处理下一个信号
type SignalFunc func(os.Signal) bool

// 正常退出，不做任何处理
func Quit() SignalFunc {
	return func(_ os.Signal) bool {
		return true
	}
}

// Nop
func Nop() SignalFunc {
	return func(_ os.Signal) bool {
		return false
	}
}

// Signal 用于监听系统信号并堵塞当前gorouting
// 参数f为信号处理函数
// 参数sig 为需要监听的系统信号，未出现在sig中的信号会被忽略
// 如果sig 为空，则监听所有信号
// 特别注意：部分操作系统的信号不可以被忽略
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

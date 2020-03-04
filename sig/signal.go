package sig

import (
	"os"
	"os/signal"
)

var (
	closeFunc []func()
)

func RegisterClose(f ...func()) {
	closeFunc = append(closeFunc, f...)
}

func HoldOn() {
	var (
		ch = make(chan os.Signal)
	)
	signal.Notify(ch, os.Interrupt, os.Kill)
	<-ch
	for _, c := range closeFunc {
		go c()
	}
}

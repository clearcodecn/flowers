package sig

import (
	"os"
	"os/signal"
	"sync"
)

var (
	o         sync.Once
	closeFunc []func()
)

func RegisterClose(f ...func()) {
	initOnce()
	closeFunc = append(closeFunc, f...)
}

func initOnce() {
	o.Do(func() {
		go func() {
			var (
				ch = make(chan os.Signal)
			)
			signal.Notify(ch, os.Interrupt, os.Kill)
			<-ch
			for _, c := range closeFunc {
				go c()
			}
		}()
	})
}

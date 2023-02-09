package sources

import (
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	log "github.com/sirupsen/logrus"
)

type CloseHandler func() error

func GracefulShutdown(timeout time.Duration, closeHandlers ...CloseHandler) <-chan struct{} {
	wait := make(chan struct{})
	go func() {
		s := make(chan os.Signal, 1)

		signal.Notify(s, syscall.SIGINT, syscall.SIGTERM, syscall.SIGHUP)
		<-s

		log.Println("shutting down")

		timeoutFunc := time.AfterFunc(timeout, func() {
			log.Printf("timeout %d ms has been elapsed, force exit", timeout.Milliseconds())
			os.Exit(0)
		})

		defer timeoutFunc.Stop()

		var wg sync.WaitGroup

		for _, closeHandler := range closeHandlers {
			wg.Add(1)
			go func(ch CloseHandler) {
				defer wg.Done()

				if err := ch(); err != nil {
					log.Error(err)

					return
				}
			}(closeHandler)
		}
		wg.Wait()
		close(wait)
	}()

	return wait
}

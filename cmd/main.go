package main

import (
	"flag"
	"context"
	"os"
	"os/signal"
	"runtime"
	"syscall"
	"time"

	"github.com/go-kratos/kratos/pkg/conf/paladin"
	"github.com/go-kratos/kratos/pkg/log"
	"github.com/itering/subscan/internal/jobs"
	"github.com/itering/subscan/internal/server/http"
	"github.com/itering/subscan/internal/service"
	"github.com/itering/subscan/internal/substrate/websocket"
)

func main() {
	defer func () {
		_ = log.Close()
		websocket.CloseWsConnection()
	}()

	// init configs
	err := flag.Set("conf", "../configs")
	if err != nil {
		panic(err)
	}
	err = paladin.Init()
	if err != nil {
		panic(err)
	}
	jobs.Init()
	log.Init(nil)
	runtime.GOMAXPROCS(runtime.NumCPU())

	// start service
	svc := service.New()
	httpSrv := http.New(svc)

	// handle signals
	c := make(chan os.Signal, 1)
	log.Info("SubScan End run ......")
	signal.Notify(c, syscall.SIGHUP, syscall.SIGQUIT, syscall.SIGTERM, syscall.SIGINT)
	for {
		s := <-c
		log.Info("get a signal %s", s.String())
		switch s {
		case syscall.SIGQUIT, syscall.SIGTERM, syscall.SIGINT:
			ctx, cancel := context.WithTimeout(context.Background(), 35*time.Second)
			_ = httpSrv.Shutdown(ctx)
			log.Info("SubScan End exit")
			svc.Close()
			cancel()
			time.Sleep(time.Second)
			return
		case syscall.SIGHUP:
		default:
			return
		}
	}
}


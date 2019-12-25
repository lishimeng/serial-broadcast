package main

import (
	"context"
	"github.com/lishimeng/go-libs/log"
	"github.com/lishimeng/go-libs/shutdown"
	"github.com/lishimeng/go-libs/stream/serial"
	"github.com/lishimeng/serial-broadcast/internal/cmd"
	"github.com/lishimeng/serial-broadcast/internal/relay"
)

var globalContext context.Context
var cancel context.CancelFunc
var worker *relay.Worker
func main() {

	log.Info("Serial broadcast")
	globalContext, cancel = context.WithCancel(context.Background())
	err := cmd.Exec(application)
	if err != nil {
		log.Info(err)
		return
	}

	shutdown.WaitExit(&shutdown.Configuration{
		BeforeExit: func(s string) {
			if len(s) > 0 {
				log.Info(s)
			}
			cancel()
			if worker != nil {
				worker.Close()
			}
		},
	})
}

func application(p cmd.Params) {

	var err error
	worker, err = relay.New(serial.Config{
		Name: p.SerialName,
		Baud: p.Baud,
	}, p.Port)

	if err != nil {
		return
	}

	go worker.StartSerial(globalContext)
	go worker.StartTcp(globalContext)
}

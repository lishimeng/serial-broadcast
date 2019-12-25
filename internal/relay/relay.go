package relay

import (
	"context"
	"fmt"
	"github.com/lishimeng/go-libs/log"
	"github.com/lishimeng/go-libs/stream/serial"
	"io"
	"net"
	"sync"
)

var index = 0

type Worker struct {

	socks *sync.Map

	ser io.ReadWriteCloser

	server net.Listener

	listen uint16

	Ser serial.Config

	bufSize int
}

func New(serialConf serial.Config, listen uint16) (w *Worker, err error) {

	w = &Worker{
		Ser: serialConf,
		listen: listen,
		socks: new(sync.Map),
	}
	return
}

func (w *Worker) StartSerial(ctx context.Context) {
	log.Info("start serial:%s:%d", w.Ser.Name, w.Ser.Baud)
	conn := serial.New(&w.Ser)
	err := conn.Connect()
	if err != nil {
		return
	}
	w.ser = conn.Ser
	w.watch(ctx)
}

func (w *Worker) StartTcp(ctx context.Context) {

	listenPort := fmt.Sprintf(":%d", w.listen)
	log.Info("start tcp:%s", listenPort)

	lis, err := net.Listen("tcp", listenPort)
	if err != nil {
		return
	}
	log.Info("start listen: %s", lis.Addr().String())
	w.server = lis
	for {
		sock, err := w.server.Accept()
		if err != nil {
			return
		}
		conn := w.register(sock)
		w.runConn(ctx, conn)
	}
}

func (w *Worker) register(sock net.Conn) *Conn {

	index++
	log.Info("new connection:%s[%s]", index, sock.RemoteAddr().String())
	conn := &Conn{
		id: fmt.Sprintf("%d", index),
		rw: sock,
		onLostConn: func(id string, err error) {
			if item, ok := w.socks.Load(id); ok {
				forRemove := item.(*Conn)
				log.Info("remove connection:%s[%s]", forRemove.id, forRemove.rw.RemoteAddr().String())
				w.socks.Delete(id)
			}
		},
		onTx: w.txCb,
	}
	w.socks.Store(conn.id, conn)
	size := 0
	w.socks.Range(func(_, _ interface{}) bool {
		size++
		return true
	})
	log.Info("connection size:%d", size)
	return conn
}

func (w *Worker) txCb(_ string, data []byte) {
	w.tx(data)
}

func (w *Worker) Close() {
	log.Info("stop application")
	if w.server != nil {
		log.Info("stop tcp")
		_ = w.server.Close()// stop accept new connection
	}

	if w.socks != nil {
		log.Info("stop sockets")
		w.socks.Range(func(key, value interface{}) bool {
			conn := value.(*Conn)
			log.Info("--%s", conn.rw.RemoteAddr().String())
			_ = conn.rw.Close()
			return true
		})
	}

	if w.ser != nil {
		log.Info("stop serial")
		_ = w.ser.Close()
	}
}

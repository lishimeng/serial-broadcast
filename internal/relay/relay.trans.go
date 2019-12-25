package relay

import (
	"context"
	"github.com/lishimeng/go-libs/log"
	"io"
)

func (w *Worker) runConn(ctx context.Context, conn *Conn) {

	go conn.tx(ctx)
}

// serial --> tcp
func (w *Worker) watch(ctx context.Context) {
	defer func() {
		if err := recover(); err != nil {
			log.Info(err)
		}
	}()

	buf := make([]byte, 1024)
	for {
		select {
		case <- ctx.Done():
			return
		default:
			n, err := w.ser.Read(buf)
			if err != nil && err != io.EOF {
				panic(err)
			}
			if n > 0 {
				data := make([]byte, n)
				copy(data, buf)
				w.socks.Range(func(key, value interface{}) bool {
					sock := value.(*Conn)
					sock.rx(data)
					return true
				})
			}
		}
	}
}

func (w *Worker) broadcast(data []byte) {

	cache := make(map[string]byte)

	w.socks.Range(func(key, _ interface{}) bool {
		cache[key.(string)] = 0x00
		return true
	})

	for key := range cache {
		w.bc(key, data)
	}
}

func (w *Worker) bc(id string, data []byte) {
	if c, ok := w.socks.Load(id); ok {
		conn := c.(*Conn)
		conn.rx(data)
	}
}

// data --> serial
func (w *Worker) tx(data []byte) {
	defer func() {
		if err := recover(); err != nil {
			log.Info(err)
		}
	}()
	_, err := w.ser.Write(data)
	if err != nil {
		log.Info(err)
	}
}

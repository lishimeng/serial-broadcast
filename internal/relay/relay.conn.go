package relay

import (
	"context"
	"io"
	"net"
)

type Conn struct {

	id string

	rw net.Conn

	onLostConn func(id string, err error)
	onTx func(id string, data []byte)
}

func (c *Conn) rx(data []byte) {
	_, err := c.rw.Write(data)
	if err != nil && err != io.EOF {
		c.lostConn(err)
	}
}

// tcp --> serial
func (c *Conn) tx(ctx context.Context) {

	buf := make([]byte, 1024)
	var err error
	var n int
	for {
		select {
		case <- ctx.Done():
			return
		default:
			n, err = c.rw.Read(buf)
			if err != nil && err != io.EOF {
				if err != nil && err != io.EOF {
					c.lostConn(err)
				}
				return
			}
			if c.onTx != nil {
				data := make([]byte, n)
				copy(data, buf)
				c.onTx(c.id, data)
			}
		}

	}
}

func (c *Conn) lostConn(err error) {

	if c.onLostConn != nil {
		c.onLostConn(c.id, err)
	}
}
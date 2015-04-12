package netutil

import (
	"crypto/cipher"
	"crypto/rand"
	"io"
	"io/ioutil"
	"net"
	"time"
)

type CipherConfig struct {
	Key             []byte
	NonceSize       int
	NewCipherStream func(key []byte, iv []byte) (cipher.Stream, error)
	Rand            io.Reader
}

type CipherConn struct {
	conn     net.Conn
	config   *CipherConfig
	isClient bool
	rs       cipher.Stream
	ws       cipher.Stream
}

func CipherClient(conn net.Conn, config *CipherConfig) *CipherConn {
	return &CipherConn{
		conn:     conn,
		config:   config,
		isClient: true,
	}
}

func CipherServer(conn net.Conn, config *CipherConfig) *CipherConn {
	return &CipherConn{
		conn:     conn,
		config:   config,
		isClient: true,
	}
}

func (c *CipherConn) Handshake() (err error) {
	var nonce []byte
	if c.isClient {
		// Rand Reader
		var r io.Reader
		if c.config.Rand != nil {
			r = c.config.Rand
		} else {
			r = rand.Reader
		}
		// Generate IV
		nonce, err = ioutil.ReadAll(&io.LimitedReader{r, int64(c.config.NonceSize)})
		if err != nil {
			return err
		}
		// Send IV
		_, err = c.conn.Write(nonce)
		if err != nil {
			return err
		}
	} else {
		// Read IV
		nonce, err = ioutil.ReadAll(&io.LimitedReader{c.conn, int64(c.config.NonceSize)})
		if err != nil {
			return err
		}
	}
	// Read Stream
	c.rs, err = c.config.NewCipherStream(c.config.Key, nonce)
	if err != nil {
		return err
	}
	// Write Stream
	c.ws, err = c.config.NewCipherStream(c.config.Key, nonce)
	if err != nil {
		return err
	}
	return nil
}

func (c *CipherConn) Read(b []byte) (n int, err error) {
	return c.conn.Read(b)
}

func (c *CipherConn) Write(b []byte) (n int, err error) {
	return c.conn.Write(b)
}

func (c *CipherConn) Close() error {
	return c.conn.Close()
}

func (c *CipherConn) LocalAddr() net.Addr {
	return c.conn.LocalAddr()
}

func (c *CipherConn) RemoteAddr() net.Addr {
	return c.conn.RemoteAddr()
}

func (c *CipherConn) SetDeadline(t time.Time) error {
	return c.conn.SetDeadline(t)
}

func (c *CipherConn) SetReadDeadline(t time.Time) error {
	return c.conn.SetReadDeadline(t)
}

func (c *CipherConn) SetWriteDeadline(t time.Time) error {
	return c.conn.SetWriteDeadline(t)
}

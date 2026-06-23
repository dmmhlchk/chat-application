package nats

import (
	"fmt"
	"time"

	"github.com/nats-io/nats.go"
)

// Config for the NATS JS client
type Config struct {
	URL            string
	MaxReconnecs   int
	ReconnecWait   time.Duration
	ConnectTimeout time.Duration
}

type Client struct {
	conn *nats.Conn
	js   nats.JetStreamContext
}

func NewClient(cfg Config) (*Client, error) {
	opts := []nats.Option{
		nats.MaxReconnects(cfg.MaxReconnecs),
		nats.ReconnectWait(cfg.ReconnecWait),
		nats.Timeout(cfg.ConnectTimeout),

		// These three handlers exist purely for observability
		// Without them, a dropped connection fails silently and you'll
		// only notice when messages stop flowing - much harder to debug
		nats.DisconnectErrHandler(func(nc *nats.Conn, err error) {
			if err != nil {
				fmt.Printf("nats - disconnected due to error: %v\n", err)
			}
		}),
		nats.ReconnectHandler(func(nc *nats.Conn) {
			fmt.Printf("nats - reconnected to %s\n", nc.ConnectedUrl())
		}),
		nats.ClosedHandler(func(nc *nats.Conn) {
			fmt.Printf("nats - connection closed\n")
		}),
	}

	conn, err := nats.Connect(cfg.URL, opts...)
	if err != nil {
		return nil, fmt.Errorf("nats - failed to connect: %v\n", err)
	}

	js, err := conn.JetStream()
	if err != nil {
		conn.Close()
		return nil, fmt.Errorf("nats - failed to get jetstream context: %v\n", err)
	}

	return &Client{conn: conn, js: js}, nil
}

func (c *Client) JetStream() *nats.JetStreamContext {
	return &c.js
}

func (c *Client) Conn() *nats.Conn {
	return c.conn
}

func (c *Client) Close() {
	// Drain - stop accepting new work, but finish processing that already in and then close
	c.conn.Drain()
}

package extra

import (
	"context"
	"log"
	"sync"

	"github.com/cooldogedev/spectral"
)

// Client represents a persistent connection to the server.
type Client struct {
	conn   spectral.Connection
	stream *spectral.Stream
	mu     sync.Mutex
}

// NewClient establishes a persistent connection to the server.
func NewClient(address string) (*Client, error) {
	ctx := context.Background()

	conn, err := spectral.Dial(ctx, address)
	if err != nil {
		return nil, err
	}

	stream, err := conn.OpenStream(ctx)
	if err != nil {
		err := conn.CloseWithError(0, "Failed to open stream")
		if err != nil {
			return nil, err
		}
		return nil, err
	}

	client := &Client{
		conn:   conn,
		stream: stream,
	}

	return client, nil
}

// Send pushes byte data to the server.
func (c *Client) Send(data []byte) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	_, err := c.stream.Write(data)
	return err
}

// Close gracefully closes the connection.
func (c *Client) Close() {
	c.mu.Lock()
	defer c.mu.Unlock()

	if err := c.stream.Close(); err != nil {
		log.Println("Error closing stream:", err)
	}

	if err := c.conn.CloseWithError(0, "Client closing"); err != nil {
		log.Println("Error closing connection:", err)
	}
}

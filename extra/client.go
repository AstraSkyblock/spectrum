package extra

import (
	"context"
	"github.com/astraskyblock/spectral"
	"log"
	"sync"
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

	// Start automatic flushing in a goroutine
	go client.autoFlush()

	return client, nil
}

// Send pushes data to the server.
func (c *Client) Send(data string) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	_, err := c.stream.Write([]byte(data))
	return err
}

// autoFlush continuously reads responses from the server.
func (c *Client) autoFlush() {
	buf := make([]byte, 1024)

	for {
		n, err := c.stream.Read(buf)
		if err != nil {
			log.Println("Error reading from stream (server likely closed):", err)
			return
		}

		log.Printf("Received echo: %s", string(buf[:n]))
	}
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

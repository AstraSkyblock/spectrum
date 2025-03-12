package extra

import (
	"context"
	"log"
	"sync"

	"github.com/cooldogedev/spectral"
)

const (
	ClientPacket = 0x01 // Identifier for client packets
	ServerPacket = 0x02 // Identifier for server packets
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
		_ = conn.CloseWithError(0, "Failed to open stream")
		return nil, err
	}

	client := &Client{
		conn:   conn,
		stream: stream,
	}

	return client, nil
}

// Send pushes byte data to the server.
func (c *Client) send(data []byte) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	_, err := c.stream.Write(data)
	return err
}

// sendClient sends data to the server, marking it as a client packet.
func (c *Client) SendClient(data []byte) error {
	packet := append([]byte{ClientPacket}, data...) // Prepend the identifier
	return c.send(packet)
}

// sendServer sends data to the server, marking it as a server packet.
func (c *Client) SendServer(data []byte) error {
	packet := append([]byte{ServerPacket}, data...) // Prepend the identifier
	return c.send(packet)
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

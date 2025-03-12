package extra

import (
	"bytes"
	"context"
	"encoding/binary"
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

// send pushes byte data to the server with length prefix.
func (c *Client) send(packetType byte, data []byte) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	// Packet format: [Size (2 bytes)] [Type (1 byte)] [Data (N bytes)]
	size := uint16(len(data) + 1) // Include type byte in size
	buf := new(bytes.Buffer)

	_ = binary.Write(buf, binary.BigEndian, size) // Write 2-byte length
	buf.WriteByte(packetType)                     // Write packet type
	buf.Write(data)                               // Write actual data

	_, err := c.stream.Write(buf.Bytes())
	return err
}

// SendClient sends data to the server, marking it as a client packet.
func (c *Client) SendClient(data []byte) error {
	return c.send(ClientPacket, data)
}

// SendServer sends data to the server, marking it as a server packet.
func (c *Client) SendServer(data []byte) error {
	return c.send(ServerPacket, data)
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

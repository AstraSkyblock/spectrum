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
func (c *Client) send(packetType byte, data []byte, id *string) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	// Ensure the identity has a null-terminator
	var identity []byte
	if id != nil {
		identity = []byte(*id + "\x00") // Add null terminator to the identity string
	} else {
		identity = []byte("\x00") // If no identity, just use the null terminator
	}

	// Calculate packet size: size of identity + size of data
	packetSize := uint16(len(identity) + len(data) + 1)

	buf := new(bytes.Buffer)

	// Write size (2 bytes)
	_ = binary.Write(buf, binary.BigEndian, packetSize)

	// Write packet type (1 byte)
	buf.WriteByte(packetType)

	// Write identity and data (identity is already null-terminated)
	buf.Write(identity) // Write identity
	buf.Write(data)     // Write actual data

	_, err := c.stream.Write(buf.Bytes()) // Write to the stream
	return err
}

// SendClient sends data to the server, marking it as a client packet.
func (c *Client) SendClient(data []byte, id *string) error {
	return c.send(ClientPacket, data, id)
}

// SendServer sends data to the server, marking it as a server packet.
func (c *Client) SendServer(data []byte, id *string) error {
	return c.send(ServerPacket, data, id)
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

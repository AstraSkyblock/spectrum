package extra

import (
	"bytes"
	"context"
	"encoding/binary"
	"github.com/AstraSkyblock/spectrum/session"
	"github.com/sandertv/gophertunnel/minecraft"
	"github.com/sandertv/gophertunnel/minecraft/protocol/packet"
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
	registry *session.Registry
	protocol minecraft.Protocol
	pool     packet.Pool

	conn   spectral.Connection
	stream *spectral.Stream
	mu     sync.Mutex
}

// NewClient establishes a persistent connection to the server.
func NewClient(address string, proto minecraft.Protocol, registry *session.Registry) (*Client, error) {
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
		conn:     conn,
		stream:   stream,
		protocol: proto,
		pool:     proto.Packets(false),
	}

	go client.readPacket()

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

// readPacket reads a packet from the server and processes it.
func (c *Client) readPacket() {
	for {
		sizeBuf := make([]byte, 2)
		_, err := c.stream.Read(sizeBuf)
		if err != nil {
			log.Println("Stream closed while reading size:", err)
			return
		}

		packetSize := binary.BigEndian.Uint16(sizeBuf)
		if packetSize < 1 {
			continue
		}

		buf := make([]byte, packetSize)
		_, err = c.stream.Read(buf)
		if err != nil {
			log.Println("Error reading packet:", err)
			return
		}

		packetType := buf[0]
		payload := buf[1:]

		switch packetType {
		case ClientPacket:
			c.readClient(payload)
		case ServerPacket:
			c.readServer(payload)
		default:
		}
	}
}

// readClientPacket processes a client packet.
func (c *Client) readClient(payload []byte) {
	// Create a buffer from the full payload
	buf := bytes.NewBuffer(payload)

	// Read the identity (null-terminated string)
	identity, err := buf.ReadString(0)
	if err != nil && err.Error() != "EOF" {
		log.Println("Failed to read identity:", err)
		return
	}

	if identity == "" {
		log.Println("Received empty identity")
		return
	}

	// Remove the null terminator from the identity
	identity = identity[:len(identity)-1]

	s := c.registry.GetSession(identity)

	// Read the packet header
	header := &packet.Header{}
	if err := header.Read(buf); err != nil {
		log.Println("Failed to read packet header:", err)
		return
	}

	// Handle panic recovery for packet decoding
	defer func() {
		if r := recover(); r != nil {
			log.Printf("Panic while decoding packet %v (Identity: %s): %v", header.PacketID, identity, r)
		}
	}()

	// Look up the packet handler from the pool
	factory, ok := c.pool[header.PacketID]
	if !ok {
		log.Println("Unknown packet ID:", header.PacketID)
		return
	}

	// Create a packet and marshal it using the buffer reader
	pk := factory()
	pk.(packet.Packet).Marshal(c.protocol.NewReader(buf, 1, false))

	s.Client().WritePacket(pk)
}

// readServerPacket processes a server packet.
func (c *Client) readServer(payload []byte) {
	// Create a buffer from the full payload
	buf := bytes.NewBuffer(payload)

	// Read the identity (null-terminated string)
	identity, err := buf.ReadString(0)
	if err != nil && err.Error() != "EOF" {
		log.Println("Failed to read identity:", err)
		return
	}

	if identity == "" {
		log.Println("Received empty identity")
		return
	}

	// Remove the null terminator from the identity
	identity = identity[:len(identity)-1]

	s := c.registry.GetSession(identity)

	// Read the packet header
	header := &packet.Header{}
	if err := header.Read(buf); err != nil {
		log.Println("Failed to read packet header:", err)
		return
	}

	// Handle panic recovery for packet decoding
	defer func() {
		if r := recover(); r != nil {
			log.Printf("Panic while decoding packet %v (Identity: %s): %v", header.PacketID, identity, r)
		}
	}()

	// Look up the packet handler from the pool
	factory, ok := c.pool[header.PacketID]
	if !ok {
		log.Println("Unknown packet ID:", header.PacketID)
		return
	}

	// Create a packet and marshal it using the buffer reader
	pk := factory()
	pk.(packet.Packet).Marshal(c.protocol.NewReader(buf, 1, true))

	s.Server().WritePacket(pk)
}

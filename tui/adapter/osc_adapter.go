package adapter

import (
	"github.com/hypebeast/go-osc/osc"
)

// OSCAdapter provides generic OSC communication
// Used for pattern control and TUI communication with SuperCollider
type OSCAdapter struct {
	client *osc.Client
	host   string
	port   int
}

// NewOSCAdapter creates a new OSC adapter
// host: target host (e.g., "localhost" or "127.0.0.1")
// port: target port (e.g., 57120 for SuperCollider sclang)
func NewOSCAdapter(host string, port int) (*OSCAdapter, error) {
	client := osc.NewClient(host, port)

	return &OSCAdapter{
		client: client,
		host:   host,
		port:   port,
	}, nil
}

// GetHost returns the current OSC host
func (o *OSCAdapter) GetHost() string {
	return o.host
}

// GetPort returns the current OSC port
func (o *OSCAdapter) GetPort() int {
	return o.port
}

// SetTarget changes the OSC target host and port
func (o *OSCAdapter) SetTarget(host string, port int) {
	o.host = host
	o.port = port
	o.client = osc.NewClient(host, port)
}

// Send sends an OSC message with the given address and arguments
func (o *OSCAdapter) Send(address string, args ...interface{}) error {
	msg := osc.NewMessage(address)
	for _, arg := range args {
		msg.Append(arg)
	}
	return o.client.Send(msg)
}

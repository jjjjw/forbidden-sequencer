package adapter

import "fmt"

// SetupSClangAdapter creates and configures an OSC adapter for sclang pattern control
// Sends control messages to SuperCollider lang (port 57120) for pattern control
func SetupSClangAdapter() (*OSCAdapter, error) {
	// Initialize OSC adapter for sclang (localhost:57120 is sclang default port)
	sclangAdapter, err := NewOSCAdapter("localhost", 57120)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize sclang OSC adapter: %w", err)
	}

	return sclangAdapter, nil
}

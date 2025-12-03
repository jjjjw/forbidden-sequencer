package adapters

import "fmt"

// SetupOSCAdapter creates and configures an OSC adapter with default mappings
// for the Forbidden Sequencer
func SetupOSCAdapter() (*OSCAdapter, error) {
	// Initialize OSC adapter (localhost:57121 is our custom SuperCollider OSC port)
	oscAdapter, err := NewOSCAdapter("localhost", 57121)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize OSC adapter: %w", err)
	}

	// Set OSC address mappings for drum sounds
	oscAdapter.SetAddressMapping("kick", "/trigger/kick")
	oscAdapter.SetAddressMapping("snare", "/trigger/snare")
	oscAdapter.SetAddressMapping("hihat", "/trigger/hihat")

	return oscAdapter, nil
}

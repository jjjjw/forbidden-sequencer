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

// SetupSuperColliderAdapter creates and configures a SuperCollider adapter
// with default mappings for the Forbidden Sequencer
func SetupSuperColliderAdapter() (*SuperColliderAdapter, error) {
	// Initialize SuperCollider adapter (localhost:57110 is scsynth default port)
	scAdapter, err := NewSuperColliderAdapter("localhost", 57110)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize SuperCollider adapter: %w", err)
	}

	// Set SynthDef mappings (event name -> SynthDef name)
	scAdapter.SetSynthDefMapping("kick", "bd")
	scAdapter.SetSynthDefMapping("snare", "cp")
	scAdapter.SetSynthDefMapping("hihat", "hh")

	// Set Group ID mappings (event name -> Group ID)
	// These must match the group IDs created in Supercollider/setup.scd
	scAdapter.SetGroupID("kick", 100)
	scAdapter.SetGroupID("snare", 101)
	scAdapter.SetGroupID("hihat", 102)

	// Set Bus ID mappings (event name -> output bus)
	// Bus 0 = master out (default)
	// Bus 10 = reverb bus (for snare/clap)
	scAdapter.SetBusID("snare", 10) // route snare to reverb
	// kick and hihat use default (bus 0)

	return scAdapter, nil
}

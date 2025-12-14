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
func SetupSuperColliderAdapter(debug bool) (*SuperColliderAdapter, error) {
	// Initialize SuperCollider adapter (localhost:57110 is scsynth default port)
	scAdapter, err := NewSuperColliderAdapter("localhost", 57110, debug)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize SuperCollider adapter: %w", err)
	}

	// Set SynthDef mappings (event name -> SynthDef name)
	scAdapter.SetSynthDefMapping("kick", "bd")
	scAdapter.SetSynthDefMapping("snare", "cp")
	scAdapter.SetSynthDefMapping("hihat", "hh")
	scAdapter.SetSynthDefMapping("fm1", "fm2op")
	scAdapter.SetSynthDefMapping("fm2", "fm2op")

	// Set Group ID mappings (event name -> Group ID)
	// These must match the group IDs created in Supercollider/setup.scd
	scAdapter.SetGroupID("kick", 100)
	scAdapter.SetGroupID("snare", 101)
	scAdapter.SetGroupID("hihat", 102)
	scAdapter.SetGroupID("fm1", 103)
	scAdapter.SetGroupID("fm2", 104)

	// Set Bus ID mappings (event name -> output bus)
	// Bus 0 = master out (default)
	// Bus 10 = reverb bus (for snare/clap and FM voices)
	scAdapter.SetBusID("snare", 10) // route snare to reverb
	scAdapter.SetBusID("fm1", 10)   // route fm1 to reverb
	scAdapter.SetBusID("fm2", 10)   // route fm2 to reverb
	// kick and hihat use default (bus 0)

	// Set Parameter mappings (event name -> parameter control names)
	// Maps Event.A/B/C/D to synth control parameter names
	scAdapter.SetParameterMapping("kick", ParameterMapping{
		A: "freq",  // Event.A controls frequency
		B: "amp",   // Event.B controls amplitude
		C: "ratio", // Event.C controls frequency ratio
		D: "sweep", // Event.D controls sweep time
	})
	scAdapter.SetParameterMapping("snare", ParameterMapping{
		A: "freq", // Event.A controls frequency (not used much in cp)
		B: "amp",  // Event.B controls amplitude
		C: "",     // Event.C unused
		D: "",     // Event.D unused
	})
	scAdapter.SetParameterMapping("hihat", ParameterMapping{
		A: "freq", // Event.A controls frequency (not used much in hh)
		B: "amp",  // Event.B controls amplitude
		C: "",     // Event.C unused
		D: "",     // Event.D unused
	})
	scAdapter.SetParameterMapping("fm1", ParameterMapping{
		A: "freq",     // Event.A controls carrier frequency
		B: "amp",      // Event.B controls amplitude
		C: "modRatio", // Event.C controls modulator ratio
		D: "modIndex", // Event.D controls modulation index
	})
	scAdapter.SetParameterMapping("fm2", ParameterMapping{
		A: "freq",     // Event.A controls carrier frequency
		B: "amp",      // Event.B controls amplitude
		C: "modRatio", // Event.C controls modulator ratio
		D: "modIndex", // Event.D controls modulation index
	})

	return scAdapter, nil
}

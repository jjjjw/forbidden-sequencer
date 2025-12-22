package sequencers

import (
	"forbidden_sequencer/sequencer/conductors"
)

// ModuleFactory creates module configs on demand
type ModuleFactory interface {
	// GetName returns the display name for this module type
	GetName() string

	// Create creates a new module config instance
	Create(conductor *conductors.Conductor) ModuleConfig
}

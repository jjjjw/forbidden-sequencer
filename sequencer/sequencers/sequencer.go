package sequencers

import (
	"fmt"
	"log"
	"time"

	"forbidden_sequencer/sequencer/adapters"
	"forbidden_sequencer/sequencer/conductors"
	"forbidden_sequencer/sequencer/events"
)

// Pattern generates scheduled events
type Pattern interface {
	GetNextScheduledEvent() (events.ScheduledEvent, error)
}

// Sequencer manages multiple patterns and outputs events through an adapter
type Sequencer struct {
	patterns  []Pattern
	adapter   adapters.EventAdapter
	conductor conductors.Conductor
	debug     bool
	running   bool
	paused    bool
	stopCh    chan struct{}
}

// NewSequencer creates a new sequencer with the given patterns, conductor, and adapter
func NewSequencer(patterns []Pattern, conductor conductors.Conductor, adapter adapters.EventAdapter, debug bool) *Sequencer {
	return &Sequencer{
		patterns:  patterns,
		conductor: conductor,
		adapter:   adapter,
		debug:     debug,
		running:   false,
	}
}

// Start starts or restarts the sequencer
func (s *Sequencer) Start() {
	// Stop any existing run
	if s.running {
		close(s.stopCh)
	}

	s.running = true
	s.stopCh = make(chan struct{})

	// Start conductor
	s.conductor.Start()

	// Start a goroutine for each pattern
	for i := range s.patterns {
		go s.runPattern(i)
	}
}

// Pause pauses playback (conductor stops ticking, patterns stop generating events)
func (s *Sequencer) Pause() {
	s.paused = true
	s.conductor.Pause()
}

// Resume resumes playback (conductor resumes ticking, patterns resume generating events)
func (s *Sequencer) Resume() {
	s.paused = false
	s.conductor.Resume()
}

// runPattern runs a single pattern's event loop
func (s *Sequencer) runPattern(index int) {
	pattern := s.patterns[index]

	for {
		select {
		case <-s.stopCh:
			return
		default:
			// Check if paused - if so, sleep briefly and continue
			if s.paused {
				time.Sleep(10 * time.Millisecond)
				continue
			}

			// Get next scheduled event from pattern
			scheduled, err := pattern.GetNextScheduledEvent()
			if err != nil {
				s.handleError(fmt.Sprintf("pattern %d error: %v", index, err))
				continue
			}

			// Wait until it's time to fire the event
			time.Sleep(scheduled.Timing.Delta)

			// Send to adapter (fires immediately)
			if s.adapter != nil {
				if err := s.adapter.Send(scheduled); err != nil {
					s.handleError(fmt.Sprintf("pattern %d adapter error: %v", index, err))
					continue
				}
			}
		}
	}
}

// handleError handles errors based on debug mode
func (s *Sequencer) handleError(msg string) {
	if s.debug {
		// Debug mode: log and panic with stack
		log.Panic(msg)
	} else {
		// Perf mode: log and sleep
		log.Println(msg)
		time.Sleep(10 * time.Millisecond)
	}
}

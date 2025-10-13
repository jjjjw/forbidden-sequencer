package sequencer

import (
	"fmt"
	"log"
	"time"

	"forbidden_sequencer/sequencer/adapters"
	"forbidden_sequencer/sequencer/event_generators"
	"forbidden_sequencer/sequencer/timing_generators"
)

// Lane pairs an event generator with a timing generator
type Lane struct {
	EventGenerator  event_generators.EventGenerator
	TimingGenerator timing_generators.TimingGenerator
}

// Sequencer manages multiple lanes and outputs events through an adapter
type Sequencer struct {
	lanes   []*Lane
	adapter adapters.EventAdapter
	debug   bool
	running bool
	paused  bool
	stopCh  chan struct{}
}

// NewSequencer creates a new sequencer with the given lanes and adapter
func NewSequencer(lanes []*Lane, adapter adapters.EventAdapter, debug bool) *Sequencer {
	return &Sequencer{
		lanes:   lanes,
		adapter: adapter,
		debug:   debug,
		running: false,
		paused:  false,
	}
}

// Start starts or restarts the sequencer
func (s *Sequencer) Start() {
	// Stop any existing run
	if s.running {
		close(s.stopCh)
	}

	s.running = true
	s.paused = false
	s.stopCh = make(chan struct{})

	// Start a goroutine for each lane
	for i := range s.lanes {
		go s.runLane(i)
	}
}

// Pause pauses playback
func (s *Sequencer) Pause() {
	s.paused = true
}

// Resume resumes playback
func (s *Sequencer) Resume() {
	s.paused = false
}

// runLane runs a single lane's event loop
func (s *Sequencer) runLane(index int) {
	lane := s.lanes[index]

	for {
		select {
		case <-s.stopCh:
			return
		default:
			// Check if paused
			if s.paused {
				time.Sleep(10 * time.Millisecond)
				continue
			}

			// Get timing from timing generator
			wait, duration, err := lane.TimingGenerator.GetNextTiming()
			if err != nil {
				s.handleError(fmt.Sprintf("lane %d timing error: %v", index, err))
				continue
			}

			// Get event from event generator
			event, err := lane.EventGenerator.GetNextEvent()
			if err != nil {
				s.handleError(fmt.Sprintf("lane %d event error: %v", index, err))
				continue
			}

			// Create scheduled event
			scheduled := ScheduledEvent{
				Event: event,
				Timing: Timing{
					Wait:     wait,
					Duration: duration,
				},
			}

			// Send to adapter
			if s.adapter != nil {
				if err := s.adapter.Send(scheduled); err != nil {
					s.handleError(fmt.Sprintf("lane %d adapter error: %v", index, err))
					continue
				}
			}

			// Wait before generating next event
			time.Sleep(wait)
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

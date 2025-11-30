package sequencers

import (
	"fmt"
	"log"
	"time"

	"forbidden_sequencer/sequencer/adapters"
	"forbidden_sequencer/sequencer/conductors"
	"forbidden_sequencer/sequencer/events"
)

// Pattern generates scheduled events for a given tick
type Pattern interface {
	GetScheduledEventsForTick(nextTickTime time.Time, tickDuration time.Duration) []events.ScheduledEvent
	Reset()
	Play()
	Stop()
}

// Sequencer manages multiple patterns and outputs events through an adapter
type Sequencer struct {
	patterns   []Pattern
	adapter    adapters.EventAdapter
	conductor  conductors.Conductor
	debug      bool
	eventsChan chan<- events.ScheduledEvent
}

// NewSequencer creates a new sequencer with the given patterns, conductor, and adapter
func NewSequencer(patterns []Pattern, conductor conductors.Conductor, adapter adapters.EventAdapter, eventsChan chan<- events.ScheduledEvent, debug bool) *Sequencer {
	return &Sequencer{
		patterns:   patterns,
		conductor:  conductor,
		adapter:    adapter,
		eventsChan: eventsChan,
		debug:      debug,
	}
}

// Start initializes and starts the sequencer
func (s *Sequencer) Start() {
	// Start conductor
	s.conductor.Start()

	// Start tick-driven event loop
	go s.runTickLoop()
}

// runTickLoop listens for conductor ticks and schedules events from patterns
func (s *Sequencer) runTickLoop() {
	for range s.conductor.Ticks() {
		// Get timing info for next tick
		nextTickTime := time.Now().Add(s.conductor.GetTickDuration())
		tickDuration := s.conductor.GetTickDuration()

		// Collect events from all patterns
		for _, pattern := range s.patterns {
			scheduledEvents := pattern.GetScheduledEventsForTick(nextTickTime, tickDuration)
			for _, scheduled := range scheduledEvents {
				s.scheduleEvent(scheduled)
			}
		}
	}
}

// scheduleEvent schedules a single event to fire at its timestamp
func (s *Sequencer) scheduleEvent(scheduled events.ScheduledEvent) {
	time.AfterFunc(time.Until(scheduled.Timing.Timestamp), func() {
		// Send to adapter
		if s.adapter != nil {
			if s.debug {
				log.Println("Sending message %", scheduled)
			}
			if err := s.adapter.Send(scheduled); err != nil {
				s.handleError(fmt.Sprintf("adapter error: %v", err))
			}
		}

		// Send event to channel for TUI display
		if s.eventsChan != nil {
			select {
			case s.eventsChan <- scheduled:
			default:
				// Don't block if channel is full
			}
		}
	})
}

// Stop pauses all patterns
func (s *Sequencer) Stop() {
	for _, pattern := range s.patterns {
		pattern.Stop()
	}
}

// Play resumes all patterns
func (s *Sequencer) Play() {
	for _, pattern := range s.patterns {
		pattern.Play()
	}
}

// Reset resets all patterns
func (s *Sequencer) Reset() {
	for _, pattern := range s.patterns {
		pattern.Reset()
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

// String returns a string representation of the sequencer
func (s *Sequencer) String() string {
	var patternStrs []string
	for _, p := range s.patterns {
		patternStrs = append(patternStrs, fmt.Sprintf("%v", p))
	}
	return fmt.Sprintf("Patterns: %v", patternStrs)
}

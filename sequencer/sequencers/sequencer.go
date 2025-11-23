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
	Reset()
	Play()
	Stop()
}

// Sequencer manages multiple patterns and outputs events through an adapter
type Sequencer struct {
	patterns  []Pattern
	adapter   adapters.EventAdapter
	conductor conductors.Conductor
	debug     bool
	Events    chan events.ScheduledEvent
}

// NewSequencer creates a new sequencer with the given patterns, conductor, and adapter
func NewSequencer(patterns []Pattern, conductor conductors.Conductor, adapter adapters.EventAdapter, debug bool) *Sequencer {
	return &Sequencer{
		patterns:  patterns,
		conductor: conductor,
		adapter:   adapter,
		debug:     debug,
		Events:    make(chan events.ScheduledEvent, 100),
	}
}

// Start initializes and starts the sequencer
func (s *Sequencer) Start() {
	// Start conductor
	s.conductor.Start()

	// Schedule first event for each pattern
	for i := range s.patterns {
		s.scheduleNextEvent(i)
	}
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

// scheduleNextEvent schedules the next event for a pattern using AfterFunc
func (s *Sequencer) scheduleNextEvent(index int) {
	pattern := s.patterns[index]

	// Get next scheduled event from pattern
	scheduled, err := pattern.GetNextScheduledEvent()
	if err != nil {
		s.handleError(fmt.Sprintf("pattern %d error: %v", index, err))
		// Retry after a short delay
		time.AfterFunc(10*time.Millisecond, func() {
			s.scheduleNextEvent(index)
		})
		return
	}

	// Schedule the event to fire at Timestamp
	time.AfterFunc(time.Until(scheduled.Timing.Timestamp), func() {
		// Send to adapter
		if s.adapter != nil {
			if s.debug {
				log.Println("Sending message %", scheduled)
			}
			if err := s.adapter.Send(scheduled); err != nil {
				s.handleError(fmt.Sprintf("pattern %d adapter error: %v", index, err))
			}
		}

		// Send event to channel for TUI display
		if s.Events != nil {
			select {
			case s.Events <- scheduled:
			default:
				// Don't block if channel is full
			}
		}

		// Schedule next event
		s.scheduleNextEvent(index)
	})
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

// GetEventsChannel returns the channel for scheduled events
func (s *Sequencer) GetEventsChannel() chan events.ScheduledEvent {
	return s.Events
}

// GetBeatsChannel returns the conductor's beats channel
func (s *Sequencer) GetBeatsChannel() chan struct{} {
	return s.conductor.GetBeatsChannel()
}

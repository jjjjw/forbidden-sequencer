package sequencers

import (
	"fmt"
	"log"
	"os"
	"time"

	"forbidden_sequencer/sequencer/adapters"
	"forbidden_sequencer/sequencer/conductors"
	"forbidden_sequencer/sequencer/events"
)

// Pattern generates tick-based events for a given tick
type Pattern interface {
	GetEventsForTick(tick int64) []events.TickEvent
	Reset()
	Play()
	Stop()
}

// Sequencer manages multiple patterns and outputs events through an adapter
type Sequencer struct {
	patterns      []Pattern
	adapter       adapters.EventAdapter
	conductor     *conductors.Conductor
	eventsChan    chan<- events.ScheduledEvent
	lookaheadMs   int   // lookahead window in milliseconds
	currentTick   int64 // current logical tick
	eventBuffers  map[int]map[int64][]events.TickEvent // pattern index -> tick -> events
	debug         bool
	debugLog      *log.Logger
}

// NewSequencer creates a new sequencer with the given patterns, conductor, and adapter
func NewSequencer(patterns []Pattern, conductor *conductors.Conductor, adapter adapters.EventAdapter, eventsChan chan<- events.ScheduledEvent, debug bool) *Sequencer {
	// Initialize event buffers for each pattern
	eventBuffers := make(map[int]map[int64][]events.TickEvent)
	for i := range patterns {
		eventBuffers[i] = make(map[int64][]events.TickEvent)
	}

	// Set up debug logging if enabled
	var debugLogger *log.Logger
	if debug {
		// Create debug log file
		debugFile, err := os.Create("debug/sequencer.log")
		if err == nil {
			debugLogger = log.New(debugFile, "", log.LstdFlags|log.Lmicroseconds)
		}
	}

	return &Sequencer{
		patterns:     patterns,
		conductor:    conductor,
		adapter:      adapter,
		eventsChan:   eventsChan,
		lookaheadMs:  25, // 25ms default lookahead
		currentTick:  0,
		eventBuffers: eventBuffers,
		debug:        debug,
		debugLog:     debugLogger,
	}
}

// Start initializes and starts the sequencer
func (s *Sequencer) Start() {
	// Start conductor
	s.conductor.Start()

	// Start tick-driven event loop
	go s.runTickLoop()
}

// runTickLoop periodically generates events to maintain lookahead window
func (s *Sequencer) runTickLoop() {
	ticker := time.NewTicker(time.Duration(s.lookaheadMs) * time.Millisecond)
	defer ticker.Stop()

	if s.debugLog != nil {
		s.debugLog.Printf("Sequencer runTickLoop started, lookaheadMs=%d", s.lookaheadMs)
	}

	for range ticker.C {
		now := time.Now()
		lastTickTime := s.conductor.GetLastTickTime()
		tickDuration := s.conductor.GetTickDuration()

		// Calculate lookahead range in ticks (minimum 1 tick)
		lookaheadDuration := time.Duration(s.lookaheadMs) * time.Millisecond
		lookaheadTicks := int64(lookaheadDuration / tickDuration)
		if lookaheadTicks < 1 {
			lookaheadTicks = 1
		}

		// Determine current tick based on time elapsed since last conductor tick
		timeSinceLastTick := now.Sub(lastTickTime)
		ticksElapsed := int64(timeSinceLastTick / tickDuration)
		s.currentTick = s.currentTick + ticksElapsed

		// Calculate the tick we need to have events scheduled through
		targetTick := s.currentTick + lookaheadTicks

		if s.debugLog != nil {
			s.debugLog.Printf("Lookahead cycle: now=%v lastTickTime=%v tickDuration=%v currentTick=%d targetTick=%d lookaheadTicks=%d",
				now.Format("15:04:05.000"), lastTickTime.Format("15:04:05.000"), tickDuration, s.currentTick, targetTick, lookaheadTicks)
		}

		// Generate and schedule events for ticks in lookahead window
		for i, pattern := range s.patterns {
			buffer := s.eventBuffers[i]

			// Find the highest tick we've already generated
			maxGenerated := s.currentTick - 1
			for tick := range buffer {
				if tick > maxGenerated {
					maxGenerated = tick
				}
			}

			// Generate events from after the last generated tick up to target
			for tick := maxGenerated + 1; tick <= targetTick; tick++ {
				// Request events for this tick from pattern
				tickEvents := pattern.GetEventsForTick(tick)

				if s.debugLog != nil && len(tickEvents) > 0 {
					s.debugLog.Printf("Pattern %d generated %d events for tick %d", i, len(tickEvents), tick)
				}

				// Schedule these events immediately (adapter handles timing)
				for _, tickEvent := range tickEvents {
					// Calculate absolute timestamp from tick-relative information
					ticksFromLastTick := tickEvent.Tick - (s.currentTick - ticksElapsed)
					timeOfEventTick := lastTickTime.Add(time.Duration(ticksFromLastTick) * tickDuration)
					timestamp := timeOfEventTick.Add(time.Duration(float64(tickDuration) * tickEvent.OffsetPercent))

					// Convert duration from ticks to wall-clock time
					duration := time.Duration(float64(tickDuration) * tickEvent.DurationTicks)

					scheduled := events.ScheduledEvent{
						Event: tickEvent.Event,
						Timing: events.Timing{
							Timestamp: timestamp,
							Duration:  duration,
						},
					}

					if s.debugLog != nil {
						s.debugLog.Printf("Scheduling event: name=%s tick=%d timestamp=%v duration=%v",
							tickEvent.Event.Name, tickEvent.Tick, timestamp.Format("15:04:05.000"), duration)
					}

					s.handleEvent(scheduled)
				}

				// Mark this tick as processed (prevent regenerating)
				buffer[tick] = tickEvents
			}

			// Clean up old buffer entries (before current tick)
			for tick := range buffer {
				if tick < s.currentTick {
					delete(buffer, tick)
				}
			}
		}
	}
}

// handleEvent sends an event to the adapter and TUI
// Adapters are responsible for scheduling (e.g., OSC bundles with timestamps)
func (s *Sequencer) handleEvent(scheduled events.ScheduledEvent) {
	// Send to adapter (adapter handles timing/scheduling)
	if s.adapter != nil {
		if err := s.adapter.Send(scheduled); err != nil {
			// Log error but don't stop playback
			fmt.Printf("adapter error: %v\n", err)
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

// SetPatterns replaces the current patterns with new ones
// This is used when switching modules
func (s *Sequencer) SetPatterns(patterns []Pattern) {
	// Stop current patterns
	s.Stop()

	// Replace patterns
	s.patterns = patterns

	// Reset currentTick to start from beginning
	s.currentTick = 0

	// Clear and reinitialize event buffers
	s.eventBuffers = make(map[int]map[int64][]events.TickEvent)
	for i := range patterns {
		s.eventBuffers[i] = make(map[int64][]events.TickEvent)
	}

	if s.debugLog != nil {
		s.debugLog.Printf("SetPatterns: loaded %d new patterns", len(patterns))
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

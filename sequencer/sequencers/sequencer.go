package sequencers

import (
	"fmt"
	"log"
	"math"
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

// Visualizer is an optional interface for patterns that can visualize their distribution
type Visualizer interface {
	Visualize() string // Returns visual pattern like "x---x--x-"
	GetPhraseLength() int
	GetTickInPhrase() int
	GetPatternName() string
}

// Sequencer manages multiple patterns and outputs events through an adapter
type Sequencer struct {
	patterns          []Pattern
	adapter           adapters.EventAdapter
	conductor         *conductors.Conductor
	eventsChan        chan<- events.ScheduledEvent
	lastScheduledTime time.Time // latest wall-clock time we've scheduled an event for (prevents overlapping on tempo change)
	debug             bool
	debugLog          *log.Logger
}

// calculateLookaheadTicks determines how many ticks ahead to schedule events
// Minimum: 2 ticks
// Target: At least 500ms of lookahead time
func calculateLookaheadTicks(tickDuration time.Duration) int64 {
	const minTicks = 2
	const targetMs = 500 * time.Millisecond

	ticks := int64(math.Ceil(float64(targetMs) / float64(tickDuration)))
	if ticks < minTicks {
		return minTicks
	}
	return ticks
}

// NewSequencer creates a new sequencer with the given patterns, conductor, and adapter
func NewSequencer(patterns []Pattern, conductor *conductors.Conductor, adapter adapters.EventAdapter, eventsChan chan<- events.ScheduledEvent, debug bool) *Sequencer {
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
		patterns:          patterns,
		conductor:         conductor,
		adapter:           adapter,
		eventsChan:        eventsChan,
		lastScheduledTime: time.Time{}, // zero time (before any real timestamp)
		debug:             debug,
		debugLog:          debugLogger,
	}
}

// Start initializes and starts the sequencer
func (s *Sequencer) Start() {
	// Register tick callback with conductor
	s.conductor.SetTickCallback(s.handleTick)

	// Start conductor (which will trigger handleTick on every tick)
	s.conductor.Start()

	if s.debugLog != nil {
		tickDuration := s.conductor.GetTickDuration()
		lookaheadTicks := calculateLookaheadTicks(tickDuration)
		lookaheadMs := time.Duration(lookaheadTicks) * tickDuration
		s.debugLog.Printf("Sequencer started with lookaheadTicks=%d (%v)", lookaheadTicks, lookaheadMs)
	}
}

// handleTick is called by the conductor on every tick
// At tick N, we generate events for tick N + lookaheadTicks
func (s *Sequencer) handleTick(currentTick int64) {
	tickDuration := s.conductor.GetTickDuration()
	lastTickTime := s.conductor.GetLastTickTime()

	// Calculate lookahead based on current tick duration
	lookaheadTicks := calculateLookaheadTicks(tickDuration)

	// Calculate which tick to generate events for
	targetTick := currentTick + lookaheadTicks

	// Calculate what time this target tick would be scheduled at
	ticksInFuture := targetTick - currentTick
	targetTickTime := lastTickTime.Add(time.Duration(ticksInFuture) * tickDuration)

	// Only generate events if we haven't already scheduled past this wall-clock time
	// This prevents overlapping events when tick duration changes
	if !s.lastScheduledTime.IsZero() && !targetTickTime.After(s.lastScheduledTime) {
		if s.debugLog != nil {
			s.debugLog.Printf("Tick %d: skipping tick %d at %v (already scheduled up to %v)",
				currentTick, targetTick, targetTickTime.Format("15:04:05.000"), s.lastScheduledTime.Format("15:04:05.000"))
		}
		return
	}

	if s.debugLog != nil {
		s.debugLog.Printf("Tick %d: generating events for tick %d at %v (lookahead=%d)",
			currentTick, targetTick, targetTickTime.Format("15:04:05.000"), lookaheadTicks)
	}

	// Track the latest timestamp we schedule in this tick
	var latestTimestamp time.Time

	// Generate events for the target tick from all patterns
	for i, pattern := range s.patterns {
		tickEvents := pattern.GetEventsForTick(targetTick)

		if s.debugLog != nil && len(tickEvents) > 0 {
			s.debugLog.Printf("Pattern %d generated %d events for tick %d", i, len(tickEvents), targetTick)
		}

		// Schedule each event
		for _, tickEvent := range tickEvents {
			// Calculate absolute timestamp from tick-relative information
			// How many ticks in the future is this event from the conductor's current position?
			ticksInFuture := tickEvent.TickTiming.Tick - currentTick
			timeOfEventTick := lastTickTime.Add(time.Duration(ticksInFuture) * tickDuration)
			timestamp := timeOfEventTick.Add(time.Duration(float64(tickDuration) * tickEvent.TickTiming.OffsetPercent))

			// Skip events that would be scheduled before our last scheduled time
			if !s.lastScheduledTime.IsZero() && !timestamp.After(s.lastScheduledTime) {
				if s.debugLog != nil {
					s.debugLog.Printf("Skipping event at %v (before lastScheduledTime %v)",
						timestamp.Format("15:04:05.000"), s.lastScheduledTime.Format("15:04:05.000"))
				}
				continue
			}

			// Convert duration from ticks to wall-clock time
			duration := time.Duration(float64(tickDuration) * tickEvent.TickTiming.DurationTicks)

			scheduled := events.ScheduledEvent{
				Event: tickEvent.Event,
				Timing: events.Timing{
					Timestamp: timestamp,
					Duration:  duration,
				},
			}

			if s.debugLog != nil {
				s.debugLog.Printf("Scheduling event: name=%s tick=%d timestamp=%v duration=%v",
					tickEvent.Event.Name, tickEvent.TickTiming.Tick, timestamp.Format("15:04:05.000"), duration)
			}

			s.sendEvent(scheduled)

			// Track the latest timestamp
			if latestTimestamp.IsZero() || timestamp.After(latestTimestamp) {
				latestTimestamp = timestamp
			}
		}
	}

	// Update last scheduled time to the latest timestamp we actually scheduled
	// (or the target tick time if we scheduled nothing)
	if !latestTimestamp.IsZero() {
		s.lastScheduledTime = latestTimestamp
	} else if s.lastScheduledTime.IsZero() || targetTickTime.After(s.lastScheduledTime) {
		s.lastScheduledTime = targetTickTime
	}
}

// sendEvent sends an event to the adapter and TUI
// Adapters are responsible for scheduling (e.g., OSC bundles with timestamps)
func (s *Sequencer) sendEvent(scheduled events.ScheduledEvent) {
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

	// Reset conductor to start from tick 0
	s.conductor.Reset()

	// Reset scheduling state
	s.lastScheduledTime = time.Time{}

	if s.debugLog != nil {
		s.debugLog.Printf("SetPatterns: loaded %d new patterns", len(patterns))
	}
}

// GetPatterns returns the current patterns
func (s *Sequencer) GetPatterns() []Pattern {
	return s.patterns
}

// String returns a string representation of the sequencer
func (s *Sequencer) String() string {
	var patternStrs []string
	for _, p := range s.patterns {
		patternStrs = append(patternStrs, fmt.Sprintf("%v", p))
	}
	return fmt.Sprintf("Patterns: %v", patternStrs)
}

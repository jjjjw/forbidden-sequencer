# Forbidden Sequencer

A pattern-based generative music system using SuperCollider for synthesis and Go for interactive control.

## Architecture

The Forbidden Sequencer uses **SuperCollider patterns** for generative music creation, controlled via **OSC messages** from a **Go TUI** (Terminal User Interface).

### System Components

```
Go TUI (Terminal UI)
    ↓ (OSC control messages)
SuperCollider sclang (port 57120)
    ↓ (pattern control, synth triggering)
SuperCollider scsynth (audio server)
    ↓ (audio output)
Audio Interface
```

**Go TUI** - Interactive terminal interface
- Pattern controller selection
- Real-time parameter control
- Play/pause/stop controls
- Sends OSC messages to sclang port 57120

**SuperCollider Patterns** - Generative music patterns
- Run in sclang (SuperCollider language)
- Receive OSC control messages
- Trigger synths on scsynth server

**OSC Communication** - Control protocol
- Go → sclang: Pattern control messages (port 57120)
- sclang → scsynth: Synth triggering (internal OSC)

## Development

### Building

```bash
go build -o forbidden-sequencer .
./forbidden-sequencer
```

### Adding New Patterns

**1. Create SuperCollider Pattern** (`supercollider/patterns/my_pattern.scd`)

```supercollider
(
// Global state
~myPattern = ~myPattern ?? ();

// Pattern parameters
~myPattern.param1 = 0.5;

// Main task
~myPattern.mainTask = Task({
    inf.do {
        // Pattern logic here
        0.125.yield;  // Wait between events
    };
}, SystemClock);

// OSC Responders
OSCdef(\myPatternPlay, {
    ~myPattern.mainTask.reset;
    ~myPattern.mainTask.start;
}, '/pattern/my_pattern/play');

// ... more OSC handlers
)
```

**2. Create Go Controller** (`tui/controllers/my_pattern.go`)

```go
type MyPatternController struct {
    sclangAdapter *adapter.OSCAdapter
    param1        float64
    isPlaying     bool
}

func (c *MyPatternController) HandleInput(msg tea.KeyMsg) bool {
    // Handle keyboard input
    // Send OSC: c.sendOSC("/pattern/my_pattern/param1", value)
}
```

**3. Register Controller** (`main.go`)

```go
m.AvailableControllers = []controllers.Controller{
    controllers.NewMyPatternController(sclangAdapter),
    // ... other controllers
}
```

## See Also

- [SuperCollider Integration](supercollider/README.md) - Details on audio setup

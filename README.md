# Forbidden Sequencer

A pattern-based generative music system using SuperCollider for synthesis with a web-based control interface.

## Architecture

```
Browser (http://localhost:5173)
    ↓ HTTP POST
OSC Bridge (port 8080)
    ↓ OSC/UDP
SuperCollider sclang (port 57120)
    ↓ pattern control, synth triggering
SuperCollider scsynth (audio server)
    ↓ audio output
Audio Interface
```

### Components

- **Web Frontend** (`web/frontend/`) - Vite + Svelte single-page application
  - Three pattern controllers with sliders and visual feedback
  - Real-time OSC communication via HTTP
  - TailwindCSS styling

- **OSC Bridge** (`web/bridge/`) - Tiny Go HTTP server (~60 lines)
  - Converts HTTP POST requests → OSC/UDP messages
  - Receives JSON from browser, forwards as OSC to SuperCollider
  - Runs on port 8080

- **SuperCollider Patterns** (`supercollider/patterns/`)
  - **Curve Time** - Rhythmic patterns with curved timing
  - **Markov Triggers** - Probabilistic Markov chain triggers
  - **Markov Chord** - Alternating chord/percussion sections

## Setup

### 1. Start SuperCollider

```supercollider
// Boot server and load patterns
s.boot;

// Load setup
"<path>/forbidden_sequencer/supercollider/setup.scd".load;

// Load patterns
"<path>/forbidden_sequencer/supercollider/patterns/curve_time.scd".load;
"<path>/forbidden_sequencer/supercollider/patterns/markov_trig.scd".load;
"<path>/forbidden_sequencer/supercollider/patterns/markov_chord.scd".load;
```

### 2. Start OSC Bridge

```bash
cd web/bridge
go run main.go
```

The bridge will start on port 8080 and forward OSC messages to SuperCollider on port 57120.

### 3. Start Web Frontend

```bash
cd web/frontend
npm install  # First time only
npm run dev
```

Open browser to http://localhost:5173

## Usage

1. Select a pattern using the tabs at the top
2. Use sliders to adjust parameters in real-time
3. Click "Play" to start the pattern
4. Adjust parameters while the pattern is playing
5. Click "Debug" to enable debug logging in SuperCollider

See [`web/README.md`](web/README.md) for detailed web GUI documentation.

## Legacy Terminal UI

A legacy terminal-based interface is available in the `tui/` directory. To use it:

```bash
cd tui
go run .
```

## Development

### Frontend Development

```bash
cd web/frontend
npm run dev    # Start dev server with hot reload
npm run build  # Build for production
npm run preview # Preview production build
```

### Bridge Development

```bash
cd web/bridge
go run main.go  # Run bridge
go build        # Build binary
```

### Project Structure

```
forbidden_sequencer/
├── web/                   # Web GUI (primary interface)
│   ├── bridge/           # HTTP→OSC bridge (Go)
│   └── frontend/         # Vite+Svelte web app
├── supercollider/        # SuperCollider patterns & setup
│   ├── setup.scd
│   ├── patterns/
│   └── lib/
├── tui/                  # Legacy Terminal UI
└── README.md
```

### Adding New Patterns

- **Web Interface** - See [`web/README.md`](web/README.md) for creating Svelte components and controls
- **SuperCollider** - See [`supercollider/README.md`](supercollider/README.md) for pattern implementation and OSC responders

## See Also

- [`web/README.md`](web/README.md) - Full web interface guide
- [`supercollider/README.md`](supercollider/README.md) - Pattern implementation and audio setup

# Forbidden Sequencer Web GUI

A web-based graphical interface for controlling SuperCollider patterns with sliders and visual feedback.

## Architecture

```
Browser (http://localhost:5173)
    ↓ HTTP POST
OSC Bridge (port 8080)
    ↓ OSC/UDP
SuperCollider sclang (port 57120)
```

## Components

### OSC Bridge (`bridge/`)
- Tiny Go HTTP server that converts HTTP POST requests → OSC/UDP messages
- Receives JSON from browser, forwards as OSC to SuperCollider
- Runs on port 8080

### Web Frontend (`frontend/`)
- Vite + Svelte single-page application
- Three pattern controllers with sliders and controls:
  - **Curve Time** - Rhythmic patterns with curved timing
  - **Markov Triggers** - Probabilistic Markov chain triggers
  - **Markov Chord** - Alternating chord/percussion sections
- TailwindCSS styling
- Real-time OSC communication

## Setup

### 1. Start SuperCollider

```supercollider
// Boot server and load patterns
s.boot;

// Load setup
"<path>/forbidden_sequencer/supercollider/setup.scd".load;

// Load patterns (load all three)
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

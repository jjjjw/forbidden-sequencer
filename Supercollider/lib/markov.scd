// Markov Chain Library for SuperCollider
// First-order Markov chain for state transitions
// Matches the Go implementation in sequencer/lib/markov.go

(
// Factory function to create a new Markov chain
~newMarkovChain = { |seed|
	var chain = ();

	// State variables
	chain.transitions = (); // Dictionary of Dictionaries: (state1: (state2: prob, ...), ...)
	chain.currentState = nil;
	chain.seed = seed ? 42;

	// Set transition probability from one state to another
	chain.setTransition = { |self, fromState, toState, probability|
		// Convert to symbols for consistent key handling
		fromState = fromState.asSymbol;
		toState = toState.asSymbol;

		if(self.transitions[fromState].isNil, {
			self.transitions[fromState] = ();
		});
		self.transitions[fromState][toState] = probability;
	};

	// Get transition probability
	chain.getTransition = { |self, fromState, toState|
		if(self.transitions[fromState].isNil, {
			0.0
		}, {
			self.transitions[fromState][toState] ? 0.0
		});
	};

	// Normalize transitions for a given state so they sum to 1.0
	chain.normalizeTransitions = { |self, fromState|
		var trans, sum;
		trans = self.transitions[fromState];
		if(trans.notNil, {
			sum = trans.values.sum;
			if(sum > 0, {
				trans.keysDo { |toState|
					trans[toState] = trans[toState] / sum;
				};
			});
		});
	};

	// Generate next state based on current state and transition probabilities
	chain.nextState = { |self|
		var trans, states, probs, cumulative, r, nextState, currentStateSym;

		// If no current state, pick a random starting state
		if(self.currentState.isNil, {
			if(self.transitions.size == 0, {
				"[Markov] Error: no states in chain".postln;
				nil
			}, {
				// Pick random starting state
				self.currentState = self.transitions.keys.choose;
				self.currentState
			});
		}, {
			// Get transitions from current state (ensure it's a symbol)
			currentStateSym = self.currentState.asSymbol;
			trans = self.transitions[currentStateSym];

			if(trans.isNil or: { trans.size == 0 }, {
				// No transitions, stay on current state
				"[Markov] No transitions from state %".format(self.currentState).postln;
				self.currentState
			}, {
				// Select next state using weighted random choice
				states = trans.keys.asArray;
				probs = trans.values.asArray;

				// Use wchoose with normalized probabilities
				nextState = states.wchoose(probs.normalizeSum);
				self.currentState = nextState;

				nextState
			});
		});
	};

	// Reset the chain to no current state
	chain.resetState = { |self|
		self.currentState = nil;
	};

	// Set current state manually
	chain.setState = { |self, stateName|
		// Convert to symbol for consistent key handling
		stateName = stateName.asSymbol;

		if(self.transitions[stateName].notNil, {
			self.currentState = stateName;
		}, {
			"[Markov] Warning: state % does not exist in chain".format(stateName).postln;
		});
	};

	// Get current state
	chain.getState = { |self|
		self.currentState
	};

	chain
};

"[Markov] Markov chain library loaded".postln;
)

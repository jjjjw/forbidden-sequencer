// OSC client for sending messages to SuperCollider via HTTP bridge

const BRIDGE_URL = 'http://localhost:8080/osc';

/**
 * Send an OSC message to SuperCollider via the HTTP bridge
 * @param {string} address - OSC address (e.g., "/pattern/curve_time/play")
 * @param {...any} args - Arguments to send with the message
 */
export async function sendOSC(address, ...args) {
	try {
		const response = await fetch(BRIDGE_URL, {
			method: 'POST',
			headers: {
				'Content-Type': 'application/json',
			},
			body: JSON.stringify({
				address,
				args
			})
		});

		if (!response.ok) {
			console.error(`OSC send failed: ${response.status} ${response.statusText}`);
		}
	} catch (error) {
		console.error('OSC send error:', error);
	}
}

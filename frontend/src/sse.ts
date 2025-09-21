// sse.ts - Server-Sent Events client with robust reconnection
export interface SSEMessage {
	type: string;
	data?: unknown;
}

export type SSEEventHandler = (message: SSEMessage) => void;

export class SSEClient {
	private eventSource: EventSource | null = null;
	private url: string;
	private handlers: Map<string, SSEEventHandler[]> = new Map();
	private reconnectTimeout: number | null = null;
	private reconnectDelay = 1000; // Start with 1 second
	private maxReconnectDelay = 30000; // Max 30 seconds
	private isConnected = false;
	private shouldReconnect = true;

	constructor(url: string) {
		this.url = url;
	}

	connect(): void {
		console.log(' SSE: connect() called, URL:', this.url);
		if (this.eventSource) {
			console.log(' SSE: Existing connection found, disconnecting first');
			this.disconnect();
		}

		console.log(' SSE: Creating new EventSource connection');
		this.eventSource = new EventSource(this.url);

		this.eventSource.onopen = () => {
			console.log(' SSE: Connection opened successfully');
			this.isConnected = true;
			this.reconnectDelay = 1000; // Reset delay on successful connection
			console.log(' SSE: isConnected set to true, reconnectDelay reset to 1000');
		};

		this.eventSource.onmessage = (event) => {
			console.log(' SSE: Raw message received', { data: event.data, type: event.type });
			try {
				const message: SSEMessage = JSON.parse(event.data);
				console.log(' SSE: Message parsed successfully', message);
				this.handleMessage(message);
			} catch (error) {
				console.error(' SSE: Failed to parse message:', error, event.data);
			}
		};

		this.eventSource.onerror = (error) => {
			console.error(' SSE: Connection error:', error);
			this.isConnected = false;
			console.log(' SSE: isConnected set to false');
			
			if (this.shouldReconnect) {
				console.log(' SSE: Should reconnect is true, scheduling reconnect');
				this.scheduleReconnect();
			} else {
				console.log(' SSE: Should reconnect is false, not reconnecting');
			}
		};

		// Handle specific event types
		this.eventSource.addEventListener('connected', (event) => {
			console.log(' SSE: Connected event received', event);
			try {
				const message: SSEMessage = {
					type: 'connected',
					data: JSON.parse(event.data).data
				};
				console.log(' SSE: Connected message parsed', message);
				this.handleMessage(message);
			} catch (error) {
				console.error(' SSE: Failed to parse connected event:', error);
			}
		});

		this.eventSource.addEventListener('ping', () => {
			console.log(' SSE: Ping received');
			// Handle keepalive ping - no action needed
		});

		// Add listeners for invalidation events
		const invalidationTypes = [
			'invalidate_grades',
			'invalidate_periods', 
			'invalidate_categories',
			'invalidate_courses',
			'invalidate_requirements',
			'invalidate_students',
			'invalidate_invitations',
			'invalidate_selection_controls',
			'invalidate_selections'
		];

		// Add listeners for new fine-grained SSE message types
		const fineGrainedTypes = [
			'course_enrollment_update',
			'selection_action',
			'invalidate_student_selections_by_period',
			'grade_selection_status_change'
		];

		const allTypes = [...invalidationTypes, ...fineGrainedTypes];

		console.log(' SSE: Setting up SSE event listeners', { types: allTypes });

		allTypes.forEach(type => {
			this.eventSource?.addEventListener(type, (event) => {
				console.log(` SSE: ${type} event received`, event);
				try {
					const message: SSEMessage = {
						type,
						data: JSON.parse(event.data).data
					};
					console.log(` SSE: ${type} message parsed`, message);
					this.handleMessage(message);
				} catch (error) {
					console.error(` SSE: Failed to parse ${type} event:`, error);
				}
			});
		});
		
		console.log(' SSE: All event listeners set up');
	}

	disconnect(): void {
		console.log(' SSE: disconnect() called');
		this.shouldReconnect = false;
		
		if (this.reconnectTimeout) {
			console.log(' SSE: Clearing reconnect timeout');
			clearTimeout(this.reconnectTimeout);
			this.reconnectTimeout = null;
		}

		if (this.eventSource) {
			console.log(' SSE: Closing EventSource');
			this.eventSource.close();
			this.eventSource = null;
		}

		this.isConnected = false;
		console.log(' SSE: Disconnected, isConnected set to false');
	}

	private scheduleReconnect(): void {
		console.log(' SSE: scheduleReconnect() called');
		if (this.reconnectTimeout) {
			console.log(' SSE: Clearing existing reconnect timeout');
			clearTimeout(this.reconnectTimeout);
		}

		console.log(` SSE: Scheduling reconnect in ${this.reconnectDelay}ms`);
		this.reconnectTimeout = window.setTimeout(() => {
			console.log(' SSE: Reconnect timeout triggered');
			if (this.shouldReconnect) {
				console.log(' SSE: Attempting reconnect');
				this.connect();
				// Exponential backoff with jitter
				this.reconnectDelay = Math.min(
					this.maxReconnectDelay,
					this.reconnectDelay * 2 + Math.random() * 1000
				);
				console.log(' SSE: Reconnect delay updated to', this.reconnectDelay);
			} else {
				console.log(' SSE: Should not reconnect, skipping');
			}
		}, this.reconnectDelay);
	}

	private handleMessage(message: SSEMessage): void {
		console.log(' SSE: handleMessage() called', { type: message.type, hasData: !!message.data });
		const handlers = this.handlers.get(message.type);
		if (handlers) {
			console.log(' SSE: Found handlers for message type', { type: message.type, handlerCount: handlers.length });
			handlers.forEach((handler, index) => {
				try {
					console.log(' SSE: Calling handler', { type: message.type, handlerIndex: index });
					handler(message);
					console.log(' SSE: Handler completed successfully', { type: message.type, handlerIndex: index });
				} catch (error) {
					console.error(` SSE: Handler error for ${message.type}:`, error);
				}
			});
		} else {
			console.log(' SSE: No handlers found for message type', { type: message.type });
		}
	}

	on(eventType: string, handler: SSEEventHandler): void {
		if (!this.handlers.has(eventType)) {
			this.handlers.set(eventType, []);
		}
		this.handlers.get(eventType)!.push(handler);
	}

	off(eventType: string, handler: SSEEventHandler): void {
		const handlers = this.handlers.get(eventType);
		if (handlers) {
			const index = handlers.indexOf(handler);
			if (index > -1) {
				handlers.splice(index, 1);
			}
		}
	}

	getConnectionStatus(): boolean {
		return this.isConnected;
	}
}

// Global SSE client instance
let sseClient: SSEClient | null = null;

export function getSSEClient(): SSEClient | null {
	console.log(' SSE: getSSEClient() called', { hasClient: !!sseClient });
	return sseClient;
}

export function initializeSSE(): void {
	console.log(' SSE: initializeSSE() called');
	if (sseClient) {
		console.log(' SSE: Existing client found, disconnecting');
		sseClient.disconnect();
	}

	console.log(' SSE: Creating new SSE client');
	sseClient = new SSEClient('/api/events');
	sseClient.connect();
	console.log(' SSE: SSE client created and connection initiated');
}

export function disconnectSSE(): void {
	console.log(' SSE: disconnectSSE() called');
	if (sseClient) {
		console.log(' SSE: Disconnecting existing client');
		sseClient.disconnect();
		sseClient = null;
		console.log(' SSE: Client disconnected and set to null');
	} else {
		console.log(' SSE: No client to disconnect');
	}
}
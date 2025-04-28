type WebSocketCallback = (event: any) => void;

class WebSocketService {
    private ws: WebSocket | null = null;
    private subscribers: Map<string, Set<Function>> = new Map();
    private reconnectAttempts = 0;
    private maxReconnectAttempts = 5;

    // Establishes WebSocket connection and handles reconnection logic
    connect() {
        if (this.ws?.readyState === WebSocket.OPEN) {
            return;
        }
        
        this.ws = new WebSocket(process.env.NEXT_PUBLIC_WS_URL || 'ws://localhost:3000/ws');
        
        this.ws.onopen = () => {
            console.log('WebSocket connected');
            this.reconnectAttempts = 0;
        };

        this.ws.onerror = (error) => {
            console.error('WebSocket error:', error);
        };

        this.ws.onmessage = (event) => {
            try {
                const data = JSON.parse(event.data);
                console.log('WebSocket message received:', data);
                const subscribers = this.subscribers.get(data.type);
                subscribers?.forEach(callback => callback(data.payload));
            } catch (error) {
                console.error('Error processing WebSocket message:', error);
            }
        };

        this.ws.onclose = () => {
            if (this.reconnectAttempts < this.maxReconnectAttempts) {
                this.reconnectAttempts++;
                setTimeout(() => this.connect(), 1000 * this.reconnectAttempts);
            }
        };
    }

    // Closes the WebSocket connection and resets reconnection attempts
    disconnect() {
        if (this.ws) {
            this.ws.close();
            this.ws = null;
            this.reconnectAttempts = 0;
        }
    }

    // Adds a callback function to the subscribers for a specific event type
    subscribe(eventType: string, callback: WebSocketCallback) {
        if (!this.subscribers.has(eventType)) {
            this.subscribers.set(eventType, new Set());
        }
        this.subscribers.get(eventType)?.add(callback);
    }

    // Removes a callback function from the subscribers for a specific event type
    unsubscribe(eventType: string, callback: WebSocketCallback) {
        this.subscribers.get(eventType)?.delete(callback);
    }
}

export const wsService = new WebSocketService();
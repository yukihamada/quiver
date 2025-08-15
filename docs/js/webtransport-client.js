// WebTransport client for direct P2P connection from browser
class QUIVerWebTransport {
    constructor() {
        this.transport = null;
        this.streams = new Map();
        this.messageId = 0;
        this.callbacks = new Map();
    }

    // Connect to QUIVer gateway via WebTransport
    async connect(url = 'https://localhost:8443/webtransport') {
        try {
            // Check if WebTransport is available
            if (!window.WebTransport) {
                throw new Error('WebTransport is not supported in this browser');
            }

            // Create WebTransport connection
            this.transport = new WebTransport(url);
            
            // Wait for connection to be ready
            await this.transport.ready;
            console.log('Connected to QUIVer via WebTransport');

            // Start handling incoming streams
            this.handleIncomingStreams();

            return true;
        } catch (error) {
            console.error('Failed to connect via WebTransport:', error);
            
            // Fallback to regular HTTP
            if (error.message.includes('not supported')) {
                console.log('Falling back to HTTP endpoints');
                return false;
            }
            
            throw error;
        }
    }

    // Handle incoming bidirectional streams
    async handleIncomingStreams() {
        const reader = this.transport.incomingBidirectionalStreams.getReader();
        
        try {
            while (true) {
                const { done, value } = await reader.read();
                if (done) break;
                
                const stream = value;
                this.handleStream(stream);
            }
        } catch (error) {
            console.error('Error reading incoming streams:', error);
        }
    }

    // Handle a single stream
    async handleStream(stream) {
        const reader = stream.readable.getReader();
        const decoder = new TextDecoder();
        
        try {
            while (true) {
                const { done, value } = await reader.read();
                if (done) break;
                
                const text = decoder.decode(value);
                const message = JSON.parse(text);
                
                // Handle response callbacks
                if (message.id && this.callbacks.has(message.id)) {
                    const callback = this.callbacks.get(message.id);
                    this.callbacks.delete(message.id);
                    callback(message);
                }
            }
        } catch (error) {
            console.error('Error reading stream:', error);
        }
    }

    // Send a generate request
    async generate(prompt, model = 'llama3.2:3b') {
        if (!this.transport || this.transport.closed) {
            throw new Error('Not connected to WebTransport');
        }

        const messageId = `msg-${++this.messageId}`;
        
        // Create a new stream
        const stream = await this.transport.createBidirectionalStream();
        const writer = stream.writable.getWriter();
        const encoder = new TextEncoder();

        // Create promise for response
        const responsePromise = new Promise((resolve, reject) => {
            this.callbacks.set(messageId, (message) => {
                if (message.type === 'error') {
                    reject(new Error(message.payload.error));
                } else {
                    resolve(message.payload);
                }
            });
        });

        // Send request
        const request = {
            type: 'generate',
            id: messageId,
            payload: {
                prompt: prompt,
                model: model
            }
        };

        await writer.write(encoder.encode(JSON.stringify(request)));
        
        // Handle the stream for responses
        this.handleStream(stream);

        return responsePromise;
    }

    // Generate with streaming support
    async generateStream(prompt, model = 'llama3.2:3b', onChunk) {
        if (!this.transport || this.transport.closed) {
            throw new Error('Not connected to WebTransport');
        }

        const messageId = `msg-${++this.messageId}`;
        
        // Create a new stream
        const stream = await this.transport.createBidirectionalStream();
        const writer = stream.writable.getWriter();
        const reader = stream.readable.getReader();
        const encoder = new TextEncoder();
        const decoder = new TextDecoder();

        // Send request
        const request = {
            type: 'generate_stream',
            id: messageId,
            payload: {
                prompt: prompt,
                model: model
            }
        };

        await writer.write(encoder.encode(JSON.stringify(request)));

        // Read streaming response
        let response = '';
        try {
            while (true) {
                const { done, value } = await reader.read();
                if (done) break;
                
                const chunk = decoder.decode(value, { stream: true });
                const messages = chunk.split('\n').filter(m => m.trim());
                
                for (const msg of messages) {
                    try {
                        const parsed = JSON.parse(msg);
                        if (parsed.type === 'chunk') {
                            response += parsed.content;
                            if (onChunk) onChunk(parsed.content);
                        } else if (parsed.type === 'complete') {
                            return {
                                completion: response,
                                receipt: parsed.receipt,
                                metrics: parsed.metrics
                            };
                        } else if (parsed.type === 'error') {
                            throw new Error(parsed.error);
                        }
                    } catch (e) {
                        // Ignore parse errors for incomplete messages
                    }
                }
            }
        } catch (error) {
            console.error('Stream error:', error);
            throw error;
        }
    }

    // Close the connection
    async close() {
        if (this.transport) {
            await this.transport.close();
            this.transport = null;
        }
    }

    // Check if connected
    isConnected() {
        return this.transport && !this.transport.closed;
    }

    // Get connection stats
    async getStats() {
        if (!this.transport) return null;
        
        const stats = await this.transport.getStats();
        return {
            bytesReceived: stats.bytesReceived,
            bytesSent: stats.bytesSent,
            packetsReceived: stats.packetsReceived,
            packetsSent: stats.packetsSent,
            rtt: stats.rtt
        };
    }
}

// Export for use in playground
window.QUIVerWebTransport = QUIVerWebTransport;
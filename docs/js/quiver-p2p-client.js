// QUIVer P2P Client for browsers
class QUIVerP2PClient {
    constructor() {
        this.peerConnection = null;
        this.dataChannel = null;
        this.connected = false;
        this.messageId = 0;
        this.callbacks = new Map();
        this.signallingUrl = 'wss://signal.quiver.network';
        this.ws = null;
    }

    // Connect to P2P network via WebRTC
    async connect() {
        try {
            // First, try to get bootstrap nodes
            const bootstrapResponse = await fetch('https://yukihamada.github.io/quiver/bootstrap/public-bootstrap.json');
            const bootstrapData = await bootstrapResponse.json();
            
            // Connect to signalling server
            await this.connectSignalling();
            
            // Create peer connection
            await this.createPeerConnection();
            
            return true;
        } catch (error) {
            console.error('Failed to connect to P2P network:', error);
            return false;
        }
    }

    // Connect to signalling server for WebRTC negotiation
    async connectSignalling() {
        return new Promise((resolve, reject) => {
            // Try multiple signalling servers
            const signallingUrls = [
                'wss://signal.localhost/signal',
                'wss://signal.quiver.network/signal',
                'wss://signal-asia.quiver.network/signal',
                'wss://34.146.216.182:8444/signal',  // GCP signalling server
                'wss://localhost:8444/signal'
            ];

            let connected = false;
            
            const tryConnect = async (url) => {
                try {
                    this.ws = new WebSocket(url);
                    
                    this.ws.onopen = () => {
                        console.log('Connected to signalling server:', url);
                        connected = true;
                        
                        // Send browser identification
                        this.ws.send(JSON.stringify({
                            type: 'browser_join',
                            userAgent: navigator.userAgent,
                            timestamp: Date.now()
                        }));
                        
                        resolve();
                    };
                    
                    this.ws.onmessage = (event) => {
                        this.handleSignallingMessage(JSON.parse(event.data));
                    };
                    
                    this.ws.onerror = (error) => {
                        console.error('WebSocket error:', error);
                    };
                    
                    this.ws.onclose = () => {
                        if (!connected) {
                            // Try next server
                            const nextIndex = signallingUrls.indexOf(url) + 1;
                            if (nextIndex < signallingUrls.length) {
                                tryConnect(signallingUrls[nextIndex]);
                            } else {
                                reject(new Error('Could not connect to any signalling server'));
                            }
                        }
                    };
                } catch (error) {
                    console.error('Failed to connect to', url, error);
                }
            };
            
            tryConnect(signallingUrls[0]);
        });
    }

    // Create WebRTC peer connection
    async createPeerConnection() {
        const config = {
            iceServers: [
                { urls: 'stun:stun.l.google.com:19302' },
                { urls: 'stun:stun1.l.google.com:19302' },
                {
                    urls: 'turn:turn.quiver.network:3478',
                    username: 'quiver',
                    credential: 'quiver-turn-2025'
                }
            ]
        };

        this.peerConnection = new RTCPeerConnection(config);

        // Create data channel
        this.dataChannel = this.peerConnection.createDataChannel('quiver', {
            ordered: true
        });

        this.dataChannel.onopen = () => {
            console.log('P2P data channel opened');
            this.connected = true;
            this.onConnected();
        };

        this.dataChannel.onmessage = (event) => {
            this.handleDataChannelMessage(event);
        };

        this.dataChannel.onclose = () => {
            console.log('P2P data channel closed');
            this.connected = false;
        };

        // Handle ICE candidates
        this.peerConnection.onicecandidate = (event) => {
            if (event.candidate) {
                this.sendSignalling({
                    type: 'ice_candidate',
                    candidate: event.candidate
                });
            }
        };

        // Create offer
        const offer = await this.peerConnection.createOffer();
        await this.peerConnection.setLocalDescription(offer);

        // Send offer to signalling server
        this.sendSignalling({
            type: 'offer',
            offer: offer
        });
    }

    // Handle signalling messages
    handleSignallingMessage(message) {
        switch (message.type) {
            case 'answer':
                this.peerConnection.setRemoteDescription(new RTCSessionDescription(message.answer));
                break;
                
            case 'ice_candidate':
                this.peerConnection.addIceCandidate(new RTCIceCandidate(message.candidate));
                break;
                
            case 'provider_available':
                console.log('Provider available:', message.providerId);
                break;
        }
    }

    // Handle data channel messages
    handleDataChannelMessage(event) {
        try {
            const message = JSON.parse(event.data);
            
            if (message.id && this.callbacks.has(message.id)) {
                const callback = this.callbacks.get(message.id);
                this.callbacks.delete(message.id);
                callback(message);
            }
            
            // Handle streaming messages
            if (message.type === 'stream_chunk' && message.streamId) {
                this.handleStreamChunk(message);
            }
        } catch (error) {
            console.error('Failed to parse message:', error);
        }
    }

    // Send message via signalling server
    sendSignalling(message) {
        if (this.ws && this.ws.readyState === WebSocket.OPEN) {
            this.ws.send(JSON.stringify(message));
        }
    }

    // Send message via data channel
    sendDataChannel(message) {
        if (this.dataChannel && this.dataChannel.readyState === 'open') {
            this.dataChannel.send(JSON.stringify(message));
            return true;
        }
        return false;
    }

    // Generate AI response
    async generate(prompt, model = 'llama3.2:3b') {
        if (!this.connected) {
            throw new Error('Not connected to P2P network');
        }

        const messageId = `msg-${++this.messageId}`;
        
        return new Promise((resolve, reject) => {
            this.callbacks.set(messageId, (response) => {
                if (response.type === 'error') {
                    reject(new Error(response.error));
                } else {
                    resolve(response.payload);
                }
            });

            const request = {
                type: 'generate',
                id: messageId,
                payload: {
                    prompt: prompt,
                    model: model
                }
            };

            if (!this.sendDataChannel(request)) {
                this.callbacks.delete(messageId);
                reject(new Error('Failed to send request'));
            }
        });
    }

    // Generate with streaming
    async generateStream(prompt, model = 'llama3.2:3b', onChunk) {
        if (!this.connected) {
            throw new Error('Not connected to P2P network');
        }

        const streamId = `stream-${++this.messageId}`;
        let response = '';
        
        return new Promise((resolve, reject) => {
            // Set up stream handler
            this.streamHandlers = this.streamHandlers || {};
            this.streamHandlers[streamId] = {
                onChunk: (chunk) => {
                    response += chunk;
                    if (onChunk) onChunk(chunk);
                },
                onComplete: (data) => {
                    delete this.streamHandlers[streamId];
                    resolve({
                        completion: response,
                        receipt: data.receipt,
                        metrics: data.metrics
                    });
                },
                onError: (error) => {
                    delete this.streamHandlers[streamId];
                    reject(new Error(error));
                }
            };

            const request = {
                type: 'generate_stream',
                id: streamId,
                payload: {
                    prompt: prompt,
                    model: model,
                    stream: true
                }
            };

            if (!this.sendDataChannel(request)) {
                delete this.streamHandlers[streamId];
                reject(new Error('Failed to send request'));
            }
        });
    }

    // Handle streaming chunks
    handleStreamChunk(message) {
        const handler = this.streamHandlers && this.streamHandlers[message.streamId];
        if (!handler) return;

        switch (message.chunkType) {
            case 'content':
                handler.onChunk(message.content);
                break;
            case 'complete':
                handler.onComplete(message);
                break;
            case 'error':
                handler.onError(message.error);
                break;
        }
    }

    // Called when connected
    onConnected() {
        // Override this method to handle connection events
        console.log('Connected to QUIVer P2P network');
    }

    // Close connection
    close() {
        if (this.dataChannel) {
            this.dataChannel.close();
        }
        if (this.peerConnection) {
            this.peerConnection.close();
        }
        if (this.ws) {
            this.ws.close();
        }
        this.connected = false;
    }

    // Check if connected
    isConnected() {
        return this.connected;
    }
}

// Export for use
window.QUIVerP2PClient = QUIVerP2PClient;
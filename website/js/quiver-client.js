/**
 * QUIVer Browser Client
 * Connects to the P2P network via WebTransport and tracks network statistics
 */

class QUIVerClient {
    constructor(gatewayUrl = 'https://gateway.quiver.network:4433') {
        this.gatewayUrl = gatewayUrl;
        this.transport = null;
        this.connected = false;
        this.nodeId = null;
        this.hllData = null;
        this.lastUpdate = null;
        this.reconnectAttempts = 0;
        this.maxReconnectAttempts = 5;
        this.reconnectDelay = 1000;
        
        // Callbacks
        this.onStatsUpdate = null;
        this.onConnectionChange = null;
        
        // Initialize
        this.init();
    }
    
    async init() {
        // Check WebTransport support
        if (!window.WebTransport) {
            console.warn('WebTransport not supported, falling back to WebSocket');
            return this.initWebSocket();
        }
        
        await this.connect();
    }
    
    async connect() {
        try {
            console.log('Connecting to QUIVer network via WebTransport...');
            
            // Create WebTransport connection
            const url = `${this.gatewayUrl}/webtransport`;
            this.transport = new WebTransport(url);
            
            // Wait for connection
            await this.transport.ready;
            
            this.connected = true;
            this.reconnectAttempts = 0;
            console.log('Connected to QUIVer network!');
            
            if (this.onConnectionChange) {
                this.onConnectionChange(true);
            }
            
            // Start periodic HLL requests
            this.startHLLPolling();
            
            // Handle connection close
            this.transport.closed.then(() => {
                console.log('Disconnected from QUIVer network');
                this.handleDisconnect();
            });
            
        } catch (error) {
            console.error('Failed to connect:', error);
            this.handleDisconnect();
        }
    }
    
    async handleDisconnect() {
        this.connected = false;
        this.transport = null;
        
        if (this.onConnectionChange) {
            this.onConnectionChange(false);
        }
        
        // Attempt reconnection
        if (this.reconnectAttempts < this.maxReconnectAttempts) {
            this.reconnectAttempts++;
            const delay = this.reconnectDelay * Math.pow(2, this.reconnectAttempts - 1);
            console.log(`Reconnecting in ${delay}ms... (attempt ${this.reconnectAttempts})`);
            setTimeout(() => this.connect(), delay);
        }
    }
    
    async sendMessage(type, payload = {}) {
        if (!this.connected || !this.transport) {
            throw new Error('Not connected to network');
        }
        
        try {
            // Create bidirectional stream
            const stream = await this.transport.createBidirectionalStream();
            const writer = stream.writable.getWriter();
            const reader = stream.readable.getReader();
            
            // Send message
            const message = {
                type: type,
                payload: payload
            };
            
            const encoder = new TextEncoder();
            await writer.write(encoder.encode(JSON.stringify(message)));
            await writer.close();
            
            // Read response
            const decoder = new TextDecoder();
            let response = '';
            
            while (true) {
                const { done, value } = await reader.read();
                if (done) break;
                response += decoder.decode(value);
            }
            
            return JSON.parse(response);
            
        } catch (error) {
            console.error('Failed to send message:', error);
            throw error;
        }
    }
    
    async requestHLLData() {
        try {
            const requestId = `req-${Date.now()}`;
            const response = await this.sendMessage('hll_request', {
                request_id: requestId
            });
            
            if (response.hll_data) {
                this.hllData = response.hll_data;
                this.lastUpdate = new Date(response.timestamp * 1000);
                
                // Notify listeners
                if (this.onStatsUpdate) {
                    this.onStatsUpdate({
                        nodeCount: response.node_count,
                        timestamp: this.lastUpdate,
                        raw: response
                    });
                }
            }
            
            return response;
            
        } catch (error) {
            console.error('Failed to request HLL data:', error);
            return null;
        }
    }
    
    async getNetworkStats() {
        try {
            const response = await this.sendMessage('get_stats');
            return response;
        } catch (error) {
            console.error('Failed to get network stats:', error);
            return null;
        }
    }
    
    startHLLPolling(intervalMs = 30000) {
        // Initial request
        this.requestHLLData();
        
        // Periodic updates
        this.pollInterval = setInterval(() => {
            if (this.connected) {
                this.requestHLLData();
            }
        }, intervalMs);
    }
    
    stopHLLPolling() {
        if (this.pollInterval) {
            clearInterval(this.pollInterval);
            this.pollInterval = null;
        }
    }
    
    // Fallback WebSocket implementation
    async initWebSocket() {
        console.log('Using WebSocket fallback...');
        
        // Connect to WebSocket gateway
        const wsUrl = this.gatewayUrl.replace('https:', 'wss:').replace(':4433', ':8084') + '/ws';
        
        try {
            this.ws = new WebSocket(wsUrl);
            
            this.ws.onopen = () => {
                this.connected = true;
                console.log('Connected via WebSocket');
                if (this.onConnectionChange) {
                    this.onConnectionChange(true);
                }
                this.startHLLPolling();
            };
            
            this.ws.onclose = () => {
                this.handleDisconnect();
            };
            
            this.ws.onmessage = (event) => {
                try {
                    const data = JSON.parse(event.data);
                    this.handleWebSocketMessage(data);
                } catch (error) {
                    console.error('Failed to parse WebSocket message:', error);
                }
            };
            
        } catch (error) {
            console.error('WebSocket connection failed:', error);
            this.handleDisconnect();
        }
    }
    
    handleWebSocketMessage(data) {
        if (data.type === 'hll_response' && this.onStatsUpdate) {
            this.onStatsUpdate({
                nodeCount: data.node_count,
                timestamp: new Date(data.timestamp * 1000),
                raw: data
            });
        }
    }
    
    // Public API
    getNodeCount() {
        return this.lastNodeCount || 0;
    }
    
    isConnected() {
        return this.connected;
    }
    
    destroy() {
        this.stopHLLPolling();
        
        if (this.transport) {
            this.transport.close();
        }
        
        if (this.ws) {
            this.ws.close();
        }
    }
}

// Helper function to create and manage a global client instance
let globalClient = null;

function getQUIVerClient(gatewayUrl) {
    if (!globalClient) {
        // Try multiple gateways for redundancy
        const gateways = [
            'https://gateway1.quiver.network:4433',
            'https://gateway2.quiver.network:4433',
            'https://gateway3.quiver.network:4433',
            gatewayUrl || 'https://localhost:4433'
        ];
        
        // Try each gateway until one works
        globalClient = new QUIVerClient(gateways[0]);
        
        // Fallback logic
        let gatewayIndex = 0;
        globalClient.onConnectionChange = (connected) => {
            if (!connected && gatewayIndex < gateways.length - 1) {
                gatewayIndex++;
                console.log(`Trying gateway ${gatewayIndex + 1}: ${gateways[gatewayIndex]}`);
                globalClient.gatewayUrl = gateways[gatewayIndex];
            }
        };
    }
    
    return globalClient;
}

// Export for use in other scripts
if (typeof module !== 'undefined' && module.exports) {
    module.exports = { QUIVerClient, getQUIVerClient };
}
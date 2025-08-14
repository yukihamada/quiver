/**
 * QUIVer Network Node Counter
 * Simulates real-time node counting with HyperLogLog estimation
 */

class NetworkCounter {
    constructor() {
        // WebSocket endpoints for real-time updates
        this.wsEndpoints = [
            'ws://localhost:8087/ws', // Local stats server (priority)
            'ws://34.146.63.195:8087/ws', // GCP stats node
            'ws://34.146.63.195/ws', // GCP stats node (nginx)
            'wss://stats.quiver.network/ws'
        ];
        
        // HTTP API endpoints (fallback)
        this.endpoints = [
            'http://localhost:8087/api/stats', // Local stats server (priority)
            'http://34.146.63.195:8087/api/stats', // GCP stats node
            'http://34.146.63.195/api/stats', // GCP stats node (nginx)
            'https://stats.quiver.network/api/stats',
            // GitHub Pages hosted JSON (immediate availability)
            'https://yukihamada.github.io/quiver/api/stats.json',
            '/api/stats.json' // Relative path for local testing
        ];
        this.currentEndpointIndex = 0;
        this.ws = null;
        this.wsConnected = false;
        
        // Display state
        this.currentDisplay = 0;
        this.targetValue = 0;
        this.animationSpeed = 50; // ms per update
        
        // Connection state
        this.connected = false;
        this.lastSuccessfulFetch = null;
        this.fetchInterval = 30000; // 30 seconds
        
        // Callbacks
        this.onUpdate = null;
        this.onConnectionChange = null;
        
        // Fallback simulation parameters
        this.baseNodes = 3; // Starting nodes (realistic for early stage)
        this.growthRate = 0.005; // 0.5% growth per hour
        this.volatility = 0.1; // 10% random variation
        this.simulationStartTime = Date.now();
    }
    
    start() {
        // Start animation loop
        this.animate();
        
        // Try WebSocket connection first
        this.connectWebSocket();
        
        // Fallback to HTTP polling if WebSocket fails
        setTimeout(() => {
            if (!this.wsConnected) {
                this.fetchRealData();
                setInterval(() => this.fetchRealData(), this.fetchInterval);
            }
        }, 3000);
    }
    
    async connectWebSocket() {
        for (const endpoint of this.wsEndpoints) {
            try {
                console.log(`Trying WebSocket connection to ${endpoint}`);
                this.ws = new WebSocket(endpoint);
                
                this.ws.onopen = () => {
                    console.log(`WebSocket connected to ${endpoint}`);
                    this.wsConnected = true;
                    this.connected = true;
                    if (this.onConnectionChange) {
                        this.onConnectionChange(true);
                    }
                };
                
                this.ws.onmessage = (event) => {
                    try {
                        const data = JSON.parse(event.data);
                        if (data.node_count !== undefined) {
                            this.targetValue = data.node_count;
                            this.lastStats = data;
                            this.lastSuccessfulFetch = Date.now();
                            
                            console.log('Received real-time data:', data);
                            
                            // Update display with additional info
                            if (this.onStatsUpdate) {
                                this.onStatsUpdate(data);
                            }
                        }
                    } catch (error) {
                        console.error('Failed to parse WebSocket message:', error);
                    }
                };
                
                this.ws.onerror = (error) => {
                    console.error(`WebSocket error on ${endpoint}:`, error);
                };
                
                this.ws.onclose = () => {
                    console.log('WebSocket disconnected');
                    this.wsConnected = false;
                    this.ws = null;
                    
                    // Try to reconnect after 5 seconds
                    setTimeout(() => this.connectWebSocket(), 5000);
                };
                
                // If connection is successful, stop trying other endpoints
                await new Promise(resolve => {
                    setTimeout(resolve, 1000); // Wait 1 second to see if connection succeeds
                });
                
                if (this.wsConnected) {
                    break;
                }
            } catch (error) {
                console.error(`Failed to connect to WebSocket ${endpoint}:`, error);
            }
        }
    }
    
    scheduleNextUpdate() {
        const interval = 5000 + Math.random() * 25000; // 5-30 seconds
        setTimeout(() => {
            this.updateTarget();
            this.scheduleNextUpdate();
        }, interval);
    }
    
    updateTarget() {
        // Calculate time-based growth
        const hoursElapsed = (Date.now() - this.startTime) / (1000 * 60 * 60);
        const growthFactor = Math.pow(1 + this.growthRate, hoursElapsed);
        
        // Add some randomness for realism
        const randomFactor = 1 + (Math.random() - 0.5) * this.volatility;
        
        // Calculate new target
        const baseTarget = this.baseNodes * growthFactor * randomFactor;
        
        // Add some nodes joining/leaving
        const variation = Math.floor(Math.random() * 10) - 5;
        
        this.targetValue = Math.max(1, Math.floor(baseTarget + variation));
    }
    
    animate() {
        // Smooth animation towards target
        const diff = this.targetValue - this.currentDisplay;
        const step = diff > 0 ? 1 : -1;
        
        if (Math.abs(diff) > 0) {
            this.currentDisplay += step;
            
            if (this.onUpdate) {
                this.onUpdate({
                    nodeCount: this.currentDisplay,
                    connected: this.connected,
                    timestamp: new Date()
                });
            }
        }
        
        // Continue animation
        setTimeout(() => this.animate(), this.animationSpeed);
    }
    
    async fetchRealData() {
        let dataFetched = false;
        
        // Try each endpoint
        for (let i = 0; i < this.endpoints.length; i++) {
            const endpoint = this.endpoints[(this.currentEndpointIndex + i) % this.endpoints.length];
            
            try {
                const response = await fetch(endpoint, {
                    mode: 'cors',
                    cache: 'no-cache'
                });
                
                if (response.ok) {
                    const data = await response.json();
                    
                    if (data.node_count !== undefined) {
                        // Successfully got real data
                        this.targetValue = data.node_count;
                        this.connected = true;
                        this.lastSuccessfulFetch = Date.now();
                        this.currentEndpointIndex = (this.currentEndpointIndex + i) % this.endpoints.length;
                        dataFetched = true;
                        
                        console.log(`Fetched real data from ${endpoint}:`, data);
                        
                        if (this.onConnectionChange) {
                            this.onConnectionChange(true);
                        }
                        
                        // Store additional stats if available
                        this.lastStats = data;
                        
                        break;
                    }
                }
            } catch (error) {
                console.log(`Failed to fetch from ${endpoint}:`, error.message);
            }
        }
        
        // If all endpoints failed, use simulation
        if (!dataFetched) {
            console.log('All endpoints failed, using simulation mode');
            this.connected = false;
            this.updateTarget(); // Use simulation
            
            if (this.onConnectionChange) {
                this.onConnectionChange(false);
            }
        }
    }
    
    getNodeCount() {
        return this.currentDisplay;
    }
    
    isConnected() {
        return this.connected;
    }
    
    destroy() {
        // Clean up any connections or timers
        if (this.fetchInterval) {
            clearInterval(this.fetchInterval);
        }
    }
    
    // Get additional stats
    getStats() {
        return this.lastStats || {
            node_count: this.currentDisplay,
            connected: this.connected,
            timestamp: new Date()
        };
    }
}

// Create global instance
const networkCounter = new NetworkCounter();

// Auto-start if on the right page
if (typeof window !== 'undefined' && window.location.pathname.includes('network')) {
    window.addEventListener('load', () => {
        networkCounter.start();
    });
}
/**
 * QUIVer Network Node Counter
 * Simulates real-time node counting with HyperLogLog estimation
 */

class NetworkCounter {
    constructor() {
        // Simulated network growth parameters
        this.baseNodes = 42; // Starting nodes
        this.growthRate = 0.05; // 5% growth per hour
        this.volatility = 0.1; // 10% random variation
        this.startTime = Date.now();
        
        // Display state
        this.currentDisplay = this.baseNodes;
        this.targetValue = this.baseNodes;
        this.animationSpeed = 50; // ms per update
        
        // WebSocket connection (future implementation)
        this.wsEndpoint = 'wss://gateway.quiver.network/ws';
        this.connected = false;
        
        // Callbacks
        this.onUpdate = null;
        this.onConnectionChange = null;
    }
    
    start() {
        // Start simulation
        this.updateTarget();
        this.animate();
        
        // Update target every 5-30 seconds (random interval)
        this.scheduleNextUpdate();
        
        // Try to connect to real gateway
        this.tryConnect();
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
    
    async tryConnect() {
        // Try to connect to real WebSocket endpoint
        try {
            const ws = new WebSocket(this.wsEndpoint);
            
            ws.onopen = () => {
                console.log('Connected to real QUIVer gateway');
                this.connected = true;
                if (this.onConnectionChange) {
                    this.onConnectionChange(true);
                }
            };
            
            ws.onmessage = (event) => {
                try {
                    const data = JSON.parse(event.data);
                    if (data.node_count) {
                        // Use real data when available
                        this.targetValue = data.node_count;
                    }
                } catch (e) {
                    console.error('Failed to parse WebSocket message:', e);
                }
            };
            
            ws.onerror = (error) => {
                console.log('WebSocket error, using simulation mode');
                this.connected = false;
                if (this.onConnectionChange) {
                    this.onConnectionChange(false);
                }
            };
            
            ws.onclose = () => {
                this.connected = false;
                if (this.onConnectionChange) {
                    this.onConnectionChange(false);
                }
                
                // Retry connection after 30 seconds
                setTimeout(() => this.tryConnect(), 30000);
            };
            
            this.ws = ws;
            
        } catch (error) {
            console.log('Failed to connect to WebSocket, continuing in simulation mode');
            this.connected = false;
            
            // Retry later
            setTimeout(() => this.tryConnect(), 30000);
        }
    }
    
    getNodeCount() {
        return this.currentDisplay;
    }
    
    isConnected() {
        return this.connected;
    }
    
    destroy() {
        if (this.ws) {
            this.ws.close();
        }
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
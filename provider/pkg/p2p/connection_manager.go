package p2p

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/libp2p/go-libp2p/core/host"
	"github.com/libp2p/go-libp2p/core/network"
	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/multiformats/go-multiaddr"
	"github.com/sirupsen/logrus"
)

// ConnectionManager manages P2P connections with retry logic
type ConnectionManager struct {
	host           host.Host
	bootstrapPeers []multiaddr.Multiaddr
	connectedPeers map[peer.ID]time.Time
	mu             sync.RWMutex
	ctx            context.Context
	cancel         context.CancelFunc
	logger         *logrus.Logger
}

// NewConnectionManager creates a new connection manager
func NewConnectionManager(h host.Host, bootstrapPeers []string, logger *logrus.Logger) (*ConnectionManager, error) {
	ctx, cancel := context.WithCancel(context.Background())
	
	// Parse bootstrap addresses
	var addrs []multiaddr.Multiaddr
	for _, peerAddr := range bootstrapPeers {
		addr, err := multiaddr.NewMultiaddr(peerAddr)
		if err != nil {
			logger.Warnf("Invalid bootstrap address %s: %v", peerAddr, err)
			continue
		}
		addrs = append(addrs, addr)
	}
	
	cm := &ConnectionManager{
		host:           h,
		bootstrapPeers: addrs,
		connectedPeers: make(map[peer.ID]time.Time),
		ctx:            ctx,
		cancel:         cancel,
		logger:         logger,
	}
	
	// Set connection notifications
	h.Network().Notify(&network.NotifyBundle{
		ConnectedF: func(n network.Network, c network.Conn) {
			cm.onPeerConnected(c.RemotePeer())
		},
		DisconnectedF: func(n network.Network, c network.Conn) {
			cm.onPeerDisconnected(c.RemotePeer())
		},
	})
	
	return cm, nil
}

// Start begins the connection manager
func (cm *ConnectionManager) Start() {
	// Initial connection attempt
	cm.connectToBootstrapPeers()
	
	// Periodic reconnection
	go cm.maintainConnections()
}

// Stop stops the connection manager
func (cm *ConnectionManager) Stop() {
	cm.cancel()
}

// connectToBootstrapPeers attempts to connect to all bootstrap peers
func (cm *ConnectionManager) connectToBootstrapPeers() {
	cm.logger.Info("Connecting to bootstrap peers...")
	
	var wg sync.WaitGroup
	for _, addr := range cm.bootstrapPeers {
		wg.Add(1)
		go func(addr multiaddr.Multiaddr) {
			defer wg.Done()
			
			peerinfo, err := peer.AddrInfoFromP2pAddr(addr)
			if err != nil {
				cm.logger.Errorf("Failed to parse peer address %s: %v", addr, err)
				return
			}
			
			// Attempt connection with timeout
			ctx, cancel := context.WithTimeout(cm.ctx, 30*time.Second)
			defer cancel()
			
			if err := cm.host.Connect(ctx, *peerinfo); err != nil {
				cm.logger.Warnf("Failed to connect to bootstrap peer %s: %v", peerinfo.ID, err)
			} else {
				cm.logger.Infof("Successfully connected to bootstrap peer %s", peerinfo.ID)
			}
		}(addr)
	}
	
	wg.Wait()
}

// maintainConnections periodically checks and maintains connections
func (cm *ConnectionManager) maintainConnections() {
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()
	
	for {
		select {
		case <-ticker.C:
			cm.checkConnections()
		case <-cm.ctx.Done():
			return
		}
	}
}

// checkConnections checks current connections and reconnects if needed
func (cm *ConnectionManager) checkConnections() {
	cm.mu.RLock()
	connectedCount := len(cm.connectedPeers)
	cm.mu.RUnlock()
	
	cm.logger.Debugf("Connected peers: %d", connectedCount)
	
	// If we have few connections, try to connect to more peers
	if connectedCount < 3 {
		cm.connectToBootstrapPeers()
		
		// Also try to find new peers via DHT
		cm.discoverPeers()
	}
	
	// Check for stale connections
	cm.pruneStaleConnections()
}

// discoverPeers attempts to discover new peers via DHT
func (cm *ConnectionManager) discoverPeers() {
	// This would integrate with DHT discovery
	cm.logger.Debug("Attempting peer discovery...")
	
	// Get connected peers and ask them for more peers
	for _, p := range cm.host.Network().Peers() {
		if cm.host.Network().Connectedness(p) == network.Connected {
			// In a real implementation, this would query the DHT
			cm.logger.Debugf("Could query peer %s for more peers", p)
		}
	}
}

// pruneStaleConnections removes connections that have been idle too long
func (cm *ConnectionManager) pruneStaleConnections() {
	cm.mu.Lock()
	defer cm.mu.Unlock()
	
	now := time.Now()
	for peerID, lastSeen := range cm.connectedPeers {
		if now.Sub(lastSeen) > 5*time.Minute {
			// Check if still connected
			if cm.host.Network().Connectedness(peerID) != network.Connected {
				delete(cm.connectedPeers, peerID)
				cm.logger.Debugf("Removed stale peer %s", peerID)
			}
		}
	}
}

// onPeerConnected handles peer connection events
func (cm *ConnectionManager) onPeerConnected(p peer.ID) {
	cm.mu.Lock()
	cm.connectedPeers[p] = time.Now()
	cm.mu.Unlock()
	
	cm.logger.Infof("Peer connected: %s", p)
}

// onPeerDisconnected handles peer disconnection events
func (cm *ConnectionManager) onPeerDisconnected(p peer.ID) {
	cm.mu.Lock()
	delete(cm.connectedPeers, p)
	cm.mu.Unlock()
	
	cm.logger.Infof("Peer disconnected: %s", p)
}

// GetConnectedPeers returns currently connected peers
func (cm *ConnectionManager) GetConnectedPeers() []peer.ID {
	cm.mu.RLock()
	defer cm.mu.RUnlock()
	
	peers := make([]peer.ID, 0, len(cm.connectedPeers))
	for p := range cm.connectedPeers {
		if cm.host.Network().Connectedness(p) == network.Connected {
			peers = append(peers, p)
		}
	}
	
	return peers
}

// IsConnected checks if we have any connections
func (cm *ConnectionManager) IsConnected() bool {
	return len(cm.GetConnectedPeers()) > 0
}

// ForceReconnect forces reconnection attempts
func (cm *ConnectionManager) ForceReconnect() {
	cm.logger.Info("Forcing reconnection to bootstrap peers...")
	cm.connectToBootstrapPeers()
}

// GetConnectionStats returns connection statistics
func (cm *ConnectionManager) GetConnectionStats() map[string]interface{} {
	cm.mu.RLock()
	defer cm.mu.RUnlock()
	
	stats := map[string]interface{}{
		"connected_peers": len(cm.connectedPeers),
		"total_peers":    len(cm.host.Network().Peers()),
		"connections":    len(cm.host.Network().Conns()),
	}
	
	// Add peer details
	peerList := make([]map[string]interface{}, 0)
	for p, lastSeen := range cm.connectedPeers {
		peerInfo := map[string]interface{}{
			"id":         p.String(),
			"last_seen":  lastSeen.Format(time.RFC3339),
			"connected":  cm.host.Network().Connectedness(p) == network.Connected,
		}
		peerList = append(peerList, peerInfo)
	}
	stats["peers"] = peerList
	
	return stats
}
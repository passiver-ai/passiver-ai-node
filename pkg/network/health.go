package network

import (
	"context"
	"sync"
	"time"
)

// HealthMonitor monitors network health and handles recovery
type HealthMonitor struct {
	networkManager *NetworkManager
	nodeStates     map[string]*NodeState
	statesMux      sync.RWMutex
	unhealthy      chan string
}

// NodeState represents the health state of a node
type NodeState struct {
	LastHeartbeat time.Time
	FailureCount  int
	IsHealthy     bool
}

// NewHealthMonitor creates a new health monitor
func NewHealthMonitor(nm *NetworkManager) *HealthMonitor {
	return &HealthMonitor{
		networkManager: nm,
		nodeStates:     make(map[string]*NodeState),
		unhealthy:      make(chan string, 100),
	}
}

// Start starts the health monitoring
func (hm *HealthMonitor) Start(ctx context.Context) {
	go hm.monitorHealth(ctx)
	go hm.handleRecovery(ctx)
}

// monitorHealth periodically checks node health
func (hm *HealthMonitor) monitorHealth(ctx context.Context) {
	ticker := time.NewTicker(time.Second * 10)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			hm.checkNodes()
		}
	}
}

// checkNodes checks the health of all nodes
func (hm *HealthMonitor) checkNodes() {
	hm.statesMux.Lock()
	defer hm.statesMux.Unlock()

	now := time.Now()
	nodes := hm.networkManager.GetConnectedNodes()

	for _, nodeID := range nodes {
		if node, exists := hm.networkManager.GetNode(nodeID); exists {
			status := node.GetStatus()
			state, exists := hm.nodeStates[nodeID]
			if !exists {
				state = &NodeState{
					LastHeartbeat: now,
					IsHealthy:     true,
				}
				hm.nodeStates[nodeID] = state
			}

			// Check if node is healthy
			if now.Sub(status.LastHeartbeat) > time.Second*30 {
				state.FailureCount++
				if state.IsHealthy && state.FailureCount >= 3 {
					state.IsHealthy = false
					hm.unhealthy <- nodeID
				}
			} else {
				state.FailureCount = 0
				state.IsHealthy = true
			}
			state.LastHeartbeat = status.LastHeartbeat
		}
	}
}

// handleRecovery handles recovery of unhealthy nodes
func (hm *HealthMonitor) handleRecovery(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			return
		case nodeID := <-hm.unhealthy:
			hm.recoverNode(ctx, nodeID)
		}
	}
}

// recoverNode attempts to recover an unhealthy node
func (hm *HealthMonitor) recoverNode(ctx context.Context, nodeID string) {
	node, exists := hm.networkManager.GetNode(nodeID)
	if !exists {
		return
	}

	// Try to restart the node
	if err := node.Stop(ctx); err != nil {
		// Log error
		return
	}

	if err := node.Start(ctx); err != nil {
		// If restart fails, remove the node
		hm.networkManager.UnregisterNode(nodeID)
		return
	}

	// Update node state
	hm.statesMux.Lock()
	if state, exists := hm.nodeStates[nodeID]; exists {
		state.FailureCount = 0
		state.IsHealthy = true
		state.LastHeartbeat = time.Now()
	}
	hm.statesMux.Unlock()
}

// GetNodeHealth returns the health status of a node
func (hm *HealthMonitor) GetNodeHealth(nodeID string) *NodeState {
	hm.statesMux.RLock()
	defer hm.statesMux.RUnlock()
	return hm.nodeStates[nodeID]
}

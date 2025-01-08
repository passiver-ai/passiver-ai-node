package router

import (
	"context"
	"sync"
	"time"

	"passiver-ai/pkg/common"
)

// RouterNode implements the router node functionality
type RouterNode struct {
	id        string
	status    common.NodeStatus
	statusMux sync.RWMutex
	routes    map[string][]string // maps destination to possible routes
	routesMux sync.RWMutex
	metrics   *RouterMetrics
}

// RouterMetrics tracks performance metrics for the router
type RouterMetrics struct {
	TotalRequests    uint64
	SuccessfulRoutes uint64
	FailedRoutes     uint64
	AverageLatency   time.Duration
	metricsMux       sync.RWMutex
}

// NewRouterNode creates a new router node
func NewRouterNode(id string) *RouterNode {
	return &RouterNode{
		id: id,
		status: common.NodeStatus{
			ID:            id,
			Type:          common.RouterNode,
			IsActive:      false,
			LastHeartbeat: time.Now(),
			Resources:     common.Resources{},
		},
		routes:  make(map[string][]string),
		metrics: &RouterMetrics{},
	}
}

// Start initializes and starts the router node
func (r *RouterNode) Start(ctx context.Context) error {
	r.setActive(true)
	go r.heartbeat(ctx)
	go r.updateMetrics(ctx)
	return nil
}

// Stop gracefully shuts down the router node
func (r *RouterNode) Stop(ctx context.Context) error {
	r.setActive(false)
	return nil
}

// GetStatus returns the current status of the node
func (r *RouterNode) GetStatus() common.NodeStatus {
	r.statusMux.RLock()
	defer r.statusMux.RUnlock()
	return r.status
}

// GetType returns the type of the node
func (r *RouterNode) GetType() common.NodeType {
	return common.RouterNode
}

// RouteMessage routes a message to its destination
func (r *RouterNode) RouteMessage(msg *common.Message) error {
	r.metrics.metricsMux.Lock()
	r.metrics.TotalRequests++
	r.metrics.metricsMux.Unlock()

	start := time.Now()

	// Implement routing logic here
	// This could include:
	// - Finding the optimal path
	// - Load balancing
	// - QoS enforcement
	// - Traffic prioritization

	r.metrics.metricsMux.Lock()
	r.metrics.SuccessfulRoutes++
	r.metrics.AverageLatency = (r.metrics.AverageLatency + time.Since(start)) / 2
	r.metrics.metricsMux.Unlock()

	return nil
}

// UpdateRoutes updates the routing table
func (r *RouterNode) UpdateRoutes(destination string, routes []string) {
	r.routesMux.Lock()
	defer r.routesMux.Unlock()
	r.routes[destination] = routes
}

// GetRoutes returns available routes for a destination
func (r *RouterNode) GetRoutes(destination string) []string {
	r.routesMux.RLock()
	defer r.routesMux.RUnlock()
	return r.routes[destination]
}

// heartbeat updates the node's last heartbeat time
func (r *RouterNode) heartbeat(ctx context.Context) {
	ticker := time.NewTicker(time.Second * 5)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			r.statusMux.Lock()
			r.status.LastHeartbeat = time.Now()
			r.statusMux.Unlock()
		}
	}
}

// updateMetrics periodically updates router metrics
func (r *RouterNode) updateMetrics(ctx context.Context) {
	ticker := time.NewTicker(time.Minute)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			// Implement metrics update logic here
			// This could include:
			// - Performance statistics
			// - Resource utilization
			// - Network health indicators
		}
	}
}

// setActive updates the active status of the node
func (r *RouterNode) setActive(active bool) {
	r.statusMux.Lock()
	defer r.statusMux.Unlock()
	r.status.IsActive = active
}

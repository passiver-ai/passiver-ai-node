package network

import (
	"context"
	"sync"
	"time"

	"passiver-ai/pkg/common"
)

// NetworkManager handles all network communications between nodes
type NetworkManager struct {
	nodes         map[string]common.Node
	nodesMux      sync.RWMutex
	msgChan       chan common.Message
	resultChan    chan common.Result
	closeChan     chan struct{}
	healthMonitor *HealthMonitor
}

// NewNetworkManager creates a new instance of NetworkManager
func NewNetworkManager() *NetworkManager {
	nm := &NetworkManager{
		nodes:      make(map[string]common.Node),
		msgChan:    make(chan common.Message, 1000),
		resultChan: make(chan common.Result, 1000),
		closeChan:  make(chan struct{}),
	}
	nm.healthMonitor = NewHealthMonitor(nm)
	return nm
}

// Start initializes the network manager and starts listening for messages
func (nm *NetworkManager) Start(ctx context.Context) error {
	go nm.handleMessages(ctx)
	nm.healthMonitor.Start(ctx)
	return nil
}

// Stop gracefully shuts down the network manager
func (nm *NetworkManager) Stop() {
	close(nm.closeChan)
}

// RegisterNode adds a new node to the network
func (nm *NetworkManager) RegisterNode(nodeID string, node common.Node) {
	nm.nodesMux.Lock()
	defer nm.nodesMux.Unlock()
	nm.nodes[nodeID] = node
}

// UnregisterNode removes a node from the network
func (nm *NetworkManager) UnregisterNode(nodeID string) {
	nm.nodesMux.Lock()
	defer nm.nodesMux.Unlock()
	delete(nm.nodes, nodeID)
}

// SendMessage sends a message to a specific node
func (nm *NetworkManager) SendMessage(msg common.Message) error {
	select {
	case nm.msgChan <- msg:
		return nil
	default:
		return ErrChannelFull
	}
}

// GetResult waits for and returns a result for a specific task
func (nm *NetworkManager) GetResult(taskID string, timeout time.Duration) (*common.Result, error) {
	timer := time.NewTimer(timeout)
	defer timer.Stop()

	for {
		select {
		case result := <-nm.resultChan:
			if result.TaskID == taskID {
				return &result, nil
			}
		case <-timer.C:
			return nil, ErrTimeout
		}
	}
}

// handleMessages processes incoming messages
func (nm *NetworkManager) handleMessages(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			return
		case <-nm.closeChan:
			return
		case msg := <-nm.msgChan:
			nm.routeMessage(msg)
		}
	}
}

// routeMessage routes the message to the appropriate node
func (nm *NetworkManager) routeMessage(msg common.Message) {
	nm.nodesMux.RLock()
	defer nm.nodesMux.RUnlock()

	// If the message is for a specific node
	if msg.To != "" {
		if node, exists := nm.nodes[msg.To]; exists {
			switch node.GetType() {
			case common.LLMNode:
				if llmNode, ok := node.(LLMNode); ok {
					go nm.handleLLMMessage(llmNode, msg)
				}
			case common.RouterNode:
				if routerNode, ok := node.(RouterNode); ok {
					go nm.handleRouterMessage(routerNode, msg)
				}
			case common.ValidatorNode:
				if validatorNode, ok := node.(ValidatorNode); ok {
					go nm.handleValidatorMessage(validatorNode, msg)
				}
			}
		}
		return
	}

	// If the message is for a type of node (load balancing)
	switch msg.Type {
	case "llm_process":
		nm.routeToLLMNode(msg)
	case "validation":
		nm.routeToValidatorNode(msg)
	}
}

// routeToLLMNode routes a message to an available LLM node
func (nm *NetworkManager) routeToLLMNode(msg common.Message) {
	var selectedNode common.Node
	var minTasks = int64(-1)

	for _, node := range nm.nodes {
		if node.GetType() != common.LLMNode {
			continue
		}

		if llmNode, ok := node.(LLMNode); ok {
			resources := llmNode.GetGPUResources()
			taskCount := int64(resources.GetUsedMemory() * 100 / resources.GetTotalMemory())

			if minTasks == -1 || taskCount < minTasks {
				minTasks = taskCount
				selectedNode = node
			}
		}
	}

	if selectedNode != nil {
		msg.To = selectedNode.GetStatus().ID
		nm.routeMessage(msg)
	}
}

// routeToValidatorNode routes a message to an available validator node
func (nm *NetworkManager) routeToValidatorNode(msg common.Message) {
	// Simple round-robin selection for validator nodes
	for _, node := range nm.nodes {
		if node.GetType() == common.ValidatorNode {
			msg.To = node.GetStatus().ID
			nm.routeMessage(msg)
			return
		}
	}
}

// Interface types for type assertions
type (
	LLMNode interface {
		common.Node
		ProcessTask(taskID string, input interface{}) (ComputationTask, error)
		GetGPUResources() GPUResources
	}

	RouterNode interface {
		common.Node
		RouteMessage(msg *common.Message) error
	}

	ValidatorNode interface {
		common.Node
		ValidateResult(result *common.Result) error
	}

	ComputationTask interface {
		GetID() string
		GetStatus() string
	}

	GPUResources interface {
		GetTotalMemory() int64
		GetUsedMemory() int64
	}
)

// Message handlers
func (nm *NetworkManager) handleLLMMessage(node LLMNode, msg common.Message) {
	if req, ok := msg.Payload.(map[string]interface{}); ok {
		task, err := node.ProcessTask(msg.ID, req)
		if err != nil {
			nm.resultChan <- common.Result{
				TaskID:    msg.ID,
				NodeID:    node.GetStatus().ID,
				Status:    "error",
				Error:     err.Error(),
				Timestamp: time.Now(),
			}
			return
		}

		nm.resultChan <- common.Result{
			TaskID:    msg.ID,
			NodeID:    node.GetStatus().ID,
			Status:    task.GetStatus(),
			Data:      task,
			Timestamp: time.Now(),
		}
	}
}

func (nm *NetworkManager) handleRouterMessage(node RouterNode, msg common.Message) {
	err := node.RouteMessage(&msg)
	if err != nil {
		// Handle routing error
		nm.resultChan <- common.Result{
			TaskID:    msg.ID,
			NodeID:    node.GetStatus().ID,
			Status:    "error",
			Error:     err.Error(),
			Timestamp: time.Now(),
		}
	}
}

func (nm *NetworkManager) handleValidatorMessage(node ValidatorNode, msg common.Message) {
	if result, ok := msg.Payload.(*common.Result); ok {
		err := node.ValidateResult(result)
		if err != nil {
			// Handle validation error
			nm.resultChan <- common.Result{
				TaskID:    msg.ID,
				NodeID:    node.GetStatus().ID,
				Status:    "error",
				Error:     err.Error(),
				Timestamp: time.Now(),
			}
		}
	}
}

// GetConnectedNodes returns a list of all connected nodes
func (nm *NetworkManager) GetConnectedNodes() []string {
	nm.nodesMux.RLock()
	defer nm.nodesMux.RUnlock()

	nodes := make([]string, 0, len(nm.nodes))
	for nodeID := range nm.nodes {
		nodes = append(nodes, nodeID)
	}
	return nodes
}

// GetNode returns a node by its ID
func (nm *NetworkManager) GetNode(nodeID string) (common.Node, bool) {
	nm.nodesMux.RLock()
	defer nm.nodesMux.RUnlock()
	node, exists := nm.nodes[nodeID]
	return node, exists
}

// GetNodesByType returns all nodes of a specific type
func (nm *NetworkManager) GetNodesByType(nodeType common.NodeType) []common.Node {
	nm.nodesMux.RLock()
	defer nm.nodesMux.RUnlock()

	var nodes []common.Node
	for _, node := range nm.nodes {
		if node.GetType() == nodeType {
			nodes = append(nodes, node)
		}
	}
	return nodes
}

// GetNodeCount returns the count of nodes by type
func (nm *NetworkManager) GetNodeCount() map[common.NodeType]int {
	nm.nodesMux.RLock()
	defer nm.nodesMux.RUnlock()

	counts := make(map[common.NodeType]int)
	for _, node := range nm.nodes {
		counts[node.GetType()]++
	}
	return counts
}

// GetHealthMonitor returns the health monitor instance
func (nm *NetworkManager) GetHealthMonitor() *HealthMonitor {
	return nm.healthMonitor
}

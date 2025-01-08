package common

import (
	"context"
	"time"
)

// NodeType represents the type of node in the network
type NodeType string

const (
	ValidatorNode NodeType = "validator"
	RouterNode    NodeType = "router"
	LLMNode       NodeType = "llm"
)

// Node represents a base interface for all node types
type Node interface {
	// Start initializes and starts the node
	Start(ctx context.Context) error
	// Stop gracefully shuts down the node
	Stop(ctx context.Context) error
	// GetStatus returns the current status of the node
	GetStatus() NodeStatus
	// GetType returns the type of the node
	GetType() NodeType
}

// NodeStatus represents the current status of a node
type NodeStatus struct {
	ID            string    `json:"id"`
	Type          NodeType  `json:"type"`
	IsActive      bool      `json:"is_active"`
	LastHeartbeat time.Time `json:"last_heartbeat"`
	Resources     Resources `json:"resources"`
}

// Resources represents the computational resources of a node
type Resources struct {
	GPUCount     int     `json:"gpu_count"`
	GPUMemory    int64   `json:"gpu_memory"` // in MB
	CPUCores     int     `json:"cpu_cores"`
	MemoryTotal  int64   `json:"memory_total"`  // in MB
	MemoryUsed   int64   `json:"memory_used"`   // in MB
	NetworkSpeed float64 `json:"network_speed"` // in Mbps
}

// Message represents a basic communication unit between nodes
type Message struct {
	ID        string      `json:"id"`
	Type      string      `json:"type"`
	From      string      `json:"from"`
	To        string      `json:"to"`
	Timestamp time.Time   `json:"timestamp"`
	Payload   interface{} `json:"payload"`
}

// Result represents the computation result
type Result struct {
	TaskID    string      `json:"task_id"`
	NodeID    string      `json:"node_id"`
	Status    string      `json:"status"`
	Data      interface{} `json:"data"`
	Error     string      `json:"error,omitempty"`
	Timestamp time.Time   `json:"timestamp"`
}

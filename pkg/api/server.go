package api

import (
	"encoding/json"
	"fmt"
	"net/http"
	"sync"
	"time"

	"passiver-ai/pkg/common"
	"passiver-ai/pkg/network"
)

// Server represents the API server
type Server struct {
	networkManager *network.NetworkManager
	port           int
	server         *http.Server
	mu             sync.RWMutex
}

// LLMRequest represents an LLM API request
type LLMRequest struct {
	Input       string                 `json:"input"`
	ModelParams map[string]interface{} `json:"model_params,omitempty"`
}

// LLMResponse represents an LLM API response
type LLMResponse struct {
	TaskID    string      `json:"task_id"`
	Status    string      `json:"status"`
	Output    interface{} `json:"output,omitempty"`
	Error     string      `json:"error,omitempty"`
	Timestamp time.Time   `json:"timestamp"`
}

// SystemStatus represents the system status response
type SystemStatus struct {
	Nodes      map[string]NodeStatus `json:"nodes"`
	NodeCounts map[string]int        `json:"node_counts"`
	Timestamp  time.Time             `json:"timestamp"`
}

// NodeStatus represents the status of a single node
type NodeStatus struct {
	Type      string    `json:"type"`
	IsHealthy bool      `json:"is_healthy"`
	LastSeen  time.Time `json:"last_seen"`
	Resources struct {
		GPUMemoryTotal int64   `json:"gpu_memory_total,omitempty"`
		GPUMemoryUsed  int64   `json:"gpu_memory_used,omitempty"`
		GPUUtilization float64 `json:"gpu_utilization,omitempty"`
	} `json:"resources,omitempty"`
}

// NewServer creates a new API server
func NewServer(networkManager *network.NetworkManager, port int) *Server {
	return &Server{
		networkManager: networkManager,
		port:           port,
	}
}

// Start starts the API server
func (s *Server) Start() error {
	mux := http.NewServeMux()

	// API endpoints
	mux.HandleFunc("/api/v1/llm/process", s.handleLLMProcess)
	mux.HandleFunc("/api/v1/llm/status", s.handleLLMStatus)
	mux.HandleFunc("/api/v1/system/status", s.handleSystemStatus)
	mux.HandleFunc("/api/v1/metrics", s.handleMetrics)

	s.server = &http.Server{
		Addr:    fmt.Sprintf(":%d", s.port),
		Handler: mux,
	}

	return s.server.ListenAndServe()
}

// Stop stops the API server
func (s *Server) Stop() error {
	if s.server != nil {
		return s.server.Close()
	}
	return nil
}

func (s *Server) handleLLMProcess(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req LLMRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Create a message for LLM processing
	msg := common.Message{
		ID:        fmt.Sprintf("task-%d", time.Now().UnixNano()),
		Type:      "llm_process",
		From:      "api",
		To:        "llm", // This will be routed to an available LLM node
		Timestamp: time.Now(),
		Payload:   req,
	}

	// Send the message through the network manager
	if err := s.networkManager.SendMessage(msg); err != nil {
		http.Error(w, "Failed to process request", http.StatusInternalServerError)
		return
	}

	// Return task ID to client
	response := LLMResponse{
		TaskID:    msg.ID,
		Status:    "processing",
		Timestamp: time.Now(),
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func (s *Server) handleLLMStatus(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	taskID := r.URL.Query().Get("task_id")
	if taskID == "" {
		http.Error(w, "Task ID is required", http.StatusBadRequest)
		return
	}

	// Create a message to query task status
	msg := common.Message{
		ID:        fmt.Sprintf("status-%d", time.Now().UnixNano()),
		Type:      "llm_status",
		From:      "api",
		To:        "llm",
		Timestamp: time.Now(),
		Payload:   taskID,
	}

	if err := s.networkManager.SendMessage(msg); err != nil {
		http.Error(w, "Failed to get status", http.StatusInternalServerError)
		return
	}

	// In a real implementation, we would wait for the response
	// For now, return a placeholder response
	response := LLMResponse{
		TaskID:    taskID,
		Status:    "processing",
		Timestamp: time.Now(),
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func (s *Server) handleSystemStatus(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Get all nodes
	nodes := s.networkManager.GetConnectedNodes()
	nodeStatuses := make(map[string]NodeStatus)

	// Get node counts by type
	nodeCounts := make(map[string]int)
	for nodeType, count := range s.networkManager.GetNodeCount() {
		nodeCounts[string(nodeType)] = count
	}

	// Get detailed status for each node
	for _, nodeID := range nodes {
		if node, exists := s.networkManager.GetNode(nodeID); exists {
			status := node.GetStatus()
			nodeStatus := NodeStatus{
				Type:     string(status.Type),
				LastSeen: status.LastHeartbeat,
			}

			// Get GPU resources for LLM nodes
			if status.Type == common.LLMNode {
				if llmNode, ok := node.(network.LLMNode); ok {
					resources := llmNode.GetGPUResources()
					nodeStatus.Resources.GPUMemoryTotal = resources.GetTotalMemory()
					nodeStatus.Resources.GPUMemoryUsed = resources.GetUsedMemory()
				}
			}

			nodeStatuses[nodeID] = nodeStatus
		}
	}

	systemStatus := SystemStatus{
		Nodes:      nodeStatuses,
		NodeCounts: nodeCounts,
		Timestamp:  time.Now(),
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(systemStatus)
}

func (s *Server) handleMetrics(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	metrics := map[string]interface{}{
		"nodes": map[string]interface{}{},
	}

	// Collect metrics from all nodes
	for _, nodeID := range s.networkManager.GetConnectedNodes() {
		if node, exists := s.networkManager.GetNode(nodeID); exists {
			status := node.GetStatus()
			nodeMetrics := map[string]interface{}{
				"type":      status.Type,
				"resources": status.Resources,
			}

			if status.Type == common.LLMNode {
				if llmNode, ok := node.(network.LLMNode); ok {
					resources := llmNode.GetGPUResources()
					nodeMetrics["gpu"] = map[string]interface{}{
						"memory_total":     resources.GetTotalMemory(),
						"memory_used":      resources.GetUsedMemory(),
						"memory_available": resources.GetTotalMemory() - resources.GetUsedMemory(),
					}
				}
			}

			metrics["nodes"].(map[string]interface{})[nodeID] = nodeMetrics
		}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(metrics)
}

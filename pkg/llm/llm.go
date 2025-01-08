package llm

import (
	"context"
	"fmt"
	"sync"
	"time"

	"passiver-ai/pkg/common"
	"passiver-ai/pkg/llm/model"
)

// LLMNode implements the LLM node functionality
type LLMNode struct {
	id           string
	status       common.NodeStatus
	statusMux    sync.RWMutex
	modelConfig  ModelConfig
	tasks        map[string]*ComputationTask
	tasksMux     sync.RWMutex
	gpuResources *GPUResources
	modelManager *model.ModelManager
}

// ModelConfig represents the configuration for the LLM model
type ModelConfig struct {
	ModelName    string
	ModelVersion string
	BatchSize    int
	MaxSequence  int
	Temperature  float32
	ModelPath    string
	LlamaPath    string
}

// ComputationTask represents a single computation task
type ComputationTask struct {
	ID        string
	Status    string
	Input     interface{}
	Output    interface{}
	StartTime time.Time
	EndTime   time.Time
}

func (t *ComputationTask) GetID() string {
	return t.ID
}

func (t *ComputationTask) GetStatus() string {
	return t.Status
}

// GPUResources tracks GPU resource utilization
type GPUResources struct {
	totalMemory     int64
	usedMemory      int64
	availableMemory int64
	utilization     float64
	temperature     float64
	resourcesMux    sync.RWMutex
}

func (g *GPUResources) GetTotalMemory() int64 {
	g.resourcesMux.RLock()
	defer g.resourcesMux.RUnlock()
	return g.totalMemory
}

func (g *GPUResources) GetUsedMemory() int64 {
	g.resourcesMux.RLock()
	defer g.resourcesMux.RUnlock()
	return g.usedMemory
}

// NewLLMNode creates a new LLM node
func NewLLMNode(id string, config ModelConfig) (*LLMNode, error) {
	modelManager, err := model.NewModelManager(config.ModelPath, config.LlamaPath)
	if err != nil {
		return nil, fmt.Errorf("failed to create model manager: %v", err)
	}

	// Initialize the model
	if err := modelManager.Initialize(config.ModelName); err != nil {
		return nil, fmt.Errorf("failed to initialize model: %v", err)
	}

	return &LLMNode{
		id: id,
		status: common.NodeStatus{
			ID:            id,
			Type:          common.LLMNode,
			IsActive:      false,
			LastHeartbeat: time.Now(),
			Resources:     common.Resources{},
		},
		modelConfig:  config,
		tasks:        make(map[string]*ComputationTask),
		gpuResources: &GPUResources{totalMemory: 8 * 1024}, // 8GB default
		modelManager: modelManager,
	}, nil
}

// Start initializes and starts the LLM node
func (l *LLMNode) Start(ctx context.Context) error {
	l.setActive(true)
	go l.heartbeat(ctx)
	go l.monitorResources(ctx)
	return nil
}

// Stop gracefully shuts down the LLM node
func (l *LLMNode) Stop(ctx context.Context) error {
	l.setActive(false)
	return nil
}

// GetStatus returns the current status of the node
func (l *LLMNode) GetStatus() common.NodeStatus {
	l.statusMux.RLock()
	defer l.statusMux.RUnlock()
	return l.status
}

// GetType returns the type of the node
func (l *LLMNode) GetType() common.NodeType {
	return common.LLMNode
}

// ProcessTask processes an LLM computation task
func (l *LLMNode) ProcessTask(taskID string, input interface{}) (ComputationTask, error) {
	l.tasksMux.Lock()
	task := &ComputationTask{
		ID:        taskID,
		Status:    "processing",
		Input:     input,
		StartTime: time.Now(),
	}
	l.tasks[taskID] = task
	l.tasksMux.Unlock()

	// Extract input and parameters
	req, ok := input.(map[string]interface{})
	if !ok {
		task.Status = "error"
		task.Output = "invalid input format"
		return *task, fmt.Errorf("invalid input format")
	}

	inputText, ok := req["input"].(string)
	if !ok {
		task.Status = "error"
		task.Output = "input text not found"
		return *task, fmt.Errorf("input text not found")
	}

	// Get model parameters
	modelParams := make(map[string]interface{})
	if params, ok := req["model_params"].(map[string]interface{}); ok {
		modelParams = params
	}

	// Update GPU resource usage
	l.gpuResources.resourcesMux.Lock()
	l.gpuResources.usedMemory += 2 * 1024 // Assume 2GB per task
	l.gpuResources.utilization += 20.0    // Assume 20% utilization increase
	l.gpuResources.resourcesMux.Unlock()

	// Run inference
	output, err := l.modelManager.RunInference(inputText, modelParams)
	if err != nil {
		task.Status = "error"
		task.Output = map[string]interface{}{
			"error": err.Error(),
		}
		// Release GPU resources
		l.gpuResources.resourcesMux.Lock()
		l.gpuResources.usedMemory -= 2 * 1024
		l.gpuResources.utilization -= 20.0
		l.gpuResources.resourcesMux.Unlock()
		return *task, err
	}

	// Process the output
	l.tasksMux.Lock()
	task.Status = "completed"
	task.EndTime = time.Now()
	task.Output = map[string]interface{}{
		"result": output,
		"metadata": map[string]interface{}{
			"duration":   task.EndTime.Sub(task.StartTime).String(),
			"model_info": l.modelManager.GetModelInfo(),
			"parameters": modelParams,
		},
	}
	l.tasksMux.Unlock()

	// Release GPU resources
	l.gpuResources.resourcesMux.Lock()
	l.gpuResources.usedMemory -= 2 * 1024
	l.gpuResources.utilization -= 20.0
	l.gpuResources.resourcesMux.Unlock()

	return *task, nil
}

// GetGPUResources returns the current GPU resource metrics
func (l *LLMNode) GetGPUResources() *GPUResources {
	return l.gpuResources
}

// UpdateGPUResources updates the GPU resource metrics
func (l *LLMNode) UpdateGPUResources(totalMemory, usedMemory int64, utilization, temperature float64) {
	l.gpuResources.resourcesMux.Lock()
	defer l.gpuResources.resourcesMux.Unlock()

	l.gpuResources.totalMemory = totalMemory
	l.gpuResources.usedMemory = usedMemory
	l.gpuResources.availableMemory = totalMemory - usedMemory
	l.gpuResources.utilization = utilization
	l.gpuResources.temperature = temperature
}

// heartbeat updates the node's last heartbeat time
func (l *LLMNode) heartbeat(ctx context.Context) {
	ticker := time.NewTicker(time.Second * 5)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			l.statusMux.Lock()
			l.status.LastHeartbeat = time.Now()
			l.statusMux.Unlock()
		}
	}
}

// monitorResources periodically monitors GPU resources
func (l *LLMNode) monitorResources(ctx context.Context) {
	ticker := time.NewTicker(time.Second * 30)
	defer ticker.Stop()

	totalMemory := int64(8 * 1024)                  // 8GB total memory
	usedMemory := int64(float64(totalMemory) * 0.3) // 30% used memory

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			// Simulate GPU resource monitoring
			l.UpdateGPUResources(
				totalMemory, // 8GB total memory
				usedMemory,  // 30% used memory
				30.0,        // 30% utilization
				50.0,        // 50°C temperature
			)
		}
	}
}

// setActive updates the active status of the node
func (l *LLMNode) setActive(active bool) {
	l.statusMux.Lock()
	defer l.statusMux.Unlock()
	l.status.IsActive = active
}

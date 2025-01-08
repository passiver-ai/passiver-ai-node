package validator

import (
	"context"
	"sync"
	"time"

	"passiver-ai/pkg/common"
)

// ValidatorNode implements the validator node functionality
type ValidatorNode struct {
	id        string
	status    common.NodeStatus
	statusMux sync.RWMutex
	tasks     map[string]*ValidationTask
	tasksMux  sync.RWMutex
}

// ValidationTask represents a single validation task
type ValidationTask struct {
	ID        string
	Status    string
	Result    *common.Result
	StartTime time.Time
	EndTime   time.Time
}

// NewValidatorNode creates a new validator node
func NewValidatorNode(id string) *ValidatorNode {
	return &ValidatorNode{
		id: id,
		status: common.NodeStatus{
			ID:            id,
			Type:          common.ValidatorNode,
			IsActive:      false,
			LastHeartbeat: time.Now(),
			Resources:     common.Resources{},
		},
		tasks: make(map[string]*ValidationTask),
	}
}

// Start initializes and starts the validator node
func (v *ValidatorNode) Start(ctx context.Context) error {
	v.setActive(true)
	go v.heartbeat(ctx)
	return nil
}

// Stop gracefully shuts down the validator node
func (v *ValidatorNode) Stop(ctx context.Context) error {
	v.setActive(false)
	return nil
}

// GetStatus returns the current status of the node
func (v *ValidatorNode) GetStatus() common.NodeStatus {
	v.statusMux.RLock()
	defer v.statusMux.RUnlock()
	return v.status
}

// GetType returns the type of the node
func (v *ValidatorNode) GetType() common.NodeType {
	return common.ValidatorNode
}

// ValidateResult validates a computation result
func (v *ValidatorNode) ValidateResult(result *common.Result) error {
	v.tasksMux.Lock()
	defer v.tasksMux.Unlock()

	task := &ValidationTask{
		ID:        result.TaskID,
		Status:    "validating",
		Result:    result,
		StartTime: time.Now(),
	}
	v.tasks[result.TaskID] = task

	// Implement validation logic here
	// This could include:
	// - Checking data integrity
	// - Verifying computation correctness
	// - Ensuring compliance with system rules

	task.Status = "completed"
	task.EndTime = time.Now()
	return nil
}

// heartbeat updates the node's last heartbeat time
func (v *ValidatorNode) heartbeat(ctx context.Context) {
	ticker := time.NewTicker(time.Second * 5)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			v.statusMux.Lock()
			v.status.LastHeartbeat = time.Now()
			v.statusMux.Unlock()
		}
	}
}

// setActive updates the active status of the node
func (v *ValidatorNode) setActive(active bool) {
	v.statusMux.Lock()
	defer v.statusMux.Unlock()
	v.status.IsActive = active
}

package model

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
)

// ModelManager handles LLM model operations
type ModelManager struct {
	modelPath    string
	llamaPath    string
	currentModel string
	mu           sync.RWMutex
}

// NewModelManager creates a new model manager
func NewModelManager(modelPath, llamaPath string) (*ModelManager, error) {
	if err := os.MkdirAll(modelPath, 0755); err != nil {
		return nil, fmt.Errorf("failed to create model directory: %v", err)
	}

	return &ModelManager{
		modelPath: modelPath,
		llamaPath: llamaPath,
	}, nil
}

// Initialize downloads and sets up the LLM model
func (mm *ModelManager) Initialize(modelName string) error {
	mm.mu.Lock()
	defer mm.mu.Unlock()

	modelFile := filepath.Join(mm.modelPath, modelName)
	if _, err := os.Stat(modelFile); os.IsNotExist(err) {
		// Download model if it doesn't exist
		if err := mm.downloadModel(modelName); err != nil {
			return fmt.Errorf("failed to download model: %v", err)
		}
	}

	mm.currentModel = modelName
	return nil
}

// RunInference runs inference using the current model
func (mm *ModelManager) RunInference(input string, params map[string]interface{}) (string, error) {
	mm.mu.RLock()
	defer mm.mu.RUnlock()

	if mm.currentModel == "" {
		return "", fmt.Errorf("no model loaded")
	}

	// Set default parameters if not provided
	if params == nil {
		params = make(map[string]interface{})
	}
	temperature := params["temperature"]
	if temperature == nil {
		temperature = 0.7
	}
	maxTokens := params["max_tokens"]
	if maxTokens == nil {
		maxTokens = 2048
	}
	threads := params["threads"]
	if threads == nil {
		threads = 4
	}
	batchSize := params["batch_size"]
	if batchSize == nil {
		batchSize = 512
	}

	// Get executable path based on OS
	execName := "main"
	if runtime.GOOS == "windows" {
		execName = "main.exe"
	}

	// Prepare command arguments
	args := []string{
		"-m", filepath.Join(mm.modelPath, mm.currentModel),
		"--prompt", input,
		"--temp", fmt.Sprintf("%v", temperature),
		"--ctx-size", fmt.Sprintf("%v", maxTokens),
		"--threads", fmt.Sprintf("%v", threads),
		"--batch-size", fmt.Sprintf("%v", batchSize),
		"--n-predict", fmt.Sprintf("%v", maxTokens),
		"--repeat_penalty", "1.1",
		"--top_k", "40",
		"--top_p", "0.9",
		"-n", "-1", // No line numbers in output
		"-r", ">>", // Use ">>" as response indicator
	}

	// Run llama.cpp inference
	cmd := exec.Command(filepath.Join(mm.llamaPath, execName), args...)
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return "", fmt.Errorf("failed to create stdout pipe: %v", err)
	}

	stderr, err := cmd.StderrPipe()
	if err != nil {
		return "", fmt.Errorf("failed to create stderr pipe: %v", err)
	}

	if err := cmd.Start(); err != nil {
		return "", fmt.Errorf("failed to start inference: %v", err)
	}

	// Read output asynchronously
	var response strings.Builder
	var errOutput strings.Builder

	// Read stdout
	go func() {
		reader := bufio.NewReader(stdout)
		for {
			line, err := reader.ReadString('\n')
			if err != nil {
				if err != io.EOF {
					fmt.Printf("Error reading stdout: %v\n", err)
				}
				return
			}
			if strings.HasPrefix(line, ">>") {
				response.WriteString(strings.TrimPrefix(line, ">>"))
			}
		}
	}()

	// Read stderr
	go func() {
		reader := bufio.NewReader(stderr)
		for {
			line, err := reader.ReadString('\n')
			if err != nil {
				if err != io.EOF {
					fmt.Printf("Error reading stderr: %v\n", err)
				}
				return
			}
			errOutput.WriteString(line)
		}
	}()

	// Wait for completion
	if err := cmd.Wait(); err != nil {
		return "", fmt.Errorf("inference failed: %v\nError output: %s", err, errOutput.String())
	}

	return strings.TrimSpace(response.String()), nil
}

// downloadModel downloads the specified model
func (mm *ModelManager) downloadModel(modelName string) error {
	downloadScript := filepath.Join(mm.llamaPath, "scripts", "download-model.sh")
	cmd := exec.Command(downloadScript, modelName, mm.modelPath)
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to download model: %v", err)
	}
	return nil
}

// GetModelInfo returns information about the current model
func (mm *ModelManager) GetModelInfo() map[string]interface{} {
	mm.mu.RLock()
	defer mm.mu.RUnlock()

	return map[string]interface{}{
		"current_model": mm.currentModel,
		"model_path":    mm.modelPath,
		"llama_path":    mm.llamaPath,
	}
}

package main

import (
	"context"
	"encoding/json"
	"flag"
	"log"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"

	"passiver-ai/pkg/api"
	"passiver-ai/pkg/common"
	"passiver-ai/pkg/llm"
	"passiver-ai/pkg/network"
	"passiver-ai/pkg/router"
	"passiver-ai/pkg/validator"
)

type Config struct {
	ModelPath    string `json:"model_path"`
	LlamaPath    string `json:"llama_path"`
	DefaultModel string `json:"default_model"`
}

func loadConfig() (*Config, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return nil, err
	}

	configPath := filepath.Join(homeDir, ".passiver", "config.json")
	data, err := os.ReadFile(configPath)
	if err != nil {
		return nil, err
	}

	var config Config
	if err := json.Unmarshal(data, &config); err != nil {
		return nil, err
	}

	return &config, nil
}

func main() {
	// Parse command line flags
	nodeType := flag.String("type", "", "Type of node to run (validator/router/llm)")
	nodeID := flag.String("id", "", "Unique identifier for the node")
	apiPort := flag.Int("api-port", 8080, "Port for the API server")
	flag.Parse()

	if *nodeType == "" || *nodeID == "" {
		log.Fatal("Node type and ID must be specified")
	}

	// Load configuration
	config, err := loadConfig()
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	// Create context with cancellation
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Initialize network manager
	networkManager := network.NewNetworkManager()
	if err := networkManager.Start(ctx); err != nil {
		log.Fatalf("Failed to start network manager: %v", err)
	}

	// Initialize API server if this is a router node
	var apiServer *api.Server
	if *nodeType == string(common.RouterNode) {
		apiServer = api.NewServer(networkManager, *apiPort)
		go func() {
			if err := apiServer.Start(); err != nil {
				log.Printf("API server error: %v", err)
			}
		}()
		log.Printf("API server listening on port %d", *apiPort)
	}

	// Initialize node based on type
	var node common.Node
	switch common.NodeType(*nodeType) {
	case common.ValidatorNode:
		node = validator.NewValidatorNode(*nodeID)
	case common.RouterNode:
		node = router.NewRouterNode(*nodeID)
	case common.LLMNode:
		modelConfig := llm.ModelConfig{
			ModelName:    config.DefaultModel,
			ModelVersion: "1.0",
			BatchSize:    32,
			MaxSequence:  1024,
			Temperature:  0.7,
			ModelPath:    config.ModelPath,
			LlamaPath:    config.LlamaPath,
		}
		node, err = llm.NewLLMNode(*nodeID, modelConfig)
		if err != nil {
			log.Fatalf("Failed to create LLM node: %v", err)
		}
	default:
		log.Fatalf("Invalid node type: %s", *nodeType)
	}

	// Start the node
	if err := node.Start(ctx); err != nil {
		log.Fatalf("Failed to start node: %v", err)
	}

	// Register node with network manager
	networkManager.RegisterNode(*nodeID, node)
	log.Printf("Started %s node with ID: %s", *nodeType, *nodeID)

	// Set up signal handling for graceful shutdown
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	// Wait for shutdown signal
	<-sigChan
	log.Println("Shutting down...")

	// Clean up
	if apiServer != nil {
		if err := apiServer.Stop(); err != nil {
			log.Printf("Error stopping API server: %v", err)
		}
	}

	networkManager.UnregisterNode(*nodeID)
	if err := node.Stop(ctx); err != nil {
		log.Printf("Error stopping node: %v", err)
	}
}

package network

import "errors"

var (
	// ErrChannelFull is returned when the message channel is full
	ErrChannelFull = errors.New("message channel is full")
	// ErrNodeNotFound is returned when the target node is not found
	ErrNodeNotFound = errors.New("node not found")
	// ErrInvalidMessage is returned when the message format is invalid
	ErrInvalidMessage = errors.New("invalid message format")
	// ErrNetworkClosed is returned when trying to use a closed network
	ErrNetworkClosed = errors.New("network is closed")
	// ErrTimeout is returned when an operation times out
	ErrTimeout = errors.New("operation timed out")
	// ErrNoAvailableNodes is returned when no suitable nodes are available
	ErrNoAvailableNodes = errors.New("no available nodes")
	// ErrInvalidNodeType is returned when the node type is invalid
	ErrInvalidNodeType = errors.New("invalid node type")
	// ErrProcessingFailed is returned when task processing fails
	ErrProcessingFailed = errors.New("task processing failed")
)

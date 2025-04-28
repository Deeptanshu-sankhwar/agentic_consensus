package registry

import (
	"fmt"
	"strings"
	"sync"
)

type NodeInfo struct {
	IsGenesis bool
	Name      string
	RPCPort   int
	P2PPort   int
	APIPort   int
}

var (
	chainNodes    = make(map[string]map[string]NodeInfo)
	registryMutex sync.RWMutex
)

// RegisterNode adds a new node to the chain registry
func RegisterNode(chainID string, nodeID string, info NodeInfo) {
	registryMutex.Lock()
	defer registryMutex.Unlock()

	if _, exists := chainNodes[chainID]; !exists {
		chainNodes[chainID] = make(map[string]NodeInfo)
	}
	chainNodes[chainID][nodeID] = info
}

// GetRPCPortForChain returns the RPC port of the genesis node for a given chain
func GetRPCPortForChain(chainID string) (int, error) {
	registryMutex.RLock()
	defer registryMutex.RUnlock()

	nodes, exists := chainNodes[chainID]
	if !exists {
		return 0, fmt.Errorf("chain %s not found", chainID)
	}

	for _, info := range nodes {
		if info.IsGenesis {
			return info.RPCPort, nil
		}
	}

	return 0, fmt.Errorf("genesis node not found for chain %s", chainID)
}

// GetNodeByAPIPort retrieves node information by its API port
func GetNodeByAPIPort(chainID string, apiPort string) (string, NodeInfo, bool) {
	registryMutex.RLock()
	defer registryMutex.RUnlock()

	chainNodes, exists := chainNodes[chainID]
	if !exists {
		return "", NodeInfo{}, false
	}

	for nodeID, info := range chainNodes {
		if fmt.Sprintf("%d", info.APIPort) == apiPort {
			return nodeID, info, true
		}
	}
	return "", NodeInfo{}, false
}

// IsValidator checks if a node is a validator
func IsValidator(chainID string, nodeID string) bool {
	registryMutex.RLock()
	defer registryMutex.RUnlock()

	if _, exists := chainNodes[chainID][nodeID]; exists {
		return strings.HasPrefix(nodeID, "validator")
	}
	return false
}

// GetNodeInfoByChainID returns all nodes for a given chain
func GetNodeInfoByChainID(chainID string) (map[string]NodeInfo, bool) {
	registryMutex.RLock()
	defer registryMutex.RUnlock()

	nodes, exists := chainNodes[chainID]
	if !exists {
		return nil, false
	}

	nodesCopy := make(map[string]NodeInfo)
	for id, info := range nodes {
		nodesCopy[id] = info
	}

	return nodesCopy, true
}

// GetNodeInfo returns information for a specific node
func GetNodeInfo(chainID string, nodeID string) (NodeInfo, bool) {
	registryMutex.RLock()
	defer registryMutex.RUnlock()

	nodes, exists := chainNodes[chainID]
	if !exists {
		return NodeInfo{}, false
	}

	info, exists := nodes[nodeID]
	return info, exists
}

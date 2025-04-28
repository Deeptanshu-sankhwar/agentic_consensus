package registry

import (
	"encoding/json"
	"log"
	"os"
	"path/filepath"
	"sync"

	"github.com/Deeptanshu-sankhwar/agentic_consensus/core"
)

var (
	agentMutex   sync.RWMutex
	registryFile = "data/agent_registry.json"
	registry     *AgentRegistry
)

type AgentRegistry struct {
	Agents       map[string]map[string]core.Agent
	ValidatorMap map[string]map[string]string
}

// Initializes or loads the registry from file
func InitRegistry() {
	agentMutex.Lock()
	defer agentMutex.Unlock()

	if err := os.MkdirAll(filepath.Dir(registryFile), 0755); err != nil {
		log.Printf("Failed to create registry directory: %v", err)
		return
	}

	registry = loadRegistry()
	log.Printf("Registry initialized with %d agents", len(registry.Agents))
}

// Loads registry from file or creates new one if file doesn't exist
func loadRegistry() *AgentRegistry {
	data, err := os.ReadFile(registryFile)
	if err != nil {
		return &AgentRegistry{
			Agents:       make(map[string]map[string]core.Agent),
			ValidatorMap: make(map[string]map[string]string),
		}
	}

	var r AgentRegistry
	if err := json.Unmarshal(data, &r); err != nil {
		log.Printf("Failed to unmarshal registry: %v", err)
		return &AgentRegistry{
			Agents:       make(map[string]map[string]core.Agent),
			ValidatorMap: make(map[string]map[string]string),
		}
	}

	return &r
}

// Saves current registry state to file
func saveRegistry() {
	data, err := json.MarshalIndent(registry, "", "  ")
	if err != nil {
		log.Printf("Failed to marshal registry: %v", err)
		return
	}

	if err := os.WriteFile(registryFile, data, 0644); err != nil {
		log.Printf("Failed to save registry: %v", err)
	}
}

// Registers a new agent in the registry for a specific chain
func RegisterAgent(chainID string, agent core.Agent) {
	agentMutex.Lock()
	defer agentMutex.Unlock()

	if registry.Agents[chainID] == nil {
		registry.Agents[chainID] = make(map[string]core.Agent)
	}
	registry.Agents[chainID][agent.ID] = agent
	saveRegistry()
}

// Links an agent to a validator address and updates its status
func LinkAgentToValidator(chainID string, agentID string, validatorAddr string) bool {
	agentMutex.Lock()
	defer agentMutex.Unlock()

	if registry.ValidatorMap[chainID] == nil {
		registry.ValidatorMap[chainID] = make(map[string]string)
	}

	registry.ValidatorMap[chainID][validatorAddr] = agentID

	if agents, exists := registry.Agents[chainID]; exists {
		if agent, exists := agents[agentID]; exists {
			agent.IsValidator = true
			agent.ValidatorAddress = validatorAddr
			agents[agentID] = agent
		}
	}

	saveRegistry()
	return true
}

// Retrieves agent information for a given validator address
func GetAgentByValidator(chainID string, validatorAddr string) (core.Agent, bool) {
	agentMutex.RLock()
	defer agentMutex.RUnlock()

	if validatorMap, exists := registry.ValidatorMap[chainID]; exists {
		if agentID, exists := validatorMap[validatorAddr]; exists {
			if agents, exists := registry.Agents[chainID]; exists {
				if agent, exists := agents[agentID]; exists {
					return agent, true
				}
			}
		}
	}
	return core.Agent{}, false
}

// Returns all registered agents for a specific chain
func GetAllAgents(chainID string) []core.Agent {
	agentMutex.RLock()
	defer agentMutex.RUnlock()

	agents := make([]core.Agent, 0)
	if chainAgents, exists := registry.Agents[chainID]; exists {
		for _, agent := range chainAgents {
			agents = append(agents, agent)
		}
	}
	return agents
}

// Returns all validator-agent mappings for a specific chain
func GetAllValidatorAgentMappings(chainID string) map[string]string {
	agentMutex.RLock()
	defer agentMutex.RUnlock()

	result := make(map[string]string)
	if chainAgents, exists := registry.Agents[chainID]; exists {
		for valAddr, agent := range chainAgents {
			result[valAddr] = agent.ID
		}
	}
	return result
}

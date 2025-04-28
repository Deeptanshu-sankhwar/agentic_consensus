package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"hash/crc32"
	"net/http"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"time"

	"github.com/Deeptanshu-sankhwar/agentic_consensus/ai"
	"github.com/Deeptanshu-sankhwar/agentic_consensus/cmd/node"
	"github.com/Deeptanshu-sankhwar/agentic_consensus/communication"
	"github.com/Deeptanshu-sankhwar/agentic_consensus/core"
	"github.com/Deeptanshu-sankhwar/agentic_consensus/registry"
	"github.com/Deeptanshu-sankhwar/agentic_consensus/utils"
	cfg "github.com/cometbft/cometbft/config"
	"github.com/cometbft/cometbft/p2p"
	"github.com/cometbft/cometbft/privval"
	rpchttp "github.com/cometbft/cometbft/rpc/client/http"
	"github.com/cometbft/cometbft/types"
	"github.com/gin-gonic/gin"
)

type RelationshipUpdate struct {
	FromID   string  `json:"fromId"`
	TargetID string  `json:"targetId"`
	Score    float64 `json:"score"`
}

// RegisterAgent registers a new AI agent (Producer or Validator)
func RegisterAgent(c *gin.Context) {
	chainID := c.GetString("chainID")
	var agent core.Agent
	if err := c.ShouldBindJSON(&agent); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid agent data"})
		return
	}

	registry.RegisterAgent(chainID, agent)

	basePort := 26656
	agentIDInt := int(crc32.ChecksumIEEE([]byte(agent.ID)))
	p2pPort := basePort + (agentIDInt % 10000)
	rpcPort := p2pPort + 1
	apiPort := p2pPort + 2

	if p2pPort == 26656 || rpcPort == 26657 {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Agent port conflicts with Genesis node"})
		return
	}

	genesisNodeKeyFile := fmt.Sprintf("./data/%s/genesis/config/node_key.json", chainID)
	genesisNodeKey, err := p2p.LoadNodeKey(genesisNodeKeyFile)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to load genesis node key"})
		return
	}

	seedNode := fmt.Sprintf("%s@127.0.0.1:26656", genesisNodeKey.ID())

	cmdStr := fmt.Sprintf("cd %s && ./agent --chain %s --agent-id %s --p2p-port %d --rpc-port %d --genesis-node-id %s --role %s --api-port %d",
		getCurrentDir(), chainID, agent.ID, p2pPort, rpcPort, seedNode, agent.Role, apiPort)

	terminalCmd := exec.Command("osascript", "-e", fmt.Sprintf(`
		tell application "Terminal"
			do script "%s"
		end tell
	`, cmdStr))

	terminalCmd.Stdout = os.Stdout
	terminalCmd.Stderr = os.Stderr

	if err := terminalCmd.Start(); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("Failed to start agent process: %v", err)})
		return
	}

	errCh := make(chan error, 1)
	go func() {
		errCh <- terminalCmd.Wait()
	}()

	select {
	case <-errCh:
	case <-time.After(3 * time.Second):
	}

	registry.RegisterNode(chainID, agent.ID, registry.NodeInfo{
		IsGenesis: false,
		Name:      agent.ID,
		P2PPort:   p2pPort,
		RPCPort:   rpcPort,
		APIPort:   apiPort,
	})

	communication.BroadcastEvent(communication.EventAgentRegistered, agent)

	c.JSON(http.StatusOK, gin.H{
		"message": "Agent registered successfully",
		"agentID": agent.ID,
		"p2pPort": p2pPort,
		"rpcPort": rpcPort,
		"apiPort": apiPort,
	})
}

func getCurrentDir() string {
	dir, err := os.Getwd()
	if err != nil {
		return "."
	}
	return dir
}

// GetBlock fetches a block by height
func GetBlock(c *gin.Context) {
	chainID := c.GetString("chainID")
	height, err := strconv.Atoi(c.Param("height"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid block height"})
		return
	}

	rpcPort, err := registry.GetRPCPortForChain(chainID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": fmt.Sprintf("Chain not found: %v", err)})
		return
	}

	client, err := rpchttp.New(fmt.Sprintf("tcp://localhost:%d", rpcPort), "/websocket")
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("Failed to connect to node: %v", err)})
		return
	}

	status, err := client.Status(context.Background())
	if err != nil || status.NodeInfo.Network != chainID {
		c.JSON(http.StatusNotFound, gin.H{"error": "Chain not found"})
		return
	}

	heightPtr := new(int64)
	*heightPtr = int64(height)
	block, err := client.Block(context.Background(), heightPtr)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": fmt.Sprintf("Failed to get block: %v", err)})
		return
	}

	blockData := gin.H{
		"height":     block.Block.Height,
		"hash":       block.Block.Hash(),
		"timestamp":  block.Block.Time,
		"numTxs":     len(block.Block.Txs),
		"proposer":   block.Block.ProposerAddress,
		"validators": block.Block.LastCommit.Signatures,
	}

	c.JSON(http.StatusOK, gin.H{"block": blockData})
}

// GetNetworkStatus returns the current status of Blockchain
func GetNetworkStatus(c *gin.Context) {
	_ = c.GetString("chainID")

	client, err := rpchttp.New("tcp://localhost:26657", "/websocket")
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to connect to node"})
		return
	}

	netInfo, err := client.NetInfo(context.Background())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get network info"})
		return
	}

	networkStatus := gin.H{
		"netInfo": netInfo,
	}

	c.JSON(http.StatusOK, gin.H{"status": networkStatus})
}

// SubmitTransaction allows an agent to submit a transaction
func SubmitTransaction(c *gin.Context) {
	chainID := c.GetString("chainID")

	host := c.Request.Host
	apiPort := ""
	if i := strings.LastIndex(host, ":"); i != -1 {
		apiPort = host[i+1:]
	}

	_, nodeInfo, found := registry.GetNodeByAPIPort(chainID, apiPort)
	if !found {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Node not recognized"})
		return
	}

	var tx core.Transaction
	if err := c.ShouldBindJSON(&tx); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid transaction format"})
		return
	}

	client, err := rpchttp.New(fmt.Sprintf("tcp://localhost:%d", nodeInfo.RPCPort), "/websocket")
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("Failed to connect to node: %v", err)})
		return
	}

	status, err := client.Status(context.Background())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("Failed to get node status: %v", err)})
		return
	}

	tx.Data = status.ValidatorInfo.PubKey.Bytes()

	txBytes, err := tx.Marshal()
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Failed to encode transaction"})
		return
	}

	result, err := client.BroadcastTxSync(context.Background(), txBytes)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("Failed to broadcast tx: %v", err)})
		return
	}

	communication.BroadcastEvent(communication.EventNewTransaction, tx)

	c.JSON(http.StatusOK, gin.H{
		"message": "Transaction submitted successfully",
		"hash":    result.Hash.String(),
	})
}

// GetValidators returns the list of registered validators
func GetValidators(c *gin.Context) {
	chainID := c.GetString("chainID")
	host := c.Request.Host
	apiPort := ""
	if i := strings.LastIndex(host, ":"); i != -1 {
		apiPort = host[i+1:]
	}

	_, nodeInfo, found := registry.GetNodeByAPIPort(chainID, apiPort)
	if !found {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error":   fmt.Sprintf("Node not recognized for port %s", apiPort),
			"chainID": chainID,
			"apiPort": apiPort,
		})
		return
	}

	client, err := rpchttp.New(fmt.Sprintf("tcp://localhost:%d", nodeInfo.RPCPort), "/websocket")
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("Failed to connect to node: %v", err)})
		return
	}

	result, err := client.Validators(context.Background(), nil, nil, nil)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("Failed to get validators: %v", err)})
		return
	}

	c.JSON(http.StatusOK, gin.H{"validators": result.Validators})
}

// GetAllThreads returns all active discussion threads for monitoring
func GetAllThreads(c *gin.Context) {
	threads := communication.GetAllThreads()
	c.JSON(http.StatusOK, threads)
}

type CreateChainRequest struct {
	ChainID       string `json:"chain_id" binding:"required"`
	GenesisPrompt string `json:"genesis_prompt" binding:"required"`
}

// LoadSampleAgents loads sample agents from a generated file
func LoadSampleAgents(genesisPrompt string) ([]core.Agent, error) {
	filename, err := ai.GenerateAgents(genesisPrompt)
	if err != nil {
		return nil, fmt.Errorf("failed to generate agents: %v", err)
	}
	filename = "examples/" + filename

	fileContent, err := os.ReadFile(filename)
	if err != nil {
		return nil, fmt.Errorf("failed to read %s: %v", filename, err)
	}

	var agents []core.Agent
	if err := json.Unmarshal(fileContent, &agents); err != nil {
		return nil, fmt.Errorf("failed to parse %s: %v", filename, err)
	}

	return agents, nil
}

// CreateChain creates a new blockchain instance
func CreateChain(c *gin.Context) {
	var req CreateChainRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request format"})
		return
	}

	if _, err := registry.GetRPCPortForChain(req.ChainID); err == nil {
		c.JSON(http.StatusConflict, gin.H{"error": "Chain already exists"})
		return
	}

	config := cfg.DefaultConfig()
	config.BaseConfig.RootDir = "./data/" + req.ChainID
	config.Moniker = "genesis-node"
	config.P2P.ListenAddress = "tcp://0.0.0.0:0"
	config.RPC.ListenAddress = "tcp://0.0.0.0:0"

	genesisNodeKeyFile := fmt.Sprintf("./data/%s/genesis/config/node_key.json", req.ChainID)
	genesisNodeKey, err := p2p.LoadNodeKey(genesisNodeKeyFile)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to load genesis node key"})
		return
	}

	config.P2P.AllowDuplicateIP = true
	config.P2P.AddrBookStrict = false
	peerString := fmt.Sprintf("%s@127.0.0.1:26656", genesisNodeKey.ID())
	config.P2P.Seeds = peerString

	config.P2P.PexReactor = true
	config.P2P.MaxNumInboundPeers = 100
	config.P2P.MaxNumOutboundPeers = 30
	config.P2P.AddrBookStrict = false
	config.P2P.AllowDuplicateIP = true

	config.P2P.HandshakeTimeout = 20 * time.Second
	config.P2P.DialTimeout = 3 * time.Second
	config.P2P.FlushThrottleTimeout = 10 * time.Millisecond

	if err := os.MkdirAll(config.BaseConfig.RootDir+"/config", 0755); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("Failed to create config directory: %v", err)})
		return
	}

	privValKeyFile := config.PrivValidatorKeyFile()
	privValStateFile := config.PrivValidatorStateFile()
	if !utils.FileExists(privValKeyFile) {
		privVal := privval.GenFilePV(privValKeyFile, privValStateFile)
		privVal.Save()
	}

	nodeKeyFile := config.NodeKeyFile()
	if !utils.FileExists(nodeKeyFile) {
		if _, err := p2p.LoadOrGenNodeKey(nodeKeyFile); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("Failed to generate node key: %v", err)})
			return
		}
	}

	genesisFile := config.GenesisFile()
	if !utils.FileExists(genesisFile) {
		privVal := privval.LoadFilePV(privValKeyFile, privValStateFile)
		pubKey, err := privVal.GetPubKey()
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("Failed to get validator public key: %v", err)})
			return
		}

		genValidator := types.GenesisValidator{
			PubKey: pubKey,
			Power:  1000000,
			Name:   "genesis",
		}

		genDoc := types.GenesisDoc{
			ChainID:         req.ChainID,
			GenesisTime:     time.Now(),
			ConsensusParams: types.DefaultConsensusParams(),
			Validators:      []types.GenesisValidator{genValidator},
		}

		if err := genDoc.ValidateAndComplete(); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("Failed to validate genesis doc: %v", err)})
			return
		}

		if err := genDoc.SaveAs(genesisFile); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("Failed to create genesis file: %v", err)})
			return
		}
	}

	genesisNode, err := node.NewNode(config, req.ChainID, "")
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("Failed to create genesis node: %v", err)})
		return
	}

	if err := genesisNode.Start(context.Background()); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to start bootstrap node"})
		return
	}

	registry.RegisterNode(req.ChainID, "genesis", registry.NodeInfo{
		IsGenesis: true,
		Name:      "genesis",
		RPCPort:   func() int { p, _ := strconv.Atoi(config.RPC.ListenAddress[10:]); return p }(),
		P2PPort:   func() int { p, _ := strconv.Atoi(config.P2P.ListenAddress[10:]); return p }(),
	})

	communication.BroadcastEvent(communication.EventChainCreated, map[string]interface{}{
		"chainId":   req.ChainID,
		"timestamp": time.Now(),
	})

	c.JSON(http.StatusCreated, gin.H{
		"message":  "Chain created successfully",
		"chain_id": req.ChainID,
		"genesis_node": map[string]int{
			"p2p_port": func() int { p, _ := strconv.Atoi(config.P2P.ListenAddress[10:]); return p }(),
			"rpc_port": func() int { p, _ := strconv.Atoi(config.RPC.ListenAddress[10:]); return p }(),
		},
	})
}

// AddValidatorToGenesis adds a validator to the genesis file
func AddValidatorToGenesis(chainID string, agent core.Agent) bool {
	dataDir := fmt.Sprintf("./data/%s/%s", chainID, agent.ID)
	genesisFile := fmt.Sprintf("./data/%s/genesis/config/genesis.json", chainID)

	if err := os.MkdirAll(dataDir+"/config", 0755); err != nil {
		return false
	}
	if err := os.MkdirAll(dataDir+"/data", 0755); err != nil {
		return false
	}

	privValKeyFile := fmt.Sprintf("%s/config/priv_validator_key.json", dataDir)
	privValStateFile := fmt.Sprintf("%s/data/priv_validator_state.json", dataDir)
	privVal := privval.GenFilePV(privValKeyFile, privValStateFile)
	pubKey, _ := privVal.GetPubKey()

	genesisBytes, err := os.ReadFile(genesisFile)
	if err != nil {
		return false
	}

	var genDoc types.GenesisDoc
	if err := json.Unmarshal(genesisBytes, &genDoc); err != nil {
		return false
	}

	validator := types.GenesisValidator{
		Address: pubKey.Address(),
		PubKey:  pubKey,
		Power:   10,
		Name:    agent.ID,
	}
	genDoc.Validators = append(genDoc.Validators, validator)

	updatedGenesisBytes, err := json.MarshalIndent(genDoc, "", "  ")
	if err != nil {
		return false
	}

	if err := os.WriteFile(genesisFile, updatedGenesisBytes, 0644); err != nil {
		return false
	}

	newGenesisFile := fmt.Sprintf("%s/config/genesis.json", dataDir)
	if err := os.WriteFile(newGenesisFile, updatedGenesisBytes, 0644); err != nil {
		return false
	}

	return true
}

// GetAllAgents returns all registered agents for a chain
func GetAllAgents(c *gin.Context) {
	chainID := c.GetString("chainID")
	agents := registry.GetAllAgents(chainID)
	c.JSON(http.StatusOK, gin.H{"agents": agents})
}

// GetRegistry returns the registry information for a chain
func GetRegistry(c *gin.Context) {
	chainID := c.GetString("chainID")
	agents := registry.GetAllAgents(chainID)
	c.JSON(http.StatusOK, gin.H{
		"agents":     agents,
		"validators": registry.GetAllValidatorAgentMappings(chainID),
	})
}

package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/Deeptanshu-sankhwar/agentic_consensus/ai"
	"github.com/Deeptanshu-sankhwar/agentic_consensus/api"
	"github.com/Deeptanshu-sankhwar/agentic_consensus/cmd/node"
	"github.com/Deeptanshu-sankhwar/agentic_consensus/core"
	"github.com/Deeptanshu-sankhwar/agentic_consensus/registry"
	cfg "github.com/cometbft/cometbft/config"
	"github.com/cometbft/cometbft/p2p"
	"github.com/cometbft/cometbft/privval"
	"github.com/cometbft/cometbft/types"
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
)

// fileExists checks if a file exists at the given path
func fileExists(filename string) bool {
	_, err := os.Stat(filename)
	return !os.IsNotExist(err)
}

// main initializes and starts the genesis node with all necessary configurations
func main() {
	chainID := flag.String("chain", "mainnet", "Chain ID")
	nodeID := flag.String("node-id", "genesis", "Node ID")
	p2pPort := flag.Int("p2p-port", 26656, "CometBFT P2P port")
	rpcPort := flag.Int("rpc-port", 26657, "CometBFT RPC port")
	apiPort := flag.Int("api-port", 3000, "API server port")
	nats := flag.String("nats", "nats://localhost:4222", "NATS URL")
	flag.Parse()

	registry.InitRegistry()
	ai.InitAI()

	if _, err := os.Stat(fmt.Sprintf("./data/%s", *chainID)); err == nil {
		os.RemoveAll(fmt.Sprintf("./data/%s", *chainID))
		os.Remove("./config/addrbook.json")
		os.RemoveAll("./config/priv_validator_state.json")
	}

	dataDir := fmt.Sprintf("./data/%s/%s", *chainID, *nodeID)
	if err := os.MkdirAll(dataDir+"/data", 0755); err != nil {
		log.Fatalf("Failed to create data directory: %v", err)
	}

	config := cfg.DefaultConfig()
	config.Consensus.TimeoutPropose = 10 * time.Second
	config.Consensus.TimeoutPrevote = 10 * time.Second
	config.Consensus.TimeoutPrecommit = 10 * time.Second
	config.Consensus.TimeoutCommit = 15 * time.Second
	config.BaseConfig.RootDir = dataDir
	config.Moniker = *chainID
	config.P2P.ListenAddress = fmt.Sprintf("tcp://0.0.0.0:%d", *p2pPort)
	config.RPC.ListenAddress = fmt.Sprintf("tcp://0.0.0.0:%d", *rpcPort)

	if err := os.MkdirAll(config.BaseConfig.RootDir+"/config", 0755); err != nil {
		log.Fatalf("Failed to create config directory: %v", err)
	}

	if err := os.MkdirAll(config.BaseConfig.RootDir+"/data", 0755); err != nil {
		log.Fatalf("Failed to create validator data directory: %v", err)
	}

	privValKeyFile := config.PrivValidatorKeyFile()
	privValStateFile := config.PrivValidatorStateFile()
	var privVal *privval.FilePV
	if !fileExists(privValKeyFile) {
		privVal = privval.GenFilePV(privValKeyFile, privValStateFile)
		privVal.Save()
	} else {
		privVal = privval.LoadFilePV(privValKeyFile, privValStateFile)
		if !fileExists(privValStateFile) {
			privVal.Save()
		}
	}

	pubKey, _ := privVal.GetPubKey()

	nodeKeyFile := config.NodeKeyFile()
	if !fileExists(nodeKeyFile) {
		if _, err := p2p.LoadOrGenNodeKey(nodeKeyFile); err != nil {
			log.Fatalf("Failed to generate node key: %v", err)
		}
	}

	genesisFile := config.GenesisFile()
	if err := os.Remove(genesisFile); err != nil && !os.IsNotExist(err) {
		log.Printf("Failed to remove existing genesis file: %v", err)
	}

	if !fileExists(genesisFile) {
		privVal := privval.LoadFilePV(privValKeyFile, privValStateFile)
		pubKey, err := privVal.GetPubKey()
		if err != nil {
			log.Fatalf("Failed to get validator public key: %v", err)
		}

		genValidator := types.GenesisValidator{
			PubKey: pubKey,
			Power:  1000000,
			Name:   "genesis",
		}

		genDoc := types.GenesisDoc{
			ChainID:         *chainID,
			GenesisTime:     time.Now(),
			ConsensusParams: types.DefaultConsensusParams(),
			Validators:      []types.GenesisValidator{genValidator},
		}

		if err := genDoc.ValidateAndComplete(); err != nil {
			log.Fatalf("Failed to validate genesis doc: %v", err)
		}

		if len(genDoc.Validators) == 0 {
			log.Fatalf("No validators in genesis document after validation")
		}

		if err := genDoc.SaveAs(genesisFile); err != nil {
			log.Fatalf("Failed to create genesis file: %v", err)
		}
	}

	config.P2P.AllowDuplicateIP = true
	config.P2P.AddrBookStrict = false
	config.P2P.ExternalAddress = fmt.Sprintf("tcp://127.0.0.1:%d", *p2pPort)
	config.P2P.ListenAddress = fmt.Sprintf("tcp://0.0.0.0:%d", *p2pPort)
	config.P2P.HandshakeTimeout = 20 * time.Second
	config.P2P.DialTimeout = 3 * time.Second
	config.P2P.FlushThrottleTimeout = 10 * time.Millisecond
	config.P2P.MaxNumInboundPeers = 40
	config.P2P.MaxNumOutboundPeers = 10
	config.P2P.SeedMode = true
	config.P2P.PexReactor = true

	genesisNode, err := node.NewNode(config, *chainID, pubKey.Address().String())
	if err != nil {
		log.Fatalf("main: Failed to create node: %v", err)
	}

	err = genesisNode.Start(context.Background())
	if err != nil {
		log.Fatalf("Failed to start node: %v", err)
	}

	registry.RegisterNode(*chainID, *nodeID, registry.NodeInfo{
		IsGenesis: true,
		RPCPort:   *rpcPort,
		P2PPort:   *p2pPort,
		APIPort:   *apiPort,
	})

	core.SetupNATS(*nats)
	defer core.CloseNATS()

	log.Printf("Genesis node for chain %s started with P2P port %d, RPC port %d, and API port %d",
		*chainID, *p2pPort, *rpcPort, *apiPort)

	router := gin.New()
	api.SetupRoutes(router, *chainID)
	log.Fatal(router.Run(fmt.Sprintf(":%d", *apiPort)))

	err = godotenv.Load("../client/.env")
	if err != nil {
		log.Fatal("Error loading .env file")
	}
}

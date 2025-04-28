package core

// Block represents a basic block structure
type Block struct {
	Height    int           `json:"height"`
	PrevHash  string        `json:"prev_hash"`
	Txs       []Transaction `json:"transactions"`
	Timestamp int64         `json:"timestamp"`
	Signature string        `json:"signature"`
	Proposer  string        `json:"proposer"`
	ChainID   string        `json:"chain_id"`
}

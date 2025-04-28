package core

import (
	"context"

	abci "github.com/cometbft/cometbft/abci/types"
)

type TxChecker struct {
	app abci.Application
}

// Creates a new TxChecker instance with the provided ABCI application
func NewTxChecker(app abci.Application) *TxChecker {
	return &TxChecker{app: app}
}

// Checks if a transaction is valid by delegating to the ABCI application
func (tc *TxChecker) CheckTx(ctx context.Context, tx []byte) (*abci.ResponseCheckTx, error) {
	resp := tc.app.CheckTx(abci.RequestCheckTx{
		Tx:   tx,
		Type: abci.CheckTxType_New,
	})
	return &resp, nil
}

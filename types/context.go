package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	db "github.com/tendermint/tm-db"
)

type contextKey uint8

const (
	_        contextKey = iota
	database contextKey = iota
)

func GetCurrentBatch(ctx sdk.Context) db.Batch {
	value := ctx.Value(database)
	if value == nil {
		return nil
	}
	if v, ok := value.(db.Batch); ok {
		return v
	}
	return nil
}
func WithCurrentBatch(ctx sdk.Context, batch db.Batch) sdk.Context {
	return ctx.WithValue(database, batch)
}

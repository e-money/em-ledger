package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	db "github.com/tendermint/tm-db"
)

type contextKey uint8

const (
	_            contextKey = iota
	currentBatch contextKey = iota
	database     contextKey = iota
)

func GetCurrentBatch(ctx sdk.Context) db.Batch {
	v, _ := ctx.Value(currentBatch).(db.Batch)
	return v
}

func GetDatabase(ctx sdk.Context) db.DB {
	v, _ := ctx.Value(database).(db.DB)
	return v
}

func WithCurrentBatch(ctx sdk.Context, batch db.Batch) sdk.Context {
	return ctx.WithValue(currentBatch, batch)
}

// WithDatabase provides access to the database used for state that should be kept out of the app_state.
func WithDatabase(ctx sdk.Context, d db.DB) sdk.Context {
	return ctx.WithValue(database, d)
}

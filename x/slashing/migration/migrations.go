package migration

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	v043 "github.com/cosmos/cosmos-sdk/x/slashing/legacy/v043"
	"github.com/e-money/em-ledger/x/slashing/keeper"
)

// Migrator is a struct for handling in-place store migrations.
type Migrator struct {
	keeper keeper.Keeper
}

// NewMigrator returns a new Migrator.
func NewMigrator(keeper keeper.Keeper) Migrator {
	return Migrator{keeper: keeper}
}

// Migrate1to2 migrates from version 1 to 2.
func (m Migrator) Migrate1to2(ctx sdk.Context) error {
	return v043.MigrateStore(ctx, m.keeper.StoreKey)
}

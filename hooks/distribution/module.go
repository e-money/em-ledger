package distribution

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/module"
	bankkeeper "github.com/cosmos/cosmos-sdk/x/bank/keeper"
	distr "github.com/cosmos/cosmos-sdk/x/distribution"
	abci "github.com/tendermint/tendermint/abci/types"
	db "github.com/tendermint/tm-db"
)

var (
	_ module.AppModule      = AppModule{}
	_ module.AppModuleBasic = AppModuleBasic{}
)

type AppModuleBasic = distr.AppModuleBasic

type AppModule struct {
	distr.AppModule
	k     DistributionKeeper
	ak    AccountKeeper
	bk    bankkeeper.ViewKeeper
	db    db.DB
	batch *db.Batch // a pointer is required a the batch object is hold and modified in app
}

func NewAppModule(nested distr.AppModule, k DistributionKeeper, ak AccountKeeper, bk bankkeeper.ViewKeeper, db db.DB, batch *db.Batch) AppModule {
	return AppModule{
		AppModule: nested,
		k:         k,
		ak:        ak,
		bk:        bk,
		db:        db,
		batch:     batch,
	}
}

func (am AppModule) BeginBlock(ctx sdk.Context, req abci.RequestBeginBlock) {
	BeginBlocker(ctx, req, am.k, am.ak, am.bk, am.db, *am.batch)
}

// todo (reviewer) : IMHO this modules would fit better into x/ than hooks as it contains an alternative/modified impl than adding callbacks

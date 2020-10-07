package queries

import (
	"encoding/json"
	"github.com/gorilla/mux"
	"github.com/spf13/cobra"

	"github.com/e-money/em-ledger/x/queries/types"

	"github.com/cosmos/cosmos-sdk/client/context"
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/module"

	abci "github.com/tendermint/tendermint/abci/types"
)

var (
	_ module.AppModule      = AppModule{}
	_ module.AppModuleBasic = AppModuleBasic{}
)

type (
	AppModuleBasic struct{}

	AppModule struct {
		AppModuleBasic

		AccountKeeper AccountKeeper
	}
)

func NewAppModule(accKeeper AccountKeeper) AppModule {
	return AppModule{
		AppModuleBasic: AppModuleBasic{},
		AccountKeeper:  accKeeper,
	}
}

func (amb AppModuleBasic) Name() string {
	return types.ModuleName
}

func (amb AppModuleBasic) RegisterCodec(_ *codec.Codec) {}

func (amb AppModuleBasic) DefaultGenesis() json.RawMessage { return nil }

func (amb AppModuleBasic) ValidateGenesis(_ json.RawMessage) error { return nil }

func (amb AppModuleBasic) RegisterRESTRoutes(cdc context.CLIContext, router *mux.Router) {
	// TODO
}

func (amb AppModuleBasic) GetTxCmd(_ *codec.Codec) *cobra.Command { return nil }

func (amb AppModuleBasic) GetQueryCmd(cdc *codec.Codec) *cobra.Command {
	// This module contains queries that should be added directly to some of the SDK standard queries.
	return nil
}

func (am AppModule) InitGenesis(_ sdk.Context, _ json.RawMessage) []abci.ValidatorUpdate { return nil }

func (am AppModule) ExportGenesis(_ sdk.Context) json.RawMessage { return nil }

func (am AppModule) RegisterInvariants(_ sdk.InvariantRegistry) {}

func (am AppModule) Route() string {
	return ""
}

func (am AppModule) NewHandler() sdk.Handler {
	return nil
}

func (am AppModule) QuerierRoute() string {
	return types.QuerierRoute
}

func (am AppModule) NewQuerierHandler() sdk.Querier {
	return NewQuerier(am.AccountKeeper)
}

func (am AppModule) BeginBlock(ctx sdk.Context, beginBlock abci.RequestBeginBlock) {}

func (am AppModule) EndBlock(ctx sdk.Context, endBlock abci.RequestEndBlock) []abci.ValidatorUpdate {
	return nil
}

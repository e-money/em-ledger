// This software is Copyright (c) 2019-2020 e-Money A/S. It is not offered under an open source license.
//
// Please contact partners@e-money.com for licensing related questions.

package market

import (
	"encoding/json"

	"github.com/gorilla/mux"
	"github.com/spf13/cobra"

	"github.com/e-money/em-ledger/x/market/client/cli"
	"github.com/e-money/em-ledger/x/market/keeper"
	"github.com/e-money/em-ledger/x/market/types"

	"github.com/cosmos/cosmos-sdk/client/context"
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/module"
	abci "github.com/tendermint/tendermint/abci/types"
)

type (
	AppModuleBasic struct{}

	AppModule struct {
		AppModuleBasic
		k *Keeper
	}
)

var (
	_ module.AppModule      = AppModule{}
	_ module.AppModuleBasic = AppModuleBasic{}
)

func NewAppModule(k *keeper.Keeper) AppModule {
	return AppModule{
		AppModuleBasic: AppModuleBasic{},
		k:              k,
	}
}

func (amb AppModuleBasic) Name() string {
	return ModuleName
}

func (amb AppModuleBasic) RegisterCodec(cdc *codec.Codec) {
	types.RegisterCodec(cdc)
}

func (amb AppModuleBasic) DefaultGenesis() (_ json.RawMessage) {
	return
}

func (amb AppModuleBasic) ValidateGenesis(json.RawMessage) error {
	// TODO
	return nil
}

func (amb AppModuleBasic) RegisterRESTRoutes(context.CLIContext, *mux.Router) {
	// TODO
}

func (amb AppModuleBasic) GetTxCmd(cdc *codec.Codec) *cobra.Command {
	return cli.GetTxCmd(cdc)
}

func (amb AppModuleBasic) GetQueryCmd(*codec.Codec) *cobra.Command {
	// TODO
	return nil
}

func (am AppModule) InitGenesis(sdk.Context, json.RawMessage) (_ []abci.ValidatorUpdate) {
	return
}

func (am AppModule) ExportGenesis(sdk.Context) (_ json.RawMessage) {
	return
}

func (am AppModule) RegisterInvariants(sdk.InvariantRegistry) {}

func (am AppModule) Route() string {
	return RouterKey
}

func (am AppModule) NewHandler() sdk.Handler {
	return NewHandler(am.k)
}

func (am AppModule) QuerierRoute() string {
	return QuerierRoute
}

func (am AppModule) NewQuerierHandler() sdk.Querier {
	return nil
}

func (am AppModule) BeginBlock(ctx sdk.Context, _ abci.RequestBeginBlock) {
	// TODO
}

func (am AppModule) EndBlock(sdk.Context, abci.RequestEndBlock) (_ []abci.ValidatorUpdate) {
	return
}

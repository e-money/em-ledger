// This software is Copyright (c) 2019-2020 e-Money A/S. It is not offered under an open source license.
//
// Please contact partners@e-money.com for licensing related questions.

package authority

import (
	"encoding/json"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/e-money/em-ledger/x/authority/keeper"
	abci "github.com/tendermint/tendermint/abci/types"

	"github.com/e-money/em-ledger/x/authority/client/rest"
	"github.com/e-money/em-ledger/x/authority/types"

	"github.com/cosmos/cosmos-sdk/client/context"
	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/types/module"
	"github.com/gorilla/mux"
	"github.com/spf13/cobra"
)

var (
	_ module.AppModule      = AppModule{}
	_ module.AppModuleBasic = AppModuleBasic{}
)

type AppModuleBasic struct{}

type AppModule struct {
	AppModuleBasic
	keeper Keeper
}

func (amb AppModuleBasic) Name() string { return ModuleName }

func (amb AppModuleBasic) RegisterCodec(cdc *codec.Codec) {
	types.RegisterCodec(cdc)
}

func (amb AppModuleBasic) DefaultGenesis() json.RawMessage {
	return ModuleCdc.MustMarshalJSON(DefaultGenesisState())
}

func (amb AppModuleBasic) ValidateGenesis(json.RawMessage) error {
	return nil
}

func (amb AppModuleBasic) RegisterRESTRoutes(clictx context.CLIContext, r *mux.Router) {
	rest.RegisterQueryRoutes(clictx, r)
}

func (amb AppModuleBasic) GetTxCmd(*codec.Codec) *cobra.Command {
	return nil
}

func (amb AppModuleBasic) GetQueryCmd(cdc *codec.Codec) *cobra.Command {
	return GetQueryCmd(cdc)
}

func NewAppModule(keeper Keeper) *AppModule {
	return &AppModule{
		keeper: keeper,
	}
}

func (am AppModule) InitGenesis(ctx sdk.Context, data json.RawMessage) (_ []abci.ValidatorUpdate) {
	var genesisState GenesisState
	ModuleCdc.MustUnmarshalJSON(data, &genesisState)
	InitGenesis(ctx, am.keeper, genesisState)

	return
}

func (am AppModule) ExportGenesis(ctx sdk.Context) json.RawMessage {
	genesis := GenesisState{
		AuthorityKey:     am.keeper.GetAuthority(ctx),
		RestrictedDenoms: am.keeper.GetRestrictedDenoms(ctx),
		MinGasPrices:     am.keeper.GetGasPrices(ctx),
	}
	return ModuleCdc.MustMarshalJSON(genesis)
}

func (am AppModule) RegisterInvariants(sdk.InvariantRegistry) {}

func (am AppModule) Route() string { return types.RouterKey }

func (am AppModule) QuerierRoute() string { return types.ModuleName }

func (am AppModule) NewHandler() sdk.Handler {
	return newHandler(am.keeper)
}

func (am AppModule) NewQuerierHandler() sdk.Querier {
	return keeper.NewQuerier(am.keeper)
}

func (am AppModule) BeginBlock(sdk.Context, abci.RequestBeginBlock) {}

func (am AppModule) EndBlock(sdk.Context, abci.RequestEndBlock) (_ []abci.ValidatorUpdate) {
	return
}

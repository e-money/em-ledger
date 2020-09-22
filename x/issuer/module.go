// This software is Copyright (c) 2019 e-Money A/S. It is not offered under an open source license.
//
// Please contact partners@e-money.com for licensing related questions.

package issuer

import (
	"encoding/json"
	"github.com/e-money/em-ledger/x/issuer/client/cli"
	"github.com/e-money/em-ledger/x/issuer/keeper"
	"github.com/e-money/em-ledger/x/issuer/types"
	"github.com/gorilla/mux"
	"github.com/spf13/cobra"

	abci "github.com/tendermint/tendermint/abci/types"

	"github.com/cosmos/cosmos-sdk/client/context"
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/module"
)

var (
	_ module.AppModule      = AppModule{}
	_ module.AppModuleBasic = AppModuleBasic{}
)

type AppModuleBasic struct{}

var _ module.AppModuleBasic = AppModuleBasic{}

func (AppModuleBasic) Name() string {
	return ModuleName
}

func (AppModuleBasic) RegisterCodec(cdc *codec.Codec) {
	types.RegisterCodec(cdc)
}

func (AppModuleBasic) DefaultGenesis() json.RawMessage {
	cdc := codec.New()
	return cdc.MustMarshalJSON(defaultGenesisState())
}

func (AppModuleBasic) ValidateGenesis(bz json.RawMessage) error {
	//var data GenesisState
	//err := ModuleCdc.UnmarshalJSON(bz, &data)
	//if err != nil {
	//	return err
	//}
	//return ValidateGenesis(data)
	return nil
}

func (AppModuleBasic) RegisterRESTRoutes(ctx context.CLIContext, rtr *mux.Router) {
	//rest.RegisterRoutes(ctx, rtr)
}

func (AppModuleBasic) GetTxCmd(cdc *codec.Codec) *cobra.Command {
	return cli.GetTxCmd(cdc)
}

func (AppModuleBasic) GetQueryCmd(cdc *codec.Codec) *cobra.Command {
	return cli.GetQueryCmd(types.StoreKey, cdc)
}

type AppModule struct {
	AppModuleBasic
	keeper keeper.Keeper
}

func NewAppModule(keeper keeper.Keeper) AppModule {
	return AppModule{
		AppModuleBasic: AppModuleBasic{},
		keeper:         keeper,
	}
}

func (AppModule) Name() string {
	return ModuleName
}

func (am AppModule) RegisterInvariants(_ sdk.InvariantRegistry) {}

func (AppModule) Route() string { return types.RouterKey }

func (am AppModule) NewHandler() sdk.Handler {
	return newHandler(am.keeper)
}

func (AppModule) QuerierRoute() string {
	return ModuleName
}

func (am AppModule) NewQuerierHandler() sdk.Querier {
	return keeper.NewQuerier(am.keeper)
}

func (am AppModule) InitGenesis(ctx sdk.Context, data json.RawMessage) []abci.ValidatorUpdate {
	var genesisState genesisState
	ModuleCdc.MustUnmarshalJSON(data, &genesisState)
	initGenesis(ctx, am.keeper, genesisState)
	return []abci.ValidatorUpdate{}
}

func (am AppModule) ExportGenesis(ctx sdk.Context) json.RawMessage {
	issuers := am.keeper.GetIssuers(ctx)
	gs := genesisState{issuers}
	return ModuleCdc.MustMarshalJSON(gs)
}

func (am AppModule) BeginBlock(_ sdk.Context, _ abci.RequestBeginBlock) {}

func (AppModule) EndBlock(_ sdk.Context, _ abci.RequestEndBlock) (_ []abci.ValidatorUpdate) {
	return
}

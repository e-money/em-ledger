package authority

import (
	"encoding/json"
	sdk "github.com/cosmos/cosmos-sdk/types"
	abci "github.com/tendermint/tendermint/abci/types"

	"emoney/x/authority/types"

	"github.com/cosmos/cosmos-sdk/client/context"
	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/types/module"
	"github.com/gorilla/mux"
	"github.com/spf13/cobra"
)

var _ module.AppModule = AppModule{}

type AppModule struct {
	keeper Keeper
}

func NewAppModule(keeper Keeper) *AppModule {
	return &AppModule{
		keeper: keeper,
	}
}

func (am AppModule) InitGenesis(sdk.Context, json.RawMessage) (_ []abci.ValidatorUpdate) {
	// TODO
	return
}

func (am AppModule) ExportGenesis(sdk.Context) json.RawMessage {
	// TODO
	return ModuleCdc.MustMarshalJSON(GenesisState{})
}

func (am AppModule) RegisterInvariants(sdk.InvariantRegistry) {}

func (am AppModule) Route() string { return types.RouterKey }

func (am AppModule) QuerierRoute() string { return types.ModuleName }

func (am AppModule) NewHandler() sdk.Handler {
	return newHandler(am.keeper)
}

func (am AppModule) NewQuerierHandler() sdk.Querier {
	return nil
}

func (am AppModule) BeginBlock(sdk.Context, abci.RequestBeginBlock) {}

func (am AppModule) EndBlock(sdk.Context, abci.RequestEndBlock) (_ []abci.ValidatorUpdate) {
	return
}

func (am AppModule) Name() string { return ModuleName }

func (am AppModule) RegisterCodec(cdc *codec.Codec) {
	types.RegisterCodec(cdc)
}

func (am AppModule) DefaultGenesis() json.RawMessage {
	return ModuleCdc.MustMarshalJSON(DefaultGenesisState())
}

func (am AppModule) ValidateGenesis(json.RawMessage) error {
	return nil
}

func (am AppModule) RegisterRESTRoutes(context.CLIContext, *mux.Router) {}

func (am AppModule) GetTxCmd(*codec.Codec) *cobra.Command {
	return nil
}

func (am AppModule) GetQueryCmd(*codec.Codec) *cobra.Command {
	return nil
}

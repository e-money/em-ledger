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
var _ module.AppModuleBasic = AppModuleBasic{}

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

func (amb AppModuleBasic) RegisterRESTRoutes(context.CLIContext, *mux.Router) {

}

func (amb AppModuleBasic) GetTxCmd(*codec.Codec) *cobra.Command {
	return nil
}

func (amb AppModuleBasic) GetQueryCmd(*codec.Codec) *cobra.Command {
	return nil
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
		AuthorityKey: am.keeper.GetAuthority(ctx),
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
	return nil
}

func (am AppModule) BeginBlock(sdk.Context, abci.RequestBeginBlock) {}

func (am AppModule) EndBlock(sdk.Context, abci.RequestEndBlock) (_ []abci.ValidatorUpdate) {
	return
}

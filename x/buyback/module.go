package buyback

import (
	"encoding/json"
	"github.com/cosmos/cosmos-sdk/client/context"
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/module"
	"github.com/gorilla/mux"
	"github.com/spf13/cobra"
	abci "github.com/tendermint/tendermint/abci/types"
)

var _ module.AppModule = AppModule{}
var _ module.AppModuleBasic = AppModuleBasic{}

type AppModuleBasic struct{}

type AppModule struct {
	AppModuleBasic
}

func (amb AppModuleBasic) Name() string {
	return ModuleName
}

func (amb AppModuleBasic) RegisterCodec(*codec.Codec) {
	panic("implement me")
}

func (amb AppModuleBasic) DefaultGenesis() json.RawMessage {
	panic("implement me")
}

func (amb AppModuleBasic) ValidateGenesis(json.RawMessage) error {
	panic("implement me")
}

func (amb AppModuleBasic) RegisterRESTRoutes(context.CLIContext, *mux.Router) {
	panic("implement me")
}

func (amb AppModuleBasic) GetTxCmd(*codec.Codec) *cobra.Command {
	panic("implement me")
}

func (amb AppModuleBasic) GetQueryCmd(*codec.Codec) *cobra.Command {
	panic("implement me")
}

func (am AppModule) InitGenesis(sdk.Context, json.RawMessage) []abci.ValidatorUpdate {
	panic("implement me")
}

func (am AppModule) ExportGenesis(sdk.Context) json.RawMessage {
	panic("implement me")
}

func (am AppModule) RegisterInvariants(sdk.InvariantRegistry) {}

func (am AppModule) Route() string {
	panic("implement me")
}

func (am AppModule) NewHandler() sdk.Handler {
	return nil
}

func (am AppModule) QuerierRoute() string {
	return QuerierRoute
}

func (am AppModule) NewQuerierHandler() sdk.Querier {
	return nil
}

func (am AppModule) BeginBlock(sdk.Context, abci.RequestBeginBlock) {

}

func (am AppModule) EndBlock(sdk.Context, abci.RequestEndBlock) (_ []abci.ValidatorUpdate) {
	return
}

package buyback

import (
	"encoding/json"
	"github.com/cosmos/cosmos-sdk/client/context"
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/module"
	"github.com/e-money/em-ledger/x/buyback/client/cli"
	"github.com/e-money/em-ledger/x/buyback/client/rest"
	"github.com/e-money/em-ledger/x/buyback/internal/keeper"
	"github.com/gorilla/mux"
	"github.com/spf13/cobra"
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

		keeper keeper.Keeper
	}
)

func NewAppModule(k keeper.Keeper) AppModule {
	return AppModule{
		AppModuleBasic: AppModuleBasic{},
		keeper:         k,
	}
}

func (amb AppModuleBasic) Name() string {
	return ModuleName
}

func (amb AppModuleBasic) RegisterCodec(*codec.Codec) {}

func (amb AppModuleBasic) DefaultGenesis() json.RawMessage {
	return nil
}

func (amb AppModuleBasic) ValidateGenesis(json.RawMessage) error {
	return nil
}

func (amb AppModuleBasic) RegisterRESTRoutes(ctx context.CLIContext, r *mux.Router) {
	rest.RegisterQueryRoutes(ctx, r)
}

func (amb AppModuleBasic) GetTxCmd(*codec.Codec) *cobra.Command {
	return nil
}

func (amb AppModuleBasic) GetQueryCmd(cdc *codec.Codec) *cobra.Command {
	return cli.GetQueryCmd(cdc)
}

func (am AppModule) InitGenesis(sdk.Context, json.RawMessage) []abci.ValidatorUpdate {
	return nil
}

func (am AppModule) ExportGenesis(sdk.Context) json.RawMessage {
	return nil
}

func (am AppModule) RegisterInvariants(sdk.InvariantRegistry) {}

func (am AppModule) Route() string {
	return QuerierRoute
}

func (am AppModule) NewHandler() sdk.Handler {
	return nil
}

func (am AppModule) QuerierRoute() string {
	return QuerierRoute
}

func (am AppModule) NewQuerierHandler() sdk.Querier {
	return keeper.NewQuerier(am.keeper)
}

func (am AppModule) BeginBlock(sdk.Context, abci.RequestBeginBlock) {}

func (am AppModule) EndBlock(sdk.Context, abci.RequestEndBlock) (_ []abci.ValidatorUpdate) {
	return
}

package slashing

import (
	"context"
	"encoding/json"
	"fmt"
	sdkslashing "github.com/cosmos/cosmos-sdk/x/slashing"
	"github.com/e-money/em-ledger/x/slashing/keeper"
	"github.com/e-money/em-ledger/x/slashing/types"
	"github.com/grpc-ecosystem/grpc-gateway/runtime"

	"github.com/gorilla/mux"
	"github.com/spf13/cobra"

	abci "github.com/tendermint/tendermint/abci/types"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/codec"
	cdctypes "github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/module"
	"github.com/cosmos/cosmos-sdk/x/slashing/client/cli"
	"github.com/cosmos/cosmos-sdk/x/slashing/client/rest"
	sdkslashingkeeper "github.com/cosmos/cosmos-sdk/x/slashing/keeper"
	sdkslashingtypes "github.com/cosmos/cosmos-sdk/x/slashing/types"
	sdkstakingkeeper "github.com/cosmos/cosmos-sdk/x/staking/keeper"
)

var (
	_ module.AppModule      = AppModule{}
	_ module.AppModuleBasic = AppModuleBasic{}
)

// AppModuleBasic defines the basic application module used by the slashing module.
type AppModuleBasic struct {
	cdc codec.Marshaler
}

var _ module.AppModuleBasic = AppModuleBasic{}

// Name returns the slashing module's name.
func (AppModuleBasic) Name() string {
	return types.ModuleName
}

// RegisterLegacyAminoCodec registers the slashing module's types for the given codec.
func (AppModuleBasic) RegisterLegacyAminoCodec(cdc *codec.LegacyAmino) {
	sdkslashingtypes.RegisterLegacyAminoCodec(cdc)
}

// RegisterInterfaces registers the module's interface types
func (b AppModuleBasic) RegisterInterfaces(registry cdctypes.InterfaceRegistry) {
	sdkslashingtypes.RegisterInterfaces(registry)
}

// DefaultGenesis returns default genesis state as raw bytes for the slashing
// module.
func (AppModuleBasic) DefaultGenesis(cdc codec.JSONMarshaler) json.RawMessage {
	genesis := &sdkslashingtypes.GenesisState{
		Params:       types.DefaultParams(),
		SigningInfos: []sdkslashingtypes.SigningInfo{},
		MissedBlocks: []sdkslashingtypes.ValidatorMissedBlocks{},
	}
	return cdc.MustMarshalJSON(genesis)
}

// ValidateGenesis performs genesis state validation for the slashing module.
func (AppModuleBasic) ValidateGenesis(cdc codec.JSONMarshaler, config client.TxEncodingConfig, bz json.RawMessage) error {
	var data sdkslashingtypes.GenesisState
	if err := cdc.UnmarshalJSON(bz, &data); err != nil {
		return fmt.Errorf("failed to unmarshal %s genesis state: %w", sdkslashingtypes.ModuleName, err)
	}

	return sdkslashingtypes.ValidateGenesis(data)
}

// RegisterRESTRoutes registers the REST routes for the slashing module.
func (AppModuleBasic) RegisterRESTRoutes(clientCtx client.Context, rtr *mux.Router) {
	rest.RegisterHandlers(clientCtx, rtr)
}

// RegisterGRPCGatewayRoutes registers the gRPC Gateway routes for the slashig module.
func (AppModuleBasic) RegisterGRPCGatewayRoutes(clientCtx client.Context, mux *runtime.ServeMux) {
	sdkslashingtypes.RegisterQueryHandlerClient(context.Background(), mux, sdkslashingtypes.NewQueryClient(clientCtx))
}

// GetTxCmd returns the root tx command for the slashing module.
func (AppModuleBasic) GetTxCmd() *cobra.Command {
	return cli.NewTxCmd()
}

// GetQueryCmd returns no root query command for the slashing module.
func (AppModuleBasic) GetQueryCmd() *cobra.Command {
	return cli.GetQueryCmd()
}

//____________________________________________________________________________

// AppModule implements an application module for the slashing module.
type AppModule struct {
	AppModuleBasic

	keeper        keeper.Keeper
	accountKeeper sdkslashingtypes.AccountKeeper
	bankKeeper    sdkslashingtypes.BankKeeper
	stakingKeeper sdkstakingkeeper.Keeper
}

// NewAppModule creates a new AppModule object
func NewAppModule(cdc codec.Marshaler, keeper keeper.Keeper, ak sdkslashingtypes.AccountKeeper, bk sdkslashingtypes.BankKeeper, sk sdkstakingkeeper.Keeper) AppModule {
	return AppModule{
		AppModuleBasic: AppModuleBasic{cdc: cdc},
		keeper:         keeper,
		accountKeeper:  ak,
		bankKeeper:     bk,
		stakingKeeper:  sk,
	}
}

// Name returns the slashing module's name.
func (AppModule) Name() string {
	return types.ModuleName
}

// RegisterInvariants registers the slashing module invariants.
func (am AppModule) RegisterInvariants(_ sdk.InvariantRegistry) {}

// Route returns the message routing key for the slashing module.
func (am AppModule) Route() sdk.Route {
	return sdk.NewRoute(types.RouterKey, sdkslashing.NewHandler(am.keeper.Keeper))
}

// QuerierRoute returns the slashing module's querier route name.
func (AppModule) QuerierRoute() string {
	return types.QuerierRoute
}

// LegacyQuerierHandler returns the slashing module sdk.Querier.
func (am AppModule) LegacyQuerierHandler(legacyQuerierCdc *codec.LegacyAmino) sdk.Querier {
	return sdkslashingkeeper.NewQuerier(am.keeper.Keeper, legacyQuerierCdc)
}

// RegisterServices registers module services.
func (am AppModule) RegisterServices(cfg module.Configurator) {
	sdkslashingtypes.RegisterMsgServer(cfg.MsgServer(), sdkslashingkeeper.NewMsgServerImpl(am.keeper.Keeper))
	sdkslashingtypes.RegisterQueryServer(cfg.QueryServer(), am.keeper)
}

// InitGenesis performs genesis initialization for the slashing module. It returns
// no validator updates.
func (am AppModule) InitGenesis(ctx sdk.Context, cdc codec.JSONMarshaler, data json.RawMessage) []abci.ValidatorUpdate {
	var genesisState sdkslashingtypes.GenesisState
	cdc.MustUnmarshalJSON(data, &genesisState)
	sdkslashing.InitGenesis(ctx, am.keeper.Keeper, am.stakingKeeper, &genesisState)
	return []abci.ValidatorUpdate{}
}

// ExportGenesis returns the exported genesis state as raw bytes for the slashing
// module.
func (am AppModule) ExportGenesis(ctx sdk.Context, cdc codec.JSONMarshaler) json.RawMessage {
	gs := sdkslashing.ExportGenesis(ctx, am.keeper.Keeper)
	return cdc.MustMarshalJSON(gs)
}

// BeginBlock returns the begin blocker for the slashing module.
func (am AppModule) BeginBlock(ctx sdk.Context, req abci.RequestBeginBlock) {
	keeper.BeginBlocker(ctx, req, am.keeper)
}

// EndBlock returns the end blocker for the slashing module. It returns no validator
// updates.
func (AppModule) EndBlock(_ sdk.Context, _ abci.RequestEndBlock) []abci.ValidatorUpdate {
	return []abci.ValidatorUpdate{}
}

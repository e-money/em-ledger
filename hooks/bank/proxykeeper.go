// This software is Copyright (c) 2019-2020 e-Money A/S. It is not offered under an open source license.
//
// Please contact partners@e-money.com for licensing related questions.

package bank

import (
	"context"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/bank/exported"
	bankkeeper "github.com/cosmos/cosmos-sdk/x/bank/keeper"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
)

var _ bankkeeper.Keeper = (*ProxyKeeper)(nil)

type ProxyKeeper struct {
	bk        bankkeeper.Keeper
	listeners []func(sdk.Context, []sdk.AccAddress)
}

func Wrap(bk bankkeeper.Keeper) *ProxyKeeper {
	return &ProxyKeeper{bk: bk}
}

func (pk *ProxyKeeper) AddBalanceListener(l func(sdk.Context, []sdk.AccAddress)) {
	pk.listeners = append(pk.listeners, l)
}

func (pk ProxyKeeper) notifyListeners(ctx sdk.Context, accounts ...sdk.AccAddress) {
	accounts = deduplicate(accounts)
	for _, l := range pk.listeners {
		l(ctx, accounts)
	}
}

func deduplicate(accounts []sdk.AccAddress) []sdk.AccAddress {
	idx := make(map[string]struct{}, len(accounts))
	r := make([]sdk.AccAddress, 0, len(accounts))
	for _, a := range accounts {
		if _, exists := idx[string(a)]; exists {
			continue
		}
		r = append(r, a)
		idx[string(a)] = struct{}{}
	}
	return r
}

func (pk ProxyKeeper) InputOutputCoins(ctx sdk.Context, inputs []banktypes.Input, outputs []banktypes.Output) error {
	err := pk.bk.InputOutputCoins(ctx, inputs, outputs)
	if err != nil {
		return err
	}

	accounts := make([]sdk.AccAddress, 0, len(inputs)+len(outputs))
	for _, a := range inputs {
		// invalid addresses were handled before in the wrapped keeper
		addr, _ := sdk.AccAddressFromBech32(a.Address)
		accounts = append(accounts, addr)
	}
	for _, a := range outputs {
		addr, _ := sdk.AccAddressFromBech32(a.Address)
		accounts = append(accounts, addr)
	}

	pk.notifyListeners(ctx, accounts...)
	return nil
}

func (pk ProxyKeeper) SendCoins(ctx sdk.Context, fromAddr sdk.AccAddress, toAddr sdk.AccAddress, amt sdk.Coins) error {
	err := pk.bk.SendCoins(ctx, fromAddr, toAddr, amt)
	if err != nil {
		return err
	}
	pk.notifyListeners(ctx, fromAddr, toAddr)
	return nil
}

func (pk ProxyKeeper) ValidateBalance(ctx sdk.Context, addr sdk.AccAddress) error {
	return pk.bk.ValidateBalance(ctx, addr)
}

func (pk ProxyKeeper) HasBalance(ctx sdk.Context, addr sdk.AccAddress, amt sdk.Coin) bool {
	return pk.bk.HasBalance(ctx, addr, amt)
}

func (pk ProxyKeeper) GetAllBalances(ctx sdk.Context, addr sdk.AccAddress) sdk.Coins {
	return pk.bk.GetAllBalances(ctx, addr)
}

func (pk ProxyKeeper) GetAccountsBalances(ctx sdk.Context) []banktypes.Balance {
	return pk.bk.GetAccountsBalances(ctx)
}

func (pk ProxyKeeper) GetBalance(ctx sdk.Context, addr sdk.AccAddress, denom string) sdk.Coin {
	return pk.bk.GetBalance(ctx, addr, denom)
}

func (pk ProxyKeeper) LockedCoins(ctx sdk.Context, addr sdk.AccAddress) sdk.Coins {
	return pk.bk.LockedCoins(ctx, addr)
}

func (pk ProxyKeeper) SpendableCoins(ctx sdk.Context, addr sdk.AccAddress) sdk.Coins {
	return pk.bk.SpendableCoins(ctx, addr)
}

func (pk ProxyKeeper) IterateAccountBalances(ctx sdk.Context, addr sdk.AccAddress, cb func(coin sdk.Coin) (stop bool)) {
	pk.bk.IterateAccountBalances(ctx, addr, cb)
}

func (pk ProxyKeeper) IterateAllBalances(ctx sdk.Context, cb func(address sdk.AccAddress, coin sdk.Coin) (stop bool)) {
	pk.bk.IterateAllBalances(ctx, cb)
}

func (pk ProxyKeeper) GetParams(ctx sdk.Context) banktypes.Params {
	return pk.bk.GetParams(ctx)
}

func (pk ProxyKeeper) SendEnabledCoin(ctx sdk.Context, coin sdk.Coin) bool {
	return pk.bk.SendEnabledCoin(ctx, coin)
}

func (pk ProxyKeeper) IsSendEnabledCoins(ctx sdk.Context, coins ...sdk.Coin) error {
	return pk.bk.IsSendEnabledCoins(ctx, coins...)
}

func (pk ProxyKeeper) BlockedAddr(addr sdk.AccAddress) bool {
	return pk.bk.BlockedAddr(addr)
}

func (pk ProxyKeeper) SetParams(ctx sdk.Context, params banktypes.Params) {
	pk.bk.SetParams(ctx, params)
}

func (pk ProxyKeeper) GetSupply(ctx sdk.Context, denom string) sdk.Coin {
	return pk.bk.GetSupply(ctx, denom)
}

func (pk *ProxyKeeper) InitGenesis(ctx sdk.Context, state *banktypes.GenesisState) {
	pk.bk.InitGenesis(ctx, state)
}

func (pk *ProxyKeeper) ExportGenesis(ctx sdk.Context) *banktypes.GenesisState {
	return pk.bk.ExportGenesis(ctx)
}

func (pk *ProxyKeeper) MintCoins(ctx sdk.Context, moduleName string, amounts sdk.Coins) error {
	return pk.bk.MintCoins(ctx, moduleName, amounts)
}

func (pk *ProxyKeeper) GetDenomMetaData(ctx sdk.Context, denom string) (banktypes.Metadata, bool) {
	return pk.bk.GetDenomMetaData(ctx, denom)
}

func (pk *ProxyKeeper) SetDenomMetaData(ctx sdk.Context, denomMetaData banktypes.Metadata) {
	pk.bk.SetDenomMetaData(ctx, denomMetaData)
}

func (pk *ProxyKeeper) IterateAllDenomMetaData(ctx sdk.Context, cb func(banktypes.Metadata) bool) {
	pk.bk.IterateAllDenomMetaData(ctx, cb)
}

func (pk *ProxyKeeper) SendCoinsFromModuleToAccount(ctx sdk.Context, senderModule string, recipientAddr sdk.AccAddress, amt sdk.Coins) error {
	err := pk.bk.SendCoinsFromModuleToAccount(ctx, senderModule, recipientAddr, amt)
	if err != nil {
		return err
	}
	pk.notifyListeners(ctx, recipientAddr)
	return nil
}

func (pk *ProxyKeeper) SendCoinsFromModuleToModule(ctx sdk.Context, senderModule, recipientModule string, amt sdk.Coins) error {
	return pk.bk.SendCoinsFromModuleToModule(ctx, senderModule, recipientModule, amt)
}

func (pk *ProxyKeeper) SendCoinsFromAccountToModule(ctx sdk.Context, senderAddr sdk.AccAddress, recipientModule string, amt sdk.Coins) error {
	err := pk.bk.SendCoinsFromAccountToModule(ctx, senderAddr, recipientModule, amt)
	if err != nil {
		return err
	}
	pk.notifyListeners(ctx, senderAddr)
	return nil
}

func (pk *ProxyKeeper) DelegateCoinsFromAccountToModule(ctx sdk.Context, senderAddr sdk.AccAddress, recipientModule string, amt sdk.Coins) error {
	err := pk.bk.DelegateCoinsFromAccountToModule(ctx, senderAddr, recipientModule, amt)
	if err != nil {
		return err
	}
	pk.notifyListeners(ctx, senderAddr)
	return nil
}

func (pk *ProxyKeeper) UndelegateCoinsFromModuleToAccount(ctx sdk.Context, senderModule string, recipientAddr sdk.AccAddress, amt sdk.Coins) error {
	err := pk.bk.UndelegateCoinsFromModuleToAccount(ctx, senderModule, recipientAddr, amt)
	if err != nil {
		return err
	}
	pk.notifyListeners(ctx, recipientAddr)
	return nil
}

func (pk *ProxyKeeper) BurnCoins(ctx sdk.Context, moduleName string, amt sdk.Coins) error {
	return pk.bk.BurnCoins(ctx, moduleName, amt)
}

func (pk *ProxyKeeper) DelegateCoins(ctx sdk.Context, delegatorAddr, moduleAccAddr sdk.AccAddress, amt sdk.Coins) error {
	err := pk.bk.DelegateCoins(ctx, delegatorAddr, moduleAccAddr, amt)
	if err != nil {
		return err
	}
	pk.notifyListeners(ctx, delegatorAddr)
	return nil
}

func (pk *ProxyKeeper) UndelegateCoins(ctx sdk.Context, moduleAccAddr, delegatorAddr sdk.AccAddress, amt sdk.Coins) error {
	err := pk.bk.UndelegateCoins(ctx, moduleAccAddr, delegatorAddr, amt)
	if err != nil {
		return err
	}
	pk.notifyListeners(ctx, delegatorAddr)
	return nil
}

func (pk *ProxyKeeper) MarshalSupply(supplyI exported.SupplyI) ([]byte, error) {
	return pk.bk.MarshalSupply(supplyI)
}

func (pk *ProxyKeeper) UnmarshalSupply(bz []byte) (exported.SupplyI, error) {
	return pk.bk.UnmarshalSupply(bz)
}

func (pk *ProxyKeeper) Balance(ctx context.Context, request *banktypes.QueryBalanceRequest) (*banktypes.QueryBalanceResponse, error) {
	return pk.bk.Balance(ctx, request)
}

func (pk *ProxyKeeper) AllBalances(ctx context.Context, request *banktypes.QueryAllBalancesRequest) (*banktypes.QueryAllBalancesResponse, error) {
	return pk.bk.AllBalances(ctx, request)
}

func (pk *ProxyKeeper) TotalSupply(ctx context.Context, request *banktypes.QueryTotalSupplyRequest) (*banktypes.QueryTotalSupplyResponse, error) {
	return pk.bk.TotalSupply(ctx, request)
}

func (pk *ProxyKeeper) SupplyOf(ctx context.Context, request *banktypes.QuerySupplyOfRequest) (*banktypes.QuerySupplyOfResponse, error) {
	return pk.bk.SupplyOf(ctx, request)
}

func (pk *ProxyKeeper) Params(ctx context.Context, request *banktypes.QueryParamsRequest) (*banktypes.QueryParamsResponse, error) {
	return pk.bk.Params(ctx, request)
}

func (pk *ProxyKeeper) DenomMetadata(ctx context.Context, request *banktypes.QueryDenomMetadataRequest) (*banktypes.QueryDenomMetadataResponse, error) {
	return pk.bk.DenomMetadata(ctx, request)
}

func (pk *ProxyKeeper) DenomsMetadata(ctx context.Context, request *banktypes.QueryDenomsMetadataRequest) (*banktypes.QueryDenomsMetadataResponse, error) {
	return pk.bk.DenomsMetadata(ctx, request)
}

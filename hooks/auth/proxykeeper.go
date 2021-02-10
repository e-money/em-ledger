// This software is Copyright (c) 2019-2020 e-Money A/S. It is not offered under an open source license.
//
// Please contact partners@e-money.com for licensing related questions.

package auth

import (
	cryptotypes "github.com/cosmos/cosmos-sdk/crypto/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	authkeeper "github.com/cosmos/cosmos-sdk/x/auth/keeper"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
)

var _ authkeeper.AccountKeeperI = (*ProxyKeeper)(nil)
var _ banktypes.AccountKeeper = (*ProxyKeeper)(nil)

type ProxyKeeper struct {
	ak        authkeeper.AccountKeeper
	listeners []func(sdk.Context, authtypes.AccountI)
}

// Wraps the auth's modules keeper in order to emit events when an account's balance changes.
func Wrap(ak authkeeper.AccountKeeper) *ProxyKeeper {
	return &ProxyKeeper{
		ak: ak,
	}
}

func (pk *ProxyKeeper) AddAccountListener(l func(sdk.Context, authtypes.AccountI)) {
	pk.listeners = append(pk.listeners, l)
}

func (pk ProxyKeeper) notifyListeners(ctx sdk.Context, acc authtypes.AccountI) {
	for _, l := range pk.listeners {
		l(ctx, acc)
	}
}

func (pk *ProxyKeeper) SetAccount(ctx sdk.Context, acc authtypes.AccountI) {
	pk.ak.SetAccount(ctx, acc)
	pk.notifyListeners(ctx, acc)
}

func (pk *ProxyKeeper) NewAccountWithAddress(ctx sdk.Context, address sdk.AccAddress) authtypes.AccountI {
	return pk.ak.NewAccountWithAddress(ctx, address)
}

func (pk *ProxyKeeper) NewAccount(ctx sdk.Context, i authtypes.AccountI) authtypes.AccountI {
	return pk.ak.NewAccount(ctx, i)
}

func (pk *ProxyKeeper) GetAccount(ctx sdk.Context, address sdk.AccAddress) authtypes.AccountI {
	return pk.ak.GetAccount(ctx, address)
}

func (pk *ProxyKeeper) RemoveAccount(ctx sdk.Context, i authtypes.AccountI) {
	pk.ak.RemoveAccount(ctx, i)
}

func (pk *ProxyKeeper) IterateAccounts(ctx sdk.Context, f func(authtypes.AccountI) bool) {
	pk.ak.IterateAccounts(ctx, f)
}

func (pk *ProxyKeeper) GetPubKey(ctx sdk.Context, address sdk.AccAddress) (cryptotypes.PubKey, error) {
	return pk.ak.GetPubKey(ctx, address)
}

func (pk *ProxyKeeper) GetSequence(ctx sdk.Context, address sdk.AccAddress) (uint64, error) {
	return pk.ak.GetSequence(ctx, address)
}

func (pk *ProxyKeeper) GetNextAccountNumber(ctx sdk.Context) uint64 {
	return pk.ak.GetNextAccountNumber(ctx)
}

func (pk *ProxyKeeper) GetAllAccounts(ctx sdk.Context) []authtypes.AccountI {
	return pk.ak.GetAllAccounts(ctx)
}

func (pk *ProxyKeeper) ValidatePermissions(macc authtypes.ModuleAccountI) error {
	return pk.ak.ValidatePermissions(macc)
}

func (pk *ProxyKeeper) GetModuleAddress(moduleName string) sdk.AccAddress {
	return pk.ak.GetModuleAddress(moduleName)
}

func (pk *ProxyKeeper) GetModuleAddressAndPermissions(moduleName string) (addr sdk.AccAddress, permissions []string) {
	return pk.ak.GetModuleAddressAndPermissions(moduleName)
}

func (pk *ProxyKeeper) GetModuleAccountAndPermissions(ctx sdk.Context, moduleName string) (authtypes.ModuleAccountI, []string) {
	return pk.ak.GetModuleAccountAndPermissions(ctx, moduleName)
}

func (pk *ProxyKeeper) GetModuleAccount(ctx sdk.Context, moduleName string) authtypes.ModuleAccountI {
	return pk.ak.GetModuleAccount(ctx, moduleName)
}

func (pk *ProxyKeeper) SetModuleAccount(ctx sdk.Context, macc authtypes.ModuleAccountI) {
	pk.ak.SetModuleAccount(ctx, macc)
}

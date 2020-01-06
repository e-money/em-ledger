// This software is Copyright (c) 2019-2020 e-Money A/S. It is not offered under an open source license.
//
// Please contact partners@e-money.com for licensing related questions.

package auth

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/auth"
	"github.com/cosmos/cosmos-sdk/x/auth/exported"
)

// Wraps the auth's modules keeper in order to emit events when an account's balance changes.

// Expected interface is copied from bank module
type AccountKeeper interface {
	NewAccountWithAddress(ctx sdk.Context, addr sdk.AccAddress) exported.Account
	NewAccount(sdk.Context, exported.Account) exported.Account

	GetAccount(ctx sdk.Context, addr sdk.AccAddress) exported.Account
	GetAllAccounts(ctx sdk.Context) []exported.Account
	SetAccount(ctx sdk.Context, acc exported.Account)

	IterateAccounts(ctx sdk.Context, process func(exported.Account) bool)

	InnerKeeper() auth.AccountKeeper
	AddAccountListener(func(sdk.Context, exported.Account))
}

var _ AccountKeeper = (*ProxyKeeper)(nil)

func Wrap(ak auth.AccountKeeper) *ProxyKeeper {
	return &ProxyKeeper{
		ak: ak,
	}
}

type ProxyKeeper struct {
	ak        auth.AccountKeeper
	listeners []func(sdk.Context, exported.Account)
}

func (pk *ProxyKeeper) AddAccountListener(l func(sdk.Context, exported.Account)) {
	pk.listeners = append(pk.listeners, l)
}

func (pk ProxyKeeper) InnerKeeper() auth.AccountKeeper {
	return pk.ak
}

func (pk ProxyKeeper) NewAccount(ctx sdk.Context, acc exported.Account) exported.Account {
	return pk.ak.NewAccount(ctx, acc)
}

func (pk ProxyKeeper) NewAccountWithAddress(ctx sdk.Context, addr sdk.AccAddress) exported.Account {
	return pk.ak.NewAccountWithAddress(ctx, addr)
}

func (pk ProxyKeeper) GetAccount(ctx sdk.Context, addr sdk.AccAddress) exported.Account {
	return pk.ak.GetAccount(ctx, addr)
}

func (pk ProxyKeeper) GetAllAccounts(ctx sdk.Context) []exported.Account {
	return pk.ak.GetAllAccounts(ctx)
}

func (pk ProxyKeeper) SetAccount(ctx sdk.Context, acc exported.Account) {
	pk.ak.SetAccount(ctx, acc)
	pk.notifyListeners(ctx, acc)
}

func (pk ProxyKeeper) IterateAccounts(ctx sdk.Context, process func(exported.Account) bool) {
	pk.ak.IterateAccounts(ctx, process)
}

func (pk ProxyKeeper) notifyListeners(ctx sdk.Context, acc exported.Account) {
	for _, l := range pk.listeners {
		l(ctx, acc)
	}
}

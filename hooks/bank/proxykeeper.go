// This software is Copyright (c) 2019-2020 e-Money A/S. It is not offered under an open source license.
//
// Please contact partners@e-money.com for licensing related questions.

package bank

import (
	"github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	bankkeeper "github.com/cosmos/cosmos-sdk/x/bank/keeper"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
)

var _ bankkeeper.SendKeeper = (*ProxyKeeper)(nil)

type ProxyKeeper struct {
	bk bankkeeper.SendKeeper
	rk RestrictedKeeper
}

func Wrap(bk bankkeeper.Keeper, rk RestrictedKeeper) ProxyKeeper {
	return ProxyKeeper{bk: bk, rk: rk}
}

func (pk ProxyKeeper) InputOutputCoins(ctx types.Context, inputs []banktypes.Input, outputs []banktypes.Output) error {
	restrictedDenoms := pk.rk.GetRestrictedDenoms(ctx)
	// Multisend does not support restricted denominations.
	for _, input := range inputs {
		for _, coin := range input.Coins {
			if _, found := restrictedDenoms.Find(coin.Denom); found {
				return sdkerrors.Wrap(ErrRestrictedDenomination, coin.Denom)
			}
		}
	}

	return pk.bk.InputOutputCoins(ctx, inputs, outputs)
}

func (pk ProxyKeeper) SendCoins(ctx types.Context, fromAddr types.AccAddress, toAddr types.AccAddress, amt types.Coins) error {
	restrictedDenoms := pk.rk.GetRestrictedDenoms(ctx)
	for _, c := range amt {
		if denom, found := restrictedDenoms.Find(c.Denom); found {
			if !denom.IsAnyAllowed(fromAddr, toAddr) {
				return sdkerrors.Wrap(ErrRestrictedDenomination, c.Denom)
			}
		}
	}

	return pk.bk.SendCoins(ctx, fromAddr, toAddr, amt)
}

func (pk ProxyKeeper) ValidateBalance(ctx types.Context, addr types.AccAddress) error {
	return pk.bk.ValidateBalance(ctx, addr)
}

func (pk ProxyKeeper) HasBalance(ctx types.Context, addr types.AccAddress, amt types.Coin) bool {
	return pk.bk.HasBalance(ctx, addr, amt)
}

func (pk ProxyKeeper) GetAllBalances(ctx types.Context, addr types.AccAddress) types.Coins {
	return pk.bk.GetAllBalances(ctx, addr)
}

func (pk ProxyKeeper) GetAccountsBalances(ctx types.Context) []banktypes.Balance {
	return pk.bk.GetAccountsBalances(ctx)
}

func (pk ProxyKeeper) GetBalance(ctx types.Context, addr types.AccAddress, denom string) types.Coin {
	return pk.bk.GetBalance(ctx, addr, denom)
}

func (pk ProxyKeeper) LockedCoins(ctx types.Context, addr types.AccAddress) types.Coins {
	return pk.bk.LockedCoins(ctx, addr)
}

func (pk ProxyKeeper) SpendableCoins(ctx types.Context, addr types.AccAddress) types.Coins {
	return pk.bk.SpendableCoins(ctx, addr)
}

func (pk ProxyKeeper) IterateAccountBalances(ctx types.Context, addr types.AccAddress, cb func(coin types.Coin) (stop bool)) {
	pk.bk.IterateAccountBalances(ctx, addr, cb)
}

func (pk ProxyKeeper) IterateAllBalances(ctx types.Context, cb func(address types.AccAddress, coin types.Coin) (stop bool)) {
	pk.bk.IterateAllBalances(ctx, cb)
}

func (pk ProxyKeeper) SubtractCoins(ctx types.Context, addr types.AccAddress, amt types.Coins) error {
	return pk.bk.SubtractCoins(ctx, addr, amt)
}

func (pk ProxyKeeper) AddCoins(ctx types.Context, addr types.AccAddress, amt types.Coins) error {
	return pk.bk.AddCoins(ctx, addr, amt)
}

func (pk ProxyKeeper) SetBalance(ctx types.Context, addr types.AccAddress, balance types.Coin) error {
	return pk.bk.SetBalance(ctx, addr, balance)
}

func (pk ProxyKeeper) SetBalances(ctx types.Context, addr types.AccAddress, balances types.Coins) error {
	return pk.bk.SetBalances(ctx, addr, balances)
}

func (pk ProxyKeeper) GetParams(ctx types.Context) banktypes.Params {
	return pk.bk.GetParams(ctx)
}

func (pk ProxyKeeper) SendEnabledCoin(ctx types.Context, coin types.Coin) bool {
	return pk.bk.SendEnabledCoin(ctx, coin)
}

func (pk ProxyKeeper) SendEnabledCoins(ctx types.Context, coins ...types.Coin) error {
	return pk.bk.SendEnabledCoins(ctx, coins...)
}

func (pk ProxyKeeper) BlockedAddr(addr types.AccAddress) bool {
	return pk.bk.BlockedAddr(addr)
}

func (pk ProxyKeeper) SetParams(ctx types.Context, params banktypes.Params) {
	pk.bk.SetParams(ctx, params)
}

// This software is Copyright (c) 2019-2020 e-Money A/S. It is not offered under an open source license.
//
// Please contact partners@e-money.com for licensing related questions.

package bank

import (
	"github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/cosmos/cosmos-sdk/x/bank"
)

var _ bank.Keeper = (*ProxyKeeper)(nil)

type ProxyKeeper struct {
	bk bank.Keeper
	rk RestrictedKeeper
}

func Wrap(bk bank.Keeper, rk RestrictedKeeper) ProxyKeeper {
	return ProxyKeeper{bk, rk}
}

func (pk ProxyKeeper) InputOutputCoins(ctx types.Context, inputs []bank.Input, outputs []bank.Output) error {
	restrictedDenoms := pk.rk.GetRestrictedDenoms(ctx)
	// Multisend does not support restricted denominations.
	for _, input := range inputs {
		for _, coin := range input.Coins {
			if _, found := restrictedDenoms.Find(coin.Denom); found {
				return sdkerrors.Wrap(ErrRestrictedDenominationUsed, coin.Denom)
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
				return sdkerrors.Wrap(ErrRestrictedDenominationUsed, c.Denom)
			}
		}
	}

	return pk.bk.SendCoins(ctx, fromAddr, toAddr, amt)
}

func (pk ProxyKeeper) GetCoins(ctx types.Context, addr types.AccAddress) types.Coins {
	return pk.bk.GetCoins(ctx, addr)
}

func (pk ProxyKeeper) HasCoins(ctx types.Context, addr types.AccAddress, amt types.Coins) bool {
	return pk.bk.HasCoins(ctx, addr, amt)
}

//func (pk ProxyKeeper) Codespace() types.CodespaceType {
//	return pk.bk.Codespace()
//}

func (pk ProxyKeeper) SubtractCoins(ctx types.Context, addr types.AccAddress, amt types.Coins) (types.Coins, error) {
	return pk.bk.SubtractCoins(ctx, addr, amt)
}

func (pk ProxyKeeper) AddCoins(ctx types.Context, addr types.AccAddress, amt types.Coins) (types.Coins, error) {
	return pk.bk.AddCoins(ctx, addr, amt)
}

func (pk ProxyKeeper) SetCoins(ctx types.Context, addr types.AccAddress, amt types.Coins) error {
	return pk.bk.SetCoins(ctx, addr, amt)
}

func (pk ProxyKeeper) GetSendEnabled(ctx types.Context) bool {
	return pk.bk.GetSendEnabled(ctx)
}

func (pk ProxyKeeper) SetSendEnabled(ctx types.Context, enabled bool) {
	pk.bk.SetSendEnabled(ctx, enabled)
}

func (pk ProxyKeeper) BlacklistedAddr(addr types.AccAddress) bool {
	return pk.bk.BlacklistedAddr(addr)
}

func (pk ProxyKeeper) DelegateCoins(ctx types.Context, delegatorAddr, moduleAccAddr types.AccAddress, amt types.Coins) error {
	return pk.bk.DelegateCoins(ctx, delegatorAddr, moduleAccAddr, amt)
}

func (pk ProxyKeeper) UndelegateCoins(ctx types.Context, moduleAccAddr, delegatorAddr types.AccAddress, amt types.Coins) error {
	return pk.bk.UndelegateCoins(ctx, moduleAccAddr, delegatorAddr, amt)
}

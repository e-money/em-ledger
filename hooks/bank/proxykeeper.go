// This software is Copyright (c) 2019-2020 e-Money A/S. It is not offered under an open source license.
//
// Please contact partners@e-money.com for licensing related questions.

package bank

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/cosmos/cosmos-sdk/x/bank/exported"
	bankkeeper "github.com/cosmos/cosmos-sdk/x/bank/keeper"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	"github.com/e-money/em-ledger/x/authority/types"
)

var _ bankingMethods = (*ProxyKeeper)(nil)

type bankingMethods interface {
	bankkeeper.SendKeeper
	GetSupply(ctx sdk.Context) exported.SupplyI
}

type ProxyKeeper struct {
	bk        bankingMethods
	rk        RestrictedKeeper
	listeners []func(sdk.Context, []sdk.AccAddress)
}

func Wrap(bk bankkeeper.Keeper, rk RestrictedKeeper) *ProxyKeeper {
	return &ProxyKeeper{bk: bk, rk: rk}
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
	restrictedDenoms := pk.rk.GetRestrictedDenoms(ctx)
	// Multisend does not support restricted denominations.
	for _, input := range inputs {
		for _, coin := range input.Coins {
			if _, found := restrictedDenoms.Find(coin.Denom); found {
				return sdkerrors.Wrap(ErrRestrictedDenomination, coin.Denom)
			}
		}
	}

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
	restrictedDenoms := pk.rk.GetRestrictedDenoms(ctx)
	for _, c := range amt {
		if denom, found := restrictedDenoms.Find(c.Denom); found {
			if !denom.IsAnyAllowed(fromAddr, toAddr) {
				return sdkerrors.Wrap(ErrRestrictedDenomination, c.Denom)
			}
		}
	}

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

func (pk ProxyKeeper) SubtractCoins(ctx sdk.Context, addr sdk.AccAddress, amt sdk.Coins) error {
	return pk.bk.SubtractCoins(ctx, addr, amt)
}

func (pk ProxyKeeper) AddCoins(ctx sdk.Context, addr sdk.AccAddress, amt sdk.Coins) error {
	return pk.bk.AddCoins(ctx, addr, amt)
}

func (pk ProxyKeeper) SetBalance(ctx sdk.Context, addr sdk.AccAddress, balance sdk.Coin) error {
	return pk.bk.SetBalance(ctx, addr, balance)
}

func (pk ProxyKeeper) SetBalances(ctx sdk.Context, addr sdk.AccAddress, balances sdk.Coins) error {
	return pk.bk.SetBalances(ctx, addr, balances)
}

func (pk ProxyKeeper) GetParams(ctx sdk.Context) banktypes.Params {
	return pk.bk.GetParams(ctx)
}

func (pk ProxyKeeper) SendEnabledCoin(ctx sdk.Context, coin sdk.Coin) bool {
	return pk.bk.SendEnabledCoin(ctx, coin)
}

func (pk ProxyKeeper) SendEnabledCoins(ctx sdk.Context, coins ...sdk.Coin) error {
	return pk.bk.SendEnabledCoins(ctx, coins...)
}

func (pk ProxyKeeper) BlockedAddr(addr sdk.AccAddress) bool {
	return pk.bk.BlockedAddr(addr)
}

func (pk ProxyKeeper) SetParams(ctx sdk.Context, params banktypes.Params) {
	pk.bk.SetParams(ctx, params)
}

func (pk ProxyKeeper) GetSupply(ctx sdk.Context) exported.SupplyI {
	return pk.bk.GetSupply(ctx)
}

// RestrictedKeeperFunc implements the RestrictedKeeper interface.
type RestrictedKeeperFunc func(sdk.Context) types.RestrictedDenoms

func (r RestrictedKeeperFunc) GetRestrictedDenoms(ctx sdk.Context) types.RestrictedDenoms {
	return r(ctx)
}

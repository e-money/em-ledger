// This software is Copyright (c) 2019-2020 e-Money A/S. It is not offered under an open source license.
//
// Please contact partners@e-money.com for licensing related questions.

package types

import (
	"fmt"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

func NewLiquidityProviderAccount(provAddr string, mintable sdk.Coins) (*LiquidityProviderAccount, error) {
	return &LiquidityProviderAccount{
		ProvAddr:  provAddr,
		Mintable: mintable,
	}, nil
}

// Validate validates the liquidity provider monetary load (Mintable) conforms
// to Cosmos' notion of Coin and provider address is a bech32 address.
func (acc LiquidityProviderAccount) Validate() error {
	if err := acc.Mintable.Validate(); err != nil {
		return sdkerrors.Wrap(err, "mintable")
	}

	_, err := sdk.AccAddressFromBech32(acc.ProvAddr)
	if err != nil {
		return sdkerrors.Wrap(sdkerrors.ErrInvalidAddress, acc.ProvAddr)
	}

	return nil
}

func (p *LiquidityProviderAccount) IncreaseMintableAmount(increase sdk.Coins) {
	p.Mintable = p.Mintable.Add(increase...)
}

// Function panics if resulting mintable amount is negative. Should be checked
// prior to invocation for cleaner handling.
func (p *LiquidityProviderAccount) DecreaseMintableAmount(decrease sdk.Coins) {
	if mintable, anyNegative := p.Mintable.SafeSub(decrease); !anyNegative {
		p.Mintable = mintable
		return
	}

	panic(fmt.Errorf("mintable amount cannot be negative"))
}

func (p LiquidityProviderAccount) String() string {
	return fmt.Sprintf(`Account:
  Address:       %s
  Mintable:      %s`,
		p.ProvAddr, p.Mintable)
}

func (p *LiquidityProviderAccount) GetAddress() (sdk.AccAddress, error) {
	acc, err := sdk.AccAddressFromBech32(p.ProvAddr)
	if err != nil {
		return acc, sdkerrors.Wrap(sdkerrors.ErrInvalidAddress, p.ProvAddr)
	}

	return acc, nil
}

func (p *LiquidityProviderAccount) SetAddress(address string) {
	p.ProvAddr = address
}
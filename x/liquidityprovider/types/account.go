package types

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

func NewLiquidityProviderAccount(provAddr string, mintable sdk.Coins) (*LiquidityProviderAccount, error) {
	return &LiquidityProviderAccount{
		Address:  provAddr,
		Mintable: mintable,
	}, nil
}

// Validate validates the liquidity provider monetary load (Mintable) conforms
// to Cosmos' notion of Coin and provider address is a bech32 address.
func (acc LiquidityProviderAccount) Validate() error {
	if err := acc.Mintable.Validate(); err != nil {
		return sdkerrors.Wrap(err, "mintable")
	}

	_, err := sdk.AccAddressFromBech32(acc.Address)
	if err != nil {
		return sdkerrors.Wrap(sdkerrors.ErrInvalidAddress, acc.Address)
	}

	return nil
}

func (acc *LiquidityProviderAccount) IncreaseMintableAmount(increase sdk.Coins) {
	acc.Mintable = acc.Mintable.Add(increase...)
}

func (acc *LiquidityProviderAccount) DecreaseMintableAmount(decrease sdk.Coins) error {
	if mintable, anyNegative := acc.Mintable.SafeSub(decrease); !anyNegative {
		acc.Mintable = mintable
		return nil
	}

	return fmt.Errorf(
		"mintable amount cannot be negative, %s - %s", acc.Mintable.String(),
		decrease.String(),
	)
}

func (acc LiquidityProviderAccount) String() string {
	return fmt.Sprintf(`Account:
  Address:       %s
  Mintable:      %s`,
		acc.Address, acc.Mintable)
}

func (acc *LiquidityProviderAccount) GetAccAddress() (sdk.AccAddress, error) {
	accAddr, err := sdk.AccAddressFromBech32(acc.Address)
	if err != nil {
		return accAddr, sdkerrors.Wrap(sdkerrors.ErrInvalidAddress, acc.Address)
	}

	return accAddr, nil
}

func (acc *LiquidityProviderAccount) SetAddress(address string) {
	acc.Address = address
}

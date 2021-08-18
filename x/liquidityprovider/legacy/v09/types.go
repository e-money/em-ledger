package v09

import (
	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/codec/legacy"
	sdk "github.com/cosmos/cosmos-sdk/types"
	auth "github.com/cosmos/cosmos-sdk/x/auth/legacy/v039"
	"github.com/tidwall/gjson"
)

type LiquidityProviderAccount struct {
	*auth.BaseAccount

	Mintable sdk.Coins `json:"mintable" yaml:"mintable"`
}

// UnmarshalJSON unmarshals raw JSON bytes into a LiquidityProviderAccount.
func (ma *LiquidityProviderAccount) UnmarshalJSON(bz []byte) error {
	js := gjson.ParseBytes(bz)

	mintable := js.Get("mintable").String()
	if err := legacy.Cdc.UnmarshalJSON([]byte(mintable), &ma.Mintable); err != nil {
		return err
	}

	baseAccount := js.Get("Account.value").String()
	ma.BaseAccount = &auth.BaseAccount{}
	if err := legacy.Cdc.UnmarshalJSON([]byte(baseAccount), ma.BaseAccount); err != nil {
		return err
	}

	return nil
}

func RegisterLegacyAminoCodec(cdc *codec.LegacyAmino) {
	cdc.RegisterConcrete(&LiquidityProviderAccount{}, "e-money/LiquidityProviderAccount", nil)
}

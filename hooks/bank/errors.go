package bank

import sdk "github.com/cosmos/cosmos-sdk/types"

type CodeType = sdk.CodeType

const (
	Codespace sdk.CodespaceType = "embank"

	CodeRestrictedDenomination sdk.CodeType = 1
)

func ErrRestrictedDenominationUsed(denom string) sdk.Error {
	return sdk.NewError(Codespace, CodeRestrictedDenomination, "'%v' is a restricted denomination", denom)
}

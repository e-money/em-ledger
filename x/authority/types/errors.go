package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
)

type CodeType = sdk.CodeType

const (
	Codespace sdk.CodespaceType = ModuleName

	CodeMissingDenomination sdk.CodeType = 1
)

func ErrNoDenomsSpecified() sdk.Error {
	return sdk.NewError(Codespace, CodeMissingDenomination, "No denominations specified in authority call")
}

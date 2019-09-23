package types

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

const (
	DefaultCodespace sdk.CodespaceType = "iss"

	CodeNegativeCredit sdk.CodeType = 1
)

func ErrNegativeCredit(lp sdk.AccAddress) sdk.Error {
	return sdk.NewError(DefaultCodespace, CodeNegativeCredit, fmt.Sprintf("credit decrease would result in negative credit for %d", lp))
}

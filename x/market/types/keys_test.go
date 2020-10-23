package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestParsePriorityKey1(t *testing.T) {
	key := GetPriorityKey("eur", "chf", sdk.NewDec(5), 14)

	src, dst, err := ParsePriorityKey(key)
	require.NoError(t, err)
	require.Equal(t, "eur", src)
	require.Equal(t, "chf", dst)
}

func TestParsePriorityKey2(t *testing.T) {
	key := GetOwnerKey("acc1", "clientoid")

	_, _, err := ParsePriorityKey(key)
	require.Error(t, err)
}

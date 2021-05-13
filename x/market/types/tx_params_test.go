package types

import (
	"github.com/stretchr/testify/require"
	"testing"
)

func Test_ParamsEqual(t *testing.T) {
	paramsA := NewTxParams(1, 0, 1)
	err := paramsA.Validate()
	require.NoError(t, err)
	paramsB := NewTxParams(1, 0, 1)
	err = paramsB.Validate()
	require.NoError(t, err)
	paramsC := NewTxParams(1, 1, 1)
	err = paramsC.Validate()
	require.NoError(t, err)

	require.True(t, paramsA.String() == paramsB.String())

	require.False(t, paramsA.String() == paramsC.String())
}

func TestValidateParams(t *testing.T) {
	require.NoError(t, DefaultTxParams().Validate())
	require.NoError(t, NewTxParams(1, 0, 5).Validate())
	require.Error(t, NewTxParams(0, 0, 5).Validate())
	require.Error(t, NewTxParams(0, 0, -1).Validate())
}
package main

import (
	"bytes"
	"testing"

	app "github.com/e-money/em-ledger"
	apptypes "github.com/e-money/em-ledger/types"

	"github.com/stretchr/testify/require"
	"github.com/tidwall/gjson"

	tmtypes "github.com/tendermint/tendermint/types"
)

func init() {
	apptypes.ConfigureSDK()
}

func TestMigrateCommand(t *testing.T) {
	cdc := app.MakeCodec()

	writer := bytes.NewBufferString("")

	cmd := MigrateGenesisCmd(cdc, writer)
	cmd.SetArgs([]string{"./testdata/emoney1-state"})
	err := cmd.Execute()
	require.NoError(t, err)
	require.NotEmpty(t, writer)

	json := gjson.ParseBytes(writer.Bytes())
	require.True(t, json.IsObject())

	// Verify that the updated genesis file can pass validation
	_, err = tmtypes.GenesisDocFromJSON(writer.Bytes())
	require.NoError(t, err)
}

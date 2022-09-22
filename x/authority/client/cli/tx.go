// This software is Copyright (c) 2019-2020 e-Money A/S. It is not offered under an open source license.
//
// Please contact partners@e-money.com for licensing related questions.

package cli

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"strings"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/client/tx"
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/cosmos/cosmos-sdk/x/params/client/utils"
	upgtypes "github.com/cosmos/cosmos-sdk/x/upgrade/types"
	"github.com/e-money/em-ledger/util"
	"github.com/e-money/em-ledger/x/authority/types"
	"github.com/spf13/cobra"
	"github.com/tendermint/tendermint/libs/os"
)

func GetTxCmd() *cobra.Command {
	authorityCmds := &cobra.Command{
		Use:                "authority",
		Short:              "Manage authority tasks",
		DisableFlagParsing: false,
	}

	authorityCmds.AddCommand(
		GetCmdCreateIssuer(),
		getCmdDestroyIssuer(),
		getCmdSetGasPrices(),
		GetCmdReplaceAuthority(),
		GetCmdScheduleUpgrade(),
		getCmdSetParameters(),
	)

	return authorityCmds
}

func getCmdSetGasPrices() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "set-gas-prices [authority_key_or_address] [minimum_gas_prices]",
		Example: "emd tx authority set-gas-prices masterkey 0.0005eeur,0.0000001ejpy",
		Short:   "Control the minimum gas prices for the chain",
		Args:    cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			cmd.Flags().Set(flags.FlagFrom, args[0])
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			gasPrices, err := sdk.ParseDecCoins(args[1])
			if err != nil {
				return err
			}

			msg := &types.MsgSetGasPrices{
				GasPrices: gasPrices,
				Authority: clientCtx.GetFromAddress().String(),
			}
			if err := msg.ValidateBasic(); err != nil {
				return err
			}
			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
		},
	}
	flags.AddTxFlagsToCmd(cmd)
	return cmd
}

const (
	DenomDescFlagName = "denominations"
	denomDescDefValue = "e-Money EUR stablecoin"
)

func GetCmdCreateIssuer() *cobra.Command {
	var denoms []string

	cmd := &cobra.Command{
		Use:     "create-issuer [authority_key_or_address] [issuer_address]",
		Example: "emd tx authority create-issuer masterkey emoney17up20gamd0vh6g9ne0uh67hx8xhyfrv2lyazgu -d 'eeur,EEUR,e-Money Euro stablecoin' -d ejpy",
		Short:   "Create a new issuer",
		Args:    cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			cmd.Flags().Set(flags.FlagFrom, args[0])
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			issuerAddr, err := sdk.AccAddressFromBech32(args[1])
			if err != nil {
				return err
			}

			denoms, err := util.ParseDenominations(denoms, denomDescDefValue)
			if err != nil {
				return err
			}

			msg := &types.MsgCreateIssuer{
				Issuer:        issuerAddr.String(),
				Denominations: denoms,
				Authority:     clientCtx.GetFromAddress().String(),
			}

			if err := msg.ValidateBasic(); err != nil {
				return err
			}

			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
		},
	}
	f := cmd.Flags()
	f.StringArrayVarP(&denoms, DenomDescFlagName, "d", []string{}, "The denominations with base i.e. eeur, display i.e. EEUR, description with default value: e-Money EUR stablecoin")
	flags.AddTxFlagsToCmd(cmd)
	return cmd
}

func getCmdDestroyIssuer() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "destroy-issuer [authority_key_or_address] [issuer_address]",
		Example: "emd tx authority destory-issuer masterkey emoney17up20gamd0vh6g9ne0uh67hx8xhyfrv2lyazgu",
		Short:   "Delete an issuer",
		Args:    cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			cmd.Flags().Set(flags.FlagFrom, args[0])
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			issuerAddr, err := sdk.AccAddressFromBech32(args[1])
			if err != nil {
				return err
			}

			msg := &types.MsgDestroyIssuer{
				Issuer:    issuerAddr.String(),
				Authority: clientCtx.GetFromAddress().String(),
			}

			if err := msg.ValidateBasic(); err != nil {
				return err
			}
			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
		},
	}
	flags.AddTxFlagsToCmd(cmd)
	return cmd
}

func GetCmdReplaceAuthority() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "replace [authority_key_or_address] new_authority_address",
		Short:   "Replace the authority key",
		Example: "emd tx authority replace emoney1n5ggspeff4fxc87dvmg0ematr3qzw5l4v20mdv emoney1hq6tnhqg4t7358f3vd9crru93lv0cgekdxrtgv",
		Long: `Replace the authority key with a new multisig address.
For a 24-hour grace period the former authority key is equivalent to the new one.`,
		Args: cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			f := cmd.Flags()

			err := f.Set(flags.FlagFrom, args[0])
			if err != nil {
				return err
			}

			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			msg := &types.MsgReplaceAuthority{
				Authority:    clientCtx.GetFromAddress().String(),
				NewAuthority: args[1],
			}

			if err := msg.ValidateBasic(); err != nil {
				return err
			}

			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
		},
	}
	flags.AddTxFlagsToCmd(cmd)
	return cmd
}

const (
	UpgHeight = "upgrade-height"
	UpgInfo   = "upgrade-info"
)

func GetCmdScheduleUpgrade() *cobra.Command {
	var (
		upgHeightVal int64
		upgInfoVal   string
	)

	cmd := &cobra.Command{
		Use:   "schedule-upgrade [authority_key_or_address] plan_name",
		Short: "Schedule a software upgrade",
		Example: `emd tx authority schedule-upgrade someplan --upgrade-height 2001 --from emoney1xue7fm6es84jze49grm4slhlmr4ffz8a3u7g3t 0.43
emd tx authority schedule-upgrade sdk-v0.43.0 --upgrade-height 2001 --from emoney1xue7fm6es84jze49grm4slhlmr4ffz8a3u7g3t --upgrade-info '{"binaries":{"linux/amd64":"http://localhost:8765/test-upg-0.2.0/emd.zip?checksum=sha256:cadd5b52fe90a04e20b2cbb93291b0d1d0204f17b64b2215eb09f5dc78a127f1"}}'`,
		Long: `Schedule a software upgrade by submitting a unique plan name that
 has not been used before with either an absolute block height or block time. An
upgrade handler should be defined at the upgraded binary. Optionally If you set DAEMON_ALLOW_DOWNLOAD_BINARIES=on pass
the upgraded binary download url with the --upgrade-info flag i.e., --upgrade-info '{"binaries":{"linux/amd64":"http://localhost:8765/test-upg-0.2.0/emd.zip?checksum=sha256:cadd5b52fe90a04e20b2cbb93291b0d1d0204f17b64b2215eb09f5dc78a127f1"}}'`,
		Args: cobra.RangeArgs(1, 2),
		RunE: func(cmd *cobra.Command, args []string) error {
			err := cmd.Flags().Set(flags.FlagFrom, args[0])
			if err != nil {
				return err
			}

			if err := validateUpgFlags(UpgHeight, upgHeightVal); err != nil {
				return err
			}

			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			msg := &types.MsgScheduleUpgrade{
				Authority: clientCtx.GetFromAddress().String(),
				Plan: upgtypes.Plan{
					Name:   args[1],
					Height: upgHeightVal,
					Info:   upgInfoVal,
				},
			}

			if err := msg.ValidateBasic(); err != nil {
				return err
			}

			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
		},
	}
	f := cmd.Flags()
	f.Int64VarP(&upgHeightVal, UpgHeight, "n", 0, "Upgrade block height number")
	f.StringVarP(
		&upgInfoVal, UpgInfo, "i", "", "Upgrade info",
	)

	flags.AddTxFlagsToCmd(cmd)
	return cmd
}

func validateUpgFlags(upgHeight string, upgHeightVal int64) error {
	if upgHeightVal == 0 {
		return sdkerrors.Wrapf(
			types.ErrMissingFlag,
			"need to specify --%s", upgHeight,
		)
	}

	flagsSet := 0
	if upgHeightVal != 0 {
		flagsSet++
	}
	if flagsSet != 1 {
		return sdkerrors.Wrapf(
			sdkerrors.ErrInvalidRequest,
			"specify only one of the flags: --%s", upgHeight)
	}

	return nil
}

func getCmdSetParameters() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "set-params [authority_key_or_address] <path/to/changes.json -- OR -- JSON snippet> --from <authority-key>",
		Short: "Set a parameter change with a JSON file or JSON snippet",
		Example: `emd tx authority set-params ./params.json --from emoney1xue7fm6es84jze49grm4slhlmr4ffz8a3u7g3t
emd tx authority set-params '[{"subspace":"staking","key":"MaxValidators","value":10}]' --from emoney1xue7fm6es84jze49grm4slhlmr4ffz8a3u7g3t`,
		Long: strings.TrimSpace(`
The parameter details must be supplied via a JSON file or JSON snippet. For values that contain
objects, only non-empty fields will be updated.
Any "value" change should be valid (ie. correct type and within bounds)
for its respective parameter, eg. "MaxValidators" should be an integer and not a
decimal.

Where proposal.json contains:

[
  {
    "subspace": "staking",
    "key": "MaxValidators",
    "value": 10
  }
]

-- OR JSON fragment e.g. [{"subspace":"staking","key":"MaxValidators","value":10}]

`),
		Args: cobra.RangeArgs(1, 2),
		RunE: func(cmd *cobra.Command, args []string) error {
			var paramsFilename string

			if len(args) == 2 {
				if err := cmd.Flags().Set(flags.FlagFrom, args[0]); err != nil {
					return err
				}

				paramsFilename = args[1]
			} else {
				paramsFilename = args[0]
			}

			if !os.FileExists(paramsFilename) && !json.Valid([]byte(paramsFilename)) {
				return fmt.Errorf("%s is not the name of an existing file nor valid json", paramsFilename)
			}

			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			paramChanges, err := parseParamChangesJSON(clientCtx.LegacyAmino, paramsFilename)
			if err != nil {
				return err
			}

			msg := &types.MsgSetParameters{
				Authority: clientCtx.GetFromAddress().String(),
				Changes:   paramChanges.ToParamChanges(),
			}

			if err := msg.ValidateBasic(); err != nil {
				return err
			}

			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
		},
	}

	flags.AddTxFlagsToCmd(cmd)
	return cmd
}

// parseParamChangesJSON reads and parses a ParamChangesJSON from file.
func parseParamChangesJSON(cdc *codec.LegacyAmino, jsonFile string) (utils.ParamChangesJSON, error) {
	params := utils.ParamChangesJSON{}

	paramsJson := []byte(jsonFile)
	if json.Valid(paramsJson) {
		return getParsedParams(cdc, paramsJson)
	}

	paramsJson, err := ioutil.ReadFile(jsonFile)
	if err != nil {
		return params, err
	}

	return getParsedParams(cdc, paramsJson)
}

func getParsedParams(cdc *codec.LegacyAmino, contents []byte) (utils.ParamChangesJSON, error) {
	var params utils.ParamChangesJSON
	err := json.Unmarshal(contents, &params)

	return params, err
}

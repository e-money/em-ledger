// This software is Copyright (c) 2019-2020 e-Money A/S. It is not offered under an open source license.
//
// Please contact partners@e-money.com for licensing related questions.

package cli

import (
	"bytes"
	"errors"
	"fmt"
	"github.com/cosmos/cosmos-sdk/client/keys"
	"gopkg.in/yaml.v2"
	"io"
	"sort"
	"strconv"
	"strings"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/client/tx"
	"github.com/cosmos/cosmos-sdk/crypto/keyring"
	"github.com/cosmos/cosmos-sdk/crypto/keys/multisig"
	cryptotypes "github.com/cosmos/cosmos-sdk/crypto/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/e-money/em-ledger/util"
	"github.com/e-money/em-ledger/x/authority/types"
	"github.com/spf13/cobra"
)

func GetTxCmd() *cobra.Command {
	authorityCmds := &cobra.Command{
		Use:                "authority",
		Short:              "Manage authority tasks",
		DisableFlagParsing: false,
	}

	authorityCmds.AddCommand(
		getCmdCreateIssuer(),
		getCmdDestroyIssuer(),
		getCmdSetGasPrices(),
		GetCmdReplaceAuthority(),
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

func getCmdCreateIssuer() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "create-issuer [authority_key_or_address] [issuer_address] [denominations]",
		Example: "emd tx authority create-issuer masterkey emoney17up20gamd0vh6g9ne0uh67hx8xhyfrv2lyazgu eeur,ejpy",
		Short:   "Create a new issuer",
		Args:    cobra.ExactArgs(3),
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

			denoms, err := util.ParseDenominations(args[2])
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
	const (
		flagAuthorities       = "authorities"
		flagMultiSigThreshold = "threshold"
		mnemonicEntropySize   = 256
		authKeyName           = "authorityKey"
	)

	cmd := &cobra.Command{
		Use:   "replace [authority_key_or_address] addr1,addr2,addr3 multisig-threshold",
		Short: "Replace the authority key",
		Example: "emd tx authority replace emoney1n5ggspeff4fxc87dvmg0ematr3qzw5l4v20mdv emoney1lagqmceycrfpkyu7y6ayrk6jyvru5mkrezacpw,emoney1gjudpa2cmwd27cjzespu2khrvy2ukje6zfevk5,emoney1mn2y0d9ugjxqevpn6k5e20wy62kcawp5523sgc 2",
		Long: `Replace the authority key by entering the list of addresses comprising the new authority multisig key 
and multisig threshold such that K out of N required authority signatures. A new authority key will be generated. 
The key will be stored in the chain state. If run with --dry-run, a key would be generated but not stored to the 
chain state. Use the --authorities flag with existing keystore users for generating the new multisig authority address.
`,
		Args:    cobra.ExactArgs(3),
		RunE: func(cmd *cobra.Command, args []string) error {
			fmt.Println("***")
			fmt.Println("***")
			fmt.Println("args:", len(args), args)
			for i, arg := range args {
				fmt.Print(i, ":", arg)
			}
			fmt.Println()

			f := cmd.Flags()
			var (
				authAddresses []string
				multisigThreshold int
			)

			err := f.Set(flags.FlagFrom, args[0])
			if err != nil {
				//return err
				panic(err)
			}

			authAddresses = strings.Split(args[1], ",")
			multisigThreshold, err = strconv.Atoi(args[2])
			if err != nil {
				//return err
				panic(err)
			}

			//// generate only for later signing
			//err = f.Set(flags.FlagGenerateOnly, "true")
			//if err != nil {
			//	return err
			//}
			//

			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				//return err
				panic(err)
			}

			kb := clientCtx.Keyring
			if len(authAddresses) == 0{
				//return errors.New("missing input multisig keys")
				panic(errors.New("missing input multisig keys"))
			}

			var pks []cryptotypes.PubKey
			if err := validateMultisigThreshold(multisigThreshold, len(authAddresses)); err != nil {
				//return err
				panic(err)
			}

			for _, keyname := range authAddresses {
				k, err := kb.Key(keyname)
				if err != nil {
					//return err
					panic(err)
				}

				pks = append(pks, k.GetPubKey())
			}

			sort.Slice(
				pks, func(i, j int) bool {
					return bytes.Compare(pks[i].Address(), pks[j].Address()) < 0
				},
			)

			pk := multisig.NewLegacyAminoPubKey(multisigThreshold, pks)
			if _, err := kb.SaveMultisig(authKeyName, pk); err != nil {
				//return err
				panic(err)
			}

			//// read entropy seed straight from tmcrypto.Rand and convert to mnemonic
			//entropySeed, err := bip39.NewEntropy(mnemonicEntropySize)
			//if err != nil {
			////	return err
			//}
			//
			//mnemonic, err := bip39.NewMnemonic(entropySeed)
			//if err != nil {
			////	return err
			//}
			//
			//info, err := kb.NewAccount(authKeyName, mnemonic, "", hdPath, algo)
			//if err != nil {
			////	return err
			//}

			msg := &types.MsgReplaceAuthority{
				Authority:      clientCtx.GetFromAddress().String(),
				NewAuthorities: nil,
				Threshold:      0,
			}

			if err := msg.ValidateBasic(); err != nil {
				//return err
				panic(err)
			}

			panic(errors.New("just before broadcasting"))

			// return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
		},
	}
	flags.AddTxFlagsToCmd(cmd)
	return cmd
}

type bechKeyOutFn func(keyInfo keyring.Info) (keyring.KeyOutput, error)

func validateMultisigThreshold(k, nKeys int) error {
	if k <= 0 {
		return fmt.Errorf("threshold must be a positive integer")
	}
	if nKeys < k {
		return fmt.Errorf(
			"threshold k of n multisignature: %d < %d", nKeys, k)
	}
	return nil
}

// available output formats.
const (
	OutputFormatText = "text"
	OutputFormatJSON = "json"
)

func printCreate(cmd *cobra.Command, info keyring.Info, showMnemonic bool, mnemonic string, outputFormat string) error {
	switch outputFormat {
	case OutputFormatText:
		cmd.PrintErrln()
		printKeyInfo(cmd.OutOrStdout(), info, keyring.Bech32KeyOutput, outputFormat)

		// print mnemonic unless requested not to.
		if showMnemonic {
			fmt.Fprintln(cmd.ErrOrStderr(), "\n**Important** write this mnemonic phrase in a safe place.")
			fmt.Fprintln(cmd.ErrOrStderr(), "It is the only way to recover your account if you ever forget your password.")
			fmt.Fprintln(cmd.ErrOrStderr(), "")
			fmt.Fprintln(cmd.ErrOrStderr(), mnemonic)
		}
	case OutputFormatJSON:
		out, err := keyring.Bech32KeyOutput(info)
		if err != nil {
			return err
		}

		if showMnemonic {
			out.Mnemonic = mnemonic
		}

		jsonString, err := keys.KeysCdc.MarshalJSON(out)
		if err != nil {
			return err
		}

		cmd.Println(string(jsonString))

	default:
		return fmt.Errorf("invalid output format %s", outputFormat)
	}

	return nil
}

// MkAccKeyOutput create a KeyOutput in with "acc" Bech32 prefixes. If the
// public key is a multisig public key, then the threshold and constituent
// public keys will be added.
func MkAccKeyOutput(keyInfo keyring.Info) keyring.KeyOutput {
	pk := keyInfo.GetPubKey()
	addr := sdk.AccAddress(pk.Address())
	return keyring.NewKeyOutput(keyInfo.GetName(), keyInfo.GetType().String(), addr.String(), pk.String())
}

func printKeyInfo(w io.Writer, keyInfo keyring.Info, bechKeyOut bechKeyOutFn, output string) {
	ko, err := bechKeyOut(keyInfo)
	if err != nil {
		panic(err)
	}

	switch output {
	case OutputFormatText:
		printTextInfos(w, []keyring.KeyOutput{ko})

	case OutputFormatJSON:

		out, err := keys.KeysCdc.MarshalJSON(ko)
		if err != nil {
			panic(err)
		}

		fmt.Fprintln(w, string(out))
	}
}

func printTextInfos(w io.Writer, kos []keyring.KeyOutput) {
	out, err := yaml.Marshal(&kos)
	if err != nil {
		panic(err)
	}
	fmt.Fprintln(w, string(out))
}
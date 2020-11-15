// This software is Copyright (c) 2019-2020 e-Money A/S. It is not offered under an open source license.
//
// Please contact partners@e-money.com for licensing related questions.

package cli

import (
	"fmt"
	"sort"
	"sync"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/e-money/em-ledger/x/queries/types"

	"github.com/cosmos/cosmos-sdk/client/context"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/x/auth/client/utils"
	"github.com/spf13/cobra"
	"golang.org/x/sync/errgroup"
)

const (
	defaultPage  = 1
	defaultLimit = 30 // should be consistent with tendermint/tendermint/rpc/core/pipe.go:19

	EventTypeTransfer     = "transfer"
	AttributeKeyRecipient = "recipient"
	sendTrx               = "Send"

	// Add uniqueness to overlapping trx timestamps
	trxHashPrefixLen = 6
)

type transferEvent struct {
	from       string
	to         string
	coinAmount string
	timestamp  string
	trxHash    string
}

type txsRespMap map[string]sdk.TxResponse

func GetQuerySpendableBalance(cdc *codec.Codec) *cobra.Command {
	spendableBalanceCmd := &cobra.Command{
		Use:   "spendable",
		Short: "Display the vested balance of an account",
		Args:  cobra.ExactArgs(1),

		RunE: func(cmd *cobra.Command, args []string) error {
			cliCtx := context.NewCLIContext().WithCodec(cdc)

			key, err := sdk.AccAddressFromBech32(args[0])
			if err != nil {
				return err
			}

			resp, _, err := cliCtx.Query(fmt.Sprintf("custom/%s/%s/%s", types.QuerierRoute, types.QuerySpendable, key))
			if err != nil {
				return err
			}

			var bal sdk.Coins
			err = cdc.UnmarshalJSON(resp, &bal)
			if err != nil {
				return err
			}

			return cliCtx.PrintOutput(bal)
		},
	}

	return flags.GetCommands(spendableBalanceCmd)[0]
}

// Meant as an extension to the "emcli query supply" queries.
func GetQueryCirculatingSupplyCmd(cdc *codec.Codec) *cobra.Command {
	circulatingSupplyCmd := &cobra.Command{
		Use:   "circulating",
		Short: "Display circulating (ie non-vesting) token supply",
		Args:  cobra.ExactArgs(0),
		RunE: func(cmd *cobra.Command, args []string) error {
			cliCtx := context.NewCLIContext().WithCodec(cdc)

			resp, _, err := cliCtx.Query(fmt.Sprintf("custom/%s/%s", types.QuerierRoute, types.QueryCirculating))
			if err != nil {
				return err
			}

			var totalSupply sdk.Coins
			err = cdc.UnmarshalJSON(resp, &totalSupply)
			if err != nil {
				return err
			}

			return cliCtx.PrintOutput(totalSupply)
		},
	}

	return flags.GetCommands(circulatingSupplyCmd)[0]
}

// GetQueryStatementCmd displays debits and credit transactions of the given
// account.
func GetQueryStatementCmd(cdc *codec.Codec) *cobra.Command {
	statementCmd := &cobra.Command{
		Use:   "statement",
		Short: "Display a statement for the given account",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			cliCtx := context.NewCLIContext().WithCodec(cdc)

			address, err := sdk.AccAddressFromBech32(args[0])
			if err != nil {
				return err
			}

			eventQueries := []string{
				fmt.Sprintf("%s.%s='%s'", EventTypeTransfer, sdk.AttributeKeySender, address),
				fmt.Sprintf("%s.%s='%s'", EventTypeTransfer, AttributeKeyRecipient, address),
			}

			txsTimestamps, addrTxs, err := postEventQueries(cliCtx, eventQueries)
			if err != nil {
				return err
			}

			// Display joined transactions
			displayResults(txsTimestamps, addrTxs)

			return nil
		},
	}

	return flags.GetCommands(statementCmd)[0]
}

// postEventQueries concurrently multiple post requests one for each event
// condition clause (till there is way to send a single combined query for
// multiple event clauses returning additive ORed results) and joins the
// results. Abandons processing in case of error immediately.
func postEventQueries(cliCtx context.CLIContext, eventQueries []string) ([]string, txsRespMap, error) {
	g := new(errgroup.Group)
	var m sync.Mutex
	addrTxs := make(txsRespMap)
	// Sort merged transactions by timestamp
	txsTimestamps := make([]string, 0)
	for _, url := range eventQueries {
		// Launch a goroutine to fetch the URL.
		req := []string{url}
		g.Go(func() error {
			// Fetch the URL.
			txsResult, err := utils.QueryTxsByEvents(cliCtx, req, defaultPage, defaultLimit)
			if err != nil {
				// do not return a partial statement
				return err
			}

			m.Lock()
			cacheTxs(txsResult, &txsTimestamps, addrTxs)
			m.Unlock()

			return nil
		})
	}

	return txsTimestamps, addrTxs, g.Wait()
}

// sortTrxKey concatenates timestamp + a hash prefix to enable uniqueness
func sortTrxKey(timestamp, hash string) string {
	return timestamp + hash[:trxHashPrefixLen]
}

// cacheTxs in map and timestamps in slice for later sorting
func cacheTxs(txsResult *sdk.SearchTxsResult, txsTimestamps *[]string, foundTxs txsRespMap) {
	for _, trx := range txsResult.Txs {
		key := sortTrxKey(trx.Timestamp, trx.TxHash)
		foundTxs[key] = trx
		*txsTimestamps = append(*txsTimestamps, key)
	}
}

// displayResults of transactions on standard output by descending chronological order.
func displayResults(txsTimestamps []string, addrTxs txsRespMap) {
	sort.Strings(txsTimestamps)

	// Present in descending order
	for i := len(txsTimestamps) - 1; i >= 0; i-- {
		trxKey := txsTimestamps[i]
		trx := addrTxs[trxKey]
		for _, log := range trx.Logs {
			for _, ev := range log.Events {
				if ev.Type != "transfer" {
					continue
				}
				transferEvent := transferEvent{
					timestamp: trxKey[:len(trxKey)-trxHashPrefixLen],
					trxHash:   trx.TxHash,
				}
				for _, attr := range ev.Attributes {
					switch attr.Key {
					case AttributeKeyRecipient:
						transferEvent.to = attr.Value
					case sdk.AttributeKeySender:
						transferEvent.from = attr.Value
					case sdk.AttributeKeyAmount:
						transferEvent.coinAmount = attr.Value
					}
				}
				fmt.Printf("%s %s %10s to\n%s\nTransacted at %s\nTransaction# %s\n\n",
					transferEvent.from, sendTrx, transferEvent.coinAmount,
					transferEvent.to,
					transferEvent.timestamp,
					transferEvent.trxHash)
			}
		}
	}
}

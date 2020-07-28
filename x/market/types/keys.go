// This software is Copyright (c) 2019-2020 e-Money A/S. It is not offered under an open source license.
//
// Please contact partners@e-money.com for licensing related questions.

package types

import (
	"bytes"
	"encoding/hex"
	"fmt"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"strings"

	"github.com/e-money/em-ledger/util"
)

const (
	ModuleName   = "market"
	StoreKey     = ModuleName
	StoreKeyIdx  = "market_indices"
	RouterKey    = ModuleName
	QuerierRoute = ModuleName

	// Query endpoints supported by the market querier
	QueryInstruments = "instruments"
	QueryInstrument  = "instrument"
	QueryByAccount   = "account"
)

var (
	// Parameter key for global order IDs
	globalOrderIDKey = []byte("globalOrderID")

	// IAVL Store prefixes
	keysPrefix       = []byte{0x01}
	instrumentPrefix = []byte{0x02}
	priorityPrefix   = []byte{0x03}
	ownerPrefix      = []byte{0x04}
)

func GetOrderIDGeneratorKey() []byte {
	return append(keysPrefix, globalOrderIDKey...)
}

func GetInstrumentsKey() []byte {
	return instrumentPrefix
}

func GetInstrumentKeyBySrcAndDst(src, dst string) []byte {
	instr := fmt.Sprintf("%v/%v", src, dst)
	return append(instrumentPrefix, []byte(instr)...)
}

func GetPriorityKeyBySrcAndDst(src, dst string) []byte {
	instr := fmt.Sprintf("%v/%v", src, dst)
	return append(priorityPrefix, []byte(instr)...)
}

func GetPriorityKeyBySource(src string) []byte {
	instr := fmt.Sprintf("%v/", src)
	return append(priorityPrefix, []byte(instr)...)
}

func GetPriorityKeyByInstrument(src, dst string) []byte {
	instr := fmt.Sprintf("%v/%v/", src, dst)
	return append(priorityPrefix, []byte(instr)...)
}

func GetPriorityKey(src, dst string, price sdk.Dec, orderId uint64) []byte {
	res := GetPriorityKeyByInstrument(src, dst)
	res = append(res, sdk.SortableDecBytes(price)...)
	res = append(res, util.Uint64ToBytes(orderId)...)
	return res
}

func MustParseInstrumentKey(key []byte) (source, destination string) {
	src, dst, err := ParseInstrumentKey(key)
	if err != nil {
		panic(err)
	}

	return src, dst
}

func ParseInstrumentKey(key []byte) (source, destination string, err error) {
	if len(key) == 0 {
		return "", "", fmt.Errorf("empty key received")
	}

	if !bytes.HasPrefix(key, instrumentPrefix) {
		return "", "", fmt.Errorf("invalid prefix: %v", hex.EncodeToString(key))
	}

	a := strings.Split(string(key[1:]), "/")
	return a[0], a[1], nil
}

func MustParsePriorityKey(key []byte) (source, destination string) {
	src, dest, err := ParsePriorityKey(key)
	if err != nil {
		panic(err)
	}

	return src, dest
}

func ParsePriorityKey(key []byte) (source, destination string, err error) {
	if len(key) == 0 {
		return "", "", fmt.Errorf("empty key received")
	}

	if !bytes.HasPrefix(key, priorityPrefix) {
		return "", "", fmt.Errorf("invalid prefix: %v", hex.EncodeToString(key))
	}

	a := strings.Split(string(key[1:]), "/")
	return a[0], a[1], nil
}

func GetOwnersPrefix() []byte {
	return ownerPrefix
}

func GetOwnerKey(acc, clientOrderId string) []byte {
	res := append(GetOwnersPrefix(), []byte(acc)...)
	res = append(res, []byte(clientOrderId)...)
	return res
}

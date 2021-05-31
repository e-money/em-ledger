// This software is Copyright (c) 2019-2020 e-Money A/S. It is not offered under an open source license.
//
// Please contact partners@e-money.com for licensing related questions.

package bank

import (
	"bytes"
	"errors"

	sdk "github.com/cosmos/cosmos-sdk/types"
	bankkeeper "github.com/cosmos/cosmos-sdk/x/bank/keeper"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/tendermint/tendermint/libs/rand"
	"testing"
)

func TestDeduplicate(t *testing.T) {
	var (
		addr1 = bytes.Repeat([]byte{0x1}, sdk.AddrLen)
		addr2 = bytes.Repeat([]byte{0x2}, sdk.AddrLen)
		addr3 = bytes.Repeat([]byte{0x3}, sdk.AddrLen)
	)
	specs := map[string]struct {
		src []sdk.AccAddress
		exp []sdk.AccAddress
	}{
		"without duplicates": {
			src: []sdk.AccAddress{addr1, addr2, addr3},
			exp: []sdk.AccAddress{addr1, addr2, addr3},
		},
		"duplicates": {
			src: []sdk.AccAddress{addr1, addr1, addr2, addr3},
			exp: []sdk.AccAddress{addr1, addr2, addr3},
		},
		"more duplicates": {
			src: []sdk.AccAddress{addr1, addr2, addr3, addr1, addr2, addr3},
			exp: []sdk.AccAddress{addr1, addr2, addr3},
		},
		"empty": {
			src: []sdk.AccAddress{},
			exp: []sdk.AccAddress{},
		},
		"nil": {
			exp: []sdk.AccAddress{},
		},
	}
	for name, spec := range specs {
		t.Run(name, func(t *testing.T) {
			got := deduplicate(spec.src)
			assert.Equal(t, spec.exp, got)
		})
	}
}

func TestSendCoinsBalanceUpdateNotification(t *testing.T) {
	var (
		ctx   sdk.Context
		addr1 = randomAddress()
		addr2 = randomAddress()
	)

	specs := map[string]struct {
		srcSender          sdk.AccAddress
		srcRecvr           sdk.AccAddress
		listenerCount      int
		nestedKeeperResult error

		expAddr []sdk.AccAddress
		expErr  bool
	}{
		"one listener called": {
			srcSender:     addr1,
			srcRecvr:      addr2,
			listenerCount: 1,
			expAddr:       []sdk.AccAddress{addr1, addr2},
		},
		"multiple listener called": {
			srcSender:     addr1,
			srcRecvr:      addr2,
			listenerCount: 2,
			expAddr:       []sdk.AccAddress{addr1, addr2},
		},
		"no listener called on error": {
			srcSender:          addr1,
			srcRecvr:           addr2,
			listenerCount:      2,
			nestedKeeperResult: errors.New("test, ignore"),
			expErr:             true,
		},
	}
	for name, spec := range specs {
		t.Run(name, func(t *testing.T) {
			receivedAddr := make([][][]sdk.AccAddress, spec.listenerCount)
			newListener := func(listenerNb int) func(sdk.Context, []sdk.AccAddress) {
				return func(_ sdk.Context, addrs []sdk.AccAddress) {
					receivedAddr[listenerNb] = append(receivedAddr[listenerNb], addrs)
				}
			}

			nestedBk := senderBankKeeperMock{
				SendCoinsFn: func(ctx sdk.Context, fromAddr sdk.AccAddress, toAddr sdk.AccAddress, amt sdk.Coins) error {
					return spec.nestedKeeperResult
				},
			}
			wrappedBankKeeper := Wrap(nestedBk)

			// register listeners
			for i := 0; i < spec.listenerCount; i++ {
				wrappedBankKeeper.AddBalanceListener(newListener(i))
			}

			// when
			gotErr := wrappedBankKeeper.SendCoins(ctx, spec.srcSender, spec.srcRecvr, coins("1token"))

			// then
			if spec.expErr {
				require.Error(t, gotErr)
				return
			}
			require.NoError(t, gotErr)
			for i := 0; i < spec.listenerCount; i++ {
				require.Len(t, receivedAddr[i], 1)
				assert.Equal(t, spec.expAddr, receivedAddr[i][0])
			}
		})
	}
}

func TestInputOutputCoinsBalanceUpdateNotification(t *testing.T) {
	var (
		ctx   sdk.Context
		addr1 = randomAddress()
		addr2 = randomAddress()
	)

	specs := map[string]struct {
		srcInput           []banktypes.Input
		srcOutput          []banktypes.Output
		listenerCount      int
		nestedKeeperResult error

		expAddr []sdk.AccAddress
		expErr  bool
	}{
		"one listener called": {
			srcInput:      []banktypes.Input{{Address: addr1.String()}},
			srcOutput:     []banktypes.Output{{Address: addr2.String()}},
			listenerCount: 1,
			expAddr:       []sdk.AccAddress{addr1, addr2},
		},
		"multiple listener called": {
			srcInput:      []banktypes.Input{{Address: addr1.String()}},
			srcOutput:     []banktypes.Output{{Address: addr2.String()}},
			listenerCount: 2,
			expAddr:       []sdk.AccAddress{addr1, addr2},
		},
		"no listener called on error": {
			srcInput:           []banktypes.Input{{Address: addr1.String()}},
			srcOutput:          []banktypes.Output{{Address: addr2.String()}},
			listenerCount:      2,
			nestedKeeperResult: errors.New("test, ignore"),
			expErr:             true,
		},
		"deduplicated": {
			srcInput:      []banktypes.Input{{Address: addr1.String()}, {Address: addr1.String()}, {Address: addr2.String()}},
			srcOutput:     []banktypes.Output{{Address: addr2.String()}, {Address: addr2.String()}, {Address: addr1.String()}},
			listenerCount: 2,
			expAddr:       []sdk.AccAddress{addr1, addr2},
		},
	}
	for name, spec := range specs {
		t.Run(name, func(t *testing.T) {
			receivedAddr := make([][][]sdk.AccAddress, spec.listenerCount)
			newListener := func(listenerNb int) func(sdk.Context, []sdk.AccAddress) {
				return func(_ sdk.Context, addrs []sdk.AccAddress) {
					receivedAddr[listenerNb] = append(receivedAddr[listenerNb], addrs)
				}
			}

			nestedBk := senderBankKeeperMock{
				InputOutputCoinsFn: func(ctx sdk.Context, in []banktypes.Input, out []banktypes.Output) error {
					return spec.nestedKeeperResult
				},
			}
			wrappedBankKeeper := Wrap(nestedBk)

			// register listeners
			for i := 0; i < spec.listenerCount; i++ {
				wrappedBankKeeper.AddBalanceListener(newListener(i))
			}

			// when
			gotErr := wrappedBankKeeper.InputOutputCoins(ctx, spec.srcInput, spec.srcOutput)

			// then
			if spec.expErr {
				require.Error(t, gotErr)
				return
			}
			require.NoError(t, gotErr)
			for i := 0; i < spec.listenerCount; i++ {
				require.Len(t, receivedAddr[i], 1)
				assert.Equal(t, spec.expAddr, receivedAddr[i][0])
			}
		})
	}
}

func TestSendCoinsFromModuleToAccount(t *testing.T) {
	var (
		ctx   sdk.Context
		addr1 = randomAddress()
	)

	specs := map[string]struct {
		senderModule       string
		recipientAddr      sdk.AccAddress
		listenerCount      int
		nestedKeeperResult error

		expAddr []sdk.AccAddress
		expErr  bool
	}{
		"one listener called": {
			senderModule:  "anyModule",
			recipientAddr: addr1,
			listenerCount: 1,
			expAddr:       []sdk.AccAddress{addr1},
		},
		"multiple listener called": {
			senderModule:  "anyModule",
			recipientAddr: addr1,
			listenerCount: 2,
			expAddr:       []sdk.AccAddress{addr1},
		},
		"no listener called on error": {
			senderModule:       "anyModule",
			recipientAddr:      addr1,
			listenerCount:      2,
			nestedKeeperResult: errors.New("test, ignore"),
			expErr:             true,
		},
	}
	for name, spec := range specs {
		t.Run(name, func(t *testing.T) {
			receivedAddr := make([][][]sdk.AccAddress, spec.listenerCount)
			newListener := func(listenerNb int) func(sdk.Context, []sdk.AccAddress) {
				return func(_ sdk.Context, addrs []sdk.AccAddress) {
					receivedAddr[listenerNb] = append(receivedAddr[listenerNb], addrs)
				}
			}

			nestedBk := senderBankKeeperMock{
				SendCoinsFromModuleToAccountFn: func(ctx sdk.Context, senderModule string, recipientAddr sdk.AccAddress, amt sdk.Coins) error {
					return spec.nestedKeeperResult
				},
			}
			wrappedBankKeeper := Wrap(nestedBk)

			// register listeners
			for i := 0; i < spec.listenerCount; i++ {
				wrappedBankKeeper.AddBalanceListener(newListener(i))
			}

			amt := sdk.NewCoins(sdk.NewCoin("alx", sdk.OneInt()), sdk.NewCoin("blx", sdk.NewInt(2)))
			// when
			gotErr := wrappedBankKeeper.SendCoinsFromModuleToAccount(ctx, spec.senderModule, spec.recipientAddr, amt)

			// then
			if spec.expErr {
				require.Error(t, gotErr)
				return
			}
			require.NoError(t, gotErr)
			for i := 0; i < spec.listenerCount; i++ {
				require.Len(t, receivedAddr[i], 1)
				assert.Equal(t, spec.expAddr, receivedAddr[i][0])
			}
		})
	}
}

func TestSendCoinsFromAccountToModule(t *testing.T) {
	var (
		ctx   sdk.Context
		addr1 = randomAddress()
	)

	specs := map[string]struct {
		reciepientModule   string
		senderAddr         sdk.AccAddress
		listenerCount      int
		nestedKeeperResult error

		expAddr []sdk.AccAddress
		expErr  bool
	}{
		"one listener called": {
			reciepientModule: "anyModule",
			senderAddr:       addr1,
			listenerCount:    1,
			expAddr:          []sdk.AccAddress{addr1},
		},
		"multiple listener called": {
			reciepientModule: "anyModule",
			senderAddr:       addr1,
			listenerCount:    2,
			expAddr:          []sdk.AccAddress{addr1},
		},
		"no listener called on error": {
			reciepientModule:   "anyModule",
			senderAddr:         addr1,
			listenerCount:      2,
			nestedKeeperResult: errors.New("test, ignore"),
			expErr:             true,
		},
	}
	for name, spec := range specs {
		t.Run(name, func(t *testing.T) {
			receivedAddr := make([][][]sdk.AccAddress, spec.listenerCount)
			newListener := func(listenerNb int) func(sdk.Context, []sdk.AccAddress) {
				return func(_ sdk.Context, addrs []sdk.AccAddress) {
					receivedAddr[listenerNb] = append(receivedAddr[listenerNb], addrs)
				}
			}

			nestedBk := senderBankKeeperMock{
				SendCoinsFromAccountToModuleFn: func(ctx sdk.Context, senderAddr sdk.AccAddress, recipientModule string, amt sdk.Coins) error {
					return spec.nestedKeeperResult
				},
			}
			wrappedBankKeeper := Wrap(nestedBk)

			// register listeners
			for i := 0; i < spec.listenerCount; i++ {
				wrappedBankKeeper.AddBalanceListener(newListener(i))
			}

			amt := sdk.NewCoins(sdk.NewCoin("alx", sdk.OneInt()), sdk.NewCoin("blx", sdk.NewInt(2)))

			// when
			gotErr := wrappedBankKeeper.SendCoinsFromAccountToModule(ctx, spec.senderAddr, spec.reciepientModule, amt)

			// then
			if spec.expErr {
				require.Error(t, gotErr)
				return
			}
			require.NoError(t, gotErr)
			for i := 0; i < spec.listenerCount; i++ {
				require.Len(t, receivedAddr[i], 1)
				assert.Equal(t, spec.expAddr, receivedAddr[i][0])
			}
		})
	}
}

func TestDelegateCoinsFromAccountToModule(t *testing.T) {
	var (
		ctx   sdk.Context
		addr1 = randomAddress()
	)

	specs := map[string]struct {
		reciepientModule   string
		senderAddr         sdk.AccAddress
		listenerCount      int
		nestedKeeperResult error

		expAddr []sdk.AccAddress
		expErr  bool
	}{
		"one listener called": {
			reciepientModule: "anyModule",
			senderAddr:       addr1,
			listenerCount:    1,
			expAddr:          []sdk.AccAddress{addr1},
		},
		"multiple listener called": {
			reciepientModule: "anyModule",
			senderAddr:       addr1,
			listenerCount:    2,
			expAddr:          []sdk.AccAddress{addr1},
		},
		"no listener called on error": {
			reciepientModule:   "anyModule",
			senderAddr:         addr1,
			listenerCount:      2,
			nestedKeeperResult: errors.New("test, ignore"),
			expErr:             true,
		},
	}
	for name, spec := range specs {
		t.Run(name, func(t *testing.T) {
			receivedAddr := make([][][]sdk.AccAddress, spec.listenerCount)
			newListener := func(listenerNb int) func(sdk.Context, []sdk.AccAddress) {
				return func(_ sdk.Context, addrs []sdk.AccAddress) {
					receivedAddr[listenerNb] = append(receivedAddr[listenerNb], addrs)
				}
			}

			nestedBk := senderBankKeeperMock{
				DelegateCoinsFromAccountToModuleFn: func(ctx sdk.Context, senderAddr sdk.AccAddress, recipientModule string, amt sdk.Coins) error {
					return spec.nestedKeeperResult
				},
			}
			wrappedBankKeeper := Wrap(nestedBk)

			// register listeners
			for i := 0; i < spec.listenerCount; i++ {
				wrappedBankKeeper.AddBalanceListener(newListener(i))
			}

			amt := sdk.NewCoins(sdk.NewCoin("alx", sdk.OneInt()), sdk.NewCoin("blx", sdk.NewInt(2)))

			// when
			gotErr := wrappedBankKeeper.DelegateCoinsFromAccountToModule(ctx, spec.senderAddr, spec.reciepientModule, amt)

			// then
			if spec.expErr {
				require.Error(t, gotErr)
				return
			}
			require.NoError(t, gotErr)
			for i := 0; i < spec.listenerCount; i++ {
				require.Len(t, receivedAddr[i], 1)
				assert.Equal(t, spec.expAddr, receivedAddr[i][0])
			}
		})
	}
}

func TestUndelegateCoinsFromModuleToAccount(t *testing.T) {
	var (
		ctx   sdk.Context
		addr1 = randomAddress()
	)

	specs := map[string]struct {
		senderModule       string
		recipientAddr      sdk.AccAddress
		listenerCount      int
		nestedKeeperResult error

		expAddr []sdk.AccAddress
		expErr  bool
	}{
		"one listener called": {
			senderModule:  "anyModule",
			recipientAddr: addr1,
			listenerCount: 1,
			expAddr:       []sdk.AccAddress{addr1},
		},
		"multiple listener called": {
			senderModule:  "anyModule",
			recipientAddr: addr1,
			listenerCount: 2,
			expAddr:       []sdk.AccAddress{addr1},
		},
		"no listener called on error": {
			senderModule:       "anyModule",
			recipientAddr:      addr1,
			listenerCount:      2,
			nestedKeeperResult: errors.New("test, ignore"),
			expErr:             true,
		},
	}
	for name, spec := range specs {
		t.Run(name, func(t *testing.T) {
			receivedAddr := make([][][]sdk.AccAddress, spec.listenerCount)
			newListener := func(listenerNb int) func(sdk.Context, []sdk.AccAddress) {
				return func(_ sdk.Context, addrs []sdk.AccAddress) {
					receivedAddr[listenerNb] = append(receivedAddr[listenerNb], addrs)
				}
			}

			nestedBk := senderBankKeeperMock{
				UndelegateCoinsFromModuleToAccountFn: func(ctx sdk.Context, senderModule string, recipientAddr sdk.AccAddress, amt sdk.Coins) error {
					return spec.nestedKeeperResult
				},
			}
			wrappedBankKeeper := Wrap(nestedBk)

			// register listeners
			for i := 0; i < spec.listenerCount; i++ {
				wrappedBankKeeper.AddBalanceListener(newListener(i))
			}

			amt := sdk.NewCoins(sdk.NewCoin("alx", sdk.OneInt()), sdk.NewCoin("blx", sdk.NewInt(2)))
			// when
			gotErr := wrappedBankKeeper.UndelegateCoinsFromModuleToAccount(ctx, spec.senderModule, spec.recipientAddr, amt)

			// then
			if spec.expErr {
				require.Error(t, gotErr)
				return
			}
			require.NoError(t, gotErr)
			for i := 0; i < spec.listenerCount; i++ {
				require.Len(t, receivedAddr[i], 1)
				assert.Equal(t, spec.expAddr, receivedAddr[i][0])
			}
		})
	}
}

func TestDelegateCoins(t *testing.T) {
	var (
		ctx       sdk.Context
		addr1     = randomAddress()
		moduleAcc = randomAddress()
	)

	specs := map[string]struct {
		moduleAddr         sdk.AccAddress
		delegatorAddr      sdk.AccAddress
		listenerCount      int
		nestedKeeperResult error

		expAddr []sdk.AccAddress
		expErr  bool
	}{
		"one listener called": {
			moduleAddr:    moduleAcc,
			delegatorAddr: addr1,
			listenerCount: 1,
			expAddr:       []sdk.AccAddress{addr1},
		},
		"multiple listener called": {
			moduleAddr:    moduleAcc,
			delegatorAddr: addr1,
			listenerCount: 2,
			expAddr:       []sdk.AccAddress{addr1},
		},
		"no listener called on error": {
			moduleAddr:         moduleAcc,
			delegatorAddr:      addr1,
			listenerCount:      2,
			nestedKeeperResult: errors.New("test, ignore"),
			expErr:             true,
		},
	}
	for name, spec := range specs {
		t.Run(name, func(t *testing.T) {
			receivedAddr := make([][][]sdk.AccAddress, spec.listenerCount)
			newListener := func(listenerNb int) func(sdk.Context, []sdk.AccAddress) {
				return func(_ sdk.Context, addrs []sdk.AccAddress) {
					receivedAddr[listenerNb] = append(receivedAddr[listenerNb], addrs)
				}
			}

			nestedBk := senderBankKeeperMock{
				DelegateCoinsFn: func(ctx sdk.Context, delegatorAddr, moduleAccAddr sdk.AccAddress, amt sdk.Coins) error {
					return spec.nestedKeeperResult
				},
			}
			wrappedBankKeeper := Wrap(nestedBk)

			// register listeners
			for i := 0; i < spec.listenerCount; i++ {
				wrappedBankKeeper.AddBalanceListener(newListener(i))
			}

			amt := sdk.NewCoins(sdk.NewCoin("alx", sdk.OneInt()), sdk.NewCoin("blx", sdk.NewInt(2)))

			// when
			gotErr := wrappedBankKeeper.DelegateCoins(ctx, spec.delegatorAddr, spec.moduleAddr, amt)

			// then
			if spec.expErr {
				require.Error(t, gotErr)
				return
			}
			require.NoError(t, gotErr)
			for i := 0; i < spec.listenerCount; i++ {
				require.Len(t, receivedAddr[i], 1)
				assert.Equal(t, spec.expAddr, receivedAddr[i][0])
			}
		})
	}
}

func TestUndelegateCoins(t *testing.T) {
	var (
		ctx       sdk.Context
		addr1     = randomAddress()
		moduleAcc = randomAddress()
	)

	specs := map[string]struct {
		moduleAddr         sdk.AccAddress
		delegatorAddr      sdk.AccAddress
		listenerCount      int
		nestedKeeperResult error

		expAddr []sdk.AccAddress
		expErr  bool
	}{
		"one listener called": {
			moduleAddr:    moduleAcc,
			delegatorAddr: addr1,
			listenerCount: 1,
			expAddr:       []sdk.AccAddress{addr1},
		},
		"multiple listener called": {
			moduleAddr:    moduleAcc,
			delegatorAddr: addr1,
			listenerCount: 2,
			expAddr:       []sdk.AccAddress{addr1},
		},
		"no listener called on error": {
			moduleAddr:         moduleAcc,
			delegatorAddr:      addr1,
			listenerCount:      2,
			nestedKeeperResult: errors.New("test, ignore"),
			expErr:             true,
		},
	}
	for name, spec := range specs {
		t.Run(name, func(t *testing.T) {
			receivedAddr := make([][][]sdk.AccAddress, spec.listenerCount)
			newListener := func(listenerNb int) func(sdk.Context, []sdk.AccAddress) {
				return func(_ sdk.Context, addrs []sdk.AccAddress) {
					receivedAddr[listenerNb] = append(receivedAddr[listenerNb], addrs)
				}
			}

			nestedBk := senderBankKeeperMock{
				UndelegateCoinsFn: func(ctx sdk.Context, moduleAccAddr, delegatorAddr sdk.AccAddress, amt sdk.Coins) error {
					return spec.nestedKeeperResult
				},
			}
			wrappedBankKeeper := Wrap(nestedBk)

			// register listeners
			for i := 0; i < spec.listenerCount; i++ {
				wrappedBankKeeper.AddBalanceListener(newListener(i))
			}

			amt := sdk.NewCoins(sdk.NewCoin("alx", sdk.OneInt()), sdk.NewCoin("blx", sdk.NewInt(2)))

			// when
			gotErr := wrappedBankKeeper.UndelegateCoins(ctx, spec.moduleAddr, spec.delegatorAddr, amt)

			// then
			if spec.expErr {
				require.Error(t, gotErr)
				return
			}
			require.NoError(t, gotErr)
			for i := 0; i < spec.listenerCount; i++ {
				require.Len(t, receivedAddr[i], 1)
				assert.Equal(t, spec.expAddr, receivedAddr[i][0])
			}
		})
	}
}

type senderBankKeeperMock struct {
	bankkeeper.Keeper
	InputOutputCoinsFn                   func(ctx sdk.Context, inputs []banktypes.Input, outputs []banktypes.Output) error
	SendCoinsFn                          func(ctx sdk.Context, fromAddr sdk.AccAddress, toAddr sdk.AccAddress, amt sdk.Coins) error
	SendCoinsFromModuleToAccountFn       func(ctx sdk.Context, senderModule string, recipientAddr sdk.AccAddress, amt sdk.Coins) error
	SendCoinsFromAccountToModuleFn       func(ctx sdk.Context, senderAddr sdk.AccAddress, recipientModule string, amt sdk.Coins) error
	DelegateCoinsFromAccountToModuleFn   func(ctx sdk.Context, senderAddr sdk.AccAddress, recipientModule string, amt sdk.Coins) error
	UndelegateCoinsFromModuleToAccountFn func(ctx sdk.Context, senderModule string, recipientAddr sdk.AccAddress, amt sdk.Coins) error
	DelegateCoinsFn                      func(ctx sdk.Context, delegatorAddr, moduleAccAddr sdk.AccAddress, amt sdk.Coins) error
	UndelegateCoinsFn                    func(ctx sdk.Context, moduleAccAddr, delegatorAddr sdk.AccAddress, amt sdk.Coins) error
}

func (pk senderBankKeeperMock) SendCoins(ctx sdk.Context, fromAddr sdk.AccAddress, toAddr sdk.AccAddress, amt sdk.Coins) error {
	if pk.SendCoinsFn == nil {
		panic("unexpected call")
	}
	return pk.SendCoinsFn(ctx, fromAddr, toAddr, amt)
}

func (pk senderBankKeeperMock) InputOutputCoins(ctx sdk.Context, inputs []banktypes.Input, outputs []banktypes.Output) error {
	if pk.InputOutputCoinsFn == nil {
		panic("unexpected call")
	}
	return pk.InputOutputCoinsFn(ctx, inputs, outputs)
}

func (m senderBankKeeperMock) SendCoinsFromModuleToAccount(ctx sdk.Context, senderModule string, recipientAddr sdk.AccAddress, amt sdk.Coins) error {
	if m.SendCoinsFromModuleToAccountFn == nil {
		panic("not expected to be called")
	}
	return m.SendCoinsFromModuleToAccountFn(ctx, senderModule, recipientAddr, amt)
}

func (m senderBankKeeperMock) SendCoinsFromAccountToModule(ctx sdk.Context, senderAddr sdk.AccAddress, recipientModule string, amt sdk.Coins) error {
	if m.SendCoinsFromAccountToModuleFn == nil {
		panic("not expected to be called")
	}
	return m.SendCoinsFromAccountToModuleFn(ctx, senderAddr, recipientModule, amt)
}

func (m senderBankKeeperMock) DelegateCoinsFromAccountToModule(ctx sdk.Context, senderAddr sdk.AccAddress, recipientModule string, amt sdk.Coins) error {
	if m.DelegateCoinsFromAccountToModuleFn == nil {
		panic("not expected to be called")
	}
	return m.DelegateCoinsFromAccountToModuleFn(ctx, senderAddr, recipientModule, amt)
}

func (m senderBankKeeperMock) UndelegateCoinsFromModuleToAccount(ctx sdk.Context, senderModule string, recipientAddr sdk.AccAddress, amt sdk.Coins) error {
	if m.UndelegateCoinsFromModuleToAccountFn == nil {
		panic("not expected to be called")
	}
	return m.UndelegateCoinsFromModuleToAccountFn(ctx, senderModule, recipientAddr, amt)
}

func (m senderBankKeeperMock) DelegateCoins(ctx sdk.Context, delegatorAddr, moduleAccAddr sdk.AccAddress, amt sdk.Coins) error {
	if m.DelegateCoinsFn == nil {
		panic("not expected to be called")
	}
	return m.DelegateCoinsFn(ctx, delegatorAddr, moduleAccAddr, amt)
}

func (m senderBankKeeperMock) UndelegateCoins(ctx sdk.Context, moduleAccAddr, delegatorAddr sdk.AccAddress, amt sdk.Coins) error {
	if m.UndelegateCoinsFn == nil {
		panic("not expected to be called")
	}
	return m.UndelegateCoinsFn(ctx, moduleAccAddr, delegatorAddr, amt)
}

func randomAddress() sdk.AccAddress {
	return rand.Bytes(sdk.AddrLen)
}

func coins(s string) sdk.Coins {
	coins, err := sdk.ParseCoinsNormalized(s)
	if err != nil {
		panic(err)
	}
	return coins
}

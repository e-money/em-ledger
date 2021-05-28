// This software is Copyright (c) 2019-2020 e-Money A/S. It is not offered under an open source license.
//
// Please contact partners@e-money.com for licensing related questions.

// +build bdd

package networktest

import (
	"context"
	"fmt"
	"time"

	abcitypes "github.com/tendermint/tendermint/abci/types"
	client "github.com/tendermint/tendermint/rpc/client/http"
	ct "github.com/tendermint/tendermint/rpc/core/types"
	"github.com/tendermint/tendermint/types"
)

type EventListener struct {
	client *client.HTTP
}

func NewEventListener() (EventListener, error) {
	httpClient, err := client.New("http://localhost:26657", "/websocket")
	if err != nil {
		return EventListener{}, err
	}

	if err := httpClient.Start(); err != nil {
		return EventListener{}, err
	}

	return EventListener{client: httpClient}, nil
}

func (el EventListener) subscribeQuery(query string, listener func(ct.ResultEvent) bool) error {
	ctx := context.Background()
	eventChannel, err := el.client.WSEvents.Subscribe(ctx, "", query)
	if err != nil {
		return err
	}

	go func() {
		defer el.client.WSEvents.Unsubscribe(ctx, "", query)
		for {
			select {
			case evt := <-eventChannel:
				if !listener(evt) {
					return
				}
			case <-time.After(time.Minute):
				fmt.Println(" *** Timeout waiting for event.", query)
				return
			}
		}
	}()

	return nil
}

func (el EventListener) awaitQuery(query string) (func() *ct.ResultEvent, error) {
	ctx := context.Background()
	eventChannel, err := el.client.WSEvents.Subscribe(ctx, "", query)
	if err != nil {
		return nil, err
	}

	res := func() *ct.ResultEvent {
		defer el.client.WSEvents.Unsubscribe(ctx, "", query)

		select {
		case evt := <-eventChannel:
			return &evt
		case <-time.After(time.Minute):
			fmt.Println(" *** Timeout waiting for event.", query)
			return nil
		}
	}

	return res, nil
}

func (el EventListener) AwaitSlash() (func() *abcitypes.Event, error) {
	// todo (reviewer): please note the new query format. See https://docs.cosmos.network/master/core/events.html#subscribing-to-events
	eventFn, err := el.awaitQuery("tm.event='NewBlock' AND slash.reason='missing_signature'")
	if err != nil {
		return nil, err
	}

	return func() *abcitypes.Event {
		event := eventFn()

		if event == nil {
			return nil
		}

		newBlockEventData, ok := event.Data.(types.EventDataNewBlock)
		if !ok {
			fmt.Printf("Unexpected event type data received: %T\n", event.Data)
			return nil
		}

		for _, e := range newBlockEventData.ResultBeginBlock.Events {
			if e.Type == "slash" {
				return &e
			}
		}

		return nil
	}, nil
}

func (el EventListener) AwaitValidatorSetChange() (func() *types.EventDataValidatorSetUpdates, error) {
	eventFn, err := el.awaitQuery("tm.event='ValidatorSetUpdates'")
	if err != nil {
		return nil, err
	}

	return func() *types.EventDataValidatorSetUpdates {
		event := eventFn()

		validatorSetUpdates, ok := event.Data.(types.EventDataValidatorSetUpdates)
		if ok {
			return &validatorSetUpdates
		}

		return nil
	}, nil
}

package manager

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/ThomasP1988/go-lambda-graphql-subscription/common"
	"github.com/aws/aws-lambda-go/events"
	"github.com/google/uuid"
	"github.com/graphql-go/graphql"
)

var CurrentManager *Manager

type Manager struct {
	Schema             *graphql.Schema
	Connection         ConnectionManager
	Subscription       SubscriptionManager
	Event              EventManager
	OnWebsocketConnect *func(*events.APIGatewayWebsocketProxyRequest) interface{}
	OnConnect          *func(*events.APIGatewayWebsocketProxyRequest) interface{}
	OnDisconnect       *func(*Connection)
}

type SetManagerArgs struct {
	Schema             *graphql.Schema
	Connection         ConnectionManager
	Subscription       SubscriptionManager
	Event              EventManager
	OnWebsocketConnect *func(*events.APIGatewayWebsocketProxyRequest) interface{}
	OnConnect          *func(*events.APIGatewayWebsocketProxyRequest) interface{}
	OnDisconnect       *func(*Connection)
}

func SetManager(args *SetManagerArgs) {
	CurrentManager = &Manager{
		Connection:         args.Connection,
		Subscription:       args.Subscription,
		Schema:             args.Schema,
		Event:              args.Event,
		OnWebsocketConnect: args.OnWebsocketConnect,
		OnConnect:          args.OnConnect,
		OnDisconnect:       args.OnDisconnect,
	}
}

func Sub(keys []string, ctx context.Context) (interface{}, error) {
	fmt.Printf("keys: %v\n", keys)

	if CurrentManager == nil {
		println("missing manager")
		return nil, errors.New("manager isn't set")
	}

	for _, v := range keys {

		sbs := &Subscription{
			Event:         v,
			ConnectionID:  ctx.Value(WSContextKey).(WSlambdaContext).ConnectionID,
			Query:         ctx.Value(WSContextKey).(WSlambdaContext).RequestString,
			Variables:     ctx.Value(WSContextKey).(WSlambdaContext).Variables,
			OperationName: ctx.Value(WSContextKey).(WSlambdaContext).OperationName,
			OperationID:   ctx.Value(WSContextKey).(WSlambdaContext).OperationID,
		}

		fmt.Printf("sbs: %v\n", sbs)
		fmt.Printf("\"ici\": %v\n", "ici")
		err := CurrentManager.Subscription.Start(ctx, sbs)
		fmt.Printf("\"la\": %v\n", "la")

		if err != nil {
			fmt.Printf("err: %v\n", err)
			println(err)
			return nil, err
		}

		fmt.Printf("sbs: %v\n", sbs)
	}

	response := make(chan interface{})

	go func() {
		response <- (func() (bool, interface{}) { return false, nil })
	}()
	return response, nil
}

func Pub(ctx context.Context, key string, payload interface{}) error {
	if CurrentManager == nil {
		return errors.New("manager isn't set")
	}

	eventId, _ := uuid.NewUUID()

	jsonPayload, err := json.Marshal(payload)

	if err != nil {
		return err
	}

	newEvent := &Event{
		ID:      eventId.String(),
		Key:     key,
		Payload: string(jsonPayload),
	}

	err = CurrentManager.Event.Add(ctx, newEvent)

	if err != nil {
		return err
	}

	return nil
}

func Execute(
	ctx context.Context,
	operationID string,
	connectionID string,
	domainName string,
	stage string,
	params *graphql.Params,
) {
	result := graphql.Do(*params)

	fmt.Printf("result: %v\n", result)
	message, err := json.Marshal(map[string]interface{}{
		"type":    "data",
		"id":      operationID,
		"payload": result,
	})

	if err != nil {
		println("err marshalling message", err)
	}

	err = common.SendMessage(ctx, connectionID, domainName, stage, message)

	if err != nil {
		fmt.Printf("err sending message: %v\n", err)
	}
}

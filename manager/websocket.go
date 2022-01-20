package manager

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/ThomasP1988/go-lambda-graphql-subscription/common"
	"github.com/aws/aws-lambda-go/events"
	"github.com/graphql-go/graphql"
)

const (
	RouteKeyConnect    string = "$connect"
	RouteKeyDisconnect string = "$disconnect"
	RouteKeyDefault    string = "$default"
)

const (
	// client
	GraphQLSubStart            string = "start"
	GraphQLSubStop             string = "stop"
	GraphQLConnectionInit      string = "connection_init"
	GraphQLConnectionTerminate string = "connection_terminate"
	// server
	GraphQLConnectionACK string = "connection_ack"
	GraphQLError         string = "error"
	GraphQLData          string = "data"
	GraphQLComplete      string = "complete"
)

type ContextKey int

const WSContextKey ContextKey = iota

func HandleWS(ctx context.Context, req events.APIGatewayWebsocketProxyRequest) (events.APIGatewayProxyResponse, error) {
	println("req.RequestContext.RouteKey", req.RequestContext.RouteKey)

	if req.RequestContext.RouteKey == RouteKeyConnect {

		// TODO: authorize to store extra data (cognito userId for example)

		newConnection := &Connection{
			Id:     req.RequestContext.ConnectionID,
			Stage:  req.RequestContext.Stage,
			Domain: req.RequestContext.DomainName,
		}

		if CurrentManager.OnWebsocketConnect != nil {
			response := (*CurrentManager.OnWebsocketConnect)(&req)
			fmt.Printf("response: %v\n", response)
			newConnection.WebsocketConnectContext = response
		}

		err := CurrentManager.Connection.OnConnect(ctx, newConnection)

		if err != nil {

			fmt.Printf("err: %v\n", err)
			println(err)

			return events.APIGatewayProxyResponse{
				Body:       "",
				StatusCode: http.StatusInternalServerError,
			}, nil
		}

		headers := map[string]string{}

		if strings.Contains(req.Headers["Sec-WebSocket-Protocol"], "graphql-ws") {
			headers["Sec-WebSocket-Protocol"] = "graphql-ws"
		}

		return events.APIGatewayProxyResponse{
			Body:       "",
			StatusCode: http.StatusOK,
			Headers:    headers,
		}, nil
	} else if req.RequestContext.RouteKey == RouteKeyDisconnect {

		connection, err := CurrentManager.Connection.Get(ctx, req.RequestContext.ConnectionID)
		if err != nil {
			log.Printf("failed to find connection, message: %v", err)
			return events.APIGatewayProxyResponse{
				Body:       "failed to find connection",
				StatusCode: http.StatusNotFound,
			}, nil
		}

		if CurrentManager.OnDisconnect != nil {
			(*CurrentManager.OnDisconnect)(connection)
		}

		CurrentManager.Connection.OnDisconnect(ctx, req.RequestContext.ConnectionID)
		return events.APIGatewayProxyResponse{
			Body:       "",
			StatusCode: http.StatusOK,
		}, nil
	} else if req.RequestContext.RouteKey == RouteKeyDefault {
		fmt.Printf("req.Body: %v\n", req.Body)
		var msg Message
		if err := json.Unmarshal([]byte(req.Body), &msg); err != nil {
			log.Printf("failed to unmarshal websocket message: %v", err)
		}

		// TODO: refresh connection TTL

		if msg.Type == GraphQLConnectionInit {
			println("connection init")
			message, err := json.Marshal(Message{
				Type: GraphQLConnectionACK,
			})

			if err != nil {
				fmt.Printf("error unmarshalling connection init message: %v\n", err)
			}

			var connectContext interface{}

			if CurrentManager.OnConnect != nil {
				connectContext = (*CurrentManager.OnConnect)(&req)
				fmt.Printf("connectContext: %v\n", connectContext)

				if connectContext == false {
					return events.APIGatewayProxyResponse{
						Body:       "connection prohibited",
						StatusCode: http.StatusForbidden,
					}, nil
				}

			}

			err = CurrentManager.Connection.Init(ctx, req.RequestContext.ConnectionID, connectContext)

			if err == nil {
				err = common.SendMessage(ctx, req.RequestContext.ConnectionID, req.RequestContext.DomainName, req.RequestContext.Stage, message)
				if err != nil {
					fmt.Printf("err sending message: %v\n", err)
				}
			} else {
				fmt.Printf("err: %v\n", err)
				println(err)
			}
		} else if msg.Type == GraphQLConnectionTerminate {
			println("connection terminate")
			CurrentManager.Connection.Terminate(ctx, req.RequestContext.ConnectionID)
		}

		err := CurrentManager.Connection.Hydrate(ctx, req.RequestContext.ConnectionID)

		if err != nil {
			fmt.Printf("error hydrating connection: %v\n", err)
		}

		connection, err := CurrentManager.Connection.Get(ctx, req.RequestContext.ConnectionID)
		if err != nil {
			log.Printf("failed to find connection, message: %v", err)
			return events.APIGatewayProxyResponse{
				Body:       "failed to find connection",
				StatusCode: http.StatusNotFound,
			}, nil
		}

		fmt.Printf("connection: %v\n", connection)
		if connection == nil || time.Unix(connection.Ttl, 0).Before(time.Now()) || !connection.IsInitialized {
			return events.APIGatewayProxyResponse{
				Body:       "connection expired",
				StatusCode: http.StatusGatewayTimeout,
			}, nil
		}

		if msg.Type == GraphQLSubStop {
			// remove from table
			println("stop")
			err := CurrentManager.Subscription.Stop(ctx, req.RequestContext.ConnectionID, msg.OperationID)
			if err != nil {
				return events.APIGatewayProxyResponse{
					Body:       "",
					StatusCode: http.StatusInternalServerError,
				}, nil
			}
		} else if msg.Type == GraphQLSubStart {
			// add to table
			println("start")

			if strings.HasPrefix(msg.Payload.Query, "subscription") {
				subscribeParams := graphql.Params{
					Context: context.WithValue(ctx, WSContextKey, WSlambdaContext{
						ConnectionID:            req.RequestContext.ConnectionID,
						Stage:                   req.RequestContext.Stage,
						DomainName:              req.RequestContext.DomainName,
						OperationID:             msg.OperationID,
						RequestString:           msg.Payload.Query,
						Variables:               msg.Payload.Variables,
						OperationName:           msg.Payload.OperationName,
						WebsocketConnectContext: connection.WebsocketConnectContext,
						ConnectContext:          connection.ConnectContext,
					}),
					RequestString:  msg.Payload.Query,
					Schema:         *CurrentManager.Schema,
					VariableValues: msg.Payload.Variables,
					OperationName:  msg.Payload.OperationName,
				}

				resultChan := graphql.Subscribe(subscribeParams)
				fmt.Printf("resultChan: %v\n", resultChan)

				result := <-resultChan

				if result.HasErrors() {
					fmt.Printf("result.Errors: %v\n", result.Errors)
				}

			} else {
				var payloadInterface map[string]interface{}
				inrec, _ := json.Marshal(msg.Payload)
				json.Unmarshal(inrec, &payloadInterface)
				Execute(
					ctx,
					msg.OperationID,
					connection.Id,
					connection.Domain,
					connection.Stage,
					&graphql.Params{
						Schema:         *CurrentManager.Schema,
						RequestString:  msg.Payload.Query,
						VariableValues: msg.Payload.Variables,
						RootObject:     payloadInterface,
						OperationName:  msg.Payload.OperationName,
						Context:        ctx,
					},
				)
			}

			// err := Pub("NEW_MESSAGE", map[string]interface{}{
			// 	"id":   "17",
			// 	"text": "hello world",
			// 	"type": "bonjour",
			// })

			// if err != nil {
			// 	fmt.Printf("err: %v\n", err)
			// 	println(err)
			// }
		}
	}

	return events.APIGatewayProxyResponse{
		StatusCode: http.StatusOK,
	}, nil
}

func DynamoDBStream(ctx context.Context, req events.DynamoDBEvent) error {

	if CurrentManager == nil {
		return errors.New("manager isn't set")
	}

	for _, record := range req.Records {
		fmt.Printf("record.Change.NewImage: %v\n", record.Change.NewImage)
		if events.DynamoDBOperationType(record.EventName) == events.DynamoDBOperationTypeInsert {

			// id := record.Change.NewImage["id"]
			key := record.Change.NewImage["key"].String()
			payload := record.Change.NewImage["payload"].String()

			payloadResult := &map[string]interface{}{}

			err := json.Unmarshal([]byte(payload), payloadResult)

			if err != nil {
				fmt.Printf("err: %v\n", err)
				println(err)
			}

			var from *string

			for {

				response, err := CurrentManager.Subscription.ListByEvents(ctx, key, from)

				if err != nil {
					fmt.Printf("err: %v\n", err)
					println(err)
				}
				from = response.Next
				fmt.Printf("from: %v\n", from)

				for _, sub := range *response.Items {

					fmt.Printf("sub: %v\n", sub)

					if time.Unix(sub.Ttl, 0).Before(time.Now()) {
						continue
					}

					connection, err := CurrentManager.Connection.Get(ctx, sub.ConnectionID)

					if err != nil {
						fmt.Printf("err: %v\n", err)
						println(err)
					}

					if connection == nil || time.Unix(connection.Ttl, 0).Before(time.Now()) || !connection.IsInitialized {
						if connection != nil {
							fmt.Printf("connection: %v\n", *connection)
						}
						fmt.Printf("\"ttl reached or connection not initialised\": %v\n", "ttl reached or connection not initialised")
						continue
					}

					ctx := context.WithValue(ctx, WSContextKey, WSlambdaContext{
						ConnectionID:            sub.ConnectionID,
						Stage:                   connection.Stage,
						DomainName:              connection.Domain,
						OperationID:             sub.OperationID,
						RequestString:           sub.Query,
						Variables:               sub.Variables,
						OperationName:           sub.OperationName,
						WebsocketConnectContext: connection.WebsocketConnectContext,
						ConnectContext:          connection.ConnectContext,
					})

					Execute(
						ctx,
						sub.OperationID,
						connection.Id,
						connection.Domain,
						connection.Stage,
						&graphql.Params{
							Schema:         *CurrentManager.Schema,
							RequestString:  sub.Query,
							VariableValues: sub.Variables,
							RootObject:     *payloadResult,
							OperationName:  sub.OperationName,
							Context:        ctx,
						},
					)
				}
				if from == nil {
					break
				}
			}
		}
	}

	return nil
}

# go-lambda-graphql-subscription
Library to use graphql subscriptions with lambda serverless

This library should be used with https://github.com/graphql-go/graphql

This library is highly inspired by https://github.com/michalkvasnicak/aws-lambda-graphql I had to adapt some part to fit with Go but tried to keep the spirit.
It is implemented to work only with DynamoDB, Redis isn't at the moment, PR welcome.

# Getting started (DynamoDB)

Creating tables

- connection table

  partition key (string): id
  
  Time To Live attribute: ttl

- events table

  partition key (string): id
  
  Time To Live attribute: ttl
  
  On this table you need to stream the table to a lambda function
  
- subscription table

  partition key (string): event
  
  sort key (string): connectionId
  
   - secondary index
 
      partition key (string): connectionId
      
      sort key (string): operationId

start by importing necessary

    import (
      "github.com/ThomasP1988/go-lambda-graphql-subscription/dynamodb"
      "github.com/ThomasP1988/go-lambda-graphql-subscription/manager"
      "github.com/aws/aws-lambda-go/events"
    )


then implement the manager

    connectionManager, err := dynamodb.NewDynamoDBConnectionManager(&dynamodb.DynamoDBConnectionManagerArgs{
        Table: "YOUR_CONNECTION_TABLE_NAME",
        Ttl:   time.Hour * 6,
      })

      if err != nil {
        return err
      }

      subscriptionManager, err := dynamodb.NewDynamoDBSubscriptionManager(&dynamodb.DynamoDBSubscriptionManagerArgs{
        Table:             "YOUR_SUBSCRIPTION_TABLE_NAME",
        Ttl:               time.Hour * 6,
        IndexConnectionID: "YOUR_SUBSCRIPTION_SECONDARY_INDEX_NAME",
      })

      if err != nil {
        return err
      }

      eventManager, err := dynamodb.NewDynamoDBEventManager(&dynamodb.DynamoDBEventManagerArgs{
        Table: "YOUR_EVENT_TABLE_NAME"
      })

      if err != nil {
        return err
      }

then simply set the manager, schema is the graphql.Schema struct that you define in graphql-go

    manager.SetManager(&manager.SetManagerArgs{
        Schema:             schema,
        Connection:         connectionManager,
        Subscription:       subscriptionManager,
        Event:              eventManager,
      })
      
 subscribing to event

    graphql.NewObject(graphql.ObjectConfig{
		Name: "RootSubscription",
		Fields: graphql.Fields{
			"messageFeed": &graphql.Field{
				Type: MessageChatType,
				Resolve: func(p graphql.ResolveParams) (interface{}, error) {
					return p.Source, nil
				},
				Subscribe: func(p graphql.ResolveParams) (interface{}, error) {
					return manager.Sub([]string{"NEW_MESSAGE"}, p.Context)
				},
			},
		},
	})
   you need to return the manager.Sub function and pass the strings of the event you wish to subscribe, pass the given context as second argument
   
   publishing an event

	err := Pub("NEW_MESSAGE", map[string]interface{}{
		"id":   "17",
		"text": "hello world",
		"type": "bonjour",
	})

	if err != nil {
		fmt.Printf("err: %v\n", err)
		println(err)
	}
   
   

On the websocket lambda function

    func HandleRequest(ctx context.Context, req events.APIGatewayWebsocketProxyRequest) (events.APIGatewayProxyResponse, error) {
      return manager.HandleWS(ctx, req)
    }

    func main() {
      // set your manager here
      lambda.Start(HandleRequest)
    }


On the DynamoDB stream function

    func HandleRequest(ctx context.Context, req events.DynamoDBEvent) error {
      return manager.DynamoDBStream(ctx, req)
    }

    func main() {
      // set your manager here
      lambda.Start(HandleRequest)
    }

# Defining context

When setting the Manager you can pass functions to store some data and use them when resolving or subscribing

        OnWebsocketConnect := func(event *events.APIGatewayWebsocketProxyRequest) interface{} {
          fmt.Printf("OnWebsocketConnect event: %v\n", event)

          return map[string]string{
            "OnWebsocketConnect": "OnWebsocketConnect",
          }
        }

        OnConnect := func(event *events.APIGatewayWebsocketProxyRequest) interface{} {
          fmt.Printf("event: %v\n", event)
          return map[string]string{
            "OnConnect": "OnConnect",
          }
        }

        OnDisconnect := func(connection *manager.Connection) {
          fmt.Printf("disconnect connection: %v\n", connection)
        }

If you pass false in OnConnect, it will refuse the connection to the user.
Then pass one of multiple function to the manager

    manager.SetManager(&manager.SetManagerArgs{
        Schema:             schema,
        Connection:         connectionManager,
        Subscription:       subscriptionManager,
        Event:              eventManager,
        OnWebsocketConnect: &OnWebsocketConnect,
        OnConnect:          &OnConnect,
        OnDisconnect:       &OnDisconnect,
      })
     
 You can access the data you've put in the context this way

	graphql.NewObject(graphql.ObjectConfig{
			Name: "RootSubscription",
			Fields: graphql.Fields{
				"messageFeed": &graphql.Field{
					Type: MessageChatType,
					Resolve: func(p graphql.ResolveParams) (interface{}, error) {
						lambdaContext := p.Context.Value(manager.WSContextKey).(manager.WSlambdaContext)
						fmt.Printf("resolve lambdaContext: %v\n", lambdaContext)

						return p.Source, nil
					},
					Subscribe: func(p graphql.ResolveParams) (interface{}, error) {
						lambdaContext := p.Context.Value(manager.WSContextKey).(manager.WSlambdaContext)
						fmt.Printf("subscribe lambdaContext: %v\n", lambdaContext)

						return manager.Sub([]string{"NEW_MESSAGE"}, p.Context)
					},
				},
			},
		})
     
# Example

If you want to see an implementation of the library, you can find it here https://github.com/ThomasP1988/go-lambda-graphql-subscription-example

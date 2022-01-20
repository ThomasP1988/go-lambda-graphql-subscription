package subscription

import (
	"context"
	"fmt"
	"strings"

	common "github.com/ThomasP1988/go-lambda-graphql-subscription/dynamodb/repositories"
	"github.com/ThomasP1988/go-lambda-graphql-subscription/manager"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
)

const separator string = "/"

type Database struct {
	Client    *dynamodb.Client
	TableName string
	IndexName string
}

func (udb *Database) GetOne(ctx context.Context, connectionID string) (*manager.Subscription, error) {
	connection := &manager.Subscription{}
	doesntExist, err := common.GetOne(ctx, udb.Client, &udb.TableName, connection, map[string]interface{}{
		"connectionId": connectionID,
	}, nil)

	if doesntExist {
		return nil, err
	}

	return connection, err
}

func (udb *Database) Add(ctx context.Context, newSubscription *manager.Subscription) error {
	println("AddAddAdd")
	fmt.Printf("udb.TableName: %v\n", udb.TableName)
	return common.AddOne(ctx, udb.Client, &udb.TableName, newSubscription)
}

func (udb *Database) List(ctx context.Context, eventKey string, from *string) (*manager.SubscriptionResponse, error) {
	if from != nil {
		fmt.Printf("from: %v\n", *from)
	}
	subscriptions := &[]manager.Subscription{}
	var limit int32 = 1
	args := common.ListArgs{
		Ctx:       ctx,
		Client:    udb.Client,
		TableName: &udb.TableName,
		Output:    subscriptions,
		Keys: map[string]interface{}{
			"event": eventKey,
		},
		Limit: &limit,
	}

	if from != nil {
		fromSplited := strings.Split(*from, separator)
		subStartKey := manager.Subscription{
			ConnectionID: fromSplited[0],
			Event:        fromSplited[1],
		}
		startExclusiveKey, err := attributevalue.MarshalMap(subStartKey)

		if err != nil {
			fmt.Printf("err: %v\n", err)
			println(err)
		} else {
			fmt.Printf("startExclusiveKey: %v\n", startExclusiveKey)
			args.From = &map[string]types.AttributeValue{
				"connectionId": startExclusiveKey["connectionId"],
				"event":        startExclusiveKey["event"],
			}
		}
	}

	_, lastEvaluatedKey, err := common.List(args)

	if err != nil {
		return nil, err
	}

	fmt.Printf("lastEvaluatedKey: %v\n", lastEvaluatedKey)

	var subLast manager.Subscription
	var next *string
	if len(lastEvaluatedKey) > 0 {
		err = attributevalue.UnmarshalMap(lastEvaluatedKey, &subLast)
		if err != nil {
			println("UnmarshalMap last evaluated key err: " + err.Error())
		} else {
			nextStr := subLast.ConnectionID + separator + subLast.Event
			next = &nextStr
		}
	}

	response := &manager.SubscriptionResponse{
		Items: subscriptions,
		Next:  next,
	}

	return response, nil
}

func (udb *Database) Delete(ctx context.Context, connectionID string, operationID string) error {
	subscription := &manager.Subscription{}

	doesntExist, err := common.GetOne(ctx, udb.Client, &udb.TableName, subscription, map[string]interface{}{
		"connectionId": connectionID,
		"operationId":  operationID,
	}, &udb.IndexName)

	if doesntExist {
		return nil
	}

	if err != nil {
		return err
	}

	output, err := udb.Client.DeleteItem(ctx, &dynamodb.DeleteItemInput{
		Key: map[string]types.AttributeValue{
			"event": &types.AttributeValueMemberS{
				Value: subscription.Event,
			},
			"connectionId": &types.AttributeValueMemberS{
				Value: subscription.ConnectionID,
			},
		},

		TableName: &udb.TableName,
	})

	if err != nil {
		fmt.Printf("err: %v\n", err)
		println(err)
		return err
	}

	fmt.Printf("deleted subscription: %v\n", output)

	return nil
}

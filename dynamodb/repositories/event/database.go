package event

import (
	"context"
	"fmt"

	common "github.com/ThomasP1988/go-lambda-graphql-subscription/dynamodb/repositories"
	"github.com/ThomasP1988/go-lambda-graphql-subscription/manager"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
)

type Database struct {
	Client    *dynamodb.Client
	TableName string
}

func (udb *Database) GetOne(eventId string) (*manager.Event, error) {
	event := &manager.Event{}
	doesntExist, err := common.GetOne(udb.Client, &udb.TableName, event, map[string]interface{}{
		"id": eventId,
	}, nil)

	if doesntExist {
		return nil, err
	}

	return event, err
}

func (udb *Database) Add(newEvent *manager.Event) error {
	return common.AddOne(udb.Client, &udb.TableName, newEvent)
}

func (udb *Database) Delete(eventID string) error {
	output, err := udb.Client.DeleteItem(context.TODO(), &dynamodb.DeleteItemInput{
		Key: map[string]types.AttributeValue{
			"id": &types.AttributeValueMemberS{
				Value: eventID,
			},
		},
		TableName: &udb.TableName,
	})

	if err != nil {
		fmt.Printf("err: %v\n", err)
		println(err)
		return err
	}

	fmt.Printf("deleted connection: %v\n", output)

	return nil
}

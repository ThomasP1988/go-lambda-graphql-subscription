package dynamodb

import (
	"errors"

	common "github.com/ThomasP1988/go-lambda-graphql-subscription/common"
	"github.com/ThomasP1988/go-lambda-graphql-subscription/dynamodb/repositories/event"
	"github.com/ThomasP1988/go-lambda-graphql-subscription/manager"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
)

type DynamoDBEventManagerArgs struct {
	Table  string
	Client *dynamodb.Client
}

type DynamoDBEventManager struct {
	Table  string
	Client *event.Database
}

func NewDynamoDBEventManager(params *DynamoDBEventManagerArgs) (*DynamoDBEventManager, error) {
	println("NewDynamoDBEventManager")
	eventManager := DynamoDBEventManager{}

	if params.Table == "" {
		return nil, errors.New("missing table in dynamodb event manager")
	}
	eventManager.Table = params.Table

	if params.Client == nil {
		ddbClient, err := common.GetDynamoDBClient()

		if err != nil {
			return nil, err
		}

		eventManager.Client = &event.Database{
			Client:    ddbClient,
			TableName: params.Table,
		}
	} else {
		eventManager.Client = &event.Database{
			Client:    params.Client,
			TableName: params.Table,
		}
	}

	return &eventManager, nil
}

func (cm *DynamoDBEventManager) Add(newEvent *manager.Event) error {
	return cm.Client.Add(newEvent)
}

func (cm *DynamoDBEventManager) Remove(eventID string) error {
	return cm.Client.Delete(eventID)
}

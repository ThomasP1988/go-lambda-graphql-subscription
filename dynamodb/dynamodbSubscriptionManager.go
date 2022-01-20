package dynamodb

import (
	"context"
	"errors"
	"time"

	common "github.com/ThomasP1988/go-lambda-graphql-subscription/common"
	"github.com/ThomasP1988/go-lambda-graphql-subscription/dynamodb/repositories/subscription"
	"github.com/ThomasP1988/go-lambda-graphql-subscription/manager"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
)

type DynamoDBSubscriptionManagerArgs struct {
	Table             string
	Client            *dynamodb.Client
	IndexConnectionID string
	Ttl               time.Duration
}

type DynamoDBSubscriptionManager struct {
	Table             string
	Client            *subscription.Database
	IndexConnectionID string
	Ttl               time.Duration
}

var SubscriptionManager *DynamoDBSubscriptionManager

func NewDynamoDBSubscriptionManager(params *DynamoDBSubscriptionManagerArgs) (*DynamoDBSubscriptionManager, error) {
	println("NewDynamoDBConnectionManager")
	SubscriptionManager = &DynamoDBSubscriptionManager{}

	if params.Table == "" {
		return nil, errors.New("missing table in dynamodb connection manager")
	}

	if params.IndexConnectionID == "" {
		return nil, errors.New("missing secondary index for connectionId")
	}

	SubscriptionManager.Table = params.Table
	SubscriptionManager.IndexConnectionID = params.IndexConnectionID

	if params.Client == nil {
		ddbClient, err := common.GetDynamoDBClient()

		if err != nil {
			return nil, err
		}

		SubscriptionManager.Client = &subscription.Database{
			Client:    ddbClient,
			TableName: params.Table,
			IndexName: params.IndexConnectionID,
		}
	} else {
		SubscriptionManager.Client = &subscription.Database{
			Client:    params.Client,
			TableName: params.Table,
			IndexName: params.IndexConnectionID,
		}
	}

	if params.Ttl != 0 {
		SubscriptionManager.Ttl = params.Ttl
	} else {
		SubscriptionManager.Ttl = defaultTTL
	}

	return SubscriptionManager, nil
}

func (cm *DynamoDBSubscriptionManager) Start(ctx context.Context, subscription *manager.Subscription) error {
	println("StartStartStart")
	subscription.Ttl = time.Now().Add(SubscriptionManager.Ttl).Unix()
	return cm.Client.Add(ctx, subscription)
}

func (cm *DynamoDBSubscriptionManager) Stop(ctx context.Context, connectionID string, operationID string) error {
	return cm.Client.Delete(ctx, connectionID, operationID)
}

func (cm *DynamoDBSubscriptionManager) ListByEvents(ctx context.Context, eventKey string, from *string) (*manager.SubscriptionResponse, error) {
	return cm.Client.List(ctx, eventKey, from)
}

package dynamodb

import (
	"errors"
	"fmt"
	"time"

	common "github.com/ThomasP1988/go-lambda-graphql-subscription/common"
	"github.com/ThomasP1988/go-lambda-graphql-subscription/dynamodb/repositories/connection"
	"github.com/ThomasP1988/go-lambda-graphql-subscription/manager"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/expression"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
)

const (
	defaultTTL time.Duration = time.Hour * 6
)

type DynamoDBConnectionManagerArgs struct {
	Table  string
	Client *dynamodb.Client
	Ttl    time.Duration
}

type DynamoDBConnectionManager struct {
	Table  string
	Client *connection.Database
	Ttl    time.Duration
}

var ConnectionManager DynamoDBConnectionManager

func NewDynamoDBConnectionManager(params *DynamoDBConnectionManagerArgs) (*DynamoDBConnectionManager, error) {
	println("NewDynamoDBConnectionManager")
	ConnectionManager = DynamoDBConnectionManager{}

	if params.Table == "" {
		return nil, errors.New("missing table in dynamodb connection manager")
	}
	ConnectionManager.Table = params.Table

	if params.Client == nil {
		ddbClient, err := common.GetDynamoDBClient()

		if err != nil {
			return nil, err
		}

		ConnectionManager.Client = &connection.Database{
			Client:    ddbClient,
			TableName: params.Table,
		}
	} else {
		ConnectionManager.Client = &connection.Database{
			Client:    params.Client,
			TableName: params.Table,
		}
	}

	if params.Ttl != 0 {
		ConnectionManager.Ttl = params.Ttl
	} else {
		ConnectionManager.Ttl = defaultTTL
	}

	return &ConnectionManager, nil
}

func (cm *DynamoDBConnectionManager) OnConnect(newconnection *manager.Connection) error {
	newconnection.Ttl = time.Now().Add(ConnectionManager.Ttl).Unix()
	newconnection.IsInitialized = false
	return cm.Client.Add(newconnection)
}

func (cm *DynamoDBConnectionManager) OnDisconnect(connectionID string) error {

	// TODO: delete all subscription for this connectionID

	return cm.Client.Delete(connectionID)
}

func (cm *DynamoDBConnectionManager) Get(connectionID string) (*manager.Connection, error) {
	return cm.Client.GetOne(connectionID)
}

func (cm *DynamoDBConnectionManager) Init(connectionId string, connectContext interface{}) error {
	updateBuilder := expression.UpdateBuilder{}.Set(expression.Name("isInitialized"), expression.Value(true))

	if connectContext != nil {
		updateBuilder.Set(expression.Name("connectContext"), expression.Value(connectContext))
	}
	expr, err := expression.NewBuilder().WithUpdate(updateBuilder).Build()

	if err != nil {
		fmt.Printf("err expression.NewBuilder().WithUpdate Init: %v\n", err)
		return err
	}

	return cm.Client.Update(connectionId, &expr)
}

func (cm *DynamoDBConnectionManager) Terminate(connectionId string) error {

	updateBuilder := expression.UpdateBuilder{}.Set(expression.Name("isInitialized"), expression.Value(false))
	expr, err := expression.NewBuilder().WithUpdate(updateBuilder).Build()

	if err != nil {
		fmt.Printf("err expression.NewBuilder().WithUpdate Terminate: %v\n", err)
		return err
	}

	return cm.Client.Update(connectionId, &expr)
}

func (cm *DynamoDBConnectionManager) Hydrate(connectionId string) error {
	updateBuilder := expression.UpdateBuilder{}.Set(expression.Name("ttl"), expression.Value(time.Now().Add(ConnectionManager.Ttl).Unix()))
	expr, err := expression.NewBuilder().WithUpdate(updateBuilder).Build()

	if err != nil {
		fmt.Printf("err expression.NewBuilder().WithUpdate Terminate: %v\n", err)
		return err
	}

	return cm.Client.Update(connectionId, &expr)
}

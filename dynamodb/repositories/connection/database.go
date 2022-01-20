package connection

import (
	"context"
	"fmt"

	common "github.com/ThomasP1988/go-lambda-graphql-subscription/dynamodb/repositories"
	"github.com/ThomasP1988/go-lambda-graphql-subscription/manager"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/expression"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
)

type Database struct {
	Client    *dynamodb.Client
	TableName string
}

func (udb *Database) GetOne(ctx context.Context, connectionID string) (*manager.Connection, error) {
	connection := &manager.Connection{}
	doesntExist, err := common.GetOne(ctx, udb.Client, &udb.TableName, connection, map[string]interface{}{
		"id": connectionID,
	}, nil)

	if doesntExist {
		return nil, err
	}

	return connection, err
}

func (udb *Database) Add(ctx context.Context, newConnection *manager.Connection) error {
	return common.AddOne(ctx, udb.Client, &udb.TableName, newConnection)
}

func (udb *Database) Update(ctx context.Context, connectionID string, expr *expression.Expression) error {

	output, err := udb.Client.UpdateItem(ctx, &dynamodb.UpdateItemInput{
		Key: map[string]types.AttributeValue{
			"id": &types.AttributeValueMemberS{
				Value: connectionID,
			},
		},
		TableName:                 aws.String(udb.TableName),
		ExpressionAttributeNames:  expr.Names(),
		ExpressionAttributeValues: expr.Values(),
		UpdateExpression:          expr.Update(),
	})

	fmt.Printf("output: %v\n", output)
	return err
}

func (udb *Database) Delete(ctx context.Context, connectionID string) error {
	output, err := udb.Client.DeleteItem(ctx, &dynamodb.DeleteItemInput{
		Key: map[string]types.AttributeValue{
			"id": &types.AttributeValueMemberS{
				Value: connectionID,
			},
		},
		TableName: aws.String(udb.TableName),
	})

	if err != nil {
		fmt.Printf("err: %v\n", err)
		println(err)
		return err
	}

	fmt.Printf("deleted connection: %v\n", output)

	return nil
}

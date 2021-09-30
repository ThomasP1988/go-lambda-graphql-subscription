package common

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/expression"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
)

func GetOne(client *dynamodb.Client, tableName *string, output interface{}, keys map[string]interface{}, indexName *string) (bool, error) {
	var keyCond *expression.KeyConditionBuilder

	for k, v := range keys {
		if keyCond == nil {
			newCond := expression.Key(k).Equal(expression.Value(v))
			keyCond = &newCond
		} else {
			newCond := expression.Key(k).Equal(expression.Value(v))
			keyCond.And(newCond)
		}
	}

	expr, err := expression.NewBuilder().WithKeyCondition(*keyCond).Build()
	if err != nil {
		return false, err
	}

	input := &dynamodb.QueryInput{
		KeyConditionExpression:    expr.KeyCondition(),
		ExpressionAttributeNames:  expr.Names(),
		ExpressionAttributeValues: expr.Values(),
		TableName:                 tableName,
	}

	if indexName != nil {
		input.IndexName = indexName
	}

	queryOutput, err := client.Query(context.TODO(), input)
	if err != nil {
		return false, err
	}

	if len(queryOutput.Items) == 0 {
		return true, nil
	}

	err = attributevalue.UnmarshalMap(queryOutput.Items[0], output)

	if err != nil {
		println("failed to unmarshal Items", err.Error())
		return false, err
	}

	return false, nil
}

func AddOne(client *dynamodb.Client, tableName *string, item interface{}) error {
	fmt.Printf("item: %v\n", item)
	marshalledItem, err := attributevalue.MarshalMap(item)
	if err != nil {
		println("dynamodb AddOne, error marshalling item", err.Error())
		return err
	}
	fmt.Printf("marshalledItem: %v\n", marshalledItem)
	input := &dynamodb.PutItemInput{
		Item:      marshalledItem,
		TableName: tableName,
	}

	putOutput, err := client.PutItem(context.TODO(), input)

	if err != nil {
		return err
	}

	attributevalue.UnmarshalMap(putOutput.Attributes, item)
	return nil
}

type ListArgs struct {
	Client    *dynamodb.Client
	TableName *string
	Output    interface{}
	Keys      map[string]interface{}
	IndexName *string
	From      *map[string]types.AttributeValue
	Limit     *int32
}

func List(args ListArgs) (bool, map[string]types.AttributeValue, error) {
	var keyCond *expression.KeyConditionBuilder

	for k, v := range args.Keys {
		if keyCond == nil {
			newCond := expression.Key(k).Equal(expression.Value(v))
			keyCond = &newCond
		} else {
			newCond := expression.Key(k).Equal(expression.Value(v))
			keyCond.And(newCond)
		}
	}

	expr, err := expression.NewBuilder().WithKeyCondition(*keyCond).Build()
	if err != nil {
		return false, nil, err
	}

	input := &dynamodb.QueryInput{
		KeyConditionExpression:    expr.KeyCondition(),
		ExpressionAttributeNames:  expr.Names(),
		ExpressionAttributeValues: expr.Values(),
		TableName:                 args.TableName,
	}

	if args.Limit != nil {
		input.Limit = args.Limit
	}

	if args.From != nil {
		input.ExclusiveStartKey = *args.From
	}

	if args.IndexName != nil {
		input.IndexName = args.IndexName
	}

	queryOutput, err := args.Client.Query(context.TODO(), input)
	if err != nil {
		return false, nil, err
	}

	if len(queryOutput.Items) == 0 {
		return true, nil, nil
	}

	err = attributevalue.UnmarshalListOfMaps(queryOutput.Items, args.Output)

	if err != nil {
		println("failed to unmarshal Items", err.Error())
		return false, nil, err
	}

	return false, queryOutput.LastEvaluatedKey, nil
}

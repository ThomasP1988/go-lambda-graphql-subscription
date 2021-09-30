package common

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/aws"
	configAWS "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
)

var AWSConfig *aws.Config

func SetAWSConfig(region *string, profile *string) error {
	var cfg aws.Config
	var err error
	if region != nil {
		cfg, err = configAWS.LoadDefaultConfig(context.TODO(), configAWS.WithRegion(*region))
	} else {
		cfg, err = configAWS.LoadDefaultConfig(context.TODO())
	}

	if err != nil {
		println("unable to load SDK config, %v", err)
	}

	AWSConfig = &cfg

	return err
}

func GetAWSConfig() (*aws.Config, error) {
	if AWSConfig == nil {
		err := SetAWSConfig(nil, nil)
		if err != nil {
			return nil, err
		}
	}
	return AWSConfig, nil
}

func GetDynamoDBClient() (*dynamodb.Client, error) {
	var err error
	if AWSConfig == nil {
		err = SetAWSConfig(nil, nil)
	}

	return dynamodb.NewFromConfig(*AWSConfig), err
}

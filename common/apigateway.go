package common

import (
	"context"
	"net/url"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/apigatewaymanagementapi"
)

var apiClient *apigatewaymanagementapi.Client
var err error

func SendMessage(ctx context.Context, connectionID, domain, stage string, data []byte) error {
	if apiClient == nil {
		apiClient, err = NewAPIGatewayManagementClient(domain, stage)
		if err != nil {
			return err
		}
	}
	_, err := apiClient.PostToConnection(ctx, &apigatewaymanagementapi.PostToConnectionInput{
		ConnectionId: aws.String(connectionID),
		Data:         data,
	})
	if err != nil {
		return err
	}

	return nil
}

// NewAPIGatewayManagementClient creates a new API Gateway Management Client instance from the provided parameters. The
// new client will have a custom endpoint that resolves to the application's deployed API.
func NewAPIGatewayManagementClient(domain, stage string) (*apigatewaymanagementapi.Client, error) {

	cfg, err := GetAWSConfig()
	if err != nil {
		return nil, err
	}
	cp := cfg.Copy()
	cp.EndpointResolver = aws.EndpointResolverFunc(func(service, region string) (aws.Endpoint, error) {
		println("service", service)
		// if service != "execute-api" {
		// 	return cfg.EndpointResolver.ResolveEndpoint(service, region)
		// }
		var endpoint url.URL
		endpoint.Path = stage
		endpoint.Host = domain
		endpoint.Scheme = "https"
		return aws.Endpoint{
			SigningRegion: region,
			URL:           endpoint.String(),
		}, nil
	})

	return apigatewaymanagementapi.NewFromConfig(cp), nil
}

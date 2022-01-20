package manager

import "context"

type SubscriptionManager interface {
	Start(ctx context.Context, subscription *Subscription) error
	Stop(ctx context.Context, connectionID string, operationID string) error
	ListByEvents(ctx context.Context, eventKey string, from *string) (*SubscriptionResponse, error)
}

type SubscriptionResponse struct {
	Items *[]Subscription
	Next  *string
}

type Subscription struct {
	Event         string                 `dynamodbav:"event"`
	ConnectionID  string                 `dynamodbav:"connectionId"`
	OperationID   string                 `dynamodbav:"operationId"`
	Query         string                 `dynamodbav:"query"`
	Variables     map[string]interface{} `dynamodbav:"variables"`
	OperationName string                 `dynamodbav:"operationName"`
	Ttl           int64                  `dynamodbav:"ttl"`
}

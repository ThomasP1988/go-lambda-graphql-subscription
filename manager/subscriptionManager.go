package manager

type SubscriptionManager interface {
	Start(subscription *Subscription) error
	Stop(connectionID string, operationID string) error
	ListByEvents(eventKey string, from *string) (*SubscriptionResponse, error)
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

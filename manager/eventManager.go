package manager

import "context"

type EventManager interface {
	Add(ctx context.Context, newEvent *Event) error
	Remove(ctx context.Context, eventId string) error
}

type Event struct {
	ID      string      `dynamodbav:"id"`
	Key     string      `dynamodbav:"key"`
	Payload interface{} `dynamodbav:"payload"`
}

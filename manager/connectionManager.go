package manager

import (
	"context"
	"time"
)

type ConnectionManager interface {
	OnConnect(ctx context.Context, newConnection *Connection) error
	OnDisconnect(ctx context.Context, connectionID string) error
	Get(ctx context.Context, connectionID string) (*Connection, error)
	Init(ctx context.Context, connectionID string, connectContext interface{}) error
	Terminate(ctx context.Context, connectionID string) error
	Hydrate(ctx context.Context, connectionId string) error
}

type Connection struct {
	Id                      string      `dynamodbav:"id"`
	Domain                  string      `dynamodbav:"domain"`
	Stage                   string      `dynamodbav:"stage"`
	Context                 string      `dynamodbav:"context"`
	IsInitialized           bool        `dynamodbav:"isInitialized"`
	CreatedAt               *time.Time  `dynamodbav:"createdAt"`
	WebsocketConnectContext interface{} `dynamodbav:"websocketConnectContext"`
	ConnectContext          interface{} `dynamodbav:"connectContext"`
	Ttl                     int64       `dynamodbav:"ttl"`
}

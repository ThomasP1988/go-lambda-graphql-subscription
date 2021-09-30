package manager

import "time"

type ConnectionManager interface {
	OnConnect(newConnection *Connection) error
	OnDisconnect(connectionID string) error
	Get(connectionID string) (*Connection, error)
	Init(connectionID string, connectContext interface{}) error
	Terminate(connectionID string) error
	Hydrate(connectionId string) error
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

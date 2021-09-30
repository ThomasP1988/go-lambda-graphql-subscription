package manager

type EventManager interface {
	Add(newEvent *Event) error
	Remove(eventId string) error
}

type Event struct {
	ID      string      `dynamodbav:"id"`
	Key     string      `dynamodbav:"key"`
	Payload interface{} `dynamodbav:"payload"`
}

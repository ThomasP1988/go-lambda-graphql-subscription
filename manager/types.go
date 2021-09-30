package manager

type Message struct {
	OperationID string `json:"id,omitempty"`
	Type        string `json:"type"`
	Payload     struct {
		Query         string                 `json:"query,omitempty"`
		Variables     map[string]interface{} `json:"variables,omitempty"`
		OperationName string                 `json:"operationName,omitempty"`
	} `json:"payload,omitempty"`
}

type WSlambdaContext struct {
	OperationID             string
	ConnectionID            string
	DomainName              string
	Stage                   string
	RequestString           string
	Variables               map[string]interface{}
	OperationName           string
	WebsocketConnectContext interface{}
	ConnectContext          interface{}
}

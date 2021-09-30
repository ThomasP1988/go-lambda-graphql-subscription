package manager

type Local struct {
	Enable    bool
	OnPublish *func(key string, payload map[string]interface{})
}

package domain

// MessageBusHandler ...
type MessageBusHandler interface {
	ExecuteHandler(Message)
}

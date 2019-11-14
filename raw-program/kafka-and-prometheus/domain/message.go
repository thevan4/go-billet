package domain

// Message struct
type Message struct {
	Type         string `json:"Type"`
	MessageID    UUID   `json:"MID"`
	InReplyTo    UUID   `json:"IRT"`
	Some         string `json:"Some-about-program"`
	To           string `json:"To"`
	InternalUUID string `json:"-"`
	Data         interface{}
}

// MessageBus sends generated messages, routes received messages
type MessageBus interface {
	Send(Message)
	Subscribe(string, MessageBusHandler)
}

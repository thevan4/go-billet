package usecase

import "../domain"

// InternalErrorType ...
type InternalErrorType string

const (
	// ErrorOneType ...
	ErrorOneType InternalErrorType = "ErrorOneType"
	// ErrorOneTwoType ...
	ErrorOneTwoType InternalErrorType = "ErrorOneTwoType"
)

// InternalErrorNotificationConfiguration interface group
type InternalErrorNotificationConfiguration struct {
	messageBus domain.MessageBus
}

// NewInternalErrorNotificationConfiguration entity
func NewInternalErrorNotificationConfiguration(messageBus domain.MessageBus) *InternalErrorNotificationConfiguration {
	return &InternalErrorNotificationConfiguration{
		messageBus: messageBus,
	}
}

// InternalErrorNotification ...
func (internalErrorNotificationConfiguration *InternalErrorNotificationConfiguration) InternalErrorNotification(incomeMessage domain.Message, internalErrorType InternalErrorType) {

	outcomeMessage := domain.Message{
		InReplyTo:    incomeMessage.MessageID,
		Type:         string(internalErrorType),
		InternalUUID: incomeMessage.InternalUUID,
	}
	internalErrorNotificationConfiguration.messageBus.Send(outcomeMessage)
}

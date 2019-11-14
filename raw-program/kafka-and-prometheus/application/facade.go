package application

import (
	"../domain"
	"../usecase"
	log "github.com/sirupsen/logrus"
)

// ProgramFacade struct
type ProgramFacade struct {
	ProgramStruct    *domain.ProgramStruct
	MessageBus       domain.MessageBus
	UUIDgenerator    domain.UUIDgenerator
	SomeSuccess      domain.SomeSuccessStruct
	SomeError        domain.SomeErrorStruct
	PrometheusConfig domain.PrometheusConfig
	Logging          *log.Logger
}

// NewProgramFacade ...
func NewProgramFacade(programStruct *domain.ProgramStruct,
	messageBus domain.MessageBus,
	uuidGenerator domain.UUIDgenerator,
	someSuccess domain.SomeSuccessStruct,
	someError domain.SomeErrorStruct,
	prometheusConfig domain.PrometheusConfig,
	logging *log.Logger) *ProgramFacade {

	return &ProgramFacade{
		ProgramStruct:    programStruct,
		MessageBus:       messageBus,
		UUIDgenerator:    uuidGenerator,
		SomeSuccess:      someSuccess,
		SomeError:        someError,
		PrometheusConfig: prometheusConfig,
		Logging:          logging,
	}
}

// SomeSuccessHandle ...
func (programFacade *ProgramFacade) SomeSuccessHandle(msg domain.Message) error {
	someSuccessConfig := usecase.NewSomeSuccessConfiguration()
	return someSuccessConfig.DoSome(msg)
}

// SomeErrHandle ...
func (programFacade *ProgramFacade) SomeErrHandle(msg domain.Message) error {
	someErrConfig := usecase.NewSomeErrConfiguration()
	return someErrConfig.DoSome(msg)
}

// InternalErrorNotification ...
func (programFacade *ProgramFacade) InternalErrorNotification(msg domain.Message, internalErrorType usecase.InternalErrorType) {
	internalErrorNotificationConfiguration := usecase.NewInternalErrorNotificationConfiguration(programFacade.MessageBus)
	internalErrorNotificationConfiguration.InternalErrorNotification(msg, internalErrorType)
}

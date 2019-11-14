package application

import (
	"../domain"
	"../usecase"
	"github.com/sirupsen/logrus"
)

const someSuccessHandlerName = "some response handler"

// SomeSuccessHandlerConfigurator ...
type SomeSuccessHandlerConfigurator struct {
	Facade *ProgramFacade
}

// ExecuteHandler ...
func (someSuccessHandlerConfigurator *SomeSuccessHandlerConfigurator) ExecuteHandler(msg domain.Message) {
	someSuccessHandlerUUID := someSuccessHandlerConfigurator.Facade.UUIDgenerator.NewUUID().UUID.String()
	msg.InternalUUID = someSuccessHandlerUUID

	someSuccessHandlerConfigurator.Facade.Logging.WithFields(logrus.Fields{
		"entity":     someSuccessHandlerName,
		"event uuid": someSuccessHandlerUUID,
	}).Infof("get job: Some Error Handler")

	err := someSuccessHandlerConfigurator.Facade.SomeSuccessHandle(msg)
	if err != nil {
		someSuccessHandlerConfigurator.Facade.Logging.WithFields(logrus.Fields{
			"entity":     someSuccessHandlerName,
			"event uuid": someSuccessHandlerUUID,
		}).Errorf("can't someSuccessHandler, got error: %v", err)
		someSuccessHandlerConfigurator.Facade.InternalErrorNotification(msg, usecase.ErrorOneType)
		return
	}
	someSuccessHandlerConfigurator.Facade.Logging.WithFields(logrus.Fields{
		"entity":     someSuccessHandlerName,
		"event uuid": someSuccessHandlerUUID,
	}).Infof("job Some Response Handler successfully completed")
}

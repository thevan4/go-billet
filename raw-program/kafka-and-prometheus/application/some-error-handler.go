package application

import (
	"../domain"
	"github.com/sirupsen/logrus"
)

const someErrorHandlerName = "some error handler"

// SomeErrorHandlerConfigurator ...
type SomeErrorHandlerConfigurator struct {
	Facade *ProgramFacade
}

// ExecuteHandler ...
func (someErrorHandlerConfigurator *SomeErrorHandlerConfigurator) ExecuteHandler(msg domain.Message) {
	someErrorHandlerUUID := someErrorHandlerConfigurator.Facade.UUIDgenerator.NewUUID().UUID.String()
	msg.InternalUUID = someErrorHandlerUUID

	someErrorHandlerConfigurator.Facade.Logging.WithFields(logrus.Fields{
		"entity":     someErrorHandlerName,
		"event uuid": someErrorHandlerUUID,
	}).Infof("get job: Some Error Handler")

	err := someErrorHandlerConfigurator.Facade.SomeErrHandle(msg)
	if err != nil {
		someErrorHandlerConfigurator.Facade.Logging.WithFields(logrus.Fields{
			"entity":     someErrorHandlerName,
			"event uuid": someErrorHandlerUUID,
		}).Errorf("can't SomeErrorHandler, got error: %v", err)
		someErrorHandlerConfigurator.Facade.InternalErrorNotification(msg, "ErrorHandler some error")
		return
	}
	someErrorHandlerConfigurator.Facade.Logging.WithFields(logrus.Fields{
		"entity":     someErrorHandlerName,
		"event uuid": someErrorHandlerUUID,
	}).Infof("job Some Error Handler successfully completed")
}

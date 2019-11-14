package usecase

import "../domain"

// SomeSuccessConfiguration ..
type SomeSuccessConfiguration struct {
	Some string
}

// NewSomeSuccessConfiguration ..
func NewSomeSuccessConfiguration() *SomeSuccessConfiguration {
	return &SomeSuccessConfiguration{Some: ""}
}

// DoSome ...
func (someSuccessConfiguration *SomeSuccessConfiguration) DoSome(msg domain.Message) error {
	return nil
}

package usecase

import "../domain"

// SomeErrConfiguration ..
type SomeErrConfiguration struct {
	Some string
}

// NewSomeErrConfiguration ..
func NewSomeErrConfiguration() *SomeErrConfiguration {
	return &SomeErrConfiguration{Some: ""}
}

// DoSome ...
func (someErrConfiguration *SomeErrConfiguration) DoSome(msg domain.Message) error {
	return nil
}

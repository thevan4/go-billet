package domain

// SomeSuccessStruct ...
type SomeSuccessStruct struct {
	Some string
}

// SomeSuccessService ...
type SomeSuccessService interface {
	SomeSuccessDo(Message) error
}

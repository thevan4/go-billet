package domain

// SomeErrorStruct ...
type SomeErrorStruct struct {
	Some string
}

// SomeErrorService ...
type SomeErrorService interface {
	SomeErrorDo(Message) error
}

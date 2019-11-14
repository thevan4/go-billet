package usecase

// DoSomeStruct ...
type DoSomeStruct struct {
	Some string
}

// NewDoSomeStruct ...
func NewDoSomeStruct() *DoSomeStruct {
	return &DoSomeStruct{}
}

// DoSome ...
func (doSomeStruct *DoSomeStruct) DoSome(some, jobUUID string) error {
	return nil
}

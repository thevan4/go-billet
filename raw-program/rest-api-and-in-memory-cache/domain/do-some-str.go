package domain

// DsomeStruct ...
type DsomeStruct struct {
	Some string
}

// JustDsome ...
type JustDsome interface {
	DoSome(string, string) error
}

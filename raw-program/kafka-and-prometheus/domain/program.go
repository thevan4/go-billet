package domain

// ProgramStruct abstract structure for containing program UUID, extra info etc
type ProgramStruct struct {
	SomeData string `json:"some"`
}

// ProgramStructConfigurator performs various actions on program abstract
type ProgramStructConfigurator interface {
	NewProgram() *ProgramStruct
	SaveProgram(ProgramStruct) error
	GetProgram() (*ProgramStruct, error)
}

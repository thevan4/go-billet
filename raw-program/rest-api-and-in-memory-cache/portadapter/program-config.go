package portadapter

import (
	"../domain"
	"github.com/sirupsen/logrus"
)

const programConfigLogPrefix = "program"

// ProgramConfigLocal ...
type ProgramConfigLocal struct {
	SomeData string
	logging  *logrus.Logger
}

// NewProgramConfigLocal ...
func NewProgramConfigLocal(someData string, logging *logrus.Logger) *ProgramConfigLocal {
	return &ProgramConfigLocal{
		SomeData: someData,
		logging:  logging,
	}
}

// GetProgram get program frome somewhere
func (programConfigLocal *ProgramConfigLocal) GetProgram() (*domain.ProgramStruct, error) {
	return &domain.ProgramStruct{}, nil

}

// NewProgram create new program entity (in memory)
func (programConfigLocal *ProgramConfigLocal) NewProgram() *domain.ProgramStruct {
	return &domain.ProgramStruct{
		SomeData: programConfigLocal.SomeData,
	}
}

// SaveProgram save program entity to somewhere
func (programConfigLocal *ProgramConfigLocal) SaveProgram(program domain.ProgramStruct) error {
	return nil
}

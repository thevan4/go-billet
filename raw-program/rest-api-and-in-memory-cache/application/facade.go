package application

import (
	"../domain"
	"../usecase"
	"github.com/patrickmn/go-cache"
	log "github.com/sirupsen/logrus"
)

// ProgramFacade struct
type ProgramFacade struct {
	ProgramStruct *domain.ProgramStruct
	UUIDgenerator domain.UUIDgenerator
	InMemoryCache *cache.Cache
	Logging       *log.Logger
}

// NewProgramFacade ...
func NewProgramFacade(programStruct *domain.ProgramStruct,
	uuidGenerator domain.UUIDgenerator,
	inMemoryCache *cache.Cache,
	logging *log.Logger) *ProgramFacade {

	return &ProgramFacade{
		ProgramStruct: programStruct,
		UUIDgenerator: uuidGenerator,
		InMemoryCache: inMemoryCache,
		Logging:       logging,
	}
}

// DoJob ...
func (programFacade *ProgramFacade) DoJob(some, jobUUID string) error {
	doSomeStruct := usecase.NewDoSomeStruct()
	return doSomeStruct.DoSome(some, jobUUID)
}

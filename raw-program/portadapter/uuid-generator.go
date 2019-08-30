package portadapter

import (
	"../domain" // FIXME: fix import path when use that billet
	"github.com/gofrs/uuid"
)

// UUIDGenerator ...
type UUIDGenerator struct {
}

// NewUUIDGenerator ...
func NewUUIDGenerator() *UUIDGenerator {
	return &UUIDGenerator{}
}

// NewUUID generate new UUID in domain uuid model/struct
func (UUIDGenerator *UUIDGenerator) NewUUID() domain.UUID {
	ud, _ := uuid.NewV4()
	newUUID := domain.UUID{
		UUID: ud,
	}
	return newUUID
}

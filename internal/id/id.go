package id

import (
	"github.com/google/uuid"
	er "github.com/mcorbin/corbierror"
)

func New() (string, error) {
	id, err := uuid.NewV6()
	if err != nil {
		return "", err
	}
	return id.String(), nil
}

func Validate(id string, message string) error {
	if err := uuid.Validate(id); err != nil {
		return er.New(message, er.BadRequest, true)
	}
	return nil
}

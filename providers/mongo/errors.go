package mongo

import (
	"errors"
	"fmt"
)

var errConfigAddrEmpty = errors.New("Config Addr must not be empty")

func errMongoConnection(err error) error {
	return fmt.Errorf("Mongo connection error: %v", err)
}

package storage

import (
	"errors"
)

type QueryRow struct {
}

var loseConnectErr = errors.New("Row not found.")

func IsLoseConnect(err error) bool {
	return err == loseConnectErr
}

func Query(querySql string) (*QueryRow, error) {
	return nil, loseConnectErr
}

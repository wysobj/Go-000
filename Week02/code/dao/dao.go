package dao

import (
	"errors"
	"fmt"
	"week2/storage"

	xerrors "github.com/pkg/errors"
)

type FooDao struct {
}

var loseConnectionErr = errors.New("Lose sql connection.")
var unknownErr = errors.New("Unknown connection.")

func transFormQueryRowToDao(queryRow *storage.QueryRow) *FooDao {
	return nil
}

func transFormErrFromStorage(storageErr error) error {
	if storage.IsLoseConnect(storageErr) {
		return loseConnectionErr
	} else {
		return unknownErr
	}
}

func Get(id int) (*FooDao, error) {
	queryRow, err := storage.Query(fmt.Sprintf("select * from TABLE where id=%d", id))
	if err != nil {
		return nil, xerrors.Wrapf(transFormErrFromStorage(err), "Query sql failed.")
	}

	return transFormQueryRowToDao(queryRow), nil
}

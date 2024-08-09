package database

import (
	"bytes"
	"fmt"

	utils "utils"

	badger "github.com/dgraph-io/badger/v4"
)

type BadgerDBStore struct {
	db *badger.DB
}

func NewBadgerDb(pathToDb string) (*BadgerDBStore, error) {
	opts := badger.DefaultOptions(pathToDb)

	opts.Logger = nil
	badgerInstance, err := badger.Open(opts)
	if err != nil {
		return nil, fmt.Errorf("opening kv: %w", err)
	}

	return &BadgerDBStore{db: badgerInstance}, nil
}

func (bs *BadgerDBStore) Close() error {
	return bs.db.Close()
}

func (bs *BadgerDBStore) setOrUpdateKeyValue(key, byteData []byte) error {
	err := bs.db.Update(func(txn *badger.Txn) error {
		err := txn.Set(key, byteData)
		return err
	})
	return err
}

func (bs *BadgerDBStore) DeleteKey(key []byte) error {
	return bs.db.Update(func(txn *badger.Txn) error {
		return txn.Delete(key)
	})
}

func (bs *BadgerDBStore) Get(key []byte) ([]byte, error) {

	var valCopy []byte

	err := bs.db.View(func(txn *badger.Txn) error {
		item, err := txn.Get(key)
		if err != nil {
			return err
		}

		err = item.Value(func(val []byte) error {
			valCopy = append([]byte{}, val...)
			return nil
		})
		if err != nil {
			return err
		}
		return nil
	})

	if err != nil {
		return nil, err
	}

	return valCopy, nil
}

func (bs *BadgerDBStore) DeleteBucket(dbPrefix string) error {

	var prefix []byte
	switch dbPrefix {
	case "day":
		prefix = dbDayPrefix
	case "week":
		prefix = dbWeekPrefix
	case "app":
		prefix = dbAppPrefix
	default:
		return fmt.Errorf("no bucket of such key prefix - %s in the db", dbPrefix)
	}

	err := bs.db.View(func(txn *badger.Txn) error {
		opts := badger.DefaultIteratorOptions
		opts.PrefetchValues = false //key-only iterations, several other magnitudes faster the docs says
		it := txn.NewIterator(opts)
		defer it.Close()

		for it.Rewind(); it.Valid(); it.Next() {
			item := it.Item()
			key := item.Key()
			if bytes.HasPrefix(key, prefix) {
				if err := bs.DeleteKey(key); err != nil {
					continue
				}
			}
		}

		return nil
	})
	if err != nil {
		return err
	}
	return nil
}

func (bs *BadgerDBStore) UpdateOpertionOnPrefix(dbPrefix string, opsFunc func([]byte) ([]byte, error)) error {

	var prefix []byte
	switch dbPrefix {
	case "day":
		prefix = dbDayPrefix
	case "week":
		prefix = dbWeekPrefix
	case "app":
		prefix = dbAppPrefix
	default:
		return fmt.Errorf("no bucket of such key prefix - %s in the db", dbPrefix)
	}

	err := bs.db.Update(func(txn *badger.Txn) error {
		opts := badger.DefaultIteratorOptions
		opts.PrefetchValues = true
		opts.PrefetchSize = 100
		it := txn.NewIterator(opts)
		defer it.Close()

		for it.Seek(prefix); it.ValidForPrefix(prefix); it.Next() {
			err := it.Item().Value(func(val []byte) error {

				updatedByteArr, err := opsFunc(val)
				if err != nil {
					return err
				}

				// if bytes.Equal(updatedByteArr, val) {
				// 	return nil
				// }

				return txn.Set(it.Item().Key(), updatedByteArr)
			})

			if err != nil {
				return err
			}
		}
		return nil // all Update successful
	})

	if err != nil {
		return err
	}
	return nil
}

func (bs *BadgerDBStore) UpdateOperationOnKey(key []byte, opsFunc func([]byte) ([]byte, error)) error {

	byteData, err := bs.Get(key)
	if err != nil {
		return err
	}
	updatedByteArr, err := opsFunc(byteData)
	if err != nil {
		return err
	}

	// if bytes.Equal(updatedByteArr, byteData) {
	// 	return nil
	// }

	return bs.setOrUpdateKeyValue(key, updatedByteArr)
}

func ExampleOf_opsFunc(v []byte) ([]byte, error) {
	var (
		app AppInfo
		err error
	)

	if app, err = utils.DecodeJSON[AppInfo](v); err != nil {
		return nil, err
	}
	fmt.Println(app.AppName, app.IsCategorySet, app.DesktopCategories, "category-", app.Category, app.IsCmdLineSet, app.CmdLine, app.IsIconSet)

	app.AppIconCategoryAndCmdLine.Category = ""
	app.AppIconCategoryAndCmdLine.IsCategorySet = false
	app.AppIconCategoryAndCmdLine.DesktopCategories = nil

	fmt.Println(app.AppName, app.IsCategorySet, app.DesktopCategories, "category-", app.Category, app.IsCmdLineSet, app.CmdLine, app.IsIconSet)
	fmt.Println()
	fmt.Println()
	return utils.EncodeJSON(app)
}

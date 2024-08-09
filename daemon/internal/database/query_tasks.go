package database

import (
	"errors"

	utils "utils"

	"slices"

	badger "github.com/dgraph-io/badger/v4"
	"github.com/google/uuid"
)

func (bs *BadgerDBStore) getAllTasks() ([]utils.Task, error) {
	byteData, err := bs.Get(dbTaskKey)

	if err != nil {
		if errors.Is(err, badger.ErrKeyNotFound) {
			if err = bs.db.Update(func(txn *badger.Txn) error {
				return txn.Set(dbTaskKey, []byte{})
			}); err != nil {
				return nil, err
			}
		}

		return nil, err
	}

	if len(byteData) == 0 {
		return []utils.Task{}, nil
	}
	return utils.DecodeJSON[[]utils.Task](byteData)
}

func (bs *BadgerDBStore) GetAllTask() ([]utils.Task, error) {
	return bs.GetTaskByAppName("all")
}

func (bs *BadgerDBStore) GetTaskByAppName(appName string) ([]utils.Task, error) {
	if appName == "" {
		return nil, errors.New("appName is empty")
	}

	taskArry, err := bs.getAllTasks()
	if err != nil {
		return nil, err
	}

	if appName == "all" {
		return taskArry, nil
	}

	requestedTaskArr := make([]utils.Task, 0, len(taskArry))
	for i := 0; i < len(taskArry); i++ {
		if taskArry[i].AppName == appName {
			requestedTaskArr = append(requestedTaskArr, taskArry[i])
		}
	}
	return slices.Clip(requestedTaskArr), nil
}

func (bs *BadgerDBStore) GetTaskByUUID(taskID uuid.UUID) (utils.Task, error) {

	taskArry, err := bs.getAllTasks()
	if err != nil {
		return utils.Task{}, err
	}

	for _, task := range taskArry {
		if task.UUID == taskID {
			return task, nil
		}
	}

	return utils.Task{}, errors.New("task does not exist")
}

func (bs *BadgerDBStore) AddTask(task utils.Task) error {
	taskArry, err := bs.getAllTasks()
	if err != nil {
		return err
	}

	if bs.checkIfLimitAppExist(task, taskArry) {
		return utils.ErrLimitAppExist
	}

	taskArry = append(taskArry, task)

	byteData, err := utils.EncodeJSON(taskArry)
	if err != nil {
		return err
	}
	return bs.setOrUpdateKeyValue(dbTaskKey, byteData)
}

func (bs BadgerDBStore) checkIfLimitAppExist(task utils.Task, tasks []utils.Task) bool {

	for i := 0; i < len(tasks); i++ {
		if tasks[i].Job == utils.DailyAppLimit {
			if tasks[i].AppName == task.AppName {
				return true
			}
		}
	}
	return false
}

func (bs *BadgerDBStore) RemoveTask(id uuid.UUID) error {
	taskArray, err := bs.getAllTasks()
	if err != nil {
		return err
	}
	newTaskArray := slices.DeleteFunc(taskArray, func(s utils.Task) bool {
		return s.UUID == id
	})

	byteData, err := utils.EncodeJSON(newTaskArray)
	if err != nil {
		return err
	}
	return bs.setOrUpdateKeyValue(dbTaskKey, byteData)
}

package handler

import (
	"context"
	"errors"
	"github.com/google/uuid"
	"sync"

	"github.com/wulfixyt/anubis-monitors/pkg/tasks/structs"
)

var (
	taskMutex           = sync.RWMutex{}
	TaskDoesNotExistErr = errors.New("Task is nonexistent")
	TaskList            = make(map[string]*structs.Task)
)

// DoesTaskExist checks if a tasks exists
func DoesTaskExist(taskId string) bool {
	taskMutex.RLock()
	defer taskMutex.RUnlock()
	_, ok := TaskList[taskId]
	return ok
}

// Creates and initializes a tasks
func Create(task *structs.Task) {
	taskMutex.Lock()
	defer taskMutex.Unlock()

	id := uuid.NewString()
	task.Id = id

	TaskList[id] = task
	return
}

// Starts a specific tasks
func Start(taskId string) error {
	if !DoesTaskExist(taskId) {
		return TaskDoesNotExistErr
	}

	taskMutex.Lock()
	defer taskMutex.Unlock()

	task := TaskList[taskId]
	if !task.Active {
		task.Ctx, task.Cancel = context.WithCancel(context.Background())
		task.Active = true

		TaskList[taskId] = task

		startModule(task)
	}

	return nil
}

// Starts all tasks
func StartAll() error {
	taskMutex.Lock()
	defer taskMutex.Unlock()

	for index, task := range TaskList {
		if !task.Active {
			task.Ctx, task.Cancel = context.WithCancel(context.Background())
			task.Active = true

			TaskList[index] = task

			startModule(task)
		}
	}

	return nil
}

// Stops a specific tasks
func Stop(taskId string) error {
	if !DoesTaskExist(taskId) {
		return TaskDoesNotExistErr
	}

	taskMutex.Lock()
	defer taskMutex.Unlock()

	task := TaskList[taskId]
	if task.Active {
		task.Active = false
		task.Cancel()
	}

	return nil
}

// Stops all tasks
func StopAll() error {
	taskMutex.Lock()
	defer taskMutex.Unlock()

	for _, task := range TaskList {
		if task.Active {
			task.Active = false
			task.Cancel()
		}
	}

	return nil
}

// Stops and deletes specific tasks
func Delete(taskId string) error {
	if !DoesTaskExist(taskId) {
		return TaskDoesNotExistErr
	}

	taskMutex.Lock()
	defer taskMutex.Unlock()

	task := TaskList[taskId]
	if task.Active {
		task.Active = false
		task.Cancel()
	}

	delete(TaskList, taskId)

	return nil
}

// Stops and deletes all tasks
func DeleteAll() error {
	taskMutex.Lock()
	defer taskMutex.Unlock()

	for _, task := range TaskList {
		if task.Active {
			task.Active = false

			task.Cancel()
		}

		delete(TaskList, task.Id)
	}

	return nil
}

// Returns a specific tasks
func GetTask(taskId string) *structs.Task {
	if !DoesTaskExist(taskId) {
		return &structs.Task{}
	}

	taskMutex.RLock()
	defer taskMutex.RUnlock()

	return TaskList[taskId]
}

// Returns all tasks
func GetTasks() []*structs.Task {
	taskMutex.RLock()
	defer taskMutex.RUnlock()

	values := []*structs.Task{}
	for _, value := range TaskList {
		values = append(values, value)
	}

	return values
}

func Filter(id string) []*structs.Task {
	taskMutex.RLock()
	defer taskMutex.RUnlock()

	var values []*structs.Task
	for _, value := range TaskList {
		if value.GroupId != id {
			continue
		}

		values = append(values, value)
	}

	return values
}

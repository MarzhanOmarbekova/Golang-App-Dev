package storage

import (
	"sync"
	"task_api/internal/models"
)

type TaskStorage struct {
	mu     sync.Mutex
	tasks  map[int]models.Task
	nextID int
}

func NewTaskStorage() *TaskStorage {
	return &TaskStorage{
		tasks:  make(map[int]models.Task),
		nextID: 1,
	}
}

func (s *TaskStorage) GetAll(doneFilter *bool) []models.Task {
	s.mu.Lock()
	defer s.mu.Unlock()

	result := []models.Task{}
	for _, task := range s.tasks {
		if doneFilter != nil && task.Done != *doneFilter {
			continue
		}
		result = append(result, task)
	}

	return result
}

func (s *TaskStorage) GetByID(id int) (models.Task, bool) {
	s.mu.Lock()
	defer s.mu.Unlock()

	task, ok := s.tasks[id]
	return task, ok
}

func (s *TaskStorage) Create(title string) models.Task {
	s.mu.Lock()
	defer s.mu.Unlock()

	task := models.Task{
		ID:    s.nextID,
		Title: title,
		Done:  false,
	}
	s.tasks[s.nextID] = task
	s.nextID++
	return task
}

func (s *TaskStorage) UpdateDone(id int, done bool) bool {
	s.mu.Lock()
	defer s.mu.Unlock()

	task, ok := s.tasks[id]
	if !ok {
		return false
	}
	task.Done = true
	s.tasks[id] = task
	return true
}

func (s *TaskStorage) Delete(id int) bool {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, ok := s.tasks[id]; !ok {
		return false
	}
	delete(s.tasks, id)
	return true
}

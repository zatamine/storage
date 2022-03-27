package storage

import (
	"errors"
	"fmt"
	"sync"

	"github.com/zatamine/uuid"
)

type memory[T model] struct {
	m    sync.Mutex
	data map[string]T
}

func NewMemory[T model](data map[string]T) Storage[T] {
	var newData map[string]T
	if len(data) <= 0 {
		newData = make(map[string]T)
	} else {
		newData = data
	}
	return &memory[T]{
		data: newData,
	}
}

func (m *memory[T]) Create(item *T) error {
	m.m.Lock()
	defer m.m.Unlock()
	id := uuid.FromString(fmt.Sprintf("%v", item)).String()
	m.data[id] = *item
	return nil
}

func (m *memory[T]) FindOne(id string) (*T, error) {
	m.m.Lock()
	defer m.m.Unlock()
	item, ok := m.data[id]
	if !ok {
		errorMsg := fmt.Sprintf("id '%s' not found in memory", id)
		return nil, errors.New(errorMsg)
	}
	return &item, nil
}

func (m *memory[T]) Update(item T) error {
	m.m.Lock()
	defer m.m.Unlock()
	id := item.ID()
	if _, ok := m.data[id]; !ok {
		errorMsg := fmt.Sprintf("Cannot update item with id '%s'", id)
		return errors.New(errorMsg)
	}
	m.data[id] = item
	return nil
}

func (m *memory[T]) Delete(id string) error {
	m.m.Lock()
	defer m.m.Unlock()
	if _, ok := m.data[id]; !ok {
		errorMsg := fmt.Sprintf("Cannot delete item with id '%s'", id)
		return errors.New(errorMsg)
	}
	delete(m.data, id)
	return nil
}

func (m *memory[T]) FindAll() ([]T, error) {
	items := make([]T, 0)
	for _, d := range m.data {
		items = append(items, d)
	}
	return items, nil
}

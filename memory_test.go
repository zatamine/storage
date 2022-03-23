package storage_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/zatamine/storage"
)

type strItem struct {
	id   string
	data string
	Stat int64
}

func (s strItem) ID() string {
	return s.id
}

func (s strItem) Status() int64 {
	return s.Stat
}

func TestCreate(t *testing.T) {
	tests := []struct {
		name   string
		input  strItem
		hasErr bool
	}{
		{
			"With empty item",
			strItem{},
			true,
		},
		{
			"With good string item",
			strItem{data: "Hello"},
			false,
		},
	}
	data := make(map[string]strItem)
	memory := storage.NewMemory(data)
	for _, tc := range tests {
		err := memory.Create(&tc.input)
		t.Run(tc.name, func(t *testing.T) {
			assert.NoError(t, err)
		})
	}
}

func TestFindOne(t *testing.T) {
	tests := []struct {
		name   string
		input  string
		want   strItem
		hasErr bool
	}{
		{
			"With empty parameter",
			"",
			strItem{},
			true,
		},
		{
			"With good id parameter",
			"1",
			strItem{"1", "Hello", 0},
			false,
		},
		{
			"With bad id parameter",
			"99",
			strItem{"99", "Hello", 0},
			true,
		},
	}
	data := map[string]strItem{
		"1": {"1", "Hello", 0},
		"2": {"2", "Holla", 0},
		"3": {"3", "Bonjour", 0},
	}
	memory := storage.NewMemory(data)
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			item, err := memory.FindOne(tc.input)
			if tc.hasErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tc.want, *item)
			}
		})
	}
}

func TestUpdate(t *testing.T) {
	tests := []struct {
		name   string
		input  strItem
		want   strItem
		hasErr bool
	}{
		{
			"With empty input",
			strItem{},
			strItem{},
			true,
		},
		{
			"With good input",
			strItem{"1", "Toto", 0},
			strItem{"1", "Toto", 0},
			false,
		},
	}
	data := map[string]strItem{
		"1": {"1", "Hello", 0},
		"2": {"2", "Holla", 0},
		"3": {"3", "Bonjour", 0},
	}
	memory := storage.NewMemory(data)
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			err := memory.Update(tc.input)
			if tc.hasErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestDelete(t *testing.T) {
	tests := []struct {
		name    string
		inputID string
		data    map[string]strItem
		hasErr  bool
	}{
		{
			"With empty storage",
			"",
			map[string]strItem{},
			true,
		},
		{
			"With 1 item storage",
			"1",
			map[string]strItem{
				"1": {data: "Hello"},
			},
			false,
		},
		{
			"With 3 items storage",
			"2",
			map[string]strItem{
				"1": {data: "Hello"},
				"2": {data: "Holla"},
				"3": {data: "Bonjour"},
			},
			false,
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			memory := storage.NewMemory(tc.data)
			err := memory.Delete(tc.inputID)
			if tc.hasErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestFindAll(t *testing.T) {
	tests := []struct {
		name      string
		data      map[string]strItem
		wantCount int
	}{
		{
			"With empty storage",
			map[string]strItem{},
			0,
		},
		{
			"With 1 item storage",
			map[string]strItem{
				"1": {data: "Hello"},
			},
			1,
		},
		{
			"With 3 items storage",
			map[string]strItem{
				"1": {data: "Hello"},
				"2": {data: "Holla"},
				"3": {data: "Bonjour"},
			},
			3,
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			memory := storage.NewMemory(tc.data)
			items, err := memory.FindAll()
			assert.NoError(t, err)
			assert.Equal(t, tc.wantCount, len(items))

		})
	}
}

package rga

import (
	"encoding/json"
	"fmt"
	"os"
	"sync"
	"time"
)

type RGAElement struct {
	Value     string
	Timestamp time.Time
	Index     int
}

type RGA struct {
	mu       sync.Mutex
	elements []RGAElement
	clock    int
}

func NewRGA() *RGA {
	return &RGA{
		elements: []RGAElement{},
		clock:    0,
	}
}

func (r *RGA) Insert(position int, value string) {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.clock++
	elem := RGAElement{
		Value:     value,
		Timestamp: time.Now(),
		Index:     r.clock,
	}

	r.elements = append(r.elements[:position], append([]RGAElement{elem}, r.elements[position:]...)...)
	fmt.Printf("Inserted: %s at position %d\n", value, position)
}

func (r *RGA) Delete(position int) {
	r.mu.Lock()
	defer r.mu.Unlock()

	if position >= 0 && position < len(r.elements) {
		deleted := r.elements[position]
		r.elements = append(r.elements[:position], r.elements[position+1:]...)
		fmt.Printf("Deleted: %s from position %d\n", deleted.Value, position)
	}
}

func (r *RGA) SaveToFile(filename string) error {
	data, err := json.Marshal(r)
	if err != nil {
		return err
	}
	return os.WriteFile(filename, data, 0644)
}

func (r *RGA) LoadFromFile(filename string) error {
	data, err := os.ReadFile(filename)
	if err != nil {
		return err
	}
	return json.Unmarshal(data, r)
}

func (r *RGA) GetDocument() []string {
	r.mu.Lock()
	defer r.mu.Unlock()

	var document []string
	for _, elem := range r.elements {
		document = append(document, elem.Value)
	}
	return document
}

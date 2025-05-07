package rga

import (
	"encoding/json"
	"os"
	"sync"
	"time"
)

type RGAMessage struct {
	Type      string `json:"type"`
	Timestamp string `json:"timestamp"`
	Position  int    `json:"position"`
	Value     string `json:"value"`
}

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
}

func (r *RGA) Delete(position int) {
	r.mu.Lock()
	defer r.mu.Unlock()

	if position >= 0 && position < len(r.elements) {
		r.elements = append(r.elements[:position], r.elements[position+1:]...)
	}
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

func (r *RGA) SaveToFile(filename string) error {
	document := r.GetDocument()
	data, err := json.Marshal(document)
	if err != nil {
		return err
	}
	return os.WriteFile(filename, data, 0644)
}

func LoadFromFile(filename string) (*RGA, error) {
	if _, err := os.Stat(filename); err == nil {
		data, err := os.ReadFile(filename)
		if err != nil {
			return nil, err
		}

		var document []string
		if err := json.Unmarshal(data, &document); err != nil {
			return nil, err
		}

		rgaDoc := NewRGA()
		for _, char := range document {
			rgaDoc.Insert(len(rgaDoc.GetDocument()), char)
		}
		return rgaDoc, nil
	}

	return NewRGA(), nil
}

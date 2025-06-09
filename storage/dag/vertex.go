package dag

import "time"

type Vertex struct {
	ID        string
	Parents   []string
	Children  []string
	Data      []byte
	Timestamp int64
}

func NewVertex(id string, parents []string, data []byte) *Vertex {
	return &Vertex{
		ID:        id,
		Parents:   parents,
		Data:      data,
		Timestamp: time.Now().Unix(),
	}
}

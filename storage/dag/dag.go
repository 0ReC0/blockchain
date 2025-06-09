package dag

import "sync"

type DAG struct {
	Vertices map[string]*Vertex
	mu       sync.Mutex
}

func NewDAG() *DAG {
	return &DAG{
		Vertices: make(map[string]*Vertex),
	}
}

func (d *DAG) AddVertex(v *Vertex) {
	d.mu.Lock()
	defer d.mu.Unlock()
	d.Vertices[v.ID] = v
	for _, parent := range v.Parents {
		if p, exists := d.Vertices[parent]; exists {
			p.Children = append(p.Children, v.ID)
		}
	}
}

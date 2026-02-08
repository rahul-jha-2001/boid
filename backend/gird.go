package main

// SpatialHashGrid bins particles into uniform cells for fast neighbor queries.
type SpatialHashGrid struct {
	CellSize float64
	Cols     int
	Rows     int
	Buckets  [][]int // Buckets[cellIndex] = particle indices
}

func NewSpatialHashGrid(w, h, cellSize float64) *SpatialHashGrid {
	if cellSize <= 0 {
		cellSize = 10
	}
	cols := int(w / cellSize)
	if cols < 1 {
		cols = 1
	}
	rows := int(h / cellSize)
	if rows < 1 {
		rows = 1
	}

	b := make([][]int, cols*rows)
	return &SpatialHashGrid{
		CellSize: cellSize,
		Cols:     cols,
		Rows:     rows,
		Buckets:  b,
	}
}

func (g *SpatialHashGrid) Clear() {
	for i := range g.Buckets {
		g.Buckets[i] = g.Buckets[i][:0]
	}
}

func (g *SpatialHashGrid) cellXY(pos Vec2) (cx, cy int) {
	cx = int(pos.X / g.CellSize)
	cy = int(pos.Y / g.CellSize)

	// clamp (world wrapping exists, but positions should already be wrapped)
	if cx < 0 {
		cx = 0
	} else if cx >= g.Cols {
		cx = g.Cols - 1
	}
	if cy < 0 {
		cy = 0
	} else if cy >= g.Rows {
		cy = g.Rows - 1
	}
	return
}

func (g *SpatialHashGrid) idx(cx, cy int) int {
	return cy*g.Cols + cx
}

func (g *SpatialHashGrid) Insert(i int, pos Vec2) {
	cx, cy := g.cellXY(pos)
	g.Buckets[g.idx(cx, cy)] = append(g.Buckets[g.idx(cx, cy)], i)
}

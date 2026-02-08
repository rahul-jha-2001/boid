package main

type World struct {
	W, H    float64
	Gravity Vec2
	Drag    float64

	Restitution float64

	Time float64
	Grid *SpatialHashGrid

	Particles []Particle
	Emitters  map[uint32]*Emitter

	GridCols      int
	SpringRestLen float64
	SpringK       float64
	SpringDamping float64
}

func (w *World) resolveCollisionsGrid() {
	g := w.Grid
	if g == nil {
		return
	}

	// Check each cell with itself + a subset of neighbors to avoid double counting
	neighborOffsets := [][2]int{
		{0, 0},  // self
		{1, 0},  // right
		{0, 1},  // down
		{1, 1},  // down-right
		{-1, 1}, // down-left
	}

	for cy := 0; cy < g.Rows; cy++ {
		for cx := 0; cx < g.Cols; cx++ {
			base := g.Buckets[g.idx(cx, cy)]
			if len(base) == 0 {
				continue
			}

			for _, off := range neighborOffsets {
				nx := cx + off[0]
				ny := cy + off[1]

				if nx < 0 {
					nx += g.Cols
				}
				if nx >= g.Cols {
					nx -= g.Cols
				}
				if ny < 0 {
					ny += g.Rows
				}
				if ny >= g.Rows {
					ny -= g.Rows
				}

				nb := g.Buckets[g.idx(nx, ny)]

				// pairwise between base and nb
				for ii := 0; ii < len(base); ii++ {
					i := base[ii]

					// if same bucket, ensure j>i within list to avoid duplicates
					startJ := 0
					if nx == cx && ny == cy {
						startJ = ii + 1
					}

					for jj := startJ; jj < len(nb); jj++ {
						j := nb[jj]
						w.resolvePair(i, j) // must be wrap-aware (you already have)
					}
				}
			}
		}
	}
}

func (w *World) Step(dt float64) {
	w.Time += dt

	// 1) clear forces
	for i := range w.Particles {
		w.Particles[i].ClearForce()
	}

	// 2) apply emitters
	for _, e := range w.Emitters {
		e.Apply(w)
	}
	w.applySprings()

	// 3) integrate + wrap
	for i := range w.Particles {
		w.Particles[i].Integrate(dt, w.Gravity, w.Drag)
		w.wrapPosition(&w.Particles[i])
	}

	// 4) build grid
	if w.Grid != nil {
		w.Grid.Clear()
		for i := range w.Particles {
			w.Grid.Insert(i, w.Particles[i].Pos)
		}

		// 5) collisions using grid
		w.resolveCollisionsGrid()
	}
}

func (w *World) wrapPosition(p *Particle) {
	if p.Pos.X < 0 {
		p.Pos.X += w.W
	}
	if p.Pos.X >= w.W {
		p.Pos.X -= w.W
	}
	if p.Pos.Y < 0 {
		p.Pos.Y += w.H
	}
	if p.Pos.Y >= w.H {
		p.Pos.Y -= w.H
	}
}

func (w *World) wrapDelta(d Vec2) Vec2 {
	if d.X > w.W/2 {
		d.X -= w.W
	} else if d.X < -w.W/2 {
		d.X += w.W
	}
	if d.Y > w.H/2 {
		d.Y -= w.H
	} else if d.Y < -w.H/2 {
		d.Y += w.H
	}
	return d
}

// If you want particle-particle collisions later, keep your wrap-aware resolvePair()
// and spatial hash grid next. For 10k, collisions O(n^2) will explode.

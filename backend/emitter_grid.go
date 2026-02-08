package main

import "math"

func (w *World) applyEmitterGrid(e *Emitter, F0, r2max, sigma2 float64) {
	g := w.Grid

	// find cell range around emitter
	cx, cy := g.cellXY(e.Pos)
	rCells := int(math.Ceil(e.Radius / g.CellSize))

	for dy := -rCells; dy <= rCells; dy++ {
		for dx := -rCells; dx <= rCells; dx++ {
			nx := cx + dx
			ny := cy + dy

			// wrap cell indices (torus world)
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

			b := g.Buckets[g.idx(nx, ny)]
			for _, i := range b {
				d := w.Particles[i].Pos.Sub(e.Pos)
				dist2 := d.Len2()
				if dist2 > r2max {
					continue
				}
				if dist2 < 1e-9 {
					continue
				}
				dist := math.Sqrt(dist2)
				dir := d.Mul(1.0 / dist)

				weight := math.Exp(-dist2 / (2 * sigma2))
				w.Particles[i].AddForce(dir.Mul(F0 * weight))
			}
		}
	}
}

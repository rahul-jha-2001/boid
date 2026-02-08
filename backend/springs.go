package main

import "math"

func (w *World) applySprings() {
	cols := int(w.GridCols)
	if cols <= 0 {
		return
	}

	L0 := w.SpringRestLen
	k := w.SpringK
	c := w.SpringDamping

	if L0 <= 0 || k == 0 {
		return
	}

	n := len(w.Particles)

	for i := 0; i < n; i++ {
		x := i % cols

		// right neighbor
		if x+1 < cols && i+1 < n {
			w.applySpringPair(i, i+1, L0, k, c)
		}

		// down neighbor
		j := i + cols
		if j < n {
			w.applySpringPair(i, j, L0, k, c)
		}
	}
}

func (w *World) applySpringPair(i, j int, L0, k, c float64) {
	pi := &w.Particles[i]
	pj := &w.Particles[j]

	d := pj.Pos.Sub(pi.Pos)
	dist2 := d.Len2()
	if dist2 < 1e-9 {
		return
	}

	dist := math.Sqrt(dist2)
	dir := d.Mul(1.0 / dist)

	// Hooke's law
	ext := dist - L0
	fs := dir.Mul(k * ext)

	// damping along spring
	relVel := pj.Vel.Sub(pi.Vel)
	fd := dir.Mul(c * relVel.Dot(dir))

	f := fs.Add(fd)

	pi.AddForce(f)
	pj.AddForce(f.Mul(-1))
}

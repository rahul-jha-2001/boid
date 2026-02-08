package main

import "math"

type Emitter struct {
	Id     uint32
	Pos    Vec2
	Radius float64 // meters
	AmpN   float64 // Newtons (peak)
	FreqHz float64
	Phase  float64 // radians

	Sigma2 float64 // m^2, controls falloff; if 0 use Radius-derived
}

func (e *Emitter) Apply(w *World) {
	if e.AmpN == 0 || e.FreqHz == 0 || e.Radius <= 0 {
		return
	}
	// drive signal
	s := math.Sin(2*math.Pi*e.FreqHz*w.Time + e.Phase)
	F0 := e.AmpN * s // Newtons

	r2max := e.Radius * e.Radius
	sigma2 := e.Sigma2
	if sigma2 <= 0 {
		// roughly match radius
		sigma2 = (e.Radius * e.Radius) / 2.0
	}

	// Use spatial grid if available to only touch nearby particles
	if w.Grid != nil {
		w.applyEmitterGrid(e, F0, r2max, sigma2)
		return
	}

	// fallback O(n)
	for i := range w.Particles {
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

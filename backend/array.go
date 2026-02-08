package main

import "math"

// Creates/updates a 1D array of N emitters centered at (cx, cy).
func (w *World) SetLineArray(baseID uint32, n int, cx, cy, spacing float64, radius float64, ampN float64, freqHz float64) {
	if n < 1 {
		n = 1
	}
	startX := cx - spacing*float64(n-1)/2.0

	for i := 0; i < n; i++ {
		id := baseID + uint32(i)
		x := startX + float64(i)*spacing
		em, ok := w.Emitters[id]
		if !ok {
			em = &Emitter{Id: id}
			w.Emitters[id] = em
		}
		em.Pos = V(x, cy)
		em.Radius = radius
		em.AmpN = ampN
		em.FreqHz = freqHz
		em.Sigma2 = 0 // auto
		em.Phase = 0  // will be set by steering
	}
}

// Updates phases to steer angle theta (radians) using assumed wave speed c (m/s).
func (w *World) SteerLineArray(baseID uint32, n int, cx float64, spacing float64, freqHz float64, c float64, thetaRad float64) {
	if n < 1 || freqHz <= 0 || c <= 0 {
		return
	}

	lambda := c / freqHz
	k := 2 * math.Pi / lambda

	// phase reference at center
	startX := cx - spacing*float64(n-1)/2.0
	for i := 0; i < n; i++ {
		id := baseID + uint32(i)
		em := w.Emitters[id]
		if em == nil {
			continue
		}
		x := startX + float64(i)*spacing
		// relative position from center
		dx := x - cx
		em.Phase = -k * dx * math.Sin(thetaRad)
	}
}

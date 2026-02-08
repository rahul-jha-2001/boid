package main

import "math"

func (w *World) resolvePair(i, j int) {
	pi := &w.Particles[i]
	pj := &w.Particles[j]

	// Wrap-aware delta from i -> j
	d := w.wrapDelta(pj.Pos.Sub(pi.Pos))
	dist2 := d.Len2()
	r := pi.Radius + pj.Radius
	r2 := r * r

	// No overlap
	if dist2 >= r2 || dist2 < 1e-12 {
		return
	}

	dist := math.Sqrt(dist2)
	n := d.Mul(1.0 / dist) // collision normal (unit)

	// --- Positional correction (separate them) ---
	penetration := r - dist

	// split correction by inverse mass
	invMi := 1.0 / pi.Mass
	invMj := 1.0 / pj.Mass
	invSum := invMi + invMj
	if invSum <= 0 {
		return
	}

	// push apart along normal (small slop helps stability)
	slop := 0.01
	corrMag := math.Max(0, penetration-slop) / invSum
	corr := n.Mul(corrMag)

	pi.Pos = pi.Pos.Sub(corr.Mul(invMi))
	pj.Pos = pj.Pos.Add(corr.Mul(invMj))

	// IMPORTANT: wrap positions back into world after correction
	w.wrapPosition(pi)
	w.wrapPosition(pj)

	// --- Velocity impulse (elastic collision) ---
	relVel := pj.Vel.Sub(pi.Vel)
	velAlongNormal := relVel.Dot(n)

	// If already separating, don't apply impulse
	if velAlongNormal > 0 {
		return
	}

	// restitution (bounce)
	e := w.Restitution
	if e < 0 {
		e = 0
	}
	if e > 1 {
		e = 1
	}

	// impulse scalar
	jImpulse := -(1 + e) * velAlongNormal / invSum
	imp := n.Mul(jImpulse)

	pi.Vel = pi.Vel.Sub(imp.Mul(invMi))
	pj.Vel = pj.Vel.Add(imp.Mul(invMj))
}

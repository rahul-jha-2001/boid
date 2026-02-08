package main

import "math"

type ForceSource interface {
	ID() uint32
	ForceOn(p *Particle, w *World) Vec2
}

type PointForce struct {
	Id         uint32
	Pos        Vec2
	ForceConst float64 // N·m^2
	Radius     float64
	Softening  float64

	WaveEnabled bool
	WaveAmp     float64 // N·m^2
	WaveFreqHz  float64 // Hz
	WavePhase   float64 // radians

	// NEW: pulsing
	PulsePeriod  float64 // seconds (e.g. 1.0)
	PulseOn      float64 // seconds ON inside each period (e.g. 0.2)
	PulseEnabled bool
}

func (pf *PointForce) ID() uint32 { return pf.Id }

func (pf *PointForce) ForceOn(p *Particle, w *World) Vec2 {
	// Wave-only force
	if !pf.WaveEnabled || pf.WaveAmp == 0 || pf.WaveFreqHz == 0 {
		return Vec2{}
	}

	K := pf.WaveAmp * math.Sin(
		2*math.Pi*pf.WaveFreqHz*w.Time+pf.WavePhase,
	)

	// Optional: clamp if you *never* want repulsion
	// if K < 0 { K = 0 }

	// Pulse gating (optional, still works)
	if pf.PulseEnabled && pf.PulsePeriod > 0 && pf.PulseOn >= 0 {
		phase := math.Mod(w.Time, pf.PulsePeriod)
		if phase > pf.PulseOn {
			return Vec2{}
		}
	}

	d := w.wrapDelta(pf.Pos.Sub(p.Pos))
	dist2 := d.Len2()
	if dist2 < 1e-9 {
		return Vec2{}
	}
	if pf.Radius > 0 && dist2 > pf.Radius*pf.Radius {
		return Vec2{}
	}

	dist := math.Sqrt(dist2)
	dir := d.Mul(1.0 / dist)

	soft := pf.Softening
	if soft <= 0 {
		soft = 25 // m²
	}

	mag := K / (dist2 + soft) // Newtons
	return dir.Mul(mag)
}

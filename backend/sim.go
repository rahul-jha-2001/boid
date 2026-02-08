package main

type Sim struct {
	World *World
	Tick  uint32
}

func SpawnGrid(world *World, n int, radius float64, spacing float64) {
	if spacing <= 2*radius {
		spacing = 2*radius + 1
	}

	cols := int(world.W / spacing)
	if cols < 1 {
		cols = 1
	}
	rows := (n + cols - 1) / cols

	// Center the grid
	gridW := float64(cols-1) * spacing
	gridH := float64(rows-1) * spacing
	startX := (world.W - gridW) / 2
	startY := (world.H - gridH) / 2

	world.Particles = world.Particles[:0]
	for i := 0; i < n; i++ {
		c := i % cols
		rw := i / cols

		x := startX + float64(c)*spacing
		y := startY + float64(rw)*spacing

		p := NewParticle(
			V(x, y),
			V(0, 0), // no random velocity
			radius,
		)
		world.Particles = append(world.Particles, p)
	}
}

func NewSim(w, h float64, n int) *Sim {
	world := &World{
		W:           w,
		H:           h,
		Gravity:     V(0, 0),
		Drag:        0.0,
		Particles:   make([]Particle, 0, n),
		Emitters:    make(map[uint32]*Emitter),
		Grid:        NewSpatialHashGrid(w, h, 10),
		Restitution: 0.9,
	}

	// ✅ Grid spawn instead of random spawn
	radius := 2.0
	spacing := 2*radius + 1.5
	SpawnGrid(world, n, radius, spacing)

	return &Sim{World: world}
}

func (s *Sim) Step(dt float64) {
	s.World.Step(dt)
	s.Tick++
}

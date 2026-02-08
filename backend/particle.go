package main

type Particle struct {
	Pos, Vel Vec2
	Mass     float64
	Radius   float64

	force Vec2
}

const density = 0.1 // kg / m^2  (tuneable)

func NewParticle(pos, vel Vec2, radius float64) Particle {
	mass := density * radius * radius // kg
	return Particle{
		Pos:    pos,
		Vel:    vel,
		Radius: radius, // m
		Mass:   mass,   // kg
	}
}


func (p *Particle) ClearForce() { p.force = Vec2{} }
func (p *Particle) AddForce(f Vec2) { p.force = p.force.Add(f) }

func (p *Particle) Integrate(dt float64, gravity Vec2, drag float64) {
	// acceleration from forces: a = F/m
	acc := p.force.Mul(1.0 / p.Mass)

	// gravity is acceleration (m/s^2)
	acc = acc.Add(gravity)

	// linear drag modeled as acceleration term: -(c/m) v
	// if you currently do acc - drag*v, then drag is 1/s
	// to make drag kg/s, scale by 1/mass:
	acc = acc.Sub(p.Vel.Mul(drag / p.Mass))

	p.Vel = p.Vel.Add(acc.Mul(dt))
	p.Pos = p.Pos.Add(p.Vel.Mul(dt))
}

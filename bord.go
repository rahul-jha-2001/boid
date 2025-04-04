package main

import (
	"math/rand"
	"math"
	"github.com/golang/geo/r3"
)


func randomVec3(Max_x, Max_y float64) r3.Vector {
	return r3.Vector{
		X: rand.Float64()*Max_x,
		Y: rand.Float64()*Max_y,
	}
}

type boid struct {
	position r3.Vector
	velocity   r3.Vector
	accelration r3.Vector
	curr_node *qtNode
}

func (b *boid) isInFOV(other *boid) bool {
    // Vector from my position to target
    distance := b.position.Distance(other.position)
    
    if distance > float64(perception) {
        return false
    }
	myForward := b.velocity.Normalize()
    toTargetNorm := other.position.Normalize()


    // Dot product to get cosine of the angle
    dot := toTargetNorm.Dot(myForward)

    // Convert view angle to cosine for comparison
    cosFOV := float64(math.Cos(float64(FOV * 0.5 * (math.Pi / 180.0)))) // Convert degrees to radians

    return dot >= cosFOV // Inside FOV if the dot product is greater than the threshold
}
// What should boid have:
// 	initial position
// 	initial velocity
// 	initial accelration
func Newboid(MaxHeight float64,MaxWidth float64) *boid{
	return &boid{
		position: randomVec3(MaxHeight,MaxWidth),
		velocity: randomVec3(1,1),
		accelration: randomVec3(1,1),
	}
}

func (b *boid) Update() {
	b.edge(screenWidth,screenHeight)
	b.position = b.position.Add(b.velocity)
	b.velocity = b.velocity.Add(b.accelration)
	speed := b.velocity.Norm()
    if speed > maxVelocity {
        // Scale the velocity vector down to maxVelocity while preserving direction
        b.velocity = b.velocity.Mul(maxVelocity / speed)
    }
}

func  (b *boid) edge(ScreenWidth,screenHeight int){

	Width := float64(ScreenWidth)
	Height := float64(screenHeight)
	if b.position.X > Width{
		b.position.X = 0
	}
	if b.position.X < 0{
		b.position.X = Width
	}
	if b.position.Y > Height{
		b.position.Y = 0
	}
	if b.position.Y < 0{
		b.position.Y = Height
	}
}

// func populate(number int,MaxHeight float64,MaxWidth float64) []*boid{
// 	population := make([]*boid, number)
// 	for i := 0; i < number; i++ {
// 		population[i] = Newboid(MaxHeight, MaxWidth)
// 	}
// 	return population
// }


func populate(number int,MaxHeight float64,MaxWidth float64) *qtNode{
	root := createNode(boundary{
		x1 : 0,
		x2 :int(MaxWidth),
		y1 :0,
		y2 :int(MaxHeight),
	})
	for i := 0; i < number; i++ {
		root.Insert(Newboid(MaxHeight, MaxWidth))
	}
	return root
}

func (b *boid) checkNode() bool{
	return b.curr_node.boundary.contains(int(b.position.X),int(b.position.Y))
}

func (b *boid) steer(flock []*boid) r3.Vector{

	desired := r3.NewPreciseVector(0,0,0).Vector()
	steer := r3.NewPreciseVector(0,0,0).Vector()
	
	count := 0
	for _,other := range flock{
		if b.isInFOV(other){
		// if b.isInFOV(other){
			desired = desired.Add(other.velocity)
			count++
		}
	}
	if count > 0 {
		div := 1.0 / float64(count) // Ensure non-zero count
		desired = desired.Mul(div)

		// Normalize desired velocity
		if desired.Norm() > 0 {
			desired = desired.Mul(maxVelocity/ desired.Norm()) // Adjust to maxSpeed
		}

		steer = desired.Sub(b.velocity) // Calculate steering force

		// Limit the steering force
		if steer.Norm() > float64(maxAlignmentForce) {
			steer = steer.Mul(float64(maxAlignmentForce) / steer.Norm())
		}
	}
	return steer
}

func (b *boid) cohesiveForce(flock []*boid) r3.Vector{

	desired := r3.NewPreciseVector(0,0,0).Vector()
	steer := r3.NewPreciseVector(0,0,0).Vector()
	
	count := 0
	for _,other := range flock{
		if other != b && b.position.Distance(other.position) <= float64(perception){
			desired = desired.Add(other.velocity.Mul(-1.00))
			count++
		}
	}
	if count > 0 {
		div := 1.0 / float64(count) // Ensure non-zero count
		desired = desired.Mul(div)

		// Normalize desired velocity
		if desired.Norm() > 0 {
			desired = desired.Mul(maxVelocity/ desired.Norm()) // Adjust to maxSpeed
		}

		steer = desired.Sub(b.velocity) // Calculate steering force

		// Limit the steering force
		if steer.Norm() > float64(MaxCohesiveFroce) {
			steer = steer.Mul(float64(MaxCohesiveFroce) / steer.Norm())
		}
	}
	return steer
}

func (b *boid) seprativeForce(flock []*boid) r3.Vector{

	desired := r3.NewPreciseVector(0,0,0).Vector()
	steer := r3.NewPreciseVector(0,0,0).Vector()
	
	count := 0
	for _,other := range flock{
		if other != b && b.position.Distance(other.position) <= float64(perception){
			diff := b.position.Sub(other.position)
			diff = diff.Mul(1.0/b.position.Distance(other.position))
			desired = desired.Add(diff)
			count++
		}
	}
	if count > 0 {
		div := 1.0 / float64(count) // Ensure non-zero count
		desired = desired.Mul(div)

		// Normalize desired velocity
		if desired.Norm() > 0 {
			desired = desired.Mul(maxVelocity/ desired.Norm()) // Adjust to maxVelocity
		}

		steer = desired.Sub(b.velocity) // Calculate steering force

		// Limit the steering force
		if steer.Norm() > float64(maxSeprationForce) {
			steer = steer.Mul(float64(maxSeprationForce) / steer.Norm())
		}
	}
	return steer

}


func (b *boid) flock(flock []*boid) {
	b.accelration = r3.NewPreciseVector(0.0,0.0,0.0).Vector()
	alignment := b.steer(flock)
	cohesion := b.cohesiveForce(flock)
	sepration := b.seprativeForce(flock)
	b.accelration = b.accelration.Add(alignment)
	b.accelration = b.accelration.Add(cohesion)
	b.accelration = b.accelration.Add(sepration)
}

func DeepCopyBoids(boids []*boid) []*boid {
    copyBoids := make([]*boid, len(boids))
    for i, b := range boids {
        copyBoids[i] = &boid{
            position: b.position,
            velocity: b.velocity,
			accelration: b.accelration,
			curr_node: b.curr_node,
        }
    }
    return copyBoids
}

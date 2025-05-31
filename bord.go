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
	raw := other.position.Sub(b.position)
	toTargetNorm := raw.Normalize()
  

    // Dot product to get cosine of the angle
    dot := toTargetNorm.Dot(myForward)

    // Convert view angle to cosine for comparison
    cosFOV := float64(math.Cos(float64(FOV * 0.5 * (math.Pi / 180.0)))) // Convert degrees to radians

    return dot >= cosFOV // Inside FOV if the dot product is greater than the threshold
}

func Newboid(MaxHeight float64,MaxWidth float64) *boid{
	return &boid{
		position: randomVec3(MaxHeight,MaxWidth),
		velocity: randomVec3(1,1),
		accelration: randomVec3(1,1),
	}
}

func (b *boid) checkCollision(other *boid) bool {
	return b.position.Distance(other.position) < float64(radius*2)
}

func (b *boid) Update() {
    b.position = b.position.Add(b.velocity)
    b.velocity = b.velocity.Add(b.accelration)
    speed := b.velocity.Norm()

    b.edge(screenWidth, screenHeight)

    if speed > maxVelocity {
        b.velocity = b.velocity.Mul(maxVelocity / speed)
    }

    if !b.checkNode() {
        // Remove from the old leaf
        oldLeaf := b.curr_node
        oldLeaf.removeBoid(b)

        // Climb up until `node.boundary` contains the boid
        node := oldLeaf
        for node != nil && !node.boundary.contains(int(b.position.X), int(b.position.Y)) {
            node = node.parent
        }

        // Re‐insert at the correct ancestor
        if node != nil {
            node.Insert(b)
        }
    }
}

func wrap(value, max float64) float64 {
    value = math.Mod(value, max)
    if value < 0 {
        value += max
    }
    return value
}

func (b *boid) edge(screenWidth, screenHeight int) {
    width := float64(screenWidth)
    height := float64(screenHeight)

    b.position.X = wrap(b.position.X, width)
    b.position.Y = wrap(b.position.Y, height)
}

func qt_populate(number int,MaxHeight float64,MaxWidth float64) *qtNode{
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

func (b *boid) steer(flock []*boid) r3.Vector {
    desired := r3.NewPreciseVector(0,0,0).Vector()
    steer   := r3.NewPreciseVector(0,0,0).Vector()
    
    count := 0
    for _, other := range flock {
        if b.isInFOV(other) {
            desired = desired.Add(other.velocity)
            count++
        }
    }
    if count > 0 {
        div := 1.0 / float64(count)
        desired = desired.Mul(div)                // average neighbor velocity
        if desired.Norm() > 0 {
            desired = desired.Mul(maxVelocity / desired.Norm()) // scale to maxVelocity
        }
        steer = desired.Sub(b.velocity)           // "steering force" = desired minus current
        // clamp to maxAlignmentForce:
        if steer.Norm() > float64(maxAlignmentForce) {
            steer = steer.Mul(float64(maxAlignmentForce) / steer.Norm())
        }
    }
    return steer
}


func (b *boid) cohesiveForce(flock []*boid) r3.Vector {
    // 1. Sum neighbor positions
    sumPos := r3.NewPreciseVector(0,0,0).Vector()
    count  := 0

    for _, other := range flock {
        if other != b {
            d := b.position.Distance(other.position)
            if d <= float64(perception) {
                sumPos = sumPos.Add(other.position)
                count++
            }
        }
    }

    if count == 0 {
        return r3.NewPreciseVector(0,0,0).Vector() // no neighbors → no cohesion force
    }

    // 2. Compute center of mass
    centerOfMass := sumPos.Mul(1.0 / float64(count))

    // 3. Desired direction = (centerOfMass − currentPosition), scaled to maxVelocity
    desired := centerOfMass.Sub(b.position)
    if desired.Norm() > 0 {
        desired = desired.Mul(maxVelocity / desired.Norm())
    } else {
        // If the boid is exactly at the center of mass (rare), no steering needed
        return r3.NewPreciseVector(0,0,0).Vector()
    }

    // 4. Steering force = desired – current velocity, clamped to MaxCohesiveFroce
    steer := desired.Sub(b.velocity)
    if steer.Norm() > float64(MaxCohesiveFroce) {
        steer = steer.Mul(float64(MaxCohesiveFroce) / steer.Norm())
    }
    return steer
}



func (b *boid) seprativeForce(flock []*boid) r3.Vector {
    desired := r3.NewPreciseVector(0,0,0).Vector()
    steer   := r3.NewPreciseVector(0,0,0).Vector()
    count   := 0

    for _, other := range flock {
        if other != b && b.position.Distance(other.position) <= float64(perception) {
            // diff is a unit vector pointing from other → b, scaled by 1/distance
            diff := b.position.Sub(other.position)
            diff = diff.Mul(1.0 / b.position.Distance(other.position))
            desired = desired.Add(diff)
            count++
        }
    }

    if count > 0 {
        invCount := 1.0 / float64(count)
        desired = desired.Mul(invCount) // average push‐away vector

        if desired.Norm() > 0 {
            desired = desired.Mul(maxVelocity / desired.Norm())
        }

        steer = desired.Sub(b.velocity)
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

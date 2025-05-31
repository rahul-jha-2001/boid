package main

import (
	rl "github.com/gen2brain/raylib-go/raylib"
)


type qtNode struct {
	boundary       boundary
	flock          []*boid
	NE, NW, SE, SW *qtNode
	divided        bool
	parent         *qtNode // Add this field
}

func createNode(b boundary) *qtNode {

	if b.x1 > b.x2 || b.y1 > b.y2 {
		return nil
	}

	return &qtNode{
		boundary: b,
		divided:  false,
		flock:    []*boid{},
	}
}



func (qt *qtNode) Insert(b *boid) bool {
	// First, check if the point is within boundary
	if !qt.boundary.contains(int(b.position.X), int(b.position.Y)) {
		return false
	}

	// If we have space and haven't divided yet
	if len(qt.flock) < boidPerQT && !qt.divided {
		b.curr_node = qt
		qt.flock = append(qt.flock, b)
		return true
	}

	// If we haven't divided yet but need to
	if !qt.divided {
		qt.subdivide()
		// Redistribute existing boids
		for _, boid := range qt.flock {
			qt.NW.Insert(boid)
			qt.NE.Insert(boid)
			qt.SW.Insert(boid)
			qt.SE.Insert(boid)
		}
		qt.flock = nil // Clear the parent's flock
		qt.divided = true
	}

	// Try to insert into children
	if qt.NW.Insert(b) {
		return true
	}
	if qt.NE.Insert(b) {
		return true
	}
	if qt.SW.Insert(b) {
		return true
	}
	if qt.SE.Insert(b) {
		return true
	}

	return false // If we couldn't insert anywhere
}

func (qt *qtNode) subdivide() {
	// Get center coordinates
	centerX := (qt.boundary.x1 + qt.boundary.x2) / 2
	centerY := (qt.boundary.y1 + qt.boundary.y2) / 2

	// Create NW quadrant (top-left)
	qt.NW = createNode(boundary{
		x1: qt.boundary.x1,
		y1: qt.boundary.y1,
		x2: centerX,
		y2: centerY,
	})
	qt.NW.parent = qt

	// Create NE quadrant (top-right)
	qt.NE = createNode(boundary{
		x1: centerX,
		y1: qt.boundary.y1,
		x2: qt.boundary.x2,
		y2: centerY,
	})
	qt.NE.parent = qt

	// Create SW quadrant (bottom-left)
	qt.SW = createNode(boundary{
		x1: qt.boundary.x1,
		y1: centerY,
		x2: centerX,
		y2: qt.boundary.y2,
	})
	qt.SW.parent = qt

	// Create SE quadrant (bottom-right)
	qt.SE = createNode(boundary{
		x1: centerX,
		y1: centerY,
		x2: qt.boundary.x2, // Added missing x2
		y2: qt.boundary.y2,
	})
	qt.SE.parent = qt

	qt.divided = true
}

func (qt *qtNode) Draw() {
	// Draw boundary
	rl.DrawRectangleLines(int32(qt.boundary.x1), int32(qt.boundary.y1),
		int32(qt.boundary.x2-qt.boundary.x1), int32(qt.boundary.y2-qt.boundary.y1), rl.LightGray)

	// Draw points in a batch
	for _, p := range qt.flock {
		rl.DrawCircle(int32(p.position.X),int32(p.position.Y), 2, rl.Red)
	}

	// Draw children recursively
	if qt.divided {
		qt.NE.Draw()
		qt.NW.Draw()
		qt.SE.Draw()
		qt.SW.Draw()
	}
}


func (qt *qtNode) query() []*boid {

	flock := make([]*boid, 0)

	// Check if current node's boundary intersects with the query area
	for _, b := range qt.flock {
		if qt.boundary.contains(int(b.position.X), int(b.position.Y))  {
			flock = append(flock, b)
		}
	}

	// If subdivided, check children recursively
	if qt.divided {
		flock = append(flock, qt.NE.query()...)
		flock = append(flock, qt.NW.query()...)
		flock = append(flock, qt.SE.query()...)
		flock = append(flock, qt.SW.query()...)
	}

	return flock
}

func (qt *qtNode) update() {
	// Update all boids in current nod
	flock := qt.query()
	for _, b := range flock {
		b.flock(qt.flock)
		b.Update()
	}
	// Update children recursively
	if qt.divided {
		qt.NE.update()
		qt.NW.update()
		qt.SE.update()
		qt.SW.update()
	}

	// Cleanup empty subdivisions
	if qt.divided && qt.isEmpty() {
		qt.merge()
	}
}

// Helper method to remove a boid from the flock
func (qt *qtNode) removeBoid(b *boid) {
	for i, boid := range qt.flock {
		if boid == b {
			qt.flock[i] = qt.flock[len(qt.flock)-1]
			qt.flock = qt.flock[:len(qt.flock)-1]
			return
		}
	}
}

// Check if node and all children are empty
func (qt *qtNode) isEmpty() bool {
	if len(qt.flock) > 0 {
		return false
	}
	if qt.divided {
		return qt.NE.isEmpty() && qt.NW.isEmpty() &&
			qt.SE.isEmpty() && qt.SW.isEmpty()
	}
	return true
}

// Merge empty subdivisions
func (qt *qtNode) merge() {
	// Clear parent references before nulling
	if qt.NE != nil {
		qt.NE.parent = nil
		qt.NE = nil
	}
	if qt.NW != nil {
		qt.NW.parent = nil
		qt.NW = nil
	}
	if qt.SE != nil {
		qt.SE.parent = nil
		qt.SE = nil
	}
	if qt.SW != nil {
		qt.SW.parent = nil
		qt.SW = nil
	}
	qt.divided = false
	qt.flock = make([]*boid, 0) // Reset flock slice
}

func (qt *qtNode) Cleanup() {
	if qt == nil {
		return
	}

	// Cleanup children first
	if qt.divided {
		qt.NE.Cleanup()
		qt.NW.Cleanup()
		qt.SE.Cleanup()
		qt.SW.Cleanup()
	}

	// Clear all references
	qt.flock = nil
	qt.NE = nil
	qt.NW = nil
	qt.SE = nil
	qt.SW = nil
	qt.parent = nil
}

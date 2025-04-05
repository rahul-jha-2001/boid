package main

import (
	rl "github.com/gen2brain/raylib-go/raylib"
)

type boundary struct {
	x1, y1 int // top-left corner
	x2, y2 int // bottom-right corner
}

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

func (b boundary) contains(x, y int) bool {
	return x >= b.x1 && x <= b.x2 &&
		y >= b.x1 && y <= b.y2
}

func (b boundary) dimensions() (width, height int) {
	return b.x2 - b.x1, b.y2 - b.y1
}

func (b boundary) center() (width, height int) {
	return b.x2 / 2, b.y2 / 2
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
		rl.DrawCircleV(rl.Vector2{X: float32(p.position.X), Y: float32(p.position.Y)}, 2, rl.Red)
	}

	// Draw children recursively
	if qt.divided {
		if qt.NE != nil {
			qt.NE.Draw()
		}
		if qt.NW != nil {
			qt.NW.Draw()
		}
		if qt.SE != nil {
			qt.SE.Draw()
		}
		if qt.SW != nil {
			qt.SW.Draw()
		}
	}
}
func (qt *qtNode) update() {
	// Update all boids in current node
	snapshot := DeepCopyBoids(qt.flock)
	for _, b := range qt.flock {
		// Store old position for boundary check
		// oldX, oldY := b.position.X, b.position.Y

		
		// Update boid behavior and position
		b.flock(snapshot)
		b.Update()

		// Check if boid moved outside current boundary
		if b.checkNode() {
			// Remove from current node
			qt.removeBoid(b)

			// Reinsert into root node
			root := qt
			for root.parent != nil {
				root = root.parent
			}
			root.Insert(b)
		}
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
			// Remove boid by swapping with last element and truncating
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
	qt.NE = nil
	qt.NW = nil
	qt.SE = nil
	qt.SW = nil
	qt.divided = false
}

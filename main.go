package main

import (


	gui "github.com/gen2brain/raylib-go/raygui"
	rl "github.com/gen2brain/raylib-go/raylib"
)
const (
    screenWidth  = 800
    screenHeight = 600
    maxVelocity  = 2.0
	// perception = 50
	// maxAlignmentForce = 1
	// MaxCohesiveFroce = 1
	// maxSeprationForce = 1
	NumberofSpeices = 5
	radius = 2
	numberOfPaticles = 100
	boidPerQT = 1
)

var maxAlignmentForce float32 = float32(0.5) // Initial slider value
var MaxCohesiveFroce float32= float32(0.5) // Initial slider value
var maxSeprationForce float32= float32(0.5) // Initial slider value
var perception float32 = 2
var FOV float32 = 20.0 
func main() {	// Initialize window
	

	rl.InitWindow(screenWidth, screenHeight, "Raylib in Go")
	defer rl.CloseWindow() // Ensure window closes properly

	// Set FPS
	rl.SetTargetFPS(60)
	sim_start := false
	flock_root :=  populate(numberOfPaticles,screenHeight,screenWidth)
	// Game loop
	for !rl.WindowShouldClose() {
		// Update game logic
		
		if rl.IsKeyPressed(rl.KeyEnter){
			sim_start = !sim_start
		}
		
		if sim_start{
		update(flock_root)
		}
		// Draw everything
		rl.BeginDrawing()
		rl.ClearBackground(rl.RayWhite)
		draw(flock_root)
		rl.EndDrawing()
	}
}



// Update game logic
func update(qt *qtNode) {

	// for _,boid := range flock{
	// 	boid.flock(flock)
	// 	boid.Update()
		
	// }
	qt.update()

	
}

// Draw everything
func draw(flock *qtNode) {
	rl.DrawFPS(0,0)
	maxAlignmentForce = gui.Slider(
		rl.NewRectangle(600, 0, 200, 20),
		"maxAlignmentForce", 
		"", 
		maxAlignmentForce, 
		0.0, 
		1.0,
	)
	MaxCohesiveFroce = gui.Slider(
		rl.NewRectangle(600, 20, 200, 20),
		"MaxCohesiveFroce", 
		"", 
		MaxCohesiveFroce, 
		0.0, 
		1.0,
	)
	maxSeprationForce = gui.Slider(
		rl.NewRectangle(600, 40, 200, 20),
		"maxSeprationForce", 
		"", 
		maxSeprationForce, 
		0.0, 
		1.0,
	)
	perception = gui.Slider(
		rl.NewRectangle(600, 60, 200, 20),
		"perception", 
		"", 
		perception, 
		1.0, 
		600.0,
	)
	flock.Draw()
	
	

}


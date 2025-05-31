package main

import (


	gui "github.com/gen2brain/raylib-go/raygui"
	rl "github.com/gen2brain/raylib-go/raylib"
)
const (
    screenWidth  = 800
    screenHeight = 600
    maxVelocity  = 1.0
	// perception = 50
	// maxAlignmentForce = 1
	// MaxCohesiveFroce = 1
	// maxSeprationForce = 1
	radius = 2
	numberOfPaticles = 1000
	boidPerQT = 10
)

var maxAlignmentForce float32 = float32(0.5) // Initial slider value
var MaxCohesiveFroce float32= float32(0.5) // Initial slider value
var maxSeprationForce float32= float32(0.5) // Initial slider value
var perception float32 = 2
var FOV float32 = 360.0 
func main() {	// Initialize window
	

	rl.InitWindow(screenWidth, screenHeight, "Raylib in Go")
	defer rl.CloseWindow() // Ensure window closes properly

	// Set FPS
	rl.SetTargetFPS(60)
	sim_start := false
	// flock :=  populate(numberOfPaticles,screenHeight,screenWidth)
	flock_root := qt_populate(numberOfPaticles,screenHeight,screenWidth)
	// Game loop
	for !rl.WindowShouldClose() {
		// Update game logic
		
		if rl.IsKeyPressed(rl.KeyEnter){
			sim_start = !sim_start
		}
		
		if sim_start{
		// update(flock)
		update_qt(flock_root)
	}
		rl.BeginDrawing()
		rl.ClearBackground(rl.RayWhite)
		// draw(flock)
		draw_qt(flock_root)
		rl.EndDrawing()
	}
}



func update_qt(qt *qtNode){
	qt.update()
}

func draw_qt(qt *qtNode) {
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
	FOV = gui.Slider(
		rl.NewRectangle(600, 80, 200, 20),
		"FOV", 
		"", 
		FOV, 
		1.0, 
		360.0,
	)
	qt.Draw()
}
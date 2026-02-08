package main

import (
	"image"
	"image/color"
	"math"
	"sync"

	"github.com/hajimehoshi/ebiten/v2"
)

var (
	circleOnce sync.Once
	circleImg  *ebiten.Image
)

// white circle alpha mask on transparent bg
func getCircleSprite() *ebiten.Image {
	circleOnce.Do(func() {
		const size = 128
		img := image.NewRGBA(image.Rect(0, 0, size, size))
		cx, cy := float64(size-1)/2, float64(size-1)/2
		r := float64(size) * 0.45

		for y := 0; y < size; y++ {
			for x := 0; x < size; x++ {
				dx := float64(x) - cx
				dy := float64(y) - cy
				if math.Sqrt(dx*dx+dy*dy) <= r {
					img.SetRGBA(x, y, color.RGBA{255, 255, 255, 255})
				} else {
					img.SetRGBA(x, y, color.RGBA{0, 0, 0, 0})
				}
			}
		}
		circleImg = ebiten.NewImageFromImage(img)
	})
	return circleImg
}

func drawParticle(screen *ebiten.Image, p Particle, col color.Color) {
	sprite := getCircleSprite()

	op := &ebiten.DrawImageOptions{}

	// scale sprite to diameter = 2*radius
	sw, sh := sprite.Size()
	scaleX := (2 * p.Radius) / float64(sw)
	scaleY := (2 * p.Radius) / float64(sh)

	op.GeoM.Scale(scaleX, scaleY)
	op.GeoM.Translate(p.Pos.X-p.Radius, p.Pos.Y-p.Radius)

	// Tint color
	r, g, b, a := col.RGBA()
	op.ColorM.Scale(
		float64(r)/65535.0,
		float64(g)/65535.0,
		float64(b)/65535.0,
		float64(a)/65535.0,
	)

	screen.DrawImage(sprite, op)
}

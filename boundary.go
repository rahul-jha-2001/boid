package main

import (

)


type boundary struct {
	x1, y1 int // top-left corner
	x2, y2 int // bottom-right corner
}

func (b boundary) contains(x, y int) bool {
	return x >= b.x1 && x < b.x2 && y >= b.y1 && y < b.y2
}

func (b boundary) dimensions() (width, height int) {
	return b.x2 - b.x1, b.y2 - b.y1
}

func (b boundary) center() (width, height int) {
	return b.x2 / 2, b.y2 / 2
}
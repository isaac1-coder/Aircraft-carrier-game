// +build js,wasm

package main

import (
	"math"
	"syscall/js"
)

// Game Constants
const (
	STATE_WALKING = 0
	STATE_FLYING  = 1
	STATE_LIFT    = 2
)

type Vector3 struct {
	X, Y, Z float64
}

type Player struct {
	Pos      Vector3
	Rot      Vector3
	State    int
	Velocity Vector3
}

var player Player
var canvas js.Value
var ctx js.Value

func main() {
	c := make(chan struct{}, 0)

	// Initialize Canvas
	doc := js.Global().Get("document")
	canvas = doc.Call("getElementById", "gameCanvas")
	ctx = canvas.Call("getContext", "2d")

	// Set resolution
	width := js.Global().Get("innerWidth").Int()
	height := js.Global().Get("innerHeight").Int()
	canvas.Set("width", width)
	canvas.Set("height", height)

	player = Player{
		Pos:   Vector3{X: 0, Y: 10, Z: 0}, // On the deck
		State: STATE_WALKING,
	}

	// Input Listeners
	js.Global().Call("addEventListener", "keydown", js.FuncOf(handleInput))

	// Start Game Loop
	var renderFrame js.Func
	renderFrame = js.FuncOf(func(this js.Value, args []js.Value) interface{} {
		update()
		draw()
		js.Global().Call("requestAnimationFrame", renderFrame)
		return nil
	})
	js.Global().Call("requestAnimationFrame", renderFrame)

	<-c
}

func handleInput(this js.Value, args []js.Value) interface{} {
	key := args[0].Get("key").String()
	
	switch key {
	case "w": player.Pos.Z += math.Cos(player.Rot.Y) * 2
	case "s": player.Pos.Z -= math.Cos(player.Rot.Y) * 2
	case "a": player.Rot.Y -= 0.1
	case "d": player.Rot.Y += 0.1
	case "f":
		if player.State == STATE_WALKING {
			player.State = STATE_LIFT // Start elevator sequence
		}
	}
	return nil
}

func update() {
	if player.State == STATE_LIFT {
		player.Pos.Y += 0.1
		if player.Pos.Y > 30 {
			player.State = STATE_FLYING
		}
	} else if player.State == STATE_FLYING {
		// Flight Physics
		player.Pos.Z += 5.0 // Constant forward thrust
		player.Pos.Y += math.Sin(player.Rot.X) * 5
	}
}

func draw() {
	width := canvas.Get("width").Float()
	height := canvas.Get("height").Float()

	// Clear screen (Ocean)
	ctx.Set("fillStyle", "#004466")
	ctx.Call("fillRect", 0, 0, width, height)

	// Draw Aircraft Carrier (Pseudo-3D Projection)
	ctx.Set("fillStyle", "#555555")
	// Drawing the flight deck
	ctx.Call("beginPath")
	ctx.Call("moveTo", width/2-200, height/2+100)
	ctx.Call("lineTo", width/2+200, height/2+100)
	ctx.Call("lineTo", width/2+150, height/2-100)
	ctx.Call("lineTo", width/2-150, height/2-100)
	ctx.Call("closePath")
	ctx.Call("fill")

	// Draw UI Overlay
	doc := js.Global().Get("document")
	statusText := "Mode: DECK WALK | Press 'F' to Enter Plane"
	if player.State == STATE_FLYING {
		statusText = "Mode: FLIGHT | W/S: Pitch | A/D: Roll"
	}
	doc.Call("getElementById", "status").Set("innerHTML", statusText)
}

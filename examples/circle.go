package main

import (
	"github.com/kbatten/httpcanvas"
	"math"
	"time"
)

func app(context *httpcanvas.Context) {
	centerX := context.Width / 2
	centerY := context.Height / 2

	context.BeginPath()
	context.Arc(centerX, centerY, 70, 0, 2*math.Pi, false)
	context.FillStyle("green")
	context.Fill()

	time.Sleep(5 * time.Second)

	context.LineWidth(5)
	context.StrokeStyle("#003300")
	context.Stroke()
}

func main() {
	httpcanvas.ListenAndServe(":8080", app)
}

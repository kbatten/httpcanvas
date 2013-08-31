package main

import (
	"github.com/kbatten/httpcanvas"
	"time"
)

func app(context *httpcanvas.Context) {
	for i := 0.0; i < 100; i++ {
		context.BeginPath()
		context.MoveTo(50+4*i, 50)
		context.LineTo(50+4*i, 100)
		context.Stroke()

		time.Sleep(1 * time.Second)
	}
}

func main() {
	httpcanvas.ListenAndServe(":8080", app)
}

package main

import (
	"github.com/kbatten/httpcanvas"
)

func app(context *httpcanvas.Context) {
	context.BeginPath()
    context.MoveTo(100, 150)
    context.LineTo(450, 50)
    context.Stroke()
}

func main() {
	httpcanvas.ListenAndServe(":8080", app)
}

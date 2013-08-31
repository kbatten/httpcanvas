package httpcanvas

import (
	"fmt"
)

type Context struct {
	command chan string
	Width   float64
	Height  float64
}

func (c *Context) BeginPath() {
	c.command <- "beginPath"
}

func (c *Context) MoveTo(x, y float64) {
	c.command <- fmt.Sprintf("moveTo|%f|%f", x, y)
}

func (c *Context) LineTo(x, y float64) {
	c.command <- fmt.Sprintf("lineTo|%f|%f", x, y)
}

func (c *Context) Stroke() {
	c.command <- "stroke"
}

func (c *Context) Arc(x, y float64, radius float64, startAngle, endAngle float64, anticlockwise bool) {
	c.command <- fmt.Sprintf("arc|%f|%f|%f|%f|%f|%v", x, y, radius, startAngle, endAngle, anticlockwise)
}

func (c *Context) FillStyle(s string) {
	c.command <- fmt.Sprintf("fillStyle|%s", s)
}

func (c *Context) Fill() {
	c.command <- "fill"
}

func (c *Context) LineWidth(f float64) {
	c.command <- fmt.Sprintf("lineWidth|%f", f)
}

func (c *Context) StrokeStyle(s string) {
	c.command <- fmt.Sprintf("strokeStyle|%s", s)
}

func (c *Context) FillRect(x, y, w, h float64) {
	c.command <- fmt.Sprintf("fillRect|%f|%f|%f|%f", x, y, w, h)
}

func (c *Context) StrokeRect(x, y, w, h float64) {
    c.command <- fmt.Sprintf("strokeRect|%f|%f|%f|%f", x, y, w, h)
}

package httpcanvas

import (
	"fmt"
)

type Context struct {
	Width  float64
	Height float64

	command chan string
	mouse   chan mouseMovement
	mouseX  float64
	mouseY  float64
	mouseClickedX float64
	mouseClickedY float64
	mouseClicked       bool
}

func newContext(w, h float64, c chan string, m chan mouseMovement) *Context {
	return &Context{w, h, c, m, 0, 0, 0, 0, false}
}

func (c *Context) updateMouse() {
	reading := true
    for reading {
        select {
        case m := <-c.mouse:
			switch m.command {
			case "MOUSECLICK":
				c.mouseClicked = true
				c.mouseClickedX = m.x
				c.mouseClickedY = m.y
			case "MOUSEMOVE":
				c.mouseX = m.x
				c.mouseY = m.y
			}
        default:
            reading = false
        }
    }
}

func (c *Context) MouseLocation() (float64, float64) {
	c.updateMouse()
	return c.mouseX, c.mouseY
}

func (c *Context) MouseClicked() (float64, float64, bool) {
	c.updateMouse()
	clicked := c.mouseClicked
	c.mouseClicked = false
	return c.mouseClickedX, c.mouseClickedY, clicked
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

func (c *Context) ClearRect(x, y, w, h float64) {
	c.command <- fmt.Sprintf("clearRect|%f|%f|%f|%f", x, y, w, h)
}

func (c *Context) ClearFrame() {
	c.command <- "CLEARFRAME" // erase buffered frame
}

func (c *Context) ShowFrame() {
	c.command <- "SHOWFRAME" // swap buffer
}

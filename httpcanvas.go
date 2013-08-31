package httpcanvas

import (
	"fmt"
	"net/http"
)

type CanvasHandler func(*Context)

type Canvas struct {
	handler CanvasHandler
	Width   float64
	Height  float64
	command chan string
}

func newCanvas(handler CanvasHandler) *Canvas {
	return &Canvas{handler, 640, 480, make(chan string)}
}

func (c *Canvas) writeContainer(w http.ResponseWriter, r *http.Request) {
	container := `<!DOCTYPE HTML>
<!-- http://www.html5canvastutorials.com/tutorials/html5-canvas-lines/ -->
<html>
  <head>
    <style>
      body {
        margin: 0px;
        padding: 0px;
      }
      .displayBox {
        border: 1px dashed rgb(170, 170, 170)
      }
    </style>
  </head>
  <body>
    <canvas id="myCanvas" class="displayBox"
      width="` + fmt.Sprintf("%d", int(c.Width)) + `"
      height="` + fmt.Sprintf("%d", int(c.Height)) + `"></canvas>
    <script>
      function getNextCommand() {
        xmlHttp = new XMLHttpRequest();
        xmlHttp.open("GET", "/command", false);
        xmlHttp.send(null);
        return xmlHttp.responseText;
      }

      function parseBool(b) {
        return b == "true"
      }

      var canvas = document.getElementById('myCanvas');
      var context = canvas.getContext('2d');

      for (;;) {
        command = getNextCommand().split(" ")
        console.log(command);
        if (command[0] == "END") {
          break;
        } else if (command[0] == "beginPath") {
          context.beginPath();
        } else if (command[0] == "moveTo") {
          context.moveTo(parseFloat(command[1]), parseFloat(command[2]));
        } else if (command[0] == "lineTo") {
          context.lineTo(parseFloat(command[1]), parseFloat(command[2]));
        } else if (command[0] == "stroke") {
          context.stroke();
        } else if (command[0] == "arc") {
          context.arc(parseFloat(command[1]),
            parseFloat(command[2]),
            parseFloat(command[3]),
            parseFloat(command[4]),
            parseFloat(command[5]),
            parseBool(command[6]));
        } else if (command[0] == "fillStyle") {
            context.fillStyle = command[1]
        } else if (command[0] == "fill") {
            context.fill()
        } else if (command[0] == "lineWidth") {
            context.lineWidth = parseFloat(command[1])
        } else if (command[0] == "strokeStyle") {
            context.strokeStyle = command[1]
        }
      }
    </script>
  </body>
</html>`
	fmt.Fprintf(w, container)
}

func (c Canvas) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.RequestURI == "/" && r.Method == "GET" {
		c.writeContainer(w, r)
		go func() {
			c.handler(&Context{c.command, c.Width, c.Height})
			c.command <- "END"
		}()
		return
	}

	if r.RequestURI == "/command" && r.Method == "GET" {
		command := <-c.command
		fmt.Fprintf(w, command)
		return
	}

	http.NotFound(w, r)
}

func ListenAndServe(addr string, handler CanvasHandler) (err error) {
	return http.ListenAndServe(addr, newCanvas(handler))
}

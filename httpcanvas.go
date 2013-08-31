package httpcanvas

import (
	"fmt"
	"net/http"
	"strings"
)

type CanvasHandler func(*Context)

type Canvas struct {
	handler CanvasHandler
	Width   float64
	Height  float64
	id      *int
	command chan string
}

func newCanvas(handler CanvasHandler) *Canvas {
	id := 0
	return &Canvas{handler, 640, 480, &id, make(chan string)}
}

func (c *Canvas) writeContainer(w http.ResponseWriter, r *http.Request) {
	// sync
	id := *c.id
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
        xmlHttp.open("GET", "/command?id=` + fmt.Sprintf("%d", id) + `", false);
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
	command, _, args := stringPartition(r.RequestURI, "?")

	if command == "/" && r.Method == "GET" {
		// TODO: sync
		(*c.id)++
		id := *c.id
		c.writeContainer(w, r)
		if id == 1 {
			go func() {
				c.handler(&Context{c.command, c.Width, c.Height})
				c.command <- "END"
			}()
		}
		return
	}

	// TODO: sync
	idExpected := fmt.Sprintf("id=%d", *c.id)

	if command == "/command" && r.Method == "GET" {
		if args == idExpected {
			command := <-c.command
			fmt.Fprintf(w, command)
			return
		} else {
			fmt.Fprintf(w, "END")
			return
		}
	}

	http.NotFound(w, r)
}

func ListenAndServe(addr string, handler CanvasHandler) (err error) {
	return http.ListenAndServe(addr, newCanvas(handler))
}

func stringPartition(s, sep string) (string, string, string) {
	sepPos := strings.Index(s, sep)
	if sepPos == -1 { // no seperator found
		return s, "", ""
	}
	split := strings.SplitN(s, sep, 2)
	return split[0], sep, split[1]
}

package httpcanvas

import (
	"fmt"
	"math/rand"
	"net/http"
	"html/template"
	"strings"
)

type CanvasHandler func(*Context)

type Canvas struct {
	handler CanvasHandler
	Width   float64
	Height  float64
	Unique  string
	started bool
	command chan string
}

func newCanvas(handler CanvasHandler) *Canvas {
	return &Canvas{handler, 640, 480, "", false, make(chan string, 1000)}
}

func (c *Canvas) updateUnique() {
	c.Unique = fmt.Sprintf("%f", rand.Float64())
}

func (c *Canvas) renderHtml(w http.ResponseWriter, r *http.Request) error {
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
    <canvas id="myCanvas" class="displayBox" width="{{.Width}}" height="{{.Height}}"></canvas>
    <script>
      xmlHttp = new XMLHttpRequest();
      currentData = []
      function getNextCommands() {
        if (currentData.length == 0) {
          try {
            xmlHttp.open("GET", "/command?id={{.Unique}}", false);
            xmlHttp.send(null);
            currentData = xmlHttp.responseText.split("~");
          } catch (e) {
            currentData = ["END"]
          }
        }
      }

      function parseBool(b) {
        return b == "true"
      }

      var canvas = document.getElementById('myCanvas');
      var context = canvas.getContext('2d');
      var intervalId = 0

      function executeNextCommands() {
        getNextCommands()
        while (currentData.length > 0) {
          command = currentData.shift().split("|")
          if (command[0] == "END") {
            clearInterval(intervalId)
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
          } else if (command[0] == "fillRect") {
              context.fillRect(parseFloat(command[1]),
                parseFloat(command[2]),
                parseFloat(command[3]),
                parseFloat(command[4]))
          } else if (command[0] == "strokeRect") {
              context.strokeRect(parseFloat(command[1]),
                parseFloat(command[2]),
                parseFloat(command[3]),
                parseFloat(command[4]))
          } else if (command[0] == "clearRect") {
              context.clearRect(parseFloat(command[1]),
                parseFloat(command[2]),
                parseFloat(command[3]),
                parseFloat(command[4]))
          }
        }
      }

      intervalId = setInterval("executeNextCommands()", 10)
    </script>
  </body>
</html>`
template, err := template.New("basic").Parse(container)
    if err != nil {
        return err
    }
    err = template.Execute(w, c)
    return err
}

func (c *Canvas) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	command, _, args := stringPartition(r.RequestURI, "?")

	if command == "/" && r.Method == "GET" {
		c.updateUnique()
		err := c.renderHtml(w, r)
		if err != nil {
			return
		}
		if !c.started {
			c.started = true
			go func() {
				c.handler(&Context{c.command, c.Width, c.Height})
				c.command <- "END"
			}()
		}
		return
	}

	uniqueExpected := fmt.Sprintf("id=%s", c.Unique)

	if command == "/command" && r.Method == "GET" {
		if args == uniqueExpected {
			commandGroup := ""
			command := " "
			for len(command) > 0 {
				select {
				case command = <-c.command:
					if len(commandGroup) > 0 {
						commandGroup += "~"
					}
					commandGroup += command
				default:
					// if we have at least one command, then send it off
					if len(commandGroup) > 0 {
						command = ""
					}
				}
			}
			fmt.Fprintf(w, commandGroup)
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

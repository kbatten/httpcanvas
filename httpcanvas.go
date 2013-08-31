package httpcanvas

import (
	"fmt"
	"math/rand"
	"net/http"
	"strings"
	"sync"
)

type CanvasHandler func(*Context)

type Canvas struct {
	handler CanvasHandler
	Width   float64
	Height  float64
	id      *int
	started *bool
	lock    *sync.RWMutex
	command chan string
}

func newCanvas(handler CanvasHandler) *Canvas {
	id := rand.Int()/2
	started := false
	return &Canvas{handler, 640, 480, &id, &started, &sync.RWMutex{}, make(chan string, 1000)}
}

func (c *Canvas) writeContainer(w http.ResponseWriter, r *http.Request) {
	// sync
	c.lock.RLock()
	id := *c.id
	c.lock.RUnlock()
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
      xmlHttp = new XMLHttpRequest();
      currentData = []
      function getNextCommands() {
        if (currentData.length == 0) {
          try {
            xmlHttp.open("GET", "/command?id=` + fmt.Sprintf("%d", id) + `", false);
            xmlHttp.send(null);
            currentData = xmlHttp.responseText.split(",");
            console.log(currentData.length)
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
          command = currentData.shift().split(" ")
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
          }else if (command[0] == "strokeRect") {
              context.strokeRect(parseFloat(command[1]),
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
	fmt.Fprintf(w, container)
}

func (c Canvas) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	command, _, args := stringPartition(r.RequestURI, "?")

	if command == "/" && r.Method == "GET" {
		// sync
		c.lock.Lock()
		(*c.id)++
		started := *c.started
		c.lock.Unlock()
		c.writeContainer(w, r)
		if !started {
			c.lock.Lock()
			(*c.started) = true
			go func() {
				c.handler(&Context{c.command, c.Width, c.Height})
				c.command <- "END"
			}()
			c.lock.Unlock()
		}
		return
	}

	// sync
	c.lock.RLock()
	id := *c.id
	c.lock.RUnlock()
	idExpected := fmt.Sprintf("id=%d", id)

	if command == "/command" && r.Method == "GET" {
		if args == idExpected {
			commandGroup := ""
			command := " "
			for len(command) > 0 {
				select {
				case command = <-c.command:
					if len(commandGroup) > 0 {
						commandGroup += ","
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

package httpcanvas

import (
	"fmt"
	"html/template"
	"log"
	"math/rand"
	"net/http"
	"net/url"
	"strings"
	"strconv"
)

type mouseMovement struct {
	command string
	x       float64
	y       float64
}

type CanvasHandler func(*Context)

type Canvas struct {
	handler CanvasHandler
	Width   float64
	Height  float64
	Unique  string
	started bool
	command chan string
	mouse   chan mouseMovement
}

func newCanvas(handler CanvasHandler) *Canvas {
	return &Canvas{handler, 640, 480, "", false,
		make(chan string, 10000),
		make(chan mouseMovement, 10000)}
}

func (c *Canvas) updateUnique() {
	c.Unique = fmt.Sprintf("%f", rand.Float64())
}

func (c *Canvas) renderHtml(w http.ResponseWriter, r *http.Request) error {
	container := `<!DOCTYPE HTML>
<!-- http://www.html5canvastutorials.com/tutorials/html5-canvas-lines/ -->
<html>
  <head>
    <script src="/jquery.js"></script>
    <style>
      body {
        margin: 0px;
        padding: 0px;
      }
      canvas {
        border: 1px dashed rgb(170, 170, 170);
        position:absolute; top:100px; left:100px;
        visibility: hidden;
      }
    </style>
  </head>
  <body>
    <canvas width="{{.Width}}" height="{{.Height}}"></canvas>
    <canvas width="{{.Width}}" height="{{.Height}}"></canvas>
    <script>
      var currentData = []
      var canvases = document.getElementsByTagName('canvas');
      var contexts = []
      contexts[0] = canvases[0].getContext('2d');
      contexts[1] = canvases[1].getContext('2d');
      var context = contexts[0]
      var bufferIndex = 0

      var intervalId = 0

      function getMoreCommands() {
        if (currentData.length == 0) {
          $.ajaxSetup({async: false});
          try {
            $.get("/command", {id:"{{.Unique}}"}, function(data) {
              currentData = data.split("~");
            })
            .fail(function() {
              currentData = ["END"]
            });
          } catch(e) {
            currentData = ["END"]
          }
        }
      }

      var nextMouseMoveEvent = 0
      function postMouseEvent(cmdName, e) {
        if (e.offsetX == undefined) {
          x = e.originalEvent.layerX;
          y = e.originalEvent.layerY;
        } else {
          x = e.offsetX;
          y = e.offsetY;
        }
        if (x == undefined || y == undefined) {
          return;
        }

        if (cmdName == "MOUSEMOVE") {
          var now = new Date().getTime();
          if (now < nextMouseMoveEvent) {
            return;
          }
          nextMouseMoveEvent = now + 30; // throttle to 30ms
        }

        var cmd = cmdName + " " + x + " " + y
        $.ajaxSetup({async: true});
        $.post("/command", {id:"{{.Unique}}", cmd:cmd})
      }

      function parseBool(b) {
        return b == "true"
      }

      function drawFrame() {
        getMoreCommands()
        while (currentData.length > 0) {
          var commandString = currentData.shift()
          var command = commandString.split("|");
          if (command[0] == "END") {
            clearInterval(intervalId)
          } else if (command[0] == "CLEARFRAME") {
            contexts[bufferIndex].clearRect(
                0, 0, canvases[bufferIndex].width,
                canvases[bufferIndex].height);
          } else if (command[0] == "SHOWFRAME") {
            canvases[bufferIndex].style.visibility='visible';
            canvases[1-bufferIndex].style.visibility='hidden';
            bufferIndex = 1 - bufferIndex
            context = contexts[bufferIndex]
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

      $(canvases[0]).click(function(e) {
        postMouseEvent("MOUSECLICK", e);
      });
	  $(canvases[0]).mousemove(function(e) {
        postMouseEvent("MOUSEMOVE", e);
      });

      $(canvases[1]).click(function(e) {
        postMouseEvent("MOUSECLICK", e);
      });
      $(canvases[1]).mousemove(function(e) {
        postMouseEvent("MOUSEMOVE", e);
      });

      intervalId = setInterval("drawFrame()", 30)
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
	u, err := url.Parse(r.RequestURI)
	if err != nil {
		http.NotFound(w, r)
		log.Println(err)
		return
	}
	command := u.Path

	if command == "/jquery.js" {
		// TODO: set mime type
		w.Write(jQuery)
		return
	}

	if command == "/" && r.Method == "GET" {
		c.updateUnique()
		err := c.renderHtml(w, r)
		if err != nil {
			return
		}
		if !c.started {
			c.started = true
			go func() {
				c.handler(newContext(c.Width, c.Height, c.command, c.mouse))
				c.command <- "END"
				close(c.command)
				close(c.mouse)
			}()
		}
		return
	}

	q := u.Query()
	unique := ""
	if _, ok := q["id"]; !ok {
		unique = r.PostFormValue("id")
		if unique == "" {
			http.NotFound(w, r)
			log.Println("missing id", r)
			return
		}
	} else {
		unique = q["id"][0]
	}

	if unique != c.Unique {
		http.NotFound(w, r)
		return
	}

	if command == "/command" && r.Method == "GET" {
		commandGroup := ""
		for command := range c.command {
			if len(commandGroup) > 0 {
				commandGroup += "~"
			}
			commandGroup += command
			if command == "SHOWFRAME" {
				break
			}
		}
		w.Write([]byte(commandGroup))
		return
	}

	if command == "/command" && r.Method == "POST" {
		cmd := strings.Fields(r.PostFormValue("cmd"))
		if len(cmd) == 0 {
			log.Println("missing command")
			http.NotFound(w, r)
			return
		}
		if cmd[0] == "MOUSEMOVE" || cmd[0] == "MOUSECLICK" {
			if len(cmd) == 3 {
				x, err := strconv.Atoi(cmd[1])
				if err != nil {
					http.NotFound(w, r)
					log.Println("invalid x")
					return
				}
				y, err := strconv.Atoi(cmd[2])
				if err != nil {
					http.NotFound(w, r)
					log.Println("invalid y")
					return
				}
				c.mouse <- mouseMovement{cmd[0], float64(x), float64(y)}
				return
			}
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

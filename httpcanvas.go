package httpcanvas

import (
//	"net/http"
)

type Canvas func(*Context)

func ListenAndServe(addr string, handler Canvas) (err error) {
	return
}

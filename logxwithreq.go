package logx

import (
	"net/http"
	"strings"

	"golang.org/x/net/context"
	"google.golang.org/appengine"
	"google.golang.org/appengine/log"
)

// Resembling, but not same
// signature as the appengine logging:
// func (c *context) Debugf(format string, args ...interface{})
func Debugf(req *http.Request, format string, args ...interface{}) {
	p := choppOffLeadingDirs(req.URL.Path)
	format = Columnify(p, 12, 4) + format
	if !IsAppengine() {
		defer SL().Incr().Decr()
		Printf(format, args...)
		return
	}

	var ctx context.Context
	func() {
		defer func() {
			if rec := recover(); rec != nil {
				Print("not an appengine request")
			}
		}()
		ctx = appengine.NewContext(req)

	}()
	log.Debugf(ctx, format, args...)
}

func choppOffLeadingDirs(p string) string {
	if strings.Index(p, "/") > -1 {
		p = p[strings.Index(p, "/")+1:]
	}
	if strings.Index(p, "/") > -1 {
		p = p[strings.Index(p, "/"):]
	}
	return p
}

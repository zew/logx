// +build appengine

// /\ keep the above blank line below the build tag
package logx

import (
	"context"
	"net/http"

	"google.golang.org/appengine"
	"google.golang.org/appengine/log"
)

type MyContext context.Context

func Debugf(req *http.Request, format string, args ...interface{}) {
	var ctx context.Context
	func() {
		defer func() {
			if rec := recover(); rec != nil {
				Print("not an appengine request")
			}
		}()
		ctx = appengine.NewContext(req)
	}()
	if ctx == nil {
		Printf(format, args...)
		return
	}
	log.Debugf(ctx, format, args...)
}

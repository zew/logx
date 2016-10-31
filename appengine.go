// +build !appengine
package logx

import (
	"context"
	"net/http"

	"google.golang.org/appengine"
	"google.golang.org/appengine/log"
)

type MyContext context.Context

func Debugf(req *http.Request, format string, args ...interface{}) {
	ctx := appengine.NewContext(req)
	if ctx == nil {
		Printf(format, args...)
		return
	}
	log.Debugf(ctx, format, args...)
}

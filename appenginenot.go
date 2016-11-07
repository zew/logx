// +build !appengine

// /\ keep the above blank line below the build tag
package logx

import (
	"net/http"
	"strings"
)

// Similar signature as the appengine logging.
// Conditional compiling for appengine
// This is the non appengine version
// where we put some simple request path info into the output
func Debugf(req *http.Request, format string, args ...interface{}) {
	p := req.URL.Path
	if strings.Index(p, "/") > -1 {
		p = p[strings.Index(p, "/")+1:]
	}
	if strings.Index(p, "/") > -1 {
		p = p[strings.Index(p, "/"):]
	}
	format = Columnify(p, 12, 4) + format
	defer SL().Incr().Decr()
	Printf(format, args...)
}

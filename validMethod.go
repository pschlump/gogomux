package gogomux

//
// Go Go Mux - Go Fast Mux / Router for HTTP requests
//
// (C) Philip Schlump, 2013-2014.
// Version: 0.4.3
// BuildNo: 803
//
// /Users/corwin/Projects/go-lib/gogomux
//

// Table of valid methods,  If other http-methods are created or used they should be added to
// this list.
var validMethod map[string]bool

func init() {
	validMethod = make(map[string]bool)
	validMethod["GET"] = true
	validMethod["PUT"] = true
	validMethod["POST"] = true
	validMethod["PATCH"] = true
	validMethod["OPTIONS"] = true
	validMethod["HEAD"] = true
	validMethod["DELETE"] = true
	validMethod["CONNECT"] = true
	validMethod["TRACE"] = true
}

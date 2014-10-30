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

// Definitions:
//	Pat			Something of the form T::T, T:T{, that is a pattern used to define where constants and patterns are in a route.
//	Route		/abc/:def/ghi
//	CleanRoute	/abc/:/ghi
//	Names		[ "def" ]
//	Url			Input from user /abc/1234/ghi
//	Values		[ "1234" ]						 		Params

import "fmt"

var dbFlag map[string]bool

func init() {
	dbFlag = make(map[string]bool)
	dbFlag["bob"] = true
}

func db(flag string, s string, x ...interface{}) {
	if b, ok := dbFlag[flag]; ok && b {
		fmt.Printf(s, x...)
	}
}

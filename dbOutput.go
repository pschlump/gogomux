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

func dumpCType(n colType) (rv string) {
	rv = "("
	com := ""
	if (n & IsWord) != 0 {
		rv += com + "IsWord"
		com = "|"
	}
	if (n & MultiUrl) != 0 {
		rv += com + "MultiUrl"
		com = "|"
	}
	if (n & SingleUrl) != 0 {
		rv += com + "SingleUrl"
		com = "|"
	}
	//if (n & BadUrl) != 0 {
	//	rv += com + "BadUrl"
	//	com = "|"
	//}
	if (n & Dummy) != 0 {
		rv += com + "Dummy"
		com = "|"
	}
	rv += ")"
	return
}

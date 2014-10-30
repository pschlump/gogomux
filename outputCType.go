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

func outputCType(c colType) (rv string) {
	rv = "{{ "
	com := ""
	if (c & IsWord) != 0 {
		rv += com + " IsWord "
		com = " |"
	}
	if (c & MultiUrl) != 0 {
		rv += com + " MultiUrl "
		com = " |"
	}
	if (c & SingleUrl) != 0 {
		rv += com + " SingleUrl "
		com = " |"
	}
	if (c & Dummy) != 0 {
		rv += com + " Dummy "
		com = " |"
	}
	rv += " }}"
	return
}

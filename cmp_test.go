package gogomux

//
// Go Go Mux - Go Fast Mux / Router for HTTP requests
//
// (C) Philip Schlump, 2013-2014.
// Version: 0.4.3
// BuildNo: 804
//
// /Users/corwin/Projects/go-lib/gogomux
//

import (
	"fmt"
	"testing"

	"./debug"
)

func Test_CmpUrlToCleanRoute(t *testing.T) {

	r := htx
	url := "/abc/:def/ghi"
	r.SplitOnSlash3(0, url, false)

	if false {
		fmt.Printf("r.Slash=%s NSl=%d %s\n", debug.SVar(r.Slash[0:r.NSl+1]), r.NSl, debug.LF())
	}

	b := r.CmpUrlToCleanRoute("T:T", "/abc/:/ghi")
	if false {
		fmt.Printf("b=%v\n", b, debug.LF())
	}
	if !b {
		t.Errorf("Not Found\n")
	}

}

// 36 ns
func OldBenchmark_CmpUrlToCleanRoute(b *testing.B) {
	r := htx

	url := "/abc/:def/ghi"
	r.SplitOnSlash3(0, url, false)

	for n := 0; n < b.N; n++ {
		b := r.CmpUrlToCleanRoute("T:T", "/abc/:/ghi")
		_ = b
	}
}

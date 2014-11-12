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

import "testing"

var testRuns_SplitOnSlash3 = []struct {
	param string
	slash string
	hash  string
	nsl   int
}{
	{"", `["/"]`, ``, -1},
	{"/", `["/"]`, ``, -1},
	{"a", `["a"]`, ``, -1},
	{"aa", `["aa"]`, ``, -1},
	{"/a", `["/","a"]`, ``, -1},
	{"/aa", `["/","aa"]`, ``, -1},
	{"//a", `["/","a"]`, ``, -1},
	{"///a", `["/","a"]`, ``, -1},
	{"///a/", `["/","a"]`, ``, -1},
	{"///a//", `["/","a"]`, ``, -1},
	{"///a///", `["/","a"]`, ``, -1},
	{"aa", `["aa"]`, ``, -1},
	{"/aa", `["/","aa"]`, ``, -1},
	{"//aa", `["/","aa"]`, ``, -1},
	{"///aa", `["/","aa"]`, ``, -1},
	{"///aa/", `["/","aa"]`, ``, -1},
	{"///aa//", `["/","aa"]`, ``, -1},
	{"./aa", `["aa"]`, ``, -1},
	{"././aa", `["aa"]`, ``, -1},
	{"./././aa", `["aa"]`, ``, -1},
	{"/./aa", `["/","aa"]`, ``, -1},
	{"/././aa", `["/","aa"]`, ``, -1},
	{"/./././aa", `["/","aa"]`, ``, -1},
	{"/aa/bb", `["/","aa","bb"]`, ``, -1},
	{"/aa//bb/cc/dd", `["/","aa","bb","cc","dd"]`, ``, -1},
	{"/aa/bb///cc/dd", `["/","aa","bb","cc","dd"]`, ``, -1},
	{"/aa/bb/./cc//.//dd", `["/","aa","bb","cc","dd"]`, ``, -1},
	{"/aa/bb.html", `["/","aa","bb.html"]`, ``, -1},
	{"/aa//bb/cc/dd.php", `["/","aa","bb","cc","dd.php"]`, ``, -1},
	{"/aa//bb/cc/dd.php/", `["/","aa","bb","cc","dd.php"]`, ``, -1},
	{"/aa//bb/cc/dd.php//", `["/","aa","bb","cc","dd.php"]`, ``, -1},
	{"/aa//bb/cc/dd.php///", `["/","aa","bb","cc","dd.php"]`, ``, -1},
	{"/aa/bb///cc.php/dd", `["/","aa","bb","cc.php","dd"]`, ``, -1},
	{"/aa/bb/./...cc//.//dd", `["/","aa","bb","...cc","dd"]`, ``, -1},
	{"/aa/bb/./.cc//.//dd", `["/","aa","bb",".cc","dd"]`, ``, -1},
	{"/../a", `["/","a"]`, ``, -1},
	{"/../../a", `["/","a"]`, ``, -1},
	{"/../../../a", `["/","a"]`, ``, -1},
	{"/../../../../a", `["/","a"]`, ``, -1},
	{"../a", `["/","a"]`, ``, -1},
	{"../../a", `["/","a"]`, ``, -1},
	{"../../../a", `["/","a"]`, ``, -1},
	{"../../../../a", `["/","a"]`, ``, -1},
	{"../../a.html", `["/","a.html"]`, ``, -1},
	{"../../../a.html", `["/","a.html"]`, ``, -1},
	{"../../../../a.html", `["/","a.html"]`, ``, -1},
	{"../bb/cc/../../a.html", `["/","a.html"]`, ``, -1},
	{"../bb/cc/dd/../../a.html", `["/","bb","a.html"]`, ``, -1},
	{"./bb/cc/dd/../../a.html", `["bb","a.html"]`, ``, -1},
	{"bb/cc/dd/../../ee/a.html", `["bb","ee","a.html"]`, ``, -1},
	{"bb/cc/dd/../../ee/../a.html", `["bb","a.html"]`, ``, -1},
	{"bb/cc/dd/../../ee/../a.html/", `["bb","a.html"]`, ``, -1},
	{"bb/cc/dd/../../ee/../a.html//", `["bb","a.html"]`, ``, -1},
	{"/./../bb/cc/dd/../../ee/../a.html//", `["/","bb","a.html"]`, ``, -1},
	{"/./../.../cc/dd/../../ee/../a.html//", `["/","...","a.html"]`, ``, -1},
	{"/redis/planb/", `["/","redis","planb"]`, ``, -1},
	{"", `["/","redis","planb"]`, ``, -1},
}

func TestSplitOnSlash3(t *testing.T) {

	for k, test := range testRuns_SplitOnSlash3 {
		htx.SplitOnSlash3(0, test.param, false)
		if false {
			t.Errorf("Test %d - Url(%v) = ", k, test.param)
		}
	}
}

/*
// 52.3 us
func BenchmarkFixPath(b *testing.B) {
	// noalloc = true
	rv := make([]string, 25)
	for n := 0; n < b.N; n++ {
		rv = rv[:25]
		// FixPath("/./../.../cc/dd/../../ee/../a.html//", &rv)
		FixPath("/cc/dd/a.html", rv, 25)
	}
}
*/

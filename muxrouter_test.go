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

import (
	"bytes"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"regexp"
	"strconv"
	"testing"

	"./debug"
)

var DataTab []RouteData

// This is used in the read-in stuff below
func getFxHandler(nfx int) Handle {
	return func(res http.ResponseWriter, req *http.Request, ps Params) {
	}
}

func (r *MuxRouter) ReadData(path string) (rv []RouteData) {
	var jsonData []RouteData
	file, err := ioutil.ReadFile(path)
	if err != nil {
		fmt.Printf("Error(10014): %v, %s, Config File:%s\n", err, debug.LF(), path)
		return jsonData
	}
	err = json.Unmarshal(file, &jsonData)
	if err != nil {
		fmt.Printf("Error(10012): %v, %s, Config File:%s\n", err, debug.LF(), path)
	}
	for i, v := range jsonData {
		jsonData[i].UseIt = !v.NotUseIt
		if v.NFxNo == 0 {
			v.NFxNo = FxTabN
			FxTabN++
			FxTab.Fx[v.NFxNo] = getFxHandler(v.NFxNo) // xyzzy102 - should be r.XXX and call user func
		}
	}
	rv = jsonData
	return
}

var testRuns2 = []struct {
	param  string
	result string
}{
	{"", `["/"]`},
	{"/", `["/"]`},
	{"a", `["a"]`},
	{"aa", `["aa"]`},
	{"/a", `["/","a"]`},
	{"/aa", `["/","aa"]`},
	{"//a", `["/","a"]`},
	{"///a", `["/","a"]`},
	{"///a/", `["/","a"]`},
	{"///a//", `["/","a"]`},
	{"///a///", `["/","a"]`},
	{"aa", `["aa"]`},
	{"/aa", `["/","aa"]`},
	{"//aa", `["/","aa"]`},
	{"///aa", `["/","aa"]`},
	{"///aa/", `["/","aa"]`},
	{"///aa//", `["/","aa"]`},
	{"./aa", `["aa"]`},
	{"././aa", `["aa"]`},
	{"./././aa", `["aa"]`},
	{"/./aa", `["/","aa"]`},
	{"/././aa", `["/","aa"]`},
	{"/./././aa", `["/","aa"]`},
	{"/aa/bb", `["/","aa","bb"]`},
	{"/aa//bb/cc/dd", `["/","aa","bb","cc","dd"]`},
	{"/aa/bb///cc/dd", `["/","aa","bb","cc","dd"]`},
	{"/aa/bb/./cc//.//dd", `["/","aa","bb","cc","dd"]`},
	{"/aa/bb.html", `["/","aa","bb.html"]`},
	{"/aa//bb/cc/dd.php", `["/","aa","bb","cc","dd.php"]`},
	{"/aa//bb/cc/dd.php/", `["/","aa","bb","cc","dd.php"]`},
	{"/aa//bb/cc/dd.php//", `["/","aa","bb","cc","dd.php"]`},
	{"/aa//bb/cc/dd.php///", `["/","aa","bb","cc","dd.php"]`},
	{"/aa/bb///cc.php/dd", `["/","aa","bb","cc.php","dd"]`},
	{"/aa/bb/./...cc//.//dd", `["/","aa","bb","...cc","dd"]`},
	{"/aa/bb/./.cc//.//dd", `["/","aa","bb",".cc","dd"]`},
	{"/../a", `["/","a"]`},
	{"/../../a", `["/","a"]`},
	{"/../../../a", `["/","a"]`},
	{"/../../../../a", `["/","a"]`},
	{"../a", `["/","a"]`},
	{"../../a", `["/","a"]`},
	{"../../../a", `["/","a"]`},
	{"../../../../a", `["/","a"]`},
	{"../../a.html", `["/","a.html"]`},
	{"../../../a.html", `["/","a.html"]`},
	{"../../../../a.html", `["/","a.html"]`},
	{"../bb/cc/../../a.html", `["/","a.html"]`},
	{"../bb/cc/dd/../../a.html", `["/","bb","a.html"]`},
	{"./bb/cc/dd/../../a.html", `["bb","a.html"]`},
	{"bb/cc/dd/../../ee/a.html", `["bb","ee","a.html"]`},
	{"bb/cc/dd/../../ee/../a.html", `["bb","a.html"]`},
	{"bb/cc/dd/../../ee/../a.html/", `["bb","a.html"]`},
	{"bb/cc/dd/../../ee/../a.html//", `["bb","a.html"]`},
	{"/./../bb/cc/dd/../../ee/../a.html//", `["/","bb","a.html"]`},
	{"/./../.../cc/dd/../../ee/../a.html//", `["/","...","a.html"]`},
	/*
	 */
}

var testMatchRuns2 = []struct {
	method string
	param  string
	result int
}{
	{"GET", "/api/table/liz/", 4},
	{"GET", "/api/table/liz/22", 3},
	{"GET", "/api/table/liz/", 4},
	{"GET", "/api/table/liz/:id", 3},
	{"GET", "/api/table/liz/1", 3},
	{"GET", "/api/table/liz/123", 3},
	{"GET", "/api/table/liz/4452-232323-2323232-232323", 3},
	{"GET", "/api/table/carbone/", 7},
	{"GET", "/index.html", 9},
	{"GET", "/api/js/jQuery-2.0.1.min.js", 9},
}

// -------------------------------------------------------------------------------------------------
var htx *MuxRouter

var test2017Data = []struct {
	LoadUrl bool
	Method  string
	Url     string
	Result  int
}{
	{true, "GET", "/planb/:vA/t1/:vB", 2},                 // 2
	{true, "GET", "/planb/:vD/t2/:vE", 3},                 // 3
	{true, "GET", "/planb/:vD/t2/xx", 4},                  // 4
	{true, "GET", "/planb/:vD/t8/{yy:^[0-9][0-9]*$}", 10}, // 4
	{true, "GET", "/planb/:vD/t8/bob", 11},                // 4
	{true, "GET", "/planb/:vD/t9/:vX", 12},                // 4
	{true, "GET", "/planb/:vD/t9/bob", 14},                // 4
	{true, "GET", "/planb/:vF/t3", 5},                     // 5
	{true, "GET", "/planb/:vG/t4", 6},                     // 6
	{true, "GET", "/planb/x3/t5", 7},                      // 7
	{true, "GET", "/planb/:vC", 1},                        // 1
	{true, "GET", "/planE/:vC", 20},                       // 20
	{true, "GET", "/users/:user/received_events", 9},      // 9
	{true, "GET", "/*vG", 8},                              // 8
	{true, "GET", "/rc/{yy:^[0-9][0-9]*$}", 15},           // 15
	{true, "GET", "/rc/:zz", 17},                          // 17
	{true, "GET", "/rc/dave", 16},                         // 16
	{true, "GET", "/rd/:z2", 18},                          // 18
	{true, "GET", "/re/{z2}", 19},                         // 19
	{true, "GET", "/js/*filename", 21},                    // 21
	{true, "GET", "/img/*filename", 22},                   // 22
	{true, "GET", "/css/*filename", 23},                   // 23
	{true, "GET", "/abc/*p1/:p2", 24},                     // 24 // test with /abc/*p1/:k2 - should be a bad pattern - should not work - should break loop at this point

	{true, "GET", "/authorizations", 31},
	{true, "GET", "/authorizations/:id", 32},
	{true, "POST", "/authorizations", 33},
	{true, "PUT", "/authorizations/clients/:client_id", 34},
	{true, "PATCH", "/authorizations/:id", 35},
	{true, "DELETE", "/authorizations/:id", 36},
	{true, "GET", "/applications/:client_id/tokens/:access_token", 37},
	{true, "DELETE", "/applications/:client_id/tokens", 38},
	{true, "DELETE", "/applications/:client_id/tokens/:access_token", 39},

	{true, "GET", "/events", 41},
	{true, "GET", "/repos/:owner/:repo/events", 42},
	{true, "GET", "/networks/:owner/:repo/events", 43},
	{true, "GET", "/orgs/:org/events", 44},
	{true, "GET", "/users/:user/received_events", 45},
	{true, "GET", "/users/:user/received_events/public", 46},
	{true, "GET", "/users/:user/events", 47},
	{true, "GET", "/users/:user/events/public", 48},
	{true, "GET", "/users/:user/events/orgs/:org", 49},
	{true, "GET", "/feeds", 50},
	{true, "GET", "/notifications", 51},
	{true, "GET", "/repos/:owner/:repo/notifications", 52},
	{true, "PUT", "/notifications", 53},
	{true, "PUT", "/repos/:owner/:repo/notifications", 54},
	{true, "GET", "/notifications/threads/:id", 55},
	{true, "PATCH", "/notifications/threads/:id", 56},
	{true, "GET", "/notifications/threads/:id/subscription", 57},
	{true, "PUT", "/notifications/threads/:id/subscription", 58},
	{true, "DELETE", "/notifications/threads/:id/subscription", 59},
	{true, "GET", "/repos/:owner/:repo/stargazers", 60},
	{true, "GET", "/users/:user/starred", 61},
	{true, "GET", "/user/starred", 62},
	{true, "GET", "/user/starred/:owner/:repo", 63},
	{true, "PUT", "/user/starred/:owner/:repo", 64},
	{true, "DELETE", "/user/starred/:owner/:repo", 65},
	{true, "GET", "/repos/:owner/:repo/subscribers", 66},
	{true, "GET", "/users/:user/subscriptions", 67},
	{true, "GET", "/user/subscriptions", 68},
	{true, "GET", "/repos/:owner/:repo/subscription", 69},
	{true, "PUT", "/repos/:owner/:repo/subscription", 70},
	{true, "DELETE", "/repos/:owner/:repo/subscription", 71},
	{true, "GET", "/user/subscriptions/:owner/:repo", 72},
	{true, "PUT", "/user/subscriptions/:owner/:repo", 73},
	{true, "DELETE", "/user/subscriptions/:owner/:repo", 74},

	{true, "GET", "/users/:user/gists", 76},
	{true, "GET", "/gists", 77},
	{true, "GET", "/gists/public", 78},
	{true, "GET", "/gists/starred", 79},
	{true, "GET", "/gists/:id", 80},
	{true, "POST", "/gists", 81},
	{true, "PATCH", "/gists/:id", 82},
	{true, "PUT", "/gists/:id/star", 83},
	{true, "DELETE", "/gists/:id/star", 84},
	{true, "GET", "/gists/:id/star", 85},
	{true, "POST", "/gists/:id/forks", 86},
	{true, "DELETE", "/gists/:id", 87},

	{true, "GET", "/repos/:owner/:repo/git/blobs/:sha", 89},
	{true, "POST", "/repos/:owner/:repo/git/blobs", 90},
	{true, "GET", "/repos/:owner/:repo/git/commits/:sha", 91},
	{true, "POST", "/repos/:owner/:repo/git/commits", 92},
	{true, "GET", "/repos/:owner/:repo/git/refs/*ref", 93},
	{true, "GET", "/repos/:owner/:repo/git/refs", 94},
	{true, "POST", "/repos/:owner/:repo/git/refs", 95},
	{true, "PATCH", "/repos/:owner/:repo/git/refs/*ref", 96},
	{true, "DELETE", "/repos/:owner/:repo/git/refs/*ref", 97},
	{true, "GET", "/repos/:owner/:repo/git/tags/:sha", 98},
	{true, "POST", "/repos/:owner/:repo/git/tags", 99},
	{true, "GET", "/repos/:owner/:repo/git/trees/:sha", 100},
	{true, "POST", "/repos/:owner/:repo/git/trees", 101},

	{true, "GET", "/issues", 103},
	{true, "GET", "/user/issues", 104},
	{true, "GET", "/orgs/:org/issues", 105},
	{true, "GET", "/repos/:owner/:repo/issues", 106},
	{true, "GET", "/repos/:owner/:repo/issues/:number", 107},
	{true, "POST", "/repos/:owner/:repo/issues", 108},
	{true, "PATCH", "/repos/:owner/:repo/issues/:number", 109},
	{true, "GET", "/repos/:owner/:repo/assignees", 110},
	{true, "GET", "/repos/:owner/:repo/assignees/:assignee", 111},
	{true, "GET", "/repos/:owner/:repo/issues/:number/comments", 112},
	{true, "GET", "/repos/:owner/:repo/issues/comments", 113},
	{true, "GET", "/repos/:owner/:repo/issues/comments/:id", 114},
	{true, "POST", "/repos/:owner/:repo/issues/:number/comments", 115},
	{true, "PATCH", "/repos/:owner/:repo/issues/comments/:id", 116},
	{true, "DELETE", "/repos/:owner/:repo/issues/comments/:id", 117},
	{true, "GET", "/repos/:owner/:repo/issues/:number/events", 118},
	{true, "GET", "/repos/:owner/:repo/issues/events", 119},
	{true, "GET", "/repos/:owner/:repo/issues/events/:id", 120},
	{true, "GET", "/repos/:owner/:repo/labels", 121},
	{true, "GET", "/repos/:owner/:repo/labels/:name", 122},
	{true, "POST", "/repos/:owner/:repo/labels", 123},
	{true, "PATCH", "/repos/:owner/:repo/labels/:name", 124},
	{true, "DELETE", "/repos/:owner/:repo/labels/:name", 125},
	{true, "GET", "/repos/:owner/:repo/issues/:number/labels", 126},
	{true, "POST", "/repos/:owner/:repo/issues/:number/labels", 127},
	{true, "DELETE", "/repos/:owner/:repo/issues/:number/labels/:name", 128},
	{true, "PUT", "/repos/:owner/:repo/issues/:number/labels", 129},
	{true, "DELETE", "/repos/:owner/:repo/issues/:number/labels", 130},
	{true, "GET", "/repos/:owner/:repo/milestones/:number/labels", 131},
	{true, "GET", "/repos/:owner/:repo/milestones", 132},
	{true, "GET", "/repos/:owner/:repo/milestones/:number", 133},
	{true, "POST", "/repos/:owner/:repo/milestones", 134},
	{true, "PATCH", "/repos/:owner/:repo/milestones/:number", 135},
	{true, "DELETE", "/repos/:owner/:repo/milestones/:number", 136},

	{true, "GET", "/emojis", 138},
	{true, "GET", "/gitignore/templates", 139},
	{true, "GET", "/gitignore/templates/:name", 140},
	{true, "POST", "/markdown", 141},
	{true, "POST", "/markdown/raw", 142},
	{true, "GET", "/meta", 143},
	{true, "GET", "/rate_limit", 144},

	{true, "GET", "/users/:user/orgs", 146},
	{true, "GET", "/user/orgs", 147},
	{true, "GET", "/orgs/:org", 148},
	{true, "PATCH", "/orgs/:org", 149},
	{true, "GET", "/orgs/:org/members", 150},
	{true, "GET", "/orgs/:org/members/:user", 151},
	{true, "DELETE", "/orgs/:org/members/:user", 152},
	{true, "GET", "/orgs/:org/public_members", 153},
	{true, "GET", "/orgs/:org/public_members/:user", 154},
	{true, "PUT", "/orgs/:org/public_members/:user", 155},
	{true, "DELETE", "/orgs/:org/public_members/:user", 156},
	{true, "GET", "/orgs/:org/teams", 157},
	{true, "GET", "/teams/:id", 158},
	{true, "POST", "/orgs/:org/teams", 159},
	{true, "PATCH", "/teams/:id", 160},
	{true, "DELETE", "/teams/:id", 161},
	{true, "GET", "/teams/:id/members", 162},
	{true, "GET", "/teams/:id/members/:user", 163},
	{true, "PUT", "/teams/:id/members/:user", 164},
	{true, "DELETE", "/teams/:id/members/:user", 165},
	{true, "GET", "/teams/:id/repos", 166},
	{true, "GET", "/teams/:id/repos/:owner/:repo", 167},
	{true, "PUT", "/teams/:id/repos/:owner/:repo", 168},
	{true, "DELETE", "/teams/:id/repos/:owner/:repo", 169},
	{true, "GET", "/user/teams", 170},

	{true, "GET", "/repos/:owner/:repo/pulls", 172},
	{true, "GET", "/repos/:owner/:repo/pulls/:number", 173},
	{true, "POST", "/repos/:owner/:repo/pulls", 174},
	{true, "PATCH", "/repos/:owner/:repo/pulls/:number", 175},
	{true, "GET", "/repos/:owner/:repo/pulls/:number/commits", 176},
	{true, "GET", "/repos/:owner/:repo/pulls/:number/files", 177},
	{true, "GET", "/repos/:owner/:repo/pulls/:number/merge", 178},
	{true, "PUT", "/repos/:owner/:repo/pulls/:number/merge", 179},
	{true, "GET", "/repos/:owner/:repo/pulls/:number/comments", 180},
	{true, "GET", "/repos/:owner/:repo/pulls/comments", 181},
	{true, "GET", "/repos/:owner/:repo/pulls/comments/:number", 182},
	{true, "PUT", "/repos/:owner/:repo/pulls/:number/comments", 183},
	{true, "PATCH", "/repos/:owner/:repo/pulls/comments/:number", 184},
	{true, "DELETE", "/repos/:owner/:repo/pulls/comments/:number", 185},

	{true, "GET", "/user/repos", 187},
	{true, "GET", "/users/:user/repos", 188},
	{true, "GET", "/orgs/:org/repos", 189},
	{true, "GET", "/repositories", 190},
	{true, "POST", "/user/repos", 191},
	{true, "POST", "/orgs/:org/repos", 192},
	{true, "GET", "/repos/:owner/:repo", 193},
	{true, "PATCH", "/repos/:owner/:repo", 194},
	{true, "GET", "/repos/:owner/:repo/contributors", 195},
	{true, "GET", "/repos/:owner/:repo/languages", 196},
	{true, "GET", "/repos/:owner/:repo/teams", 197},
	{true, "GET", "/repos/:owner/:repo/tags", 198},
	{true, "GET", "/repos/:owner/:repo/branches", 199},
	{true, "GET", "/repos/:owner/:repo/branches/:branch", 200},
	{true, "DELETE", "/repos/:owner/:repo", 201},
	{true, "GET", "/repos/:owner/:repo/collaborators", 202},
	{true, "GET", "/repos/:owner/:repo/collaborators/:user", 203},
	{true, "PUT", "/repos/:owner/:repo/collaborators/:user", 204},
	{true, "DELETE", "/repos/:owner/:repo/collaborators/:user", 205},
	{true, "GET", "/repos/:owner/:repo/comments", 206},
	{true, "GET", "/repos/:owner/:repo/commits/:sha/comments", 207},
	{true, "POST", "/repos/:owner/:repo/commits/:sha/comments", 208},
	{true, "GET", "/repos/:owner/:repo/comments/:id", 209},
	{true, "PATCH", "/repos/:owner/:repo/comments/:id", 210},
	{true, "DELETE", "/repos/:owner/:repo/comments/:id", 211},
	{true, "GET", "/repos/:owner/:repo/commits", 212},
	{true, "GET", "/repos/:owner/:repo/commits/:sha", 213},
	{true, "GET", "/repos/:owner/:repo/readme", 214},
	{true, "GET", "/repos/:owner/:repo/contents/*path", 215},
	{true, "PUT", "/repos/:owner/:repo/contents/*path", 216},
	{true, "DELETE", "/repos/:owner/:repo/contents/*path", 217},
	{true, "GET", "/repos/:owner/:repo/:archive_format/:ref", 218},
	{true, "GET", "/repos/:owner/:repo/keys", 219},
	{true, "GET", "/repos/:owner/:repo/keys/:id", 220},
	{true, "POST", "/repos/:owner/:repo/keys", 221},
	{true, "PATCH", "/repos/:owner/:repo/keys/:id", 222},
	{true, "DELETE", "/repos/:owner/:repo/keys/:id", 223},
	{true, "GET", "/repos/:owner/:repo/downloads", 224},
	{true, "GET", "/repos/:owner/:repo/downloads/:id", 225},
	{true, "DELETE", "/repos/:owner/:repo/downloads/:id", 226},
	{true, "GET", "/repos/:owner/:repo/forks", 227},
	{true, "POST", "/repos/:owner/:repo/forks", 228},
	{true, "GET", "/repos/:owner/:repo/hooks", 229},
	{true, "GET", "/repos/:owner/:repo/hooks/:id", 230},
	{true, "POST", "/repos/:owner/:repo/hooks", 231},
	{true, "PATCH", "/repos/:owner/:repo/hooks/:id", 232},
	{true, "POST", "/repos/:owner/:repo/hooks/:id/tests", 233},
	{true, "DELETE", "/repos/:owner/:repo/hooks/:id", 234},
	{true, "POST", "/repos/:owner/:repo/merges", 235},
	{true, "GET", "/repos/:owner/:repo/releases", 236},
	{true, "GET", "/repos/:owner/:repo/releases/:id", 237},
	{true, "POST", "/repos/:owner/:repo/releases", 238},
	{true, "PATCH", "/repos/:owner/:repo/releases/:id", 239},
	{true, "DELETE", "/repos/:owner/:repo/releases/:id", 240},
	{true, "GET", "/repos/:owner/:repo/releases/:id/assets", 241},
	{true, "GET", "/repos/:owner/:repo/stats/contributors", 242},
	{true, "GET", "/repos/:owner/:repo/stats/commit_activity", 243},
	{true, "GET", "/repos/:owner/:repo/stats/code_frequency", 244},
	{true, "GET", "/repos/:owner/:repo/stats/participation", 245},
	{true, "GET", "/repos/:owner/:repo/stats/punch_card", 246},
	{true, "GET", "/repos/:owner/:repo/statuses/:ref", 247},
	{true, "POST", "/repos/:owner/:repo/statuses/:ref", 248},

	{true, "GET", "/search/repositories", 250},
	{true, "GET", "/search/code", 251},
	{true, "GET", "/search/issues", 252},
	{true, "GET", "/search/users", 253},
	{true, "GET", "/legacy/issues/search/:owner/:repository/:state/:keyword", 254},
	{true, "GET", "/legacy/repos/search/:keyword", 255},
	{true, "GET", "/legacy/user/search/:keyword", 256},
	{true, "GET", "/legacy/user/email/:email", 257},

	{true, "GET", "/users/:user", 259},
	{true, "GET", "/user", 260},
	{true, "PATCH", "/user", 261},
	{true, "GET", "/users", 262},
	{true, "GET", "/user/emails", 263},
	{true, "POST", "/user/emails", 264},
	{true, "DELETE", "/user/emails", 265},
	{true, "GET", "/users/:user/followers", 266},
	{true, "GET", "/user/followers", 267},
	{true, "GET", "/users/:user/following", 268},
	{true, "GET", "/user/following", 269},
	{true, "GET", "/user/following/:user", 270},
	{true, "GET", "/users/:user/following/:target_user", 271},
	{true, "PUT", "/user/following/:user", 272},
	{true, "DELETE", "/user/following/:user", 273},
	{true, "GET", "/users/:user/keys", 274},
	{true, "GET", "/user/keys", 275},
	{true, "GET", "/user/keys/:id", 276},
	{true, "POST", "/user/keys", 277},
	{true, "PATCH", "/user/keys/:id", 278},
	{true, "DELETE", "/user/keys/:id", 279},

	{true, "GET", "/xyz/xyz/xyz", 300},
	{true, "GET", "/xyz01", 301},
	{true, "delete", "/bad01", 302},
	{true, "delete", "bad02", 303},
}

type NV struct {
	Type  string
	Name  string
	Value string
}

var test2017Run = []struct {
	RunTest        bool
	Method         string
	Url            string
	Expect         int
	ShouldBeFound  bool
	ExpectedParams []NV
}{
	/*  00 */ {true, "GET", "/planb/vD-Data", 1, true, []NV{NV{Type: ":", Name: "vC", Value: "vD-Data"}}},
	/*  01 */ {true, "GET", "/planb/vD-data/t2/xx", 4, true, []NV{NV{Type: ":", Name: "vD", Value: "vD-data"}}},
	/*  02 */ {true, "GET", "/planb/vD-data/t2/yy", 3, true, []NV{NV{Type: ":", Name: "vD", Value: "vD-data"}, {Type: ":", Name: "vE", Value: "yy"}}},
	/*  03 */ {true, "GET", "/planb/x3/t5", 7, true, nil},
	/*  04 */ {true, "GET", "/planb/x4/t5", 8, true, []NV{NV{Type: "*", Name: "vG", Value: "planb/x4/t5"}}}, // <<<< STAR >>>>	<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<< this one >>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>
	/*  05 */ {true, "GET", "//planb/x/.././/////x3/t5", 7, true, nil},
	/*  06 */ {true, "GET", "/planb/x/.././/////x3/t5", 7, true, nil},
	/*  07 */ {true, "GET", "/index.html", 8, true, nil},
	/*  08 */ {true, "GET", "/p/a/b/c/d/e/f/g/h/i/j/k/l/m/n/o/p/q/r/s/t/u/v/w/x/y/z/a/b/c/d/e/f", 8, true, []NV{NV{Type: "*", Name: "vG",
		Value: "p/a/b/c/d/e/f/g/h/i/j/k/l/m/n/o/p/q/r/s/t/u/v/w/x/y/z/a/b/c/d/e/f"}}},
	/*  09 */ {true, "GET", "/", 8, true, []NV{NV{Type: "*", Name: "vG", Value: ""}}},
	/*  10 */ {true, "GET", "/planb/:vD/t2/xx", 4, true, nil},
	/*  11 */ {true, "GET", "/planb/:vF/t3", 5, true, nil},
	/*  12 */ {true, "GET", "/planb/:vG/t4", 6, true, nil},
	/*  13 */ {true, "GET", "/planb/x3/t5", 7, true, nil},
	/*  14 */ {true, "GET", "/planb/:vC", 1, true, nil},
	/*  15 */ {true, "GET", "/*vG", 8, true, []NV{NV{Type: "*", Name: "vG", Value: "*vG"}, NV{Type: "*", Name: "vX", Value: ""}}},
	/*  16 */ {true, "GET", "/planb/vD-data/t8/11", 10, true, []NV{NV{Type: ":", Name: "vD", Value: "vD-data"}, NV{Type: "{", Name: "yy", Value: "11"}}},
	/*  17 */ {true, "GET", "/planb/vD-data/t9/bob", 14, true, []NV{NV{Type: ":", Name: "vD", Value: "vD-data"}}},
	/*  18 */ {true, "GET", "/planb/vD-data/t9/jane", 12, true, []NV{NV{Type: ":", Name: "vD", Value: "vD-data"}, NV{Type: ":", Name: "vX", Value: "jane"}}},
	/*  19 */ {true, "GET", "/rc/dave", 16, false, nil},
	/*  20 */ {true, "GET", "/rc/11", 15, true, []NV{NV{Type: "{", Name: "yy", Value: "11"}}},
	/*  21 */ {true, "GET", "/rc/jane", 17, true, nil},
	/*  22 */ {true, "GET", "/rd/jane", 18, true, []NV{NV{Type: ":", Name: "z2", Value: "jane"}}},
	/*  23 */ {true, "GET", "/re/jane", 19, true, []NV{NV{Type: ":", Name: "z2", Value: "jane"}}},
	/*  24 */ {true, "GET", "/planb/more-Data", 1, true, []NV{NV{Type: ":", Name: "vC", Value: "more-Data"}}},
	/*  25 */ {true, "GET", "/abc/cbs/bbc", 24, true, []NV{NV{Type: "*", Name: "p1", Value: "cbs/bbc"}, NV{Type: ":", Name: "p2", Value: ""}}},
	/*  26 */ {true, "GET", "/repos/cbs/bbc/assignees", 110, true, []NV{NV{Type: ":", Name: "owner", Value: "cbs"}, NV{Type: ":", Name: "repo", Value: "bbc"}}},
	/*  27 */ {true, "GET", "/users/Auser", 259, true, []NV{NV{Type: ":", Name: "user", Value: "Auser"}}},
	/*  28 */ {true, "GET", "/user", 260, true, nil},
	/*  29 */ {true, "PATCH", "/user", 261, true, nil},
	/*  30 */ {true, "GET", "/users", 262, true, nil},
	/*  31 */ {true, "GET", "/user/emails", 263, true, nil},
	/*  32 */ {true, "POST", "/user/emails", 264, true, nil},
	/*  33 */ {true, "DELETE", "/user/emails", 265, true, nil},
	/*  34 */ {true, "GET", "/users/Auser/followers", 266, true, []NV{NV{Type: ":", Name: "user", Value: "Auser"}}},
	/*  35 */ {true, "GET", "/user/followers", 267, true, nil},
	/*  36 */ {true, "GET", "/users/Auser/following", 268, true, []NV{NV{Type: ":", Name: "user", Value: "Auser"}}},
	/*  37 */ {true, "GET", "/user/following", 269, true, nil},
	/*  38 */ {true, "GET", "/user/following/Auser", 270, true, []NV{NV{Type: ":", Name: "user", Value: "Auser"}}},
	/*  39 */ {true, "GET", "/users/Auser/following/Atarget_user", 271, true, []NV{NV{Type: ":", Name: "user", Value: "Auser"}, NV{Type: ":", Name: "target_user", Value: "Atarget_user"}}},
	/*  40 */ {true, "PUT", "/user/following/Auser", 272, true, nil},
	/*  41 */ {true, "DELETE", "/user/following/Auser", 273, true, []NV{NV{Type: ":", Name: "user", Value: "Auser"}}},
	/*  42 */ {true, "GET", "/js/angular-animate.min.js", 21, false, nil},
	/*  43 */ {true, "GET", "/js/angular-animate.min.js.map", 21, false, nil},
	/*  44 */ {true, "GET", "/js/angular-cookies.min.js", 21, false, nil},
	/*  45 */ {true, "GET", "/js/angular-cookies.min.js.map", 21, false, nil},
	/*  46 */ {true, "GET", "/js/angular-loader.min.js", 21, false, nil},
	/*  47 */ {true, "GET", "/js/angular-loader.min.js.map", 21, false, nil},
	/*  48 */ {true, "GET", "/js/angular-resource.min.js", 21, false, nil},
	/*  49 */ {true, "GET", "/js/angular-resource.min.js.map", 21, false, nil},
	/*  50 */ {true, "GET", "/js/angular-route.min.js", 21, false, nil},
	/*  51 */ {true, "GET", "/js/angular-route.min.js.map", 21, false, nil},
	/*  52 */ {true, "GET", "/js/angular-sanitize.min.js", 21, false, nil},
	/*  53 */ {true, "GET", "/js/angular-sanitize.min.js.map", 21, false, nil},
	/*  54 */ {true, "GET", "/js/angular-touch.min.js", 21, false, nil},
	/*  55 */ {true, "GET", "/js/angular-touch.min.js.map", 21, false, nil},
	/*  56 */ {true, "GET", "/js/angular-translate.min.js", 21, false, nil},
	/*  57 */ {true, "GET", "/js/angular.1.2.10.js", 21, false, nil},
	/*  58 */ {true, "GET", "/js/angular.js", 21, false, nil},
	/*  59 */ {true, "GET", "/js/angular.min.js", 21, false, nil},
	/*  60 */ {true, "GET", "/js/angular.min.js.map", 21, false, nil},
	/*  61 */ {true, "GET", "/js/bootstrap-3.1.1-mod", 21, false, nil},
	/*  62 */ {true, "GET", "/js/date.js", 21, false, nil},
	/*  63 */ {true, "GET", "/js/dialog-4.2.0", 21, false, nil},
	/*  64 */ {true, "GET", "/js/dialogs.js", 21, false, nil},
	/*  65 */ {true, "GET", "/js/dialogs.min.js", 21, false, nil},
	/*  66 */ {true, "GET", "/js/jquery-1.10.2.js", 21, false, nil},
	/*  67 */ {true, "GET", "/js/jquery-1.10.2.min.js", 21, false, nil},
	/*  68 */ {true, "GET", "/js/jquery-1.10.2.min.map", 21, false, nil},
	/*  69 */ {true, "GET", "/js/jquery-1.11.0.min.js", 21, false, nil},
	/*  70 */ {true, "GET", "/js/jquery.complexify.banlist.js", 21, false, nil},
	/*  71 */ {true, "GET", "/js/jquery.complexify.js", 21, false, nil},
	/*  72 */ {true, "GET", "/js/jquery.complexify.min.js", 21, false, nil},
	/*  73 */ {true, "GET", "/js/jquery.hotkeys.js", 21, false, nil},
	/*  74 */ {true, "GET", "/js/jsoneditor.js", 21, false, nil},
	/*  75 */ {true, "GET", "/js/jsoneditor.min.js", 21, false, nil},
	/*  76 */ {true, "GET", "/js/jstorage.min.js", 21, false, nil},
	/*  77 */ {true, "GET", "/js/libmp3lame.min.js", 21, false, nil},
	/*  78 */ {true, "GET", "/js/mobile-angular-ui.js", 21, false, nil},
	/*  79 */ {true, "GET", "/js/mobile-angular-ui.min.js", 21, false, nil},
	/*  80 */ {true, "GET", "/js/moment.min.js", 21, false, nil},
	/*  81 */ {true, "GET", "/js/mp3Worker.js", 21, false, nil},
	/*  82 */ {true, "GET", "/js/ng-grid-2.0.11", 21, false, nil},
	/*  83 */ {true, "GET", "/js/recorderWorker.js", 21, false, nil},
	/*  84 */ {true, "GET", "/js/recordmp3.js", 21, false, nil},
	/*  85 */ {true, "GET", "/js/so.js", 21, false, nil},
	/*  86 */ {true, "GET", "/js/so.m4.js", 21, false, nil},
	/*  87 */ {true, "GET", "/js/ui-bootstrap-tpls-0.10.0-SNAPSHOT.js", 21, false, nil},
	/*  88 */ {true, "GET", "/js/ui-bootstrap-tpls-0.10.0-SNAPSHOT.min.js", 21, false, nil},
	/*  89 */ {true, "GET", "/css/dialogs.css", 23, false, nil},
	/*  90 */ {true, "GET", "/css/mobile-angular-ui-base.css", 23, false, nil},
	/*  91 */ {true, "GET", "/css/mobile-angular-ui-base.min.css", 23, false, nil},
	/*  92 */ {true, "GET", "/css/mobile-angular-ui-desktop.css", 23, false, nil},
	/*  93 */ {true, "GET", "/css/mobile-angular-ui-desktop.min.css", 23, false, nil},
	/*  94 */ {true, "GET", "/img/SafteyIcon-v1-114x114.png", 22, false, nil},
	/*  95 */ {true, "GET", "/img/SafteyIcon-v1-144x144.png", 22, false, nil},
	/*  96 */ {true, "GET", "/img/SafteyIcon-v1-57x57.png", 22, false, nil},
	/*  97 */ {true, "GET", "/img/SafteyIcon-v1-72x72.png", 22, false, nil},
	/*  98 */ {true, "GET", "/img/SafteyIcon-v1.png", 22, false, nil},
	/*  99 */ {true, "GET", "/img/ajax-loader-small.gif", 22, false, nil},
	/* 100 */ {true, "GET", "/img/bg_strength_gradient.jpg", 22, false, nil},
	/* 101 */ {true, "GET", "/img/checkbox_yes.png", 22, false, nil},
	/* 102 */ {true, "GET", "/img/checkbox_yes.svg", 22, false, nil},
	/* 103 */ {true, "GET", "/img/clear.gif", 22, false, nil},
	/* 104 */ {true, "GET", "/img/favicon.ico", 22, false, nil},
	/* 105 */ {true, "GET", "/img/favicon.png", 22, false, nil},
	/* 106 */ {true, "GET", "/img/icons.png", 22, false, nil},
	/* 107 */ {true, "GET", "/app.html", 8, false, nil},
	// {true, "GET", "/repos/:owner/:repo/pulls/:number/files", 177},
	/* 108 */ {true, "GET", "/repos/--owner--/--repo--/pulls/--number--/files", 177, true,
		[]NV{
			NV{Type: ":", Name: "owner", Value: "--owner--"},
			NV{Type: ":", Name: "repo", Value: "--repo--"},
			NV{Type: ":", Name: "number", Value: "--number--"},
		}},
	/* 109 */ {true, "GET", "/gists", 77, true, nil},
	/* 110 */ {true, "HEAD", "/r1", 7, true, nil},
	/* 111 */ {true, "GET", "/r2/4/4", 11, true, nil},
	/* 111 */ {true, "GET", "/r2/a/4", 10, true, nil},
	/* 111 */ {true, "GET", "/r2/A/4", 12, true, nil},
}

var rpTest string
var rpTest2 string

func rptParams(w http.ResponseWriter, r *http.Request, ps Params) {
	arrived = 4000
	s := ps.DumpParam()
	rpTest = s
	// fmt.Printf("\nrptParams: %s\n", s)
	w.Write([]byte("Hello Silly World<br>"))
}
func rptParams2(w http.ResponseWriter, r *http.Request, ps Params) {
	arrived = 4008
	s := ps.DumpParam()
	rpTest2 = s
	// fmt.Printf("\nrptParams: %s\n", s)
	done := false
	t := ""
	n := ""
	for i := 0; !done; i++ {
		n, t, done = ps.ByPostion(i)
		fmt.Printf("At [%d] %s ->%s<-\n", i, n, t)
	}
	// func (ps Params) ByPostion(pos int) ( s string, inRange bool ) {
	w.Write([]byte("Hello Silly World<br>"))
}

func xHandle(w http.ResponseWriter, r *http.Request, ps Params) {
	arrived = 4001
}
func xHandleR2_2(w http.ResponseWriter, r *http.Request, ps Params) {
	arrived = 4002
}
func xHandleR2_3(w http.ResponseWriter, r *http.Request, ps Params) {
	arrived = 4003
}
func xHandleR2_4(w http.ResponseWriter, r *http.Request, ps Params) {
	arrived = 4004
}

func init() {
	htx = New()
	htx.AttachWidget(Before, ParseQueryParams)
	htx.AttachWidget(Before, ParseBodyAsParams)
	htx.AttachWidget(Before, MethodParam)
	htx.AttachWidget(Before, ParseCookiesAsParams)
	//htx.AttachWidget(Before, SimpleLogingBefore)
	//htx.AttachWidget(After, SimpleLogingAfter)
	htx.AttachWidget(Before, ApacheLogingBefore)
	htx.AttachWidget(After, ApacheLogingAfter)
	for i, test := range test2017Data {
		if test.LoadUrl {
			// fmt.Printf("\n ==================== AddRoue, %d, %s\n", test.Result, test.Url)
			if test.Url == "/xyz/xyz/xyz" {
				// fmt.Printf("adding new route, %s\n", debug.LF())
				// htx.AddRoute(test.Method, test.Url, test.Result, func(w http.ResponseWriter, r *http.Request, ps Params) { j := i; arrived = 2000 + j }).SetHost("localhost:8090")
				htx.AddRoute(test.Method, test.Url, test.Result, func(w http.ResponseWriter, r *http.Request, ps Params) { j := i; arrived = 2000 + j }).SetPort("8090")
				// htx.AddRoute(test.Method, test.Url, test.Result, emptyTestingHandle)
			} else if test.Url == "/xyz01" {
				// fmt.Printf("adding new route, %s\n", debug.LF())
				// /*  00 */ {true, "GET", "https", "localhost:2000", "/xyz01", 1, true},
				htx.AddRoute(test.Method, test.Url, test.Result, func(w http.ResponseWriter, r *http.Request, ps Params) { arrived = 2000 }).SetPort("2000").SetHost("localhost:2000").SetHTTPSOnly()
			} else if test.Url == "/user/keys/:id" {
				htx.AddRoute(test.Method, test.Url, test.Result, rptParams)

			} else if test.Url == "/repos/:owner/:repo/merges" {
				htx.AddRoute(test.Method, test.Url, test.Result, rptParams2)

			} else {
				htx.AddRoute(test.Method, test.Url, test.Result, func(w http.ResponseWriter, r *http.Request, ps Params) { j := i; arrived = 1000 + j }) // 0
			}
		}
	}
	htx.GET("/r1", xHandle)
	htx.POST("/r1", xHandle)
	htx.PUT("/r1", xHandle)
	htx.PATCH("/r1", xHandle)
	htx.DELETE("/r1", xHandle)
	htx.OPTIONS("/r1", xHandle)
	htx.HEAD("/r1", xHandle)
	htx.CONNECT("/r1", xHandle)
	htx.TRACE("/r1", xHandle)
	htx.GET("/r2/{blah:[a-z]}/:goo", xHandleR2_2)
	htx.GET("/r2/{blah:[1-9]}/:goo", xHandleR2_3)
	htx.GET("/r2/{blah:[A-Z]}/:goo", xHandleR2_4)

	htx.OutputStatus = false
	htx.CompileRoutes()
	htx.OutputStatusInfo()
}

// func (r *MuxRouter) LookupUrlViaHash2(w http.ResponseWriter, req *http.Request, m *int, data GoGoData) (found bool, ln int, rv Collision2) {
func Test2017_newHash2(t *testing.T) {
	// fmt.Printf("\nTest2017_newHash2\n")
	r := htx

	var url string
	var found bool
	var ln int
	var item Collision2
	var data GoGoData
	var w http.ResponseWriter
	var req http.Request

	// dbHash2 = true
	for i, test := range test2017Run {
		if test.RunTest {
			url = test.Url
			m := (int(test.Method[0]) + (int(test.Method[1]) << 1))
			// fmt.Printf("m=%d\n", m)
			r.AllParam.NParam = 0
			r.SplitOnSlash3(m, url, true)
			db("Test2017_newHash2", "After SplitOnSlash3 ->%s<- ->%s<-\n r.Hash=%s r.Slash=%s r.NSl=%d\n", test.Url, r.CurUrl,
				debug.SVar(r.Hash[0:r.NSl]), debug.SVar(r.Slash[0:r.NSl+1]), r.NSl)
			found, ln, item = r.LookupUrlViaHash2(w, &req, &m, data)
			// fmt.Printf("ln=%d\n", ln)
			fail := false
			if !found {
				if test.ShouldBeFound {
					t.Errorf("Not Found\n")
					fail = true
				}
			} else if item.Hdlr != test.Expect {
				t.Errorf("Test: %d, ->%s<- Expected Result = %d, got %d\n", i, url, test.Expect, item.Hdlr)
				fail = true
			} else {
				// fmt.Printf("Test = %d\n", i)
				r.GetArgs3(url, item.ArgPattern, item.ArgNames, ln)
				// item.HandleFunc(w, req, r.AllParam)
				// fmt.Printf("item.ArgPattern=%s Names=%s, r.AllParams=%s\n", item.ArgPattern, debug.SVar(item.ArgNames), r.AllParam.DumpParam())
				if test.ExpectedParams != nil {
					for j, eparam := range test.ExpectedParams {
						if eparam.Type == ":" || eparam.Type == "*" || eparam.Type == "{" {
							pv := r.AllParam.ByName(eparam.Name)
							if pv != eparam.Value {
								t.Errorf("Test: %d/%d, Expected Param %s==->%s<-, got ->%s<-\n", i, j, eparam.Name, eparam.Value, pv)
								fail = true
							}
						}
					}
				}
			}
			_ = fail
			//if dbFlag["dbMach"] {
			//	if fail {
			//		fmt.Printf("    Test Failed\n")
			//	} else {
			//		fmt.Printf("    Test PASSed\n")
			//	}
			//}
		}
	}

	_, _, _ = found, ln, item
}

var ServeHTTP_Tests = []struct {
	RunTest       bool
	Method        string
	HTTPS         string
	Host          string
	Url           string
	Expect        int
	ShouldBeFound bool
	RawQuery      string
	Rp            string
	CookieName    string
	CookieValue   string
}{
	/*  00 */ {true, "GET", "https", "localhost:2000", "/xyz01", 2000, true, "", "", "", ""},
	/*  01 */ {true, "GET", "http", "localhost:2000", "/xyz01", 1265, false, "", "", "", ""},
	/*  02 */ {true, "GET", "https", "localhost:2001", "/xyz01", 1265, false, "", "", "", ""},
	/*  03 */ {true, "GET", "http", "localhost:2001", "/xyz01", 1265, false, "", "", "", ""},
	/*  04 */ {true, "GET", "https", "127.0.0.1:2001", "/xyz01", 1265, false, "", "", "", ""},
	/*  05 */ {true, "GET", "http", "127.0.0.1:2001", "/xyz01", 1265, false, "", "", "", ""},
	/*  06 */ {true, "GET", "http", "localhost:8090", "/img/checkbox_yes.png", 1265, true, "", "", "", ""},
	/*  07 */ {true, "GET", "http", "localhost:8090", "/user/keys/1234", 4000, true, "left=55&right=22&id=IdOnUrlWrong&x=12&x=15",
		`[{"Name":"right","Value":"22","From":1,"Type":113},{"Name":"id","Value":"1234","From":0,"Type":58},{"Name":"left","Value":"55","From":1,"Type":113}]`, "top", "999"},
	/*  08 */ {true, "GET", "http", "localhost:8090", "/user", 1265, true, "left=66&right=22&id=IdOnUrlWrong&METHOD=PATCH", "", "", ""},
}

// func (r *MuxRouter) ServeHTTP(w http.ResponseWriter, req *http.Request) {
func Test_1_ServeHTTP(t *testing.T) {
	// fmt.Printf("\nTest2017_newHash2\n")
	r := htx
	var req http.Request
	// var w http.ResponseWriter
	var url url.URL
	var tls tls.ConnectionState

	disableOutput = true

	w := new(mockResponseWriter)

	// router.ServeHTTP(w, req)

	r.NotFound = func(w http.ResponseWriter, req *http.Request) {
		arrived = -1
	}
	req.URL = &url

	for i, v := range ServeHTTP_Tests {
		req.URL.Path = v.Url
		req.URL.RawQuery = v.RawQuery
		req.Method = v.Method
		req.Host = v.Host
		req.TLS = nil
		req.RemoteAddr = "[::1]:53248"
		// req.RequestURI = v.HTTPS + "://" + v.Host + v.Url + "?" + v.RawQuery
		if v.RawQuery != "" {
			req.RequestURI = v.Url + "?" + v.RawQuery
		} else {
			req.RequestURI = v.Url
		}
		req.Proto = "HTTP/1.1"
		req.Header = make(http.Header)
		if v.HTTPS == "https" {
			req.TLS = &tls
		}
		if v.CookieName != "" {
			c := http.Cookie{Name: v.CookieName, Value: v.CookieValue}
			req.AddCookie(&c)
		}
		arrived = 0
		// ---------------------------------------------------------------------------------------------------
		r.ServeHTTP(w, &req)
		// ---------------------------------------------------------------------------------------------------
		if arrived != v.Expect {
			t.Errorf("Test[%d]: Loop Expected to have handler called. %d\n", i, arrived)
		} else if v.Rp != "" {
			fmt.Printf("%s\n", rpTest)
			//var aa, bb map[int]map[string]interface{}
			//err0 := json.Unmarshal([]byte(v.Rp), &aa)
			//err1 := json.Unmarshal([]byte(rpTest), &bb)
			//eq := reflect.DeepEqual(aa, bb)
			//if !eq || err0 != nil || err1 != nil {
			//	t.Errorf("Test[%d]: Loop Expected:\n%s\nGot:\n%s, err0=%v err1=%v\n", i, v.Rp, rpTest, err0, err1)
			//}
		}
	}

	// func (r *MuxRouter) UrlToCleanRoute(UsePat string) (rv string) {
	url2 := "/abc/:def/:ghi/jkl"
	Method := "GET"
	m := (int(Method[0]) + (int(Method[1]) << 1))
	r.SplitOnSlash3(m, url2, true)
	rv := r.UrlToCleanRoute("T::T")
	if rv != "/abc/:/:/jkl" {
		t.Errorf("Test: Expected to have clean pattern\n")
	}

	if false {
		req.URL = &url
		req.URL.Path = "/img/checkbox_yes.png"
		req.Method = "GET"
		req.Host = "localhost:8090" // new to HashHost test
		//var tls tls.ConnectionState
		//req.TLS = &tls
		r.NotFound = func(w http.ResponseWriter, req *http.Request) {
			arrived = -1
			// fmt.Printf("Test2017_ServeHTTP - not found %s\n", debug.LF())
		}

		// ----------------------------------------------------------------------------------------------
		// Validate a path that is un-related to the host of test.
		// ----------------------------------------------------------------------------------------------
		arrived = 0
		// fmt.Printf("This One\n")
		r.ServeHTTP(w, &req)
		if arrived != 1265 {
			t.Errorf("Test: Expected to have handler called. %d\n", arrived)
		}

		// ----------------------------------------------------------------------------------------------
		// Do a test that requreis the HOST test and should be found
		// ----------------------------------------------------------------------------------------------
		// /* 108 */ {true, "GET", "/xyz/xyz/xyz", 300, true, nil},
		arrived = 0
		req.URL.Path = "/xyz/xyz/xyz"
		r.ServeHTTP(w, &req)
		if arrived != 2265 {
			t.Errorf("Test: Expected to have handler called. %d\n", arrived)
		}

		// ----------------------------------------------------------------------------------------------
		// Do a test that requreis the HOST test and should NOT be found
		// ----------------------------------------------------------------------------------------------
		arrived = 0
		req.URL.Path = "/xyz/xyz/xyz"
		req.Host = "192.168.0.140:8000" // new to HashHost test
		r.ServeHTTP(w, &req)
		if arrived != 1265 {
			t.Errorf("Test: Expected to have FILE handler called, arrived=%d.\n", arrived)
		}
	}
}

func Test_2_ServeHTTP(t *testing.T) {

	r := htx
	var w http.ResponseWriter
	url := "/repos/OWNER/REPO/merges"

	var jsonStr = []byte(`{"title":"Buy cheese and bread for breakfast.","tInt":100,"bBool":false,"fFloat":1.2,"cplx":{"a":2,"b":[1,2,3]}}`)
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonStr))
	if err != nil {
		t.Errorf("Test: Got an error creating a request, %s\n", err)
	}
	req.Header.Set("X-Custom-Header", "myvalue")
	req.Header.Set("Content-Type", "application/json")

	arrived = 0
	rpTest = ""
	r.ServeHTTP(w, req)
	fmt.Printf("POST-1 (json): %s\n", rpTest)
	if arrived != 4008 {
		t.Errorf("Test: %d\n", arrived)
	}

}

func Test_3_ServeHTTP(t *testing.T) {

	r := htx
	var w http.ResponseWriter
	uuu := "/repos/OWNER/REPO/merges"

	data := url.Values{}
	data.Set("name", "foo")
	data.Add("surname", "bar")

	req, err := http.NewRequest("POST", uuu, bytes.NewBufferString(data.Encode()))
	if err != nil {
		t.Errorf("Test: Got an error creating a request, %s\n", err)
	}
	req.Header.Set("X-Custom-Header", "myvalue")
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("Content-Length", strconv.Itoa(len(data.Encode())))

	arrived = 0
	rpTest = ""
	r.ServeHTTP(w, req)
	fmt.Printf("POST-2 (strd): %s\n", rpTest2)
	if arrived != 4008 {
		t.Errorf("Test: %d\n", arrived)
	}

}

// 283 ns - 0 allocations per op.
func Benchmark_ServeHTTP(b *testing.B) {
	r := htx
	var req http.Request
	var w http.ResponseWriter
	var url url.URL

	req.URL = &url
	req.URL.Path = "/i/yes.png"            // 92 ns - miss
	req.URL.Path = "/img/checkbox_yes.png" // 283 ns - match with wild card
	req.Method = "GET"

	for n := 0; n < b.N; n++ {
		r.ServeHTTP(w, &req)
	}
}

func Test_db(t *testing.T) {
	db("bob", "%s %s\n", "test", "of debug and dump functions")
	dbFlag["dbMatch"] = false
	var s string
	s = dumpCType(IsWord | MultiUrl | SingleUrl | Dummy)
	if s != "(IsWord|MultiUrl|SingleUrl|Dummy)" {
		t.Errorf("Test: dumpCType not working correctly\n")
	}
	// ToDo
	//	 prams.go
	//		 func (ps Params) DumpParam() (rv string) {
	//		 func (ps Params) HasName(name string) (rv bool) {
}

func Test_MatchPort(t *testing.T) {
	var req http.Request
	var url url.URL

	req.URL = &url
	req.URL.Path = "/img/checkbox_yes.png" // 283 ns - match with wild card
	req.Method = "GET"
	sData[0] = "8000"
	sData[1] = "localhost:8001"

	req.Host = "localhost:8000"
	b := MatchPortFunc(&req, 0)
	if !b {
		t.Errorf("Test: MatchPort not working correctly\n")
	}

	req.Host = "[::1]:8000"
	b = MatchPortFunc(&req, 0)
	if !b {
		t.Errorf("Test: MatchPort not working correctly\n")
	}

	req.Host = "[::1]:8001"
	b = MatchPortFunc(&req, 0)
	if b {
		t.Errorf("Test: MatchPort not working correctly\n")
	}

	req.Host = "[::1]:8001"
	b = MatchHostFunc(&req, 1)
	if b {
		t.Errorf("Test: MatchPort not working correctly\n")
	}

	req.Host = "localhost:8001"
	b = MatchHostFunc(&req, 1)
	if !b {
		t.Errorf("Test: MatchPort not working correctly\n")
	}

	b = MatchTlsFunc(&req, 1)
	if b {
		t.Errorf("Test: MatchTls not working correctly\n")
	}

	var tls tls.ConnectionState
	req.TLS = &tls

	b = MatchTlsFunc(&req, 1)
	if !b {
		t.Errorf("Test: MatchTls not working correctly\n")
	}

}

// 25 ns - for 20 bytes
func oldBenchmark_StrCmp(b *testing.B) {

	var aa [20]uint8
	var bb [20]uint8

	for n := 0; n < b.N; n++ {
		for i := 0; i < 20; i++ {
			if aa[i] == bb[i] {
			}
		}
	}
}

// 47.9 ns - No Early Exit
// 46.1 ns - Early Exit
func Benchmark2017_newHash2_hash(b *testing.B) {
	var m int

	Method := "GET"
	url := "/index.html"

	for n := 0; n < b.N; n++ {
		m = (int(Method[0]) + (int(Method[1]) << 1))
		htx.SplitOnSlash3(m, url, true)
	}
}

// /index.html 126 ns
// /planb/x3.x5 158 ns
/*
PASS
BenchmarkOfSplitOnSlash3_long	    10000000	       151 ns/op	       0 B/op	       0 allocs/op
BenchmarkOfSplitOnSlash3_short	   100000000	        26.0 ns/op	       0 B/op	       0 allocs/op
Benchmark2017_newHash2_hash	   	    50000000	        42.5 ns/op	       0 B/op	       0 allocs/op
BenchmarkFullUrlLookupWithParams	20000000	       128 ns/op	       0 B/op	       0 allocs/op
*/
func BenchmarkFullUrlLookupWithParams(b *testing.B) {

	r := htx

	var url string
	var found bool
	var ln, m int
	var item Collision2
	var data GoGoData
	var w http.ResponseWriter
	var req http.Request

	Method := "GET"
	// 							               Regular		Early Exit
	url = "/planb/x3/x5"                    // 158 ns		160 ns (found)
	url = "/planb/vD-data/t2/yy"            // 230 ns		232 ns (found)
	url = "/index.html"                     // 126 ns		113 ns (early exit)
	url = "/static/js/jquery-1.10.4.min.js" // 162 ns       119 ns (early exit)

	for n := 0; n < b.N; n++ {
		m = (int(Method[0]) + (int(Method[1]) << 1))
		r.SplitOnSlash3(m, url, true)
		found, ln, item = r.LookupUrlViaHash2(w, &req, &m, data)
		if found {
			r.GetArgs3(url, item.ArgPattern, item.ArgNames, ln)
		}
	}

	_, _, _ = found, ln, item
}

//

var aRe *regexp.Regexp

func init() {
	aRe = regexp.MustCompile("^[0-9][0-9]*$")
}

// about 150 ns per RE match
func OldBenchmarkSpeedOfRe(b *testing.B) {
	Pat0 := "abcdefghi"
	Pat1 := "123abcdefghi"
	for n := 0; n < b.N; n++ {
		matched := aRe.MatchString(Pat0)
		_ = matched
		matched = aRe.MatchString(Pat1)
		_ = matched
	}
}

var xxx [800]int

// 0.32 ns to do array index
func OldBenchmarkOfArrayAccess(b *testing.B) {
	for n := 0; n < b.N; n++ {
		matched := xxx[5]
		_ = matched
	}
}

// 17.2 to 20 ns for a map[string]int lookup
func OldBenchmarkOfMapLookup(b *testing.B) {
	for n := 0; n < b.N; n++ {
		matched := validMethod["GET"]
		_ = matched
	}
}

func HashTest4(m uint8, Pat string) (ln int, h int, ht int) {
	ln = len(Pat)
	h = int(m)
	for i := 1; i < len(Pat); i++ {
		// fmt.Printf("Used ->%c<-\n", Pat[i])
		h += int(Pat[i])
		h += (h << 10)
		h = h ^ (h >> 6)
	}
	ht = h
	h += (h << 3)
	h = h ^ (h >> 11)
	h += (h << 15)
	h = ((h & bitMask) ^ ((h >> nBits) & bitMask))
	return
}

// 59.3 ns
func OldBenchmarkOfMapHash(b *testing.B) {
	for n := 0; n < b.N; n++ {
		ln, h, ht := HashTest4(12, "/repos/:1232323/:1232323/releases")
		_, _, _ = ln, h, ht
	}
}

// 69.3 ns - old before converting to DFSA
// 29.9 ns now
func OldBenchmarkOfSplitOnSlash3(b *testing.B) {
	htx := New()

	url := "/repos/julienschmidt/httprouter/stargazers"
	for n := 0; n < b.N; n++ {
		htx.SplitOnSlash3(1, url, true)
	}
}

//
// From the most common T::T patter for 4 long, to the least common
// You only have to go through the number of patterns
// We can use a bigger pattern thatn just T::T, could be T* or *
// 0 can map to most common, by sort of frequencey of T::T - so 0 T::T, 1 T:T:, no other patterns
//

func OldTestSplitOfSlash(t *testing.T) {
	// t.Errorf("Test: %d, Expected Result = %s, got %s\n", i, test.ExpectRe, re)
	htx := New()

	url := "/repos/julienschmidt/httprouter/stargazers"
	htx.SplitOnSlash3(1, url, false)
	// fmt.Printf("NSl = %d: ->%s<- Slash %s\n", htx.NSl, htx.CurUrl, debug.SVar(htx.Slash[0:htx.NSl]))
	var r [20]string
	for i := 0; i < 32; i++ {
		for j := 0; j < (htx.NSl - 1); j++ {
			// fmt.Printf("i=%d j=%d\n", i, j)
			r[j] = htx.CurUrl[htx.Slash[j]+1 : htx.Slash[j+1]]
			if (i & (1 << uint(j))) != 0 {
				r[j] = ":"
			}
		}
		// fmt.Printf("%d: %s\n", i, debug.SVar(r[0:htx.NSl-1]))
	}
}

// About 15 ns to gen a new array, 0 alloc
func OldBenchmarkSplitOfSlash_GenURL(b *testing.B) {
	htx := New()

	url := "/repos/julienschmidt/httprouter/stargazers"
	htx.SplitOnSlash3(1, url, false)
	var r [20]string

	for n := 0; n < b.N; n++ {
		// htx.SplitOnSlash(1, url)
		for i := 0; i < 2; i++ {
			for j := 0; j < (htx.NSl - 1); j++ {
				r[j] = htx.CurUrl[htx.Slash[j]+1 : htx.Slash[j+1]]
				if (i & (1 << uint(j))) != 0 {
					r[j] = ":"
				}
			}
		}
	}
	_ = r

}

// Slower than a string hash -
func OldBenchmarkOfMapLookupInt(b *testing.B) {
	mapOfNo := make(map[int]int)
	mapOfNo[5] = 1
	mapOfNo[3] = 1
	mapOfNo[7] = 1
	mapOfNo[9] = 1
	mapOfNo[10] = 1
	mapOfNo[15] = 1
	mapOfNo[17] = 1
	mapOfNo[22] = 1
	mapOfNo[18] = 1
	mapOfNo[55] = 1
	for n := 0; n < b.N; n++ {
		matched := mapOfNo[12]
		_ = matched
	}
}

// Time used based on number of keys, 83.3 ns - with 3 keys
type Key struct {
	p1, p2, p3 string
}

// 83.3 ns - grows 20+ ns per key in struct
func OldBenchmarkKeyMap(b *testing.B) {
	lookup := make(map[Key]int)
	lookup[Key{"repos", "stargazers", "a"}] = 1
	lookup[Key{"repos", "starga1ers", "a"}] = 1
	lookup[Key{"repos", "starga2ers", "a"}] = 1
	lookup[Key{"repos", "starga3ers", "a"}] = 1
	lookup[Key{"repos", "starga4ers", "a"}] = 1
	lookup[Key{"repos", "starga5ers", "a"}] = 1
	lookup[Key{"repos", "starg6zers", "a"}] = 1
	lookup[Key{"repos", "starg7zers", "a"}] = 1
	lookup[Key{"repos", "starg8zers", "a"}] = 1
	lookup[Key{"r9pos", "stargazers", "a"}] = 1
	lookup[Key{"rapos", "stargazers", "a"}] = 1
	lookup[Key{"rbpos", "starAazers", "a"}] = 1
	lookup[Key{"rcpos", "stargazers", "a"}] = 1
	for n := 0; n < b.N; n++ {
		x := lookup[Key{"repos", "stargazers", "a"}]
		_ = x
	}
}

// 101 ns - per concat - Plus 1 memory alloc
func OldBenchmarkStringConcat(b *testing.B) {
	a := []string{"repos", ":", ":", "stargazers"}
	for n := 0; n < b.N; n++ {
		x := a[0] + a[1] + a[2] + a[3]
		_ = x
	}
}

func concatArray(n int, s []string, out []uint8) (no int) {
	no = 0
	k := 0
	for i := 0; i < n; i++ {
		for j := 0; j < len(s[i]); j++ {
			out[k] = s[i][j]
			k++
		}
	}
	no = k
	return
}

// 42 ns
func OldBenchmarkStringConcat2(b *testing.B) {
	a := []string{"repos", ":", ":", "stargazers"}
	var x [100]uint8
	for n := 0; n < b.N; n++ {
		l := concatArray(4, a, x[:])
		_ = l
	}
}

func add(m map[string]map[string]int, path, country string) {
	mm, ok := m[path]
	if !ok {
		mm = make(map[string]int)
		m[path] = mm
	}
	mm[country] = 1
}

// 20.6 ns - lookup multiple keys
func OldBenchmark2KeyMap(b *testing.B) {
	lookup := make(map[string]map[string]int)

	add(lookup, "a", "b")
	add(lookup, "aaaaa", "baaaa")
	add(lookup, "abbbb", "bcccc")
	add(lookup, "a", "bddd")
	add(lookup, "a", "bddd")
	add(lookup, "a", "star")
	add(lookup, "a", "beee")
	x, ok := lookup["a"]["star"]

	for n := 0; n < b.N; n++ {
		x, ok = lookup["a"]["star"]
		_ = x
	}
	if !ok {
		fmt.Printf("\n! Ok\n")
	}
}

func Test_UrlToCleanRoute(t *testing.T) {
	url := "/abc/def/ghi"
	pat := "T:T"

	htx.CurUrl = url
	htx.Slash[0] = 0
	htx.Slash[1] = 4
	htx.Slash[2] = 8
	htx.Slash[3] = len(url)
	htx.NSl = 3
	Result := htx.UrlToCleanRoute(pat)
	if Result != "/abc/:/ghi" {
		t.Errorf("Test: %d, Expected Result = %d, got %d\n", 0, "/abc/:/ghi", Result)
	}
}

func Test_MinInt(t *testing.T) {
	a := minInt(1, 2)
	if a != 1 {
		t.Errorf("Test: failed\n")
	}
	a = minInt(2, 1)
	if a != 1 {
		t.Errorf("Test: failed\n")
	}

}

func dumpReArray(tmpRe []Re) (rv string) {
	rv = "{{ "
	com := ""
	for i, v := range tmpRe {
		rv += com + fmt.Sprintf(" at[%d] Pos=%d Re=%s Name=%s ", i, v.Pos, v.Re, v.Name)
		com = ","
	}
	rv += " }}"
	return
}

func (r *MuxRouter) OutputStatusInfo() {
	if !r.OutputStatus {
		return
	}
	nc := 0
	for i, v := range r.Hash2Test {
		if v > 0 {
			fmt.Printf("[%4d] %x %s\n", i, r.LookupResults[v].cType, dumpCType(r.LookupResults[v].cType))
			if (r.LookupResults[v].cType & MultiUrl) != 0 {
				nc++
				for j, w := range r.LookupResults[v].Multi {
					ns := numChar(j, '/')
					fmt.Printf("   %2d %s\n", ns, j)
					_ = w
				}
			}
		}
	}
	fmt.Printf("Number of Collisions = %d\n", nc)
}

// -----------------------------------------------------------------------------------------------------

type mockResponseWriter struct{}

func (m *mockResponseWriter) Header() (h http.Header) {
	return http.Header{}
}

func (m *mockResponseWriter) Write(p []byte) (n int, err error) {
	return len(p), nil
}

func (m *mockResponseWriter) WriteString(s string) (n int, err error) {
	return len(s), nil
}

func (m *mockResponseWriter) WriteHeader(int) {}

/* vim: set noai ts=4 sw=4: */

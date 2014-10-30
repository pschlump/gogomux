package gogomux

import (
	"net/http"
	"testing"
)

// http://developer.github.com/v3/
var testGithubPat1 = []struct {
	Method     string
	UrlPattern string
}{

	{"GET", "/authorizations"},
	{"GET", "/authorizations/:id"},
	{"POST", "/authorizations"},
	{"PUT", "/authorizations/clients/:client_id"},
	{"PATCH", "/authorizations/:id"},
	{"DELETE", "/authorizations/:id"},
	{"GET", "/applications/:client_id/tokens/:access_token"},
	{"DELETE", "/applications/:client_id/tokens"},
	{"DELETE", "/applications/:client_id/tokens/:access_token"},

	{"GET", "/events"},
	{"GET", "/repos/:owner/:repo/events"},
	{"GET", "/networks/:owner/:repo/events"},
	{"GET", "/orgs/:org/events"},
	{"GET", "/users/:user/received_events"},
	{"GET", "/users/:user/received_events/public"},
	{"GET", "/users/:user/events"},
	{"GET", "/users/:user/events/public"},
	{"GET", "/users/:user/events/orgs/:org"},
	{"GET", "/feeds"},
	{"GET", "/notifications"},
	{"GET", "/repos/:owner/:repo/notifications"},
	{"PUT", "/notifications"},
	{"PUT", "/repos/:owner/:repo/notifications"},
	{"GET", "/notifications/threads/:id"},
	{"PATCH", "/notifications/threads/:id"},
	{"GET", "/notifications/threads/:id/subscription"},
	{"PUT", "/notifications/threads/:id/subscription"},
	{"DELETE", "/notifications/threads/:id/subscription"},
	{"GET", "/repos/:owner/:repo/stargazers"},
	{"GET", "/users/:user/starred"},
	{"GET", "/user/starred"},
	{"GET", "/user/starred/:owner/:repo"},
	{"PUT", "/user/starred/:owner/:repo"},
	{"DELETE", "/user/starred/:owner/:repo"},
	{"GET", "/repos/:owner/:repo/subscribers"},
	{"GET", "/users/:user/subscriptions"},
	{"GET", "/user/subscriptions"},
	{"GET", "/repos/:owner/:repo/subscription"},
	{"PUT", "/repos/:owner/:repo/subscription"},
	{"DELETE", "/repos/:owner/:repo/subscription"},
	{"GET", "/user/subscriptions/:owner/:repo"},
	{"PUT", "/user/subscriptions/:owner/:repo"},
	{"DELETE", "/user/subscriptions/:owner/:repo"},

	{"GET", "/users/:user/gists"},
	{"GET", "/gists"},
	{"GET", "/gists/public"},
	{"GET", "/gists/starred"},
	{"GET", "/gists/:id"},
	{"POST", "/gists"},
	{"PATCH", "/gists/:id"},
	{"PUT", "/gists/:id/star"},
	{"DELETE", "/gists/:id/star"},
	{"GET", "/gists/:id/star"},
	{"POST", "/gists/:id/forks"},
	{"DELETE", "/gists/:id"},

	{"GET", "/repos/:owner/:repo/git/blobs/:sha"},
	{"POST", "/repos/:owner/:repo/git/blobs"},
	{"GET", "/repos/:owner/:repo/git/commits/:sha"},
	{"POST", "/repos/:owner/:repo/git/commits"},
	{"GET", "/repos/:owner/:repo/git/refs/*ref"},
	{"GET", "/repos/:owner/:repo/git/refs"},
	{"POST", "/repos/:owner/:repo/git/refs"},
	{"PATCH", "/repos/:owner/:repo/git/refs/*ref"},
	{"DELETE", "/repos/:owner/:repo/git/refs/*ref"},
	{"GET", "/repos/:owner/:repo/git/tags/:sha"},
	{"POST", "/repos/:owner/:repo/git/tags"},
	{"GET", "/repos/:owner/:repo/git/trees/:sha"},
	{"POST", "/repos/:owner/:repo/git/trees"},

	{"GET", "/issues"},
	{"GET", "/user/issues"},
	{"GET", "/orgs/:org/issues"},
	{"GET", "/repos/:owner/:repo/issues"},
	{"GET", "/repos/:owner/:repo/issues/:number"},
	{"POST", "/repos/:owner/:repo/issues"},
	{"PATCH", "/repos/:owner/:repo/issues/:number"},
	{"GET", "/repos/:owner/:repo/assignees"},
	{"GET", "/repos/:owner/:repo/assignees/:assignee"},
	{"GET", "/repos/:owner/:repo/issues/:number/comments"},
	{"GET", "/repos/:owner/:repo/issues/comments"},
	{"GET", "/repos/:owner/:repo/issues/comments/:id"},
	{"POST", "/repos/:owner/:repo/issues/:number/comments"},
	{"PATCH", "/repos/:owner/:repo/issues/comments/:id"},
	{"DELETE", "/repos/:owner/:repo/issues/comments/:id"},
	{"GET", "/repos/:owner/:repo/issues/:number/events"},
	{"GET", "/repos/:owner/:repo/issues/events"},
	{"GET", "/repos/:owner/:repo/issues/events/:id"},
	{"GET", "/repos/:owner/:repo/labels"},
	{"GET", "/repos/:owner/:repo/labels/:name"},
	{"POST", "/repos/:owner/:repo/labels"},
	{"PATCH", "/repos/:owner/:repo/labels/:name"},
	{"DELETE", "/repos/:owner/:repo/labels/:name"},
	{"GET", "/repos/:owner/:repo/issues/:number/labels"},
	{"POST", "/repos/:owner/:repo/issues/:number/labels"},
	{"DELETE", "/repos/:owner/:repo/issues/:number/labels/:name"},
	{"PUT", "/repos/:owner/:repo/issues/:number/labels"},
	{"DELETE", "/repos/:owner/:repo/issues/:number/labels"},
	{"GET", "/repos/:owner/:repo/milestones/:number/labels"},
	{"GET", "/repos/:owner/:repo/milestones"},
	{"GET", "/repos/:owner/:repo/milestones/:number"},
	{"POST", "/repos/:owner/:repo/milestones"},
	{"PATCH", "/repos/:owner/:repo/milestones/:number"},
	{"DELETE", "/repos/:owner/:repo/milestones/:number"},

	{"GET", "/emojis"},
	{"GET", "/gitignore/templates"},
	{"GET", "/gitignore/templates/:name"},
	{"POST", "/markdown"},
	{"POST", "/markdown/raw"},
	{"GET", "/meta"},
	{"GET", "/rate_limit"},

	{"GET", "/users/:user/orgs"},
	{"GET", "/user/orgs"},
	{"GET", "/orgs/:org"},
	{"PATCH", "/orgs/:org"},
	{"GET", "/orgs/:org/members"},
	{"GET", "/orgs/:org/members/:user"},
	{"DELETE", "/orgs/:org/members/:user"},
	{"GET", "/orgs/:org/public_members"},
	{"GET", "/orgs/:org/public_members/:user"},
	{"PUT", "/orgs/:org/public_members/:user"},
	{"DELETE", "/orgs/:org/public_members/:user"},
	{"GET", "/orgs/:org/teams"},
	{"GET", "/teams/:id"},
	{"POST", "/orgs/:org/teams"},
	{"PATCH", "/teams/:id"},
	{"DELETE", "/teams/:id"},
	{"GET", "/teams/:id/members"},
	{"GET", "/teams/:id/members/:user"},
	{"PUT", "/teams/:id/members/:user"},
	{"DELETE", "/teams/:id/members/:user"},
	{"GET", "/teams/:id/repos"},
	{"GET", "/teams/:id/repos/:owner/:repo"},
	{"PUT", "/teams/:id/repos/:owner/:repo"},
	{"DELETE", "/teams/:id/repos/:owner/:repo"},
	{"GET", "/user/teams"},

	{"GET", "/repos/:owner/:repo/pulls"},
	{"GET", "/repos/:owner/:repo/pulls/:number"},
	{"POST", "/repos/:owner/:repo/pulls"},
	{"PATCH", "/repos/:owner/:repo/pulls/:number"},
	{"GET", "/repos/:owner/:repo/pulls/:number/commits"},
	{"GET", "/repos/:owner/:repo/pulls/:number/files"},
	{"GET", "/repos/:owner/:repo/pulls/:number/merge"},
	{"PUT", "/repos/:owner/:repo/pulls/:number/merge"},
	{"GET", "/repos/:owner/:repo/pulls/:number/comments"},
	{"GET", "/repos/:owner/:repo/pulls/comments"},
	{"GET", "/repos/:owner/:repo/pulls/comments/:number"},
	{"PUT", "/repos/:owner/:repo/pulls/:number/comments"},
	{"PATCH", "/repos/:owner/:repo/pulls/comments/:number"},
	{"DELETE", "/repos/:owner/:repo/pulls/comments/:number"},

	{"GET", "/user/repos"},
	{"GET", "/users/:user/repos"},
	{"GET", "/orgs/:org/repos"},
	{"GET", "/repositories"},
	{"POST", "/user/repos"},
	{"POST", "/orgs/:org/repos"},
	{"GET", "/repos/:owner/:repo"},
	{"PATCH", "/repos/:owner/:repo"},
	{"GET", "/repos/:owner/:repo/contributors"},
	{"GET", "/repos/:owner/:repo/languages"},
	{"GET", "/repos/:owner/:repo/teams"},
	{"GET", "/repos/:owner/:repo/tags"},
	{"GET", "/repos/:owner/:repo/branches"},
	{"GET", "/repos/:owner/:repo/branches/:branch"},
	{"DELETE", "/repos/:owner/:repo"},
	{"GET", "/repos/:owner/:repo/collaborators"},
	{"GET", "/repos/:owner/:repo/collaborators/:user"},
	{"PUT", "/repos/:owner/:repo/collaborators/:user"},
	{"DELETE", "/repos/:owner/:repo/collaborators/:user"},
	{"GET", "/repos/:owner/:repo/comments"},
	{"GET", "/repos/:owner/:repo/commits/:sha/comments"},
	{"POST", "/repos/:owner/:repo/commits/:sha/comments"},
	{"GET", "/repos/:owner/:repo/comments/:id"},
	{"PATCH", "/repos/:owner/:repo/comments/:id"},
	{"DELETE", "/repos/:owner/:repo/comments/:id"},
	{"GET", "/repos/:owner/:repo/commits"},
	{"GET", "/repos/:owner/:repo/commits/:sha"},
	{"GET", "/repos/:owner/:repo/readme"},
	{"GET", "/repos/:owner/:repo/contents/*path"},
	{"PUT", "/repos/:owner/:repo/contents/*path"},
	{"DELETE", "/repos/:owner/:repo/contents/*path"},
	{"GET", "/repos/:owner/:repo/:archive_format/:ref"},
	{"GET", "/repos/:owner/:repo/keys"},
	{"GET", "/repos/:owner/:repo/keys/:id"},
	{"POST", "/repos/:owner/:repo/keys"},
	{"PATCH", "/repos/:owner/:repo/keys/:id"},
	{"DELETE", "/repos/:owner/:repo/keys/:id"},
	{"GET", "/repos/:owner/:repo/downloads"},
	{"GET", "/repos/:owner/:repo/downloads/:id"},
	{"DELETE", "/repos/:owner/:repo/downloads/:id"},
	{"GET", "/repos/:owner/:repo/forks"},
	{"POST", "/repos/:owner/:repo/forks"},
	{"GET", "/repos/:owner/:repo/hooks"},
	{"GET", "/repos/:owner/:repo/hooks/:id"},
	{"POST", "/repos/:owner/:repo/hooks"},
	{"PATCH", "/repos/:owner/:repo/hooks/:id"},
	{"POST", "/repos/:owner/:repo/hooks/:id/tests"},
	{"DELETE", "/repos/:owner/:repo/hooks/:id"},
	{"POST", "/repos/:owner/:repo/merges"},
	{"GET", "/repos/:owner/:repo/releases"},
	{"GET", "/repos/:owner/:repo/releases/:id"},
	{"POST", "/repos/:owner/:repo/releases"},
	{"PATCH", "/repos/:owner/:repo/releases/:id"},
	{"DELETE", "/repos/:owner/:repo/releases/:id"},
	{"GET", "/repos/:owner/:repo/releases/:id/assets"},
	{"GET", "/repos/:owner/:repo/stats/contributors"},
	{"GET", "/repos/:owner/:repo/stats/commit_activity"},
	{"GET", "/repos/:owner/:repo/stats/code_frequency"},
	{"GET", "/repos/:owner/:repo/stats/participation"},
	{"GET", "/repos/:owner/:repo/stats/punch_card"},
	{"GET", "/repos/:owner/:repo/statuses/:ref"},
	{"POST", "/repos/:owner/:repo/statuses/:ref"},

	{"GET", "/search/repositories"},
	{"GET", "/search/code"},
	{"GET", "/search/issues"},
	{"GET", "/search/users"},
	{"GET", "/legacy/issues/search/:owner/:repository/:state/:keyword"},
	{"GET", "/legacy/repos/search/:keyword"},
	{"GET", "/legacy/user/search/:keyword"},
	{"GET", "/legacy/user/email/:email"},

	{"GET", "/users/:user"},
	{"GET", "/user"},
	{"PATCH", "/user"},
	{"GET", "/users"},
	{"GET", "/user/emails"},
	{"POST", "/user/emails"},
	{"DELETE", "/user/emails"},
	{"GET", "/users/:user/followers"},
	{"GET", "/user/followers"},
	{"GET", "/users/:user/following"},
	{"GET", "/user/following"},
	{"GET", "/user/following/:user"},
	{"GET", "/users/:user/following/:target_user"},
	{"PUT", "/user/following/:user"},
	{"DELETE", "/user/following/:user"},
	{"GET", "/users/:user/keys"},
	{"GET", "/user/keys"},
	{"GET", "/user/keys/:id"},
	{"POST", "/user/keys"},
	{"PATCH", "/user/keys/:id"},
	{"DELETE", "/user/keys/:id"},
}

func OldTestGithub2(t *testing.T) {
	r := New()
	for i, test := range testGithubPat1 {
		// fmt.Printf("i=%d Url=%s\n", i, test.UrlPattern)
		r.AddRoute(test.Method, test.UrlPattern, i+1, emptyTestingHandle)
	}
	// fmt.Printf("Calling Compile\n")
	r.CompileRoutes()
}

var ss = []string{"DELETE :",
	"DELETE applications",
	"DELETE authorizations",
	"DELETE comments",
	"DELETE commits",
	"DELETE downloads",
	"DELETE gists",
	"DELETE hooks",
	"DELETE issues",
	"DELETE keys",
	"DELETE labels",
	"DELETE members",
	"DELETE milestones",
	"DELETE notifications",
	"DELETE orgs",
	"DELETE public_members",
	"DELETE pulls",
	"DELETE releases",
	"DELETE repos",
	"DELETE stargazers",
	"DELETE starred",
	"DELETE subscribers",
	"DELETE subscription",
	"DELETE subscriptions",
	"DELETE teams",
	"DELETE templates",
	"DELETE threads",
	"DELETE tokens",
	"DELETE user",
	"GET *path",
	"GET *ref",
	"GET :",
	"GET applications",
	"GET assets",
	"GET assignees",
	"GET authorizations",
	"GET branches",
	"GET code",
	"GET code_frequency",
	"GET collaborators",
	"GET comments",
	"GET commit_activity",
	"GET commits",
	"GET contents",
	"GET contributors",
	"GET downloads",
	"GET email",
	"GET emails",
	"GET emojis",
	"GET events",
	"GET feeds",
	"GET files",
	"GET followers",
	"GET following",
	"GET forks",
	"GET gists",
	"GET git",
	"GET gitignore",
	"GET hooks",
	"GET issues",
	"GET keys",
	"GET labels",
	"GET languages",
	"GET legacy",
	"GET members",
	"GET merge",
	"GET meta",
	"GET milestones",
	"GET networks",
	"GET notifications",
	"GET orgs",
	"GET participation",
	"GET public",
	"GET pulls",
	"GET punch_card",
	"GET rate_limit",
	"GET readme",
	"GET received_events",
	"GET refs",
	"GET releases",
	"GET repos",
	"GET repositories",
	"GET search",
	"GET starred",
	"GET stats",
	"GET statuses",
	"GET subscription",
	"GET subscriptions",
	"GET tags",
	"GET teams",
	"GET tests",
	"GET threads",
	"GET tokens",
	"GET trees",
	"GET user",
	"GET users",
	"PATCH :",
	"PATCH authorizations",
	"PATCH comments",
	"PATCH gists",
	"PATCH issues",
	"PATCH notifications",
	"PATCH orgs",
	"PATCH releases",
	"PATCH repos",
	"PATCH teams",
	"PATCH threads",
	"PATCH user",
	"POST :",
	"POST authorizations",
	"POST blobs",
	"POST forks",
	"POST gists",
	"POST git",
	"POST issues",
	"POST markdown",
	"POST merges",
	"POST orgs",
	"POST raw",
	"POST repos",
	"POST trees",
	"POST user",
	"PUT :",
	"PUT authorizations",
	"PUT blobs",
	"PUT clients",
	"PUT forks",
	"PUT gists",
	"PUT git",
	"PUT issues",
	"PUT labels",
	"PUT notifications",
	"PUT orgs",
	"PUT repos",
	"PUT star",
	"PUT subscription",
	"PUT subscriptions",
	"PUT teams",
	"PUT threads",
	"PUT user",
}

// 23 ns
func OldBenchmarkMapLookupOfStringInMap(b *testing.B) {

	mm := make(map[string]int)

	for i, v := range ss {
		mm[v] = i + 1
	}

	j := 0
	for n := 0; n < b.N; n++ {
		x := mm[ss[j]]
		_ = x
		j++
		if j > 50 {
			j = 0
		}
	}

}

var words = []string{
	"*path",
	"*ref",
	":",
	"applications",
	"assets",
	"assignees",
	"authorizations",
	"blobs",
	"branches",
	"clients",
	"code",
	"code_frequency",
	"collaborators",
	"comments",
	"commit_activity",
	"commits",
	"contents",
	"contributors",
	"downloads",
	"email",
	"emails",
	"emojis",
	"events",
	"feeds",
	"files",
	"followers",
	"following",
	"forks",
	"gists",
	"git",
	"gitignore",
	"hooks",
	"issues",
	"keys",
	"labels",
	"languages",
	"legacy",
	"markdown",
	"members",
	"merge",
	"merges",
	"meta",
	"milestones",
	"networks",
	"notifications",
	"orgs",
	"participation",
	"public",
	"public_members",
	"pulls",
	"punch_card",
	"rate_limit",
	"raw",
	"readme",
	"received_events",
	"refs",
	"releases",
	"repos",
	"repositories",
	"search",
	"star",
	"stargazers",
	"starred",
	"stats",
	"statuses",
	"subscribers",
	"subscription",
	"subscriptions",
	"tags",
	"teams",
	"templates",
	"tests",
	"threads",
	"tokens",
	"trees",
	"user",
	"users",
}

func FastHash(m uint8, Pat string, Mx int) (h int) {
	ln := len(Pat)
	h = int(m) + ln
	for i := 0; i < len(Pat) && i < Mx; i++ {
		h += int(Pat[i])
		h += (h << 10)
		h = h ^ (h >> 6)
	}
	h += (h << 3)
	h = h ^ (h >> 11)
	h += (h << 15)
	h = ((h & bitMask) ^ ((h >> nBits) & bitMask))
	return
}

var colision [bitMask]int

func TestGithub_WordCollisions(t *testing.T) {
	for i := range colision {
		colision[i] = 0
	}
	for _, test := range words {
		h := FastHash(0, test, 4)
		colision[h]++
	}

	//fmt.Printf("Collisions in Words\n")
	nCol := 0
	for i, v := range colision {
		if v > 1 {
			nCol++
			//fmt.Printf("\t[%d] = %d, %s\n", i, v, debug.LF())
			_ = i
		}
	}
	if nCol != 2 {
		t.Errorf("Should have been 2 collision found, found %d instead\n", nCol)
	}
}

func TestGithub_AllUrlCollisions(t *testing.T) {
	r := New()
	for i := range colision {
		colision[i] = 0
	}
	for _, test := range testGithubPat1 {
		Pat := test.UrlPattern
		// m := (int(test.Method[0]) + (int(test.Method[1]) << 2) + (int(test.Method[2]) << 4))
		m := (int(test.Method[0]) + (int(test.Method[1]) << 1))
		r.SplitOnSlash3(m, test.UrlPattern, false)
		ss := m
		pp := ""
		for i := 0; i < r.NSl; i++ {
			// fmt.Printf("i=%d ->%c<-, ->%s<-", i, Pat[r.Slash[i]+1], Pat[r.Slash[i]+1:r.Slash[i+1]])
			if Pat[r.Slash[i]+1] == ':' {
				ss += 153
				pp += ":"
				// fmt.Printf(" ss=%d after : 153\n", ss)
			} else if Pat[r.Slash[i]+1] == '*' {
				ss += 51
				pp += "*"
				// fmt.Printf(" ss=%d after * 51\n", ss)
				break
			} else {
				ss = ss ^ r.Hash[i]
				// fmt.Printf(" ss=%d after adding %d\n", ss, r.Hash[i])
				pp += "T"
			}
		}
		ss = ((ss & bitMask) ^ ((ss >> nBits) & bitMask) ^ ((ss >> (nBits * 2)) & bitMask))
		colision[ss]++
	}

	// fmt.Printf("Collisions in Words\n")
	nCol := 0
	for i, v := range colision {
		if v > 1 {
			nCol++
			// fmt.Printf("\t[%d] = %d, %s\n", i, v, debug.LF())
			_ = i
		}
	}
	if nCol != 7 {
		t.Errorf("Should have been %d collision found, found %d instead\n", 7, nCol)
	}
}

// 14.2 ns per word
func OldBenchmarkFashHash1(b *testing.B) {

	for n := 0; n < b.N; n++ {
		x := FastHash(0, "tests", 4)
		_ = x
	}

}

// 31.3 ms, 95% success rate
// 163 - full lookup, w/o extradting params
// 127 - SplitOnSlash3
func OldBenchmarkGithub_LookupUrlMapHash2(b *testing.B) {
	r := New()
	for i, test := range testGithubPat1 {
		// fmt.Printf("i=%d Url=%s\n", i, test.UrlPattern)
		r.AddRoute(test.Method, test.UrlPattern, i+1, emptyTestingHandle)
	}
	// fmt.Printf("Calling Compile\n")
	r.CompileRoutes()

	var found bool
	var ln int
	var item Collision2
	var data GoGoData
	var w http.ResponseWriter
	var req http.Request

	url := "/repos/julienschmidt/httprouter/stargazers"

	Method := "GET"

	for n := 0; n < b.N; n++ {
		m := (int(Method[0]) + (int(Method[1]) << 1))
		r.SplitOnSlash3(m, url, false)
		found, ln, item = r.LookupUrlViaHash2(w, &req, &m, data)
		if found {
			r.GetArgs3(url, item.ArgPattern, item.ArgNames, ln)
		}
	}

	_, _, _ = found, ln, item

}

func OldTestGithub_LookkupViaHash2(t *testing.T) {
	r := New()
	for i, test := range testGithubPat1 {
		// fmt.Printf("i=%d Url=%s\n", i, test.UrlPattern)
		r.AddRoute(test.Method, test.UrlPattern, i+1, emptyTestingHandle)
	}
	// fmt.Printf("Calling Compile\n")
	r.CompileRoutes()

	var found bool
	var ln int
	var item Collision2
	var data GoGoData
	var w http.ResponseWriter
	var req http.Request

	url := "/repos/julienschmidt/httprouter/stargazers"

	Method := "GET"

	m := (int(Method[0]) + (int(Method[1]) << 1))
	r.SplitOnSlash3(m, url, false)
	found, ln, item = r.LookupUrlViaHash2(w, &req, &m, data)

	// fmt.Printf("This Test: %v\n", found)
	if !found {
		t.Errorf("Should have been found << New >> Hash code\n")
	}

	_, _, _ = found, ln, item
}

// 127 ns - old
// 142 ns - New with calls to FixUrl in plac
func OldBenchmarkSplitOnSlash3(b *testing.B) {
	r := New()
	for i, test := range testGithubPat1 {
		r.AddRoute(test.Method, test.UrlPattern, i+1, emptyTestingHandle)
	}
	r.CompileRoutes()

	url := "/repos/julienschmidt/httprouter/stargazers"

	for n := 0; n < b.N; n++ {
		r.SplitOnSlash3(1, url, false)
	}

}

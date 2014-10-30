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

import (
	"fmt"
	"net/http"
	"regexp"
	"sort"
	"time"

	"./debug"
)

// -------------------------------------------------------------------------------------------------
const MaxSlashInUrl = 20

/*
This is the maximum number of "/" that can be processed in a URL.  If you have more than this number
of slashes then it will try to route based on what it has found and it may succeed.  With more it
will just disregard the trailing end of the URL - and probably return a 404.

Realistically more than 20 slashes in a route is probably a mistake somewhere.  I have looked at
a bunch of APIs and none of them had more than 9 slashes in any route.
*/

// -------------------------------------------------------------------------------------------------
const nBits = 12
const bitMask = 0xfff

/*
Hash Table Size Constants

	bitMask				nBits									Probability of Miss
	Hex		Decimal		Number of Bits		Number of URLs 		resoled in single compare
	-------	-------		--------------		--------------		-------------------
	0x7ff	2048		11					75					96.34%
	0xfff	4096		12					175					95.73%
	0x1fff	8192 		13					350					95.73%
	0x3fff	16384		14					700					95.73%

If you have more URLs then you should probably increase the size of nBits and bitMask to match.
This will help to keep the performance to a near constant.
*/

//
// The average world length in English is 5.1 chars.  7 chars seems to be enough to capture
// the uniqueness of most API tokens.   I have looked at 9 different APIs and more than 7
// chars took more time but did not decrease the number of collisions in the hashing table.
// In most cases more than 7 chars increased the collisions and slowed down the processing.
//
// Note: Set to 6 for a 4x and a 3x collision - then test collision handling - in the Github test.
//
const nKeyChar = 7

type RouteData struct {
	Method   string
	Route    string
	Hdlr     int
	NFxNo    int
	NotUseIt bool
	UseIt    bool
	set      int
	Ns       int
	FileName string
	LineNo   int
	MatchIt  []Match
}

const optionEarlyExit = true

// -------------------------------------------------------------------------------------------------
/*

Jenkins Hash: In C Code from Wikipedia.   This code is in the public domain.   This is the
old Jenkins hash.    A newer function exits - but I have not profiled the results of using
it.  This hash produces a low collision rate in a small amount of time.   I may switch to
using the newer (supposedly better) hash but not until I have done timing on it to very
that it will work with a similar speed.

uint32_t jenkins_one_at_a_time_hash(char *key, size_t len)
{
    uint32_t hash, i;
    for(hash = i = 0; i < len; ++i)
    {
        hash += key[i];
        hash += (hash << 10);
        hash ^= (hash >> 6);
    }
    hash += (hash << 3);
    hash ^= (hash >> 11);
    hash += (hash << 15);
    return hash;
}

*/

// Lookup table for Handle Functions - this is used because it is much easier to keep function
// pointers out of the structures that you are working with.
// xyzzy - change name of "NFx" to something better
type NFx struct {
	Fx map[int]Handle
}

//
// The main router type.
//
type MuxRouter struct {
	// ------------------------------------------------------------------------------------------------------
	// The hash of paths/URLs
	// HashItems []HashItem
	Hash2Test []int

	// ------------------------------------------------------------------------------------------------------
	// Info used during processing of a URL
	CurUrl   string                 // The current URL being processed.
	Hash     [MaxSlashInUrl]int     // The set of hash keys in the current operation.
	Slash    [MaxSlashInUrl + 1]int // Array of locaitons for the '/' in the url.  For /abc/def, it would be [ 0, 4, 8 ]
	NSl      int                    // Number of slashes in the URL for /abc/def it would be 2
	allParam [MaxParams]Param       // The parameters for the current operation
	AllParam Params                 // Slice that pints into allParam
	UsePat   string                 // The used T::T pattern for matching - at URL time.
	cRoute   string                 //

	// Maximum number of slashes found in any route
	MaxSlash int

	widgetBefore   []GoGoWidgetSet
	widgetAfter    []GoGoWidgetSet
	widgetHashNewM []GoGoWidgetSet

	// ------------------------------------------------------------------------------------------------------
	// Setup Info
	routeData []RouteData // Raw routes - this is where the routes are kept before compile

	// ------------------------------------------------------------------------------------------------------
	// User settable handler called when no match is found.  Type: http.HandlerFunc.
	// If not set then http.NotFound will be called.
	NotFound http.HandlerFunc

	gen_hdlr int

	// ------------------------------------------------------------------------------------------------------
	// Function to handle panics recovered from http handlers.
	// It should be used to generate a error page and return the http error code
	// 500 (Internal Server Error).
	// The handler can be used to keep your server from crashing because of
	// unrecovered panics.
	PanicHandler func(http.ResponseWriter, *http.Request, interface{})

	// ------------------------------------------------------------------------------------------------------
	HasBeenCompiled bool //	Flag, set to true when the routes are compiled.

	LookupResults  []Collision2
	nLookupResults int

	OutputStatus bool // If true will output info on collision etc.
}

// Handle is a function that can be registered to a route to handle HTTP
// requests. Like http.HandlerFunc, but has a third parameter for the values of
// wildcards (variables).
type Handle func(http.ResponseWriter, *http.Request, Params)

// -------------------------------------------------------------------------------------------------
// Verify the MuxRouter works with the http.Handler interface.  This will produce a syntax
// error if there is a mismatch in interface.
// See:  https://blog.splice.com/golang-verify-type-implements-interface-compile-time/
var _ http.Handler = New()

// -------------------------------------------------------------------------------------------------
func New() *MuxRouter {
	r := &MuxRouter{
		HasBeenCompiled: false,
		MaxSlash:        1,
		gen_hdlr:        1,
		nLookupResults:  1,
		OutputStatus:    false,
	}
	fn, ln := LineFile()
	r.LookupResults = append(r.LookupResults, Collision2{cType: Dummy, FileName: fn, LineNo: ln})
	r.AllParam.Data = r.allParam[:]
	r.Hash2Test = make([]int, bitMask+1, bitMask+1)
	r.NotFound = http.NotFound // Set to default, http.NotFound handler.
	return r
}

// -------------------------------------------------------------------------------------------------
// Internal Functions
// -------------------------------------------------------------------------------------------------
func MethodToCode(Method string, AddM int) int {
	return ((int(Method[0]) + (int(Method[1]) << 1)) + AddM)
}

// -------------------------------------------------------------------------------------------------
// USED - in related GET/PUT/POST functions.

// Add a route to be handed.
//
// Handle registers a new request handle with the given path and method.
//
// For GET, POST, PUT, PATCH, OPTIONS, HEAD and DELETE requests the
// respective shortcut functions can be used.
//
// This function is intended for bulk loading and to allow the usage of less
// frequently used, non-standardized or custom methods (e.g. for internal
// communication with a proxy).
func (r *MuxRouter) AddRoute(Method string, Route string, hdlr int, fx Handle) *MuxRouter {
	if Route[0] != '/' {
		fmt.Printf("Path should begin with '/', passed %s, at %s, %s", Route, debug.LF())
		fmt.Printf("Called from: %s\n", debug.LF(1))
		fmt.Printf("Called from: %s\n", debug.LF(2))
		return r
	}
	if !validMethod[Method] {
		fmt.Printf("Method invalid, should be one of: GET, POST, PUT, PATCH, OPTIONS, HEAD, CONNECT, TRACE or DELETE, instead got %s, %s\n", Method, debug.LF())
		fmt.Printf("Called from: %s\n", debug.LF(1))
		fmt.Printf("Called from: %s\n", debug.LF(2))
		return r
	}

	var n int
	n = FxTabN
	FxTabN++
	FxTab.Fx[n] = fx
	fn, ln := LineFile(2)
	r.routeData = append(r.routeData, RouteData{
		Method:   Method,
		Route:    Route,
		Hdlr:     hdlr,
		NFxNo:    n,
		NotUseIt: false,
		UseIt:    true,
		FileName: fn,
		LineNo:   ln,
	})
	return r
}

func MatchTlsFunc(req *http.Request, data int) bool {
	return req.TLS != nil
}
func (r *MuxRouter) SetHTTPSOnly() *MuxRouter {
	r.routeData[len(r.routeData)-1].MatchIt = append(r.routeData[len(r.routeData)-1].MatchIt, Match{Data: -1, MatchFunc: MatchTlsFunc})
	return r
}

var sData []string
var Dn int

func init() {
	Dn = 0
}

func MatchHostFunc(req *http.Request, data int) bool {
	// fmt.Printf("!!!!!!!!!!!!!!!!!!!!!! host compare called for %s/%s\n", req.Host, req.RequestURI)
	return req.Host == sData[data]
}

func (r *MuxRouter) SetHost(Host string) *MuxRouter {
	r.routeData[len(r.routeData)-1].MatchIt = append(r.routeData[len(r.routeData)-1].MatchIt, Match{Data: Dn, MatchFunc: MatchHostFunc})
	sData = append(sData, Host)
	// fmt.Printf("Setting (before sort) MatchIt on %d %s\n", len(r.routeData)-1, debug.LF())
	Dn++
	return r
}

func MatchPortFunc(req *http.Request, data int) bool {
	var i, st int
	Port := "80"
	st = 0
	for i = 0; i < len(req.Host); i++ {
		if req.Host[i] == '[' {
			st = 1
		} else if st == 1 && req.Host[i] == ']' {
			st = 0
		} else if st == 0 && req.Host[i] == ':' {
			Port = req.Host[i+1:]
		}
	}
	// fmt.Printf("$$$$$$$$$$$$$$$$$$$$$$ port compare called for %s/%s\n", Port, req.RequestURI)
	return Port == sData[data]
}

func (r *MuxRouter) SetPort(Port string) *MuxRouter {
	r.routeData[len(r.routeData)-1].MatchIt = append(r.routeData[len(r.routeData)-1].MatchIt, Match{Data: Dn, MatchFunc: MatchPortFunc})
	sData = append(sData, Port)
	// fmt.Printf("Setting (before sort) MatchIt on %d %s\n", len(r.routeData)-1, debug.LF())
	Dn++
	return r
}

var FxTab NFx
var FxTabN int = 1

func init() {
	if FxTab.Fx == nil {
		FxTab.Fx = make(map[int]Handle)
	}
}

type UrlAPat struct {
	Pat  string
	Star bool
}

type UrlPat struct {
	PatList []UrlAPat
	// Pat    []string
	// Star   []bool
	PatOcc map[string]int
}

var nMatch []UrlPat // Index by Length ( NSl )
// var starPat []string // Longer than max NSl => only match to * items

func init() {
	nMatch = make([]UrlPat, MaxSlashInUrl, MaxSlashInUrl)
}

// Return the minimum of 2 integers.
func minInt(a int, b int) int {
	if a < b {
		return a
	}
	return b
}

// Post process nMatch adding all of the "*" patterns where they need to be during LookupUrlViaHash2.
//
// for the longest pattern with a star
//    for each pattern that is longer than that up to min(MaxSlashInUrl,r.MasSlash+1)
//		Add that pattern (the * one) to the set of the others.
//
func (r *MuxRouter) addStarPat() {
	// fmt.Printf("nMatch=%s\n", debug.SVarI(nMatch))
	for i := minInt(MaxSlashInUrl-1, r.MaxSlash+1); i > 0; i-- { // nothing at 0 so skip it.
		// fmt.Printf("i=%d\n", i)
		if nMatch[i].PatList != nil {
			// fmt.Printf("Star is not nil\n")
			for ii := 0; ii < len(nMatch[i].PatList); ii++ {
				// fmt.Printf("ii=%d\n", ii)
				if nMatch[i].PatList[ii].Star {
					mm := minInt(MaxSlashInUrl, r.MaxSlash+1)
					for j := i + 1; j < mm; j++ {
						// do add
						p := nMatch[i].PatList[ii]
						nMatch[j].PatList = append(nMatch[j].PatList, p)
						if nMatch[j].PatOcc == nil {
							nMatch[j].PatOcc = make(map[string]int)
						}
						nMatch[j].PatOcc[p.Pat] = 1
					}
				}
			}
		}
	}
}

// Return true if the sting 'p' has a '*' in it.
// Possible Improvement: this should be a better scanner - this, just look for '*' character is probably buggy.
func hasStar(p string) bool {
	// func numChar(s string, c rune) (rv int) {
	// fmt.Printf("p ->%s<- numChcar=%d\n", p, numChar(p, '*'))
	if numChar(p, '*') > 0 {
		return true
	}
	return false
}

// Add a pattern to nMatch - check to see if it is already there.
// Possible Improvement - inefficient/slow but it works.
func addPat2(NSl int, p string, FileName string, LineNo int) {
	f := false
	// fmt.Printf("NSl=%d ->%s<- %s\n", NSl, p, debug.LF())
	for _, v := range nMatch[NSl].PatList {
		if v.Pat == p {
			f = true
			break
		}
	}
	if !f {
		nMatch[NSl].PatList = append(nMatch[NSl].PatList, UrlAPat{Pat: p, Star: hasStar(p)})
	}
	if nMatch[NSl].PatOcc == nil {
		nMatch[NSl].PatOcc = make(map[string]int)
	}
	nMatch[NSl].PatOcc[p]++
}

type colType int

const (
	IsWord    colType = 1 << iota
	MultiUrl          = 1 << iota
	SingleUrl         = 1 << iota
	Dummy             = 1 << iota
)

type MatchFunc func(req *http.Request, data int) bool

type Match struct {
	Data      int
	MatchFunc MatchFunc
}

type Re struct {
	Pos  int
	Re   string
	cRe  *regexp.Regexp
	Name string
}

// xyzzy - add match/match-data
type ReList struct {
	Hdlr     int      // User specified int, mostly for testing
	Fx       Handle   // Function to call to handle this request
	ArgNames []string //
	ReSet    []Re
	MatchIt  []Match
}

type Collision2 struct {
	cType      colType               // if IsWord, then this is just a marker that the prefix is a valid m+word hash
	Url        string                // /path/:v1/:v2/whatever
	NSL        int                   // number of / in the URL/Route
	CleanUrl   string                // /path/:/:/whatever
	Hdlr       int                   // User specified int, mostly for testing
	Fx         Handle                // Function to call to handle this request
	TPat       string                // T::T
	ArgNames   []string              //
	ArgPattern string                // T::T dup?? not used
	LineNo     int                   // Location this was created
	FileName   string                // Location this was created
	HasRe      []ReList              // Set of RE that is required to match this Collision2
	MatchIt    []Match               // If additional matching criteria are used
	Multi      map[string]Collision2 // if (cType&MultiUrl)!=0, then use string to disambiguate collisions
}

// With a known bad URL, that has //, /./, or /../ in it, fix the URL.
func (r *MuxRouter) FixBadUrl(Url string) (rv string, fixed bool) {
	fixed = false
	rv = Url
	var tmpUrlArray [MaxSlashInUrl]string // Decleration takes 200 ns - Should move to r. so declare at startup.
	tmpUrl := tmpUrlArray[:]
	n := FixPath(Url, tmpUrl, MaxSlashInUrl)
	tmpUrl = tmpUrl[:n]
	nUrl := ""
	for i, v := range tmpUrl {
		if i > 0 {
			nUrl += "/" + v
		}
	}
	// fmt.Printf("Trailing Slash: nUrl ->%s<- Url ->%s<-\n", nUrl, Url)
	if nUrl != Url {
		r.CurUrl = nUrl
		rv = nUrl
		fixed = true
	}
	return
}

func (r *MuxRouter) addHash2Map(Method string, Route string, cleanRoute string, hdlr int, fx Handle, names []string, AddToM []Match, ns int, FileName string, LineNo int) {
	//if dbMatch2 {
	//	fmt.Printf("addHash2Map: len(AddToM) = %d %s\n", len(AddToM), debug.LF())
	//}
	var i int
	var ss int
	ss = 0
	var pp string
	tmpRe := make([]Re, 0, MaxSlashInUrl)
	reNames := make([]string, 0, MaxSlashInUrl)
	pp = ""
	//if dbHash2 || dbMatch2 {
	//	fmt.Printf("\nTOP(addHash2Map): %s %s (%s) => %d, %s\n", Method, Route, cleanRoute, hdlr, debug.LF())
	//}
	// m := ((int(Method[0]) + (int(Method[1]) << 1)) + AddToM) ^ (ns << 2)
	m := MethodToCode(Method, 0)
	// fmt.Printf("m=%d\n", m)
	r.SplitOnSlash3(m, Route, false)
	if optionEarlyExit {
		hh := (r.Hash[0] ^ m) & bitMask
		// fmt.Printf("hh=%d bitMask=%x\n", m, bitMask)
		if r.Hash2Test[hh] == 0 {
			r.Hash2Test[hh] = r.nLookupResults
			r.LookupResults = append(r.LookupResults, Collision2{cType: IsWord})
			r.nLookupResults++
		}
	}
	//if dbHash2 {
	//	fmt.Printf("After SplitOnSlash3 Orig:->%s<- Fixed:->%s<-\n r.Hash=%s r.Slash=%s r.NSl=%d\n", Route, r.CurUrl, debug.SVar(r.Hash[0:r.NSl]), debug.SVar(r.Slash[0:r.NSl+1]), r.NSl)
	//}
	haveRealRe := false
	for i = 0; i < r.NSl; i++ {
		//if dbHash2 {
		//	fmt.Printf("i=%d ->%c<-, ->%s<-", i, Route[r.Slash[i]+1], Route[r.Slash[i]+1:r.Slash[i+1]])
		//}
		if Route[r.Slash[i]+1] == ':' {
			ss += 153
			pp += ":"
			reNames = append(reNames, Route[r.Slash[i]+2:r.Slash[i+1]])
			//if dbHash2 {
			//	fmt.Printf(" ss=%d after : 153\n", ss)
			//}
		} else if Route[r.Slash[i]+1] == '*' {
			ss += 51
			pp += "*"
			//if dbHash2 {
			//	fmt.Printf(" ss=%d after * 51\n", ss)
			//}
			reNames = append(reNames, Route[r.Slash[i]+2:r.Slash[i+1]])
			break
		} else if Route[r.Slash[i]+1] == '{' {
			name, re, valid, convertToColon := parseReFromToken3(Route[r.Slash[i]+1 : r.Slash[i+1]])
			_, _, _, _ = name, re, valid, convertToColon
			if convertToColon {
				ss += 153
				pp += ":"
				reNames = append(reNames, name)
				//if dbHash2 {
				//	fmt.Printf(" ss=%d after : 153 / convertToColon, name=%s re=%s valid=%v \n", ss, name, re, valid)
				//}
			} else {
				haveRealRe = true
				ss += 211
				pp += "{"
				aRe := regexp.MustCompile(re)
				tmpRe = append(tmpRe, Re{Pos: i, Re: re, Name: name, cRe: aRe})
				reNames = append(reNames, name)
				//if dbHash2 {
				//	fmt.Printf(" ss=%d after { 211\n", ss)
				//}
			}
		} else {
			ss = ss ^ r.Hash[i]
			//if dbHash2 {
			//	fmt.Printf(" ss=%d after adding %d\n", ss, r.Hash[i])
			//}
			pp += "T"
		}
	}
	ss = ((ss & bitMask) ^ ((ss >> nBits) & bitMask) ^ ((ss >> (nBits * 2)) & bitMask))
	//if dbHash2 || dbMatch2 {
	//	fmt.Printf("ss=%-5d m=%4d/%s Url=%s, %s %d\n", ss, m, Method, Route, FileName, LineNo)
	//}
	//if haveRealRe {
	//	if dbHash2 {
	//		fmt.Printf("****** Have a real re *******\nreNames=%s tmpRe=%s\n", debug.SVar(reNames), dumpReArray(tmpRe))
	//	}
	//}
	if r.Hash2Test[ss] == 0 {
		// fmt.Printf("At %s\n", debug.LF())
		// fmt.Printf("Adding to empty locaiton in table, r.Hash2Test[ss]==0\n")
		r.Hash2Test[ss] = r.nLookupResults
		if haveRealRe {
			// fmt.Printf("At %s\n", debug.LF())
			r.LookupResults = append(r.LookupResults, Collision2{cType: SingleUrl, Url: Route, NSL: ns, CleanUrl: cleanRoute, Hdlr: hdlr, Fx: fx, TPat: pp, FileName: FileName,
				LineNo: LineNo, ArgNames: names, MatchIt: AddToM,
				HasRe: []ReList{ReList{Hdlr: hdlr, Fx: fx, ArgNames: reNames, ReSet: tmpRe}}})
		} else {
			// fmt.Printf("At %s\n", debug.LF())
			r.LookupResults = append(r.LookupResults, Collision2{cType: SingleUrl, Url: Route, NSL: ns, CleanUrl: cleanRoute, Hdlr: hdlr, Fx: fx, TPat: pp, FileName: FileName, LineNo: LineNo, ArgNames: names, MatchIt: AddToM})
		}
		// fmt.Printf("At %s\n", debug.LF())
		r.nLookupResults++
	} else { // Have a collision on our hands.
		c := r.Hash2Test[ss]
		// fmt.Printf("At %s\n", debug.LF())
		if optionEarlyExit && r.LookupResults[c].cType == IsWord { // No biggie - just a IsWord marker.
			if haveRealRe {
				// fmt.Printf("At %s\n", debug.LF())
				r.LookupResults[c] = Collision2{cType: (SingleUrl | IsWord), Url: Route, NSL: ns, CleanUrl: cleanRoute, Hdlr: hdlr, Fx: fx, TPat: pp, FileName: FileName,
					LineNo: LineNo, ArgNames: names, MatchIt: AddToM,
					HasRe: []ReList{ReList{Hdlr: hdlr, Fx: fx, ArgNames: reNames, ReSet: tmpRe, MatchIt: AddToM}}}
			} else {
				// fmt.Printf("At %s\n", debug.LF())
				r.LookupResults[c] = Collision2{cType: (SingleUrl | IsWord), Url: Route, NSL: ns, CleanUrl: cleanRoute, Hdlr: hdlr, Fx: fx, TPat: pp, FileName: FileName,
					LineNo: LineNo, MatchIt: AddToM, ArgNames: names}
			}
		} else {
			// fmt.Printf("At %s\n", debug.LF())
			//if dbHash2 {
			//	fmt.Printf("+==========================================+\n| Just a collision                         |\n+==========================================+\n")
			//}
			old := r.LookupResults[c]
			if old.HasRe != nil && haveRealRe { // need to check to see if is alreay a RE in old.  If so just append
				// fmt.Printf("At %s\n", debug.LF())
				//if dbHash2 {
				//	fmt.Printf("Old - is just a RE, so append it\n")
				//}
				old.HasRe = append(old.HasRe, ReList{Hdlr: hdlr, Fx: fx, ArgNames: reNames, ReSet: tmpRe, MatchIt: AddToM})
			} else {
				//if dbHash2 {
				//	fmt.Printf("Before Multi Check:cType=%04x,%s %s\n", old.cType, dumpCType(old.cType), debug.LF())
				//}
				// fmt.Printf("At %s\n", debug.LF())
				if old.Multi == nil {
					//if dbHash2 {
					//	fmt.Printf("Multi is NIL\n")
					//}
					old.Multi = make(map[string]Collision2)
					//if dbHash2 {
					//	fmt.Printf("Before:cType=%04x,%s %s\n", old.cType, dumpCType(old.cType), debug.LF())
					//}
					old.cType = (old.cType & (^SingleUrl))
					old.cType |= MultiUrl
					//if dbHash2 {
					//	fmt.Printf("After:cType=%04x,%s\n", old.cType, dumpCType(old.cType))
					//}
					// fmt.Printf("At %s\n", debug.LF())
					old.Multi[old.CleanUrl] = r.LookupResults[c]
				}

				// One of the cleanRoute items is the colliding one - we will need to set the RE on that one.
				if haveRealRe {
					// fmt.Printf("At %s\n", debug.LF())
					//if dbHash2 {
					//	fmt.Printf("Have RE in multi colision\n")
					//}
					if xx, multiOk := old.Multi[cleanRoute]; multiOk { // If we have already seen this cleanRoute, then just append the RE
						//if dbHash2 {
						//	fmt.Printf("HasRe - append case\n")
						//}
						xx.HasRe = append(xx.HasRe, ReList{Hdlr: hdlr, Fx: fx, ArgNames: reNames, ReSet: tmpRe, MatchIt: AddToM})
						old.Multi[cleanRoute] = xx
					} else {
						//if dbHash2 {
						//	fmt.Printf("Crea new entry in multi\n")
						//}
						old.Multi[cleanRoute] = Collision2{cType: MultiUrl, Url: Route, NSL: ns, CleanUrl: cleanRoute, Hdlr: hdlr, Fx: fx, TPat: pp, FileName: FileName, LineNo: LineNo, ArgNames: names, MatchIt: AddToM,
							HasRe: []ReList{ReList{Hdlr: hdlr, Fx: fx, ArgNames: reNames, ReSet: tmpRe, MatchIt: AddToM}}}
					}
				} else {
					// fmt.Printf("At %s\n", debug.LF())
					//if dbHash2 {
					//	fmt.Printf("NON Re Case - just insert into Multi\n")
					//}
					old.Multi[cleanRoute] = Collision2{cType: MultiUrl, Url: Route, NSL: ns, CleanUrl: cleanRoute, Hdlr: hdlr, Fx: fx, TPat: pp, FileName: FileName, LineNo: LineNo, MatchIt: AddToM, ArgNames: names}
				}
			}
			r.LookupResults[c] = old
		}
	}
	//if dbLookupUrlMap {
	//	fmt.Printf("%-10s : %s\n", pp, Route)
	//}
	// fmt.Printf("At -- that's all folks -- %s\n", debug.LF())
}

// Input:  Pattern like T::T and the current URL with Slash locaiton information.
// So... /abc/:def/ghi is the Route, /abc/x/ghi is the ULR, Slash is [ 0, 4, 6, 10 ]
// The output is /abc/:/ghi - Sutiable for lookup in a map of cleanRoute
func (r *MuxRouter) UrlToCleanRoute(UsePat string) (rv string) {
	for i, v := range UsePat { // Pat is T::T format pattern
		if v == ':' {
			rv += "/:"
		} else if v == '*' {
			rv += "/*"
			break
		} else if v == '{' {
			rv += "/{"
		} else {
			rv += "/" + r.CurUrl[r.Slash[i]+1:r.Slash[i+1]]
		}
	}
	return
}

// compate r.CurUrl to a Pattern
func (r *MuxRouter) CmpUrlToCleanRoute(UsePat string, CleanUrl string) (rv bool) {
	rv = true
	// for i, v := range UsePat { // Pat is T::T format pattern
	k := 1 // Index into CleanUrl
	//if dbCmp {
	//	fmt.Printf("UsePat ->%s<-\n", UsePat)
	//}
	for i := 0; i < len(UsePat); i++ {
		v := UsePat[i]
		//if dbCmp {
		//	fmt.Printf("k=%d i=%d, UsePat[i]=%c\n", k, i, UsePat[i])
		//}
		if v == ':' {
			k += 2
		} else if v == '*' {
			break
		} else if v == '{' {
			k += 2
		} else {
			// r.cRoute += "/" + r.CurUrl[r.Slash[i]+1:r.Slash[i+1]]
			l := r.Slash[i+1] - r.Slash[i] - 1
			//if dbCmp {
			//	fmt.Printf("T: l=%d\n", l)
			//	fmt.Printf("T: left=->%s<-\n", r.CurUrl[r.Slash[i]+1:r.Slash[i+1]])
			//	fmt.Printf("T: rght=->%s<-\n", CleanUrl[k:k+l])
			//}
			if r.CurUrl[r.Slash[i]+1:r.Slash[i+1]] != CleanUrl[k:k+l] {
				//if dbCmp {
				//	fmt.Printf("match failed\n")
				//}
				rv = false
				return
			}
			k += l + 1
		}
	}
	return
}

// xyzzy - not take into account ReList -
// func (r *MuxRouter) WidgetMatch(c2 Collision2, w http.ResponseWriter, req *http.Request, m *int, data GoGoData) bool {
func (r *MuxRouter) WidgetMatch(MatchIt []Match, w http.ResponseWriter, req *http.Request, m *int, data GoGoData) bool {
	if MatchIt != nil {
		for i, v := range MatchIt {
			_ = i
			b := v.MatchFunc(req, v.Data)
			if !b {
				return false
			}
		}
	}
	return true
}

func (r *MuxRouter) LookupUrlViaHash2(w http.ResponseWriter, req *http.Request, m *int, data GoGoData) (found bool, ln int, rv Collision2) {
	Url := r.CurUrl
	found = false
	ln = len(Url)
	var ss int
	var cRoute string
	var reMatch bool

	//if dbLookupUrlMap {
	//	fmt.Printf("\n\nLookupUrlViaHash2: Top of Lookup test %s\n", debug.LF())
	//}
	if r.NSl > minInt(MaxSlashInUrl-1, r.MaxSlash+1) {
		r.NSl = minInt(MaxSlashInUrl-1, r.MaxSlash+1)
	}
	//if dbLookupUrlMap2 {
	//	fmt.Printf("LookupUrlViaHash2: %s, r.NSl=%d, len(nMatch[%d].PatList)=%d\n", debug.LF(), r.NSl, r.NSl, len(nMatch[r.NSl].PatList))
	//	fmt.Printf("LookupUrlViaHash2: nMatch[%d]=%s\n", r.NSl, debug.SVarI(nMatch[r.NSl]))
	//	// fmt.Printf("nMatch=%s\n", debug.SVarI(nMatch))
	//}
	k2 := len(nMatch[r.NSl].PatList)
	for jj := 0; jj < k2; jj++ {
		ss = 0
		xPat := nMatch[r.NSl].PatList[jj].Pat
		//if dbHash2 {
		//	fmt.Printf("Top of Pat Match Loop, jj=%d pat=%s\n", jj, xPat)
		//}
		r.UsePat = xPat
		for i := 0; i < r.NSl; i++ {
			if xPat[i] == ':' {
				ss += 153
				//if dbHash2 {
				//	fmt.Printf(" ss=%d after : 153\n", ss)
				//}
			} else if xPat[i] == '*' {
				ss += 51
				//if dbHash2 {
				//	fmt.Printf(" ss=%d after * 51\n", ss)
				//}
				break
			} else if xPat[i] == '{' {
				ss += 211
				//if dbHash2 {
				//	fmt.Printf(" ss=%d after { 211\n", ss)
				//}
			} else {
				ss = ss ^ r.Hash[i]
				//if dbHash2 {
				//	fmt.Printf(" ss=%d after adding %d\n", ss, r.Hash[i])
				//}
			}
		}
		ss = ((ss & bitMask) ^ ((ss >> nBits) & bitMask) ^ ((ss >> (nBits * 2)) & bitMask))
		//if dbLookupUrlMap2 {
		//	fmt.Printf("ss=%s, %s\n", debug.SVar(ss), debug.LF())
		//}
		if r.Hash2Test[ss] > 0 {
			//if dbHash2 {
			//	fmt.Printf("Found [%d] at ss=%d, %s\n", r.Hash2Test[ss], ss, debug.LF())
			//}
			// 1. match all the constants and disambiguate.  If Collision2 - has REFlag set to true ( > 0 Re Sets ) then iterate over them to see if we get a match.
			// xyzzy-widget
			c := r.LookupResults[r.Hash2Test[ss]]
			if (c.cType & SingleUrl) != 0 {
				//if dbHash2 {
				//	fmt.Printf("Found a SingleUrl to be true - just return success?? what about RE, %s\n", debug.LF())
				//}
				if c.HasRe != nil { // if we have a RE-List - then Iterate over it for a match.
					//if dbHash2 {
					//	fmt.Printf("Ok, has RE\n")
					//}
					for k, ww := range c.HasRe { // deal RE at this point too
						//if dbHash2 {
						//	fmt.Printf("k=%d, len()=%d\n", k, len(c.HasRe))
						//}
						_ = k
						reMatch = true
						for m, x := range ww.ReSet {
							_ = m
							if !x.cRe.MatchString(Url[r.Slash[x.Pos]+1 : r.Slash[x.Pos+1]]) {
								//if dbHash2 {
								//	fmt.Printf("Found false match on set k=%d\n", k)
								//}
								reMatch = false
								goto next
							}
						}
						if reMatch {
							//if dbHash2 {
							//	fmt.Printf("RE match is found\n")
							//}
							// xyzzy-widget -- Final matching on user stuff
							if ww.MatchIt != nil {
								if r.WidgetMatch(ww.MatchIt, w, req, m, data) {
									found = true
									rv.Hdlr = ww.Hdlr
									rv.Fx = ww.Fx
									rv.ArgNames = ww.ArgNames
									return
								}
							} else {
								found = true
								rv.Hdlr = ww.Hdlr
								rv.Fx = ww.Fx
								rv.ArgNames = ww.ArgNames
								return
							}
						}
					next:
					}
				} else {
					/* Problem at this locaiton xyzzy xyzzy */
					// r.UrlToCleanRoute3(r.UsePat) // URL to cRoute? -- URL --
					// if dbHash2 {
					// 	fmt.Printf("NO RE match is found?? TPat=->%s<-, match v.s. URL ->%s<- CleanUrl ->%s<- cRoute ->%s<-\n", c.TPat, c.Url, c.CleanUrl, cRoute)
					// }
					// if c.CleanUrl == r.cRoute {
					//if dbHash2 {
					//	fmt.Printf("NO RE match is found?? TPat=->%s<-, match v.s. URL ->%s<- CleanUrl ->%s<-\n", c.TPat, c.Url, c.CleanUrl)
					//}
					if r.CmpUrlToCleanRoute(r.UsePat, c.CleanUrl) {
						//if dbHash2 {
						//	fmt.Printf("   Matched on absolute pattern, returning success\n")
						//}
						// xyzzy-widget -- Final matching on user stuff
						if c.MatchIt != nil {
							if r.WidgetMatch(c.MatchIt, w, req, m, data) {
								found = true
								rv = c
								return
							}
						} else {
							found = true
							rv = c
							return
						}
					} else {
						//if dbHash2 {
						//	fmt.Printf("   Did not match\n")
						//}
					}
				}
			} else if (c.cType & MultiUrl) != 0 {
				// xyzzy - this is where to use NSL
				cRoute = r.UrlToCleanRoute(r.UsePat) // xyzzy- will alloc memory on call- URL to cRoute? -- URL --
				if c2, ok := c.Multi[cRoute]; ok {
					if c2.HasRe != nil { // if we have a RE-List - then Iterate over it for a match.
						for k, ww := range c2.HasRe { // deal RE at this point too
							_ = k
							reMatch = true
							for m, x := range ww.ReSet {
								_ = m
								if !x.cRe.MatchString(Url[r.Slash[x.Pos]+1 : r.Slash[x.Pos+1]]) {
									reMatch = false
									goto next2
								}
							}
							if reMatch {
								// xyzzy-widget -- Final matching on user stuff
								if ww.MatchIt != nil {
									if r.WidgetMatch(ww.MatchIt, w, req, m, data) {
										found = true
										rv.Hdlr = ww.Hdlr
										rv.Fx = ww.Fx
										rv.ArgNames = ww.ArgNames
										return
									}
								} else {
									found = true
									rv.Hdlr = ww.Hdlr
									rv.Fx = ww.Fx
									rv.ArgNames = ww.ArgNames
									return
								}
							}
						next2:
						}
					} else {
						// xyzzy-widget -- Final matching on user stuff
						if c2.MatchIt != nil {
							if r.WidgetMatch(c2.MatchIt, w, req, m, data) {
								found = true
								rv = c2
								return
							}
						} else {
							found = true
							rv = c2
							return
						}
					}
				}
			}
		}
	}
	return
}

// -------------------------------------------------------------------------------------------------
// Profile of routes.  Use to optimize processing.
// -------------------------------------------------------------------------------------------------

// Count the number of characters 'c' in the string 's', return that value.
func numChar(s string, c rune) (rv int) {
	rv = 0
	for _, v := range s {
		if v == c {
			rv++
		}
	}
	return
}

// Iterate over the set of routes and calculate the number of '/' in each route.
func (r *MuxRouter) calcNumSlash() {
	for i, v := range r.routeData {
		ns := numChar(v.Route, '/')
		r.routeData[i].Ns = ns
		// fmt.Printf("ns=%2d %s\n", ns, v.Route)
	}
	r.MaxSlash = 1
	for _, v := range r.routeData {
		if r.MaxSlash < v.Ns {
			r.MaxSlash = v.Ns
		}
	}
	// fmt.Printf("MaxSlash=%d\n", r.MaxSlash)
}

// -------------------------------------------------------------------------------------------------
// Sorting of routes.   Use to improve the average case - at the expense of more rare cases.
// -------------------------------------------------------------------------------------------------

type lessFunc func(p1, p2 *RouteData) bool

// multiSorter implements the Sort interface, sorting the the_data within.
type multiSorter struct {
	the_data []RouteData
	less     []lessFunc
}

// Sort sorts the argument slice according to the less functions passed to OrderedBy.
func (ms *multiSorter) Sort(the_data []RouteData) {
	ms.the_data = the_data
	sort.Sort(ms)
}

// OrderedBy returns a Sorter that sorts using the less functions, in order.
// Call its Sort method to sort the data.
func OrderedBy(less ...lessFunc) *multiSorter {
	return &multiSorter{
		less: less,
	}
}

// Len is part of sort.Interface.
func (ms *multiSorter) Len() int {
	return len(ms.the_data)
}

// Swap is part of sort.Interface.
func (ms *multiSorter) Swap(i, j int) {
	ms.the_data[i], ms.the_data[j] = ms.the_data[j], ms.the_data[i]
}

// Less is part of sort.Interface. It is implemented by looping along the
// less functions until it finds a comparison that is either Less or
// !Less. Note that it can call the less functions twice per call. We
// could change the functions to return -1, 0, 1 and reduce the
// number of calls for greater efficiency: an exercise for the reader.
func (ms *multiSorter) Less(i, j int) bool {
	p, q := &ms.the_data[i], &ms.the_data[j]
	// Try all but the last comparison.
	var k int
	for k = 0; k < len(ms.less)-1; k++ {
		less := ms.less[k]
		switch {
		case less(p, q):
			// p < q, so we have a decision.
			return true
		case less(q, p):
			// p > q, so we have a decision.
			return false
		}
		// p == q; try the next comparison.
	}
	// All comparisons to here said "equal", so just return whatever
	// the final comparison reports.
	return ms.less[k](p, q)
}

// Build the route pattern table.  A route of /abc/:def/ghi will become T:T for the fixed tokens and return
// the string /abc/:/ghi for a matching patter for colision resolution.   The T:T patterns are stored by
// addPat2().
func (r *MuxRouter) addPatT__T(Route string, hdlr int, fx Handle, FileName string, LineNo int) (ss string, names []string) {
	i, k := 0, 0
	pp := ""
	r.SplitOnSlash3(1, Route, false)
	for i = 0; i < r.NSl; i++ {
		if r.CurUrl[r.Slash[i]+1] == ':' {
			ss += "/:"
			pp += ":"
			names = append(names, r.CurUrl[r.Slash[i]+2:r.Slash[i+1]])
			k++
		} else if r.CurUrl[r.Slash[i]+1] == '*' {
			ss += "/*"
			pp += "*"
			names = append(names, r.CurUrl[r.Slash[i]+2:r.Slash[i+1]])
			k++
			break
		} else if r.CurUrl[r.Slash[i]+1] == '{' {
			name, re, valid, convertToColon := parseReFromToken3(Route[r.Slash[i]+1 : r.Slash[i+1]])
			names = append(names, name)
			_, _, _, _ = name, re, valid, convertToColon
			if convertToColon {
				ss += "/:"
				pp += ":"
				k++
				//if dbHash2 {
				//	fmt.Printf(" ss=%s after : {nanme} / convertToColon, name=%s re=%s valid=%v \n", ss, name, re, valid)
				//}
			} else {
				ss += "/{"
				pp += "{"
				k++
				//if dbHash2 {
				//	fmt.Printf(" ss=%s after {name:re}\n", ss)
				//}
			}
		} else {
			ss += "/" + r.CurUrl[r.Slash[i]+1:r.Slash[i+1]]
			pp += "T"
		}
	}
	addPat2(r.NSl, pp, FileName, LineNo)
	r.UsePat = pp
	// ss = pp

	//if dbHash2 {
	//	fmt.Printf("names=%s, %s\n", debug.SVar(names), debug.LF())
	//}
	return
}

// -------------------------------------------------------------------------------------------------
/*
	Degrees of Freedom, 			Lo .. Hi		sortDf(Pat[i])
		T 		0
		{		1
		:		2
		*		3
*/
func sortDf(Pat string) (rv int) {
	rv = 0
	for _, v := range Pat {
		if v == 'T' {
		} else if v == '{' {
			rv += 1
		} else if v == ':' {
			rv += 2
		} else if v == '*' {
			rv += 3
		}
	}
	return
}

// -------------------------------------------------------------------------------------------------
const dbMap1 = true

func (r *MuxRouter) CompileRoutes() {

	if r.HasBeenCompiled {
		return
	}
	r.HasBeenCompiled = true // Mark that the compilation has taken place.
	r.calcNumSlash()         // Use this to find over MaxSlashInUrl of slashes and report error/warn.

	//if dbCompileRoutes2 {
	//	fmt.Printf("%s\n%s\n%s\n", debug.LF(1), debug.LF(2), debug.LF(3))
	//	for i, v := range r.routeData {
	//		fmt.Printf("BEF: %3d: %s => %d   ns=%d\n", i, v.Route, v.Hdlr, v.Ns)
	//	}
	//}

	// -------------------------------------------------------------------------------------------------
	sf_NumSlash_Desc := func(c1, c2 *RouteData) bool {
		return c1.Ns > c2.Ns
	}
	sf_Length_Desc := func(c1, c2 *RouteData) bool {
		return len(c1.Route) > len(c2.Route)
	}
	sf_Text := func(c1, c2 *RouteData) bool {
		return c1.Route < c2.Route
	}
	sf_MethodHash := func(c1, c2 *RouteData) bool {
		return MethodToCode(c1.Method, 0) < MethodToCode(c2.Method, 0)
	}
	// -------------------------------------------------------------------------------------------------

	OrderedBy(sf_MethodHash, sf_NumSlash_Desc, sf_Length_Desc, sf_Text).Sort(r.routeData)
	//if dbCompileRoutes2 {
	//	fmt.Printf("\n\n")
	//	for i, v := range r.routeData {
	//		fmt.Printf("AFT: %3d: %s => %d   ns=%d\n", i, v.Route, v.Hdlr, v.Ns)
	//	}
	//}

	for i, v := range r.routeData {
		fx := FxTab.Fx[v.NFxNo]
		_ = i
		//if dbHash2 {
		//	fmt.Printf("\n=====================================================================\n")
		//	fmt.Printf("At [%d] Sorted Url %s %s, from NFxNo=%d\n", i, v.Method, v.Route, v.NFxNo)
		//}
		cleanRoute, names := r.addPatT__T(v.Route, v.Hdlr, fx, v.FileName, v.LineNo)
		// cleanRoute := r.UrlToCleanRoute(r.UsePat) // URL to cRoute? -- URL --
		ns := numChar(v.Route, '/')
		//if len(v.MatchIt) > 0 {
		//	dbMatch2 = true
		//}
		r.addHash2Map(v.Method, v.Route, cleanRoute, v.Hdlr, fx, names, v.MatchIt, ns, v.FileName, v.LineNo) // AddToM
		//dbMatch2 = false
	}
	r.addStarPat()

	// fmt.Printf("Before Sort nMatch=%s %s\n", debug.SVarI(nMatch), debug.LF())
	r.sortPat()
	// fmt.Printf("After Sort nMatch=%s %s\n", debug.SVarI(nMatch), debug.LF())
}

// -------------------------------------------------------------------------------------------------
/*

	Length Longest to Shortest, 	Hi .. Lo		len(Pat[i])
	Degrees of Freedom, 			Lo .. Hi		sortDf(Pat[i])
		T 		0
		{		1
		:		2
		*		3
	Frequencey in PatOcc, 			Hi .. Lo


This is dependent on having r.MaxSlash set properly.   That is set in calcNumSlash.

*/
func (r *MuxRouter) sortPat() {
	var CurPatOcc map[string]int
	sp_Length_Desc := func(c1, c2 *UrlAPat) bool {
		return len(c1.Pat) > len(c2.Pat)
	}
	sp_DF := func(c1, c2 *UrlAPat) bool {
		return sortDf(c1.Pat) < sortDf(c2.Pat)
	}
	sp_Frequency := func(c1, c2 *UrlAPat) bool {
		return CurPatOcc[c1.Pat] < CurPatOcc[c2.Pat]
	}
	sp_Text := func(c1, c2 *UrlAPat) bool {
		return c1.Pat < c2.Pat
	}
	for i := 0; i < minInt(MaxSlashInUrl, r.MaxSlash+1); i++ {
		if nMatch[i].PatList != nil && len(nMatch[i].PatList) > 1 {
			CurPatOcc = nMatch[i].PatOcc
			// fmt.Printf("sortPat: (before) nMatch[%d]=%s\n", i, debug.SVarI(nMatch[i]))
			OrderedByPat(sp_Length_Desc, sp_DF, sp_Frequency, sp_Text).Sort(nMatch[i].PatList)
			// fmt.Printf("sortPat: (after) nMatch[%d]=%s\n", i, debug.SVarI(nMatch[i]))
		}
	}
}

// -------------------------------------------------------------------------------------------------
func (r *MuxRouter) GetArgs3(Url string, _ string, names []string, _ int) {
	k := 0
	// db("GetArgs3","names=%s r.UsePat=%s %s\n", debug.SVar(names), r.UsePat, debug.LF())
	for i, v := range r.UsePat {
		// db("GetArgs3","k=%d\n", k)
		if i < MaxSlashInUrl-1 {
			vv := ""
			if v == ':' {
				if r.Slash[i]+1 < len(Url) && r.Slash[i+1] <= len(Url) {
					vv = Url[r.Slash[i]+1 : r.Slash[i+1]]
				}
				AddValueToParams(names[k], vv, ':', FromURL, &r.AllParam)
				k++
			} else if v == '{' {
				if r.Slash[i]+1 < len(Url) && r.Slash[i+1] <= len(Url) {
					vv = Url[r.Slash[i]+1 : r.Slash[i+1]]
				}
				AddValueToParams(names[k], vv, '{', FromURL, &r.AllParam)
				k++
			} else if v == '*' {
				if r.Slash[i]+1 < len(Url) {
					vv = Url[r.Slash[i]+1:]
				}
				AddValueToParams(names[k], vv, '{', FromURL, &r.AllParam)
				k++
			}
		}
	}
}

// -------------------------------------------------------------------------------------------------
func (r *MuxRouter) recv(w http.ResponseWriter, req *http.Request) {
	if rcv := recover(); rcv != nil {
		r.PanicHandler(w, req, rcv)
	}
}

// GET is a shortcut for router.Handle("GET", path, handle)
func (r *MuxRouter) GET(path string, handle Handle) {
	r.AddRoute("GET", path, r.gen_hdlr, handle)
	r.gen_hdlr++
}

// POST is a shortcut for router.AddRoute("POST", path, r.gen_hdlr, handle)
func (r *MuxRouter) POST(path string, handle Handle) {
	r.AddRoute("POST", path, r.gen_hdlr, handle)
	r.gen_hdlr++
}

// PUT is a shortcut for router.AddRoute("PUT", path, r.gen_hdlr, handle)
func (r *MuxRouter) PUT(path string, handle Handle) {
	r.AddRoute("PUT", path, r.gen_hdlr, handle)
	r.gen_hdlr++
}

// PATCH is a shortcut for router.AddRoute("PATCH", path, r.gen_hdlr, handle)
func (r *MuxRouter) PATCH(path string, handle Handle) {
	r.AddRoute("PATCH", path, r.gen_hdlr, handle)
	r.gen_hdlr++
}

// DELETE is a shortcut for router.AddRoute("DELETE", path, r.gen_hdlr, handle)
func (r *MuxRouter) DELETE(path string, handle Handle) {
	r.AddRoute("DELETE", path, r.gen_hdlr, handle)
	r.gen_hdlr++
}

// OPTIONS is a shortcut for router.AddRoute("OPTIONS", path, r.gen_hdlr, handle)
func (r *MuxRouter) OPTIONS(path string, handle Handle) {
	r.AddRoute("OPTIONS", path, r.gen_hdlr, handle)
	r.gen_hdlr++
}

// HEAD is a shortcut for router.AddRoute("HEAD", path, r.gen_hdlr, handle)
func (r *MuxRouter) HEAD(path string, handle Handle) {
	r.AddRoute("HEAD", path, r.gen_hdlr, handle)
	r.gen_hdlr++
}

// CONNECT is a shortcut for router.AddRoute("CONNECT", path, r.gen_hdlr, handle)
func (r *MuxRouter) CONNECT(path string, handle Handle) {
	r.AddRoute("CONNECT", path, r.gen_hdlr, handle)
	r.gen_hdlr++
}

// TRACE is a shortcut for router.AddRoute("TRACE", path, r.gen_hdlr, handle)
func (r *MuxRouter) TRACE(path string, handle Handle) {
	r.AddRoute("TRACE", path, r.gen_hdlr, handle)
	r.gen_hdlr++
}

// ServeFiles serves files from the given file system root.
// The path must end with "/*filepath", files are then served from the local
// path /defined/root/dir/*filepath.
// For example if root is "/etc" and *filepath is "passwd", the local file
// "/etc/passwd" would be served.
// Internally a http.FileServer is used, therefore http.NotFound is used instead
// of the Router's NotFound handler.
// To use the operating system's file system implementation,
// use http.Dir:
//     router.ServeFiles("/src/*filepath", http.Dir("/var/www"))
func (r *MuxRouter) ServeFiles(path string, root http.FileSystem) {
	if len(path) < 10 || path[len(path)-10:] != "/*filepath" {
		panic("path must end with /*filepath")
	}

	fileServer := http.FileServer(root)

	r.GET(path, func(w http.ResponseWriter, req *http.Request, ps Params) {
		req.URL.Path = ps.ByName("filepath")
		// fmt.Printf("Path Requested: %s\n", ps.ByName("filepath"))
		fileServer.ServeHTTP(w, req)
	})
}

// Lookup allows the manual lookup of a method + path combo.
// This is e.g. useful to build a framework around this router.
func (r *MuxRouter) Lookup(method, path string) (Handle, Params, bool) {
	//if root := r.trees[method]; root != nil {
	//	return root.getValue(path)
	//}
	r.AllParam.NParam = 0
	for _, v := range r.routeData {
		if v.Route == path {
			return FxTab.Fx[v.NFxNo], r.AllParam, true
		}
	}
	return nil, r.AllParam, false
}

// -------------------------------------------------------------------------------------------------

/*
Old - Slower code, structured - see SplitOnSlash3
Benchmark: 148 ns - long, 28.3 ns short.

func (r *MuxRouter) SplitOnSlash2(m uint8, Url string) {
	r.CurUrl = Url
	ln := len(Url)
	NSl := 0
	h := int(m)
	j := 0
	im1 := 0
	for i := 1; i < ln && NSl < MaxSlashInUrl-1; i++ {
		if Url[im1] == '/' || j == nKeyChar {
			h += (h << 3)
			h = h ^ (h >> 11)
			h += (h << 15)
		}
		if Url[im1] == '/' {
			h += j
			h = ((h & bitMask) ^ ((h >> nBits) & bitMask))
			r.Hash[NSl] = h
			r.Slash[NSl] = im1
			NSl++
			h = int(m)
			j = 0
		}
		if j < nKeyChar {
			h += int(Url[i])
			h += (h << 10)
			h = h ^ (h >> 6)
		}
		j++
		im1++
	}
	r.Slash[NSl] = ln
	h += (h << 3)
	h = h ^ (h >> 11)
	h += (h << 15)
	h += j
	h = ((h & bitMask) ^ ((h >> nBits) & bitMask))
	r.Hash[NSl] = h
	r.NSl = NSl + 1
	return
}

*/

/*
// USED
// Test: SplitSlash3_test.go
// Benchmark: 118 for long, 21.1 for short.

PASS
BenchmarkOfSplitOnSlash3_long	20000000	       118 ns/op	       0 B/op	       0 allocs/op
BenchmarkOfSplitOnSlash3_short	100000000	        21.1 ns/op	       0 B/op	       0 allocs/op
*/

// xyzzy-hash
func (r *MuxRouter) SplitOnSlash3(m int, Url string, isUrl bool) {
	fixed := false
	var ln, NSl, h, wLen, i int
	var p, eem bool
	eem = true
s0:
	r.CurUrl = Url
	ln = len(Url)
	NSl = 0
	h = m
	wLen = 0
	i = 0
	r.Hash[NSl] = 0
	r.Slash[NSl] = 0
	NSl++
	p = true
s1:
	i++
	if i >= ln {
		goto s10
	}
	// if Url[i-1] == '/' && (Url[i] == '.' || Url[i] == '/') {
	if p && (Url[i] == '.' || Url[i] == '/') {
		Url, fixed = r.FixBadUrl(r.CurUrl)
		if fixed {
			goto s0
		}
	}
	// fmt.Printf("s0: i=%d url ->%s<- wLen=%d\n", i, Url[i:], wLen)
	if wLen >= nKeyChar {
		goto s2a
	}
	if Url[i] == '/' {
		h += wLen
		h += (h << 3)
		h = h ^ (h >> 11)
		h += (h << 15)
		r.Hash[NSl-1] = h
		r.Slash[NSl] = i
		NSl++
		if optionEarlyExit {
			if eem && isUrl {
				eem = false
				if r.Hash2Test[(h^m)&bitMask] == 0 {
					r.NSl = 1
					goto s11
				}
			}
			eem = false
		}
		if NSl >= MaxSlashInUrl-1 {
			goto s10
		}
		h = m
		wLen = 0
		p = true
		goto s1
	}
	h += int(Url[i])
	h += (h << 10)
	h = h ^ (h >> 6)
	wLen++
	p = false
	goto s1

s2:
	i++
	if i >= ln {
		goto s10
	}
s2a:
	if p && (Url[i] == '.' || Url[i] == '/') {
		Url, fixed = r.FixBadUrl(r.CurUrl)
		if fixed {
			goto s0
		}
	}
	// fmt.Printf("s2a: i=%d url ->%s<- wLen=%d\n", i, Url[i:], wLen)
	if Url[i] == '/' {
		h += wLen
		h += (h << 3)
		h = h ^ (h >> 11)
		h += (h << 15)
		r.Hash[NSl-1] = h
		r.Slash[NSl] = i
		NSl++
		if optionEarlyExit {
			if eem && isUrl {
				eem = false
				if r.Hash2Test[(h^m)&bitMask] == 0 {
					r.NSl = 1
					goto s11
				}
			}
			eem = false
		}
		if NSl >= MaxSlashInUrl-1 {
			goto s10
		}
		h = m
		wLen = 0
		p = true
		goto s1
	}
	wLen++
	p = false
	goto s2

s10:
	// fmt.Printf("s10: wLen=%d\n", wLen)
	if wLen > 0 {
		h += (h << 3)
		h = h ^ (h >> 11)
		h += (h << 15)
		h += wLen
		r.Hash[NSl-1] = h
		r.Slash[NSl] = ln
		r.NSl = NSl
	}
	//if false {
	//	fmt.Printf("Hash=%s Slash=%s NSl=%d\n", debug.SVar(r.Hash[0:r.NSl]), debug.SVar(r.Slash[0:r.NSl+1]), r.NSl)
	//}
s11:
	return
}

/*
-----------------------------------------------------------------------------------------------------

	+6. Add in ability to macth DOMAIN, Headers and Options - Protocal HTTP/https
		Use initial 'm' value - with a chain of calls like middleware
			func (r *xxx) Match ( req ) (new-m) { ... }
		Combine 'm' with some sort of XOR and Shift/Shift

	+5. Add Middleware
		 Json Body
		 XML Body
		 Parameters
		 Cookies
		 PrefixOutput
		 TokenAuth
		 InjectParams
		SendDataAsJSON
		SendDataAsXML
		SendDataAsHTML
		SendDataAsText

*/

// -------------------------------------------------------------------------------------------------
/*

type Middlware struct {
	data		interface{}
	whenToCall	int
	Init		func () ( interface{} )
	Before		func ( data interface{}, req, resp, param ) ( action )
	After		func ( data interface{}, req, resp, param ) ( action )
}

1. Should be able to keep data as in-memory struct and then have a "After" that will JSON-stringify the data.
2. Should be able to add a pre-fix/post-fix to this data.
3. Should be able to do timing of calls - logging - monetering
4. Should be able to extract params - in some order for (cookies|Headers|Get|POST(encode|JSON|XML)|Inject)
5. Should be able to handle AUTH in different forms
5. Should be able to return errors in diff forms - as status - as JSON etc.
5. Should be able to log errors in diff forms - as status - as JSON etc.
5. Should be able to use as a static file cache - early exit form call loop
5. Should be able to set/remove headers/cookies etc.

1. Must be able to attache to differen spots in proces
	At Top 'm', before 'm' (10,000)
	After 'm', before hash - to modify 'm' (10,002)
	After lookup , before params (10,003)
	Before/after call to handler func (10,004) (10,005)
	Before/after not found (10,006) (10,007)
	Before return (10,008)
1. Msut be really fast

*/
type Where int

const (
	HashNewM Where = iota
	Before         = iota
	After          = iota
)

type GoGoData struct {
	UserData []interface{}
}

type GoGoWidgetSet struct {
	w  Where
	fx GoGoWidgetFunc
}

// type GoGoWidgetFunc func(http.ResponseWriter, *http.Request, *Params, *GoGoData, int) int
type GoGoWidgetFunc func(*ApacheLogRecord, *http.Request, *Params, *GoGoData, int) int

type GoGoWidgetSetMatch struct {
	w  Where
	fx GoGoWidgetMatchFunc
}

type GoGoWidgetMatchFunc func(http.ResponseWriter, *http.Request, Params, *int, GoGoData, int, *[]string) bool

// Attach middlewhare widget to the handler.
func (r *MuxRouter) AttachWidget(w Where, fx GoGoWidgetFunc) {
	switch w {
	case HashNewM:
		r.widgetHashNewM = append(r.widgetHashNewM, GoGoWidgetSet{w: w, fx: fx})
	case Before:
		r.widgetBefore = append(r.widgetBefore, GoGoWidgetSet{w: w, fx: fx})
	case After:
		r.widgetAfter = append(r.widgetAfter, GoGoWidgetSet{w: w, fx: fx})
	}
}

// -------------------------------------------------------------------------------------------------
// Widget return values -
// Early exit from widget streem -
// success exit 200
// failure exit 404
// how do we create hash entries with the extra 311 in them

/*
// -------------------------------------------------------------------------------------------------
// req.TLS != nil => https
// xyzzy-widget
func HashTLSRequired(_ http.ResponseWriter, req *http.Request, _ Params, m *int, _ GoGoData, _ int, extra *[]string) {
	//if req.TLS != nil {
	//	(*m) += 311
	//}
}

func MatchTLSRequired(_ http.ResponseWriter, req *http.Request, _ Params, m *int, _ GoGoData, _ int, extra *[]string) bool {
	return req.TLS != nil
}

// xyzzy-widget
func HashHost(_ http.ResponseWriter, req *http.Request, _ Params, m *int, _ GoGoData, _ int, extra *[]string) {
	*extra = append(*extra, req.Host)
	//(*m) = HashString((*m), req.Host)
}

func MatchHost(_ http.ResponseWriter, req *http.Request, _ Params, m *int, _ GoGoData, _ int, extra *[]string) bool {
	return len(*extra) > 0 && (*extra)[0] == req.Host
}
*/
var disableOutput bool = false

type ApacheLogRecord struct {
	http.ResponseWriter

	startTime     time.Time
	status        int
	responseBytes int64
}

func (r *ApacheLogRecord) Write(p []byte) (written int, err error) {
	if disableOutput {
		written = len(string(p))
		r.responseBytes += int64(written)
		return written, nil
	}
	written, err = r.ResponseWriter.Write(p)
	r.responseBytes += int64(written)
	return written, err
}

func (r *ApacheLogRecord) WriteHeader(status int) {
	r.status = status
	if !disableOutput {
		r.ResponseWriter.WriteHeader(status)
	}
}

// -------------------------------------------------------------------------------------------------
func (r *MuxRouter) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	var m, status int
	m = 0
	status = 0
	_ = status
	var data GoGoData

	if r.PanicHandler != nil {
		defer r.recv(w, req)
	}
	if !r.HasBeenCompiled {
		r.CompileRoutes()
	}

	record := &ApacheLogRecord{
		ResponseWriter: w,
		status:         http.StatusOK,
	}

	r.AllParam.NParam = 0
	if r.widgetBefore != nil {
		for _, x := range r.widgetBefore {
			// xyzzy - problem with &m, - 8 bytes allocated, 50 ns used
			m = x.fx(record, req, &r.AllParam, &data, status)
		}
	}

	path := req.URL.Path
	Method := req.Method
	// fmt.Printf("%s at %s\n", Method, debug.LF())
	m = (int(Method[0]) + (int(Method[1]) << 1))

	if r.widgetHashNewM != nil {
		for _, x := range r.widgetHashNewM {
			m = x.fx(record, req, &r.AllParam, &data, status)
		}
	}

	r.SplitOnSlash3(m, path, true)
	found, ln, item := r.LookupUrlViaHash2(w, req, &m, data)
	if found {
		status = 200
		r.GetArgs3(path, item.ArgPattern, item.ArgNames, ln) // Get the Args if Any? - or defer?
		item.Fx(record, req, r.AllParam)
	} else {
		status = 404
		r.NotFound(w, req)
	}

	if r.widgetAfter != nil {
		for _, x := range r.widgetAfter {
			m = x.fx(record, req, &r.AllParam, &data, status)
		}
	}

	return
}

func (r *MuxRouter) ServeHTTP_old(w http.ResponseWriter, req *http.Request) {
	var data GoGoData
	if r.PanicHandler != nil {
		defer r.recv(w, req)
	}
	if !r.HasBeenCompiled {
		r.CompileRoutes()
	}

	path := req.URL.Path // 10,000
	Method := req.Method

	m := (int(Method[0]) + (int(Method[1]) << 1))
	r.AllParam.NParam = 0
	r.SplitOnSlash3(m, path, true)
	found, ln, item := r.LookupUrlViaHash2(w, req, &m, data)
	if found {
		r.GetArgs3(path, item.ArgPattern, item.ArgNames, ln) // Get the Args if Any? - or defer?
		item.Fx(w, req, r.AllParam)
	} else {
		// fmt.Printf("path=%s not found\n", path)
		r.NotFound(w, req)
	}

	return
}

// const dbLookupUrlMap = false
// const dbCompileRoutes2 = false
// const dbLookupUrlMap2 = false

// const dbHash2 = false // Print out info for Hash based multi lookup

/* vim: set noai ts=4 sw=4: */

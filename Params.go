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
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"strings"
	"sync"
	"time"

	debug "github.com/pschlump/godebug"
)

type FromType int

const (
	FromURL      FromType = iota
	FromParams            = iota
	FromCookie            = iota
	FromBody              = iota
	FromBodyJson          = iota
	FromInject            = iota
	FromHeader            = iota
	FromOther             = iota
	FromDefault           = iota
)

// Param is a single URL parameter, consisting of a key and a value.
type Param struct {
	Name  string
	Value string
	From  FromType
	Type  uint8
}

// Params is a Param-slice, as returned by the router.
// The slice is ordered, the first URL parameter is also the first slice value.
// It is therefore safe to read values by the index.
type Params struct {
	NParam       int
	Data         []Param
	route_i      int // What matched
	parent       *MuxRouter
	search       map[string]int
	search_ready bool
}

func FromTypeToString(ff FromType) string {
	switch ff {
	case FromURL:
		return "FromURL"
	case FromParams:
		return "FromParams"
	case FromCookie:
		return "FromCookie"
	case FromBody:
		return "FromBody"
	case FromBodyJson:
		return "FromBodyJson"
	case FromInject:
		return "FromInject"
	case FromHeader:
		return "FromHeader"
	case FromOther:
		return "FromOther"
	case FromDefault:
		return "FromDefault"
	default:
		return "Unknown-FromType"
	}
}

func (ps *Params) MakeStringMap(mdata map[string]string) {
	for _, v := range ps.Data {
		mdata[v.Name] = v.Value
	}
}

func (ps *Params) CreateSearch() {
	// fmt.Printf("CreateSearch - called\n")
	if ps.search_ready {
		return
	}

	for i, v := range ps.Data {
		ps.search[v.Name] = i
	}
	// fmt.Printf("CreateSearch - set to true\n")
	ps.search_ready = true
}

func (ps *Params) DumpParam() (rv string) {
	var Data2 []Param
	Data2 = ps.Data[0:ps.NParam]
	rv = debug.SVar(Data2)
	return
}

func (ps *Params) DumpParamDB() (rv string) {
	var Data2 []Param
	Data2 = ps.Data[0:ps.NParam]
	rv = debug.SVarI(Data2)
	return
}

// ByName returns the value of the first Param which key matches the given name.
// If no matching Param is found, an empty string is returned.
func (ps *Params) ByName(name string) (rv string) {
	rv = ""
	// xyzzy100 Change this to use a map[string]int - build maps on setup.
	// fmt.Printf("Looking For: %s, ps = %s\n", name, debug.SVarI(ps.Data[0:ps.NParam]))
	// fmt.Printf("ByName ------------------------\n")
	if ps.search_ready {
		// fmt.Printf("Is True ------------------------\n")
		if i, ok := ps.search[name]; ok {
			rv = ps.Data[i].Value
		}
		return
	}

	for i := 0; i < ps.NParam; i++ {
		if ps.Data[i].Name == name {
			rv = ps.Data[i].Value
			return
		}
	}
	return
}

func (ps *Params) GetByName(name string) (rv string, found bool) {
	rv = ""
	found = false
	// xyzzy100 Change this to use a map[string]int - build maps on setup.
	// fmt.Printf("Looking For: %s, ps = %s\n", name, debug.SVarI(ps.Data[0:ps.NParam]))
	// fmt.Printf("ByName ------------------------\n")
	if ps.search_ready {
		// fmt.Printf("Is True ------------------------\n")
		if i, ok := ps.search[name]; ok {
			rv = ps.Data[i].Value
			found = true
		}
		return
	}

	for i := 0; i < ps.NParam; i++ {
		if ps.Data[i].Name == name {
			rv = ps.Data[i].Value
			found = true
			return
		}
	}
	return
}

func (ps *Params) ByNameDflt(name string, dflt string) (rv string) {
	if ps.HasName(name) {
		return ps.ByName(name)
	}
	return dflt
}

func (ps *Params) HasName(name string) (rv bool) {
	rv = false
	if ps.search_ready {
		if _, ok := ps.search[name]; ok {
			rv = true
		}
		return
	}
	for i := 0; i < ps.NParam; i++ {
		if ps.Data[i].Name == name {
			rv = true
			return
		}
	}
	return
}

func (ps *Params) SetValue(name string, val string) {
	x := ps.PositionOf(name)
	if x >= 0 {
		ps.Data[x].Value = val
	}
}

func (ps *Params) PositionOf(name string) (rv int) {
	rv = -1
	for i := 0; i < ps.NParam; i++ {
		if ps.Data[i].Name == name {
			rv = i
			return
		}
	}
	return
}

func (ps *Params) ByPostion(pos int) (name string, val string, outRange bool) {
	// xyzzy101 Change this to use a map[string]int - build maps on setup.
	if pos >= 0 && pos < ps.NParam {
		return ps.Data[pos].Name, ps.Data[pos].Value, false
	}
	return "", "", true
}

// -------------------------------------------------------------------------------------------------
func AddValueToParams(Name string, Value string, Type uint8, From FromType, ps *Params) (k int) {
	ps.search_ready = false
	j := ps.PositionOf(Name)
	k = ps.NParam
	// db("AddValueToParams","j=%d k=%d %s\n", j, k, debug.LF())
	if j >= 0 {
		ps.Data[j].Value = Value
		ps.Data[j].Type = Type
		ps.Data[j].From = From
	} else {
		ps.Data[k].Value = Value
		ps.Data[k].Name = Name
		ps.Data[k].Type = Type
		ps.Data[k].From = From
		k++
		// xyzzy - check for more than MaxParams
	}
	ps.NParam = k
	// db(A"AddValueToParams","At end: NParam=%d %s\n", ps.NParam, debug.SVar(ps.Data[0:ps.NParam]))
	return
}

// -------------------------------------------------------------------------------------------------
func ParseBodyAsParams(w *MyResponseWriter, req *http.Request, ps *Params) int {

	ct := req.Header.Get("Content-Type")
	if req.Method == "POST" || req.Method == "PUT" || req.Method == "PATCH" || req.Method == "DELETE" {
		if req.PostForm == nil {
			if strings.HasPrefix(ct, "application/json") {
				body, err2 := ioutil.ReadAll(req.Body)
				if err2 != nil {
					fmt.Printf("Error(20008): Malformed JSON body, RequestURI=%s err=%v\n", req.RequestURI, err2)
				}
				var jsonData map[string]interface{}
				err := json.Unmarshal(body, &jsonData)
				if err == nil {
					for Name, v := range jsonData {
						Value := ""
						switch v.(type) {
						case bool:
							Value = fmt.Sprintf("%v", v)
						case float64:
							Value = fmt.Sprintf("%v", v)
						case int64:
							Value = fmt.Sprintf("%v", v)
						case int32:
							Value = fmt.Sprintf("%v", v)
						case time.Time:
							Value = fmt.Sprintf("%v", v)
						case string:
							Value = fmt.Sprintf("%v", v)
						default:
							Value = fmt.Sprintf("%s", debug.SVar(v))
						}
						AddValueToParams(Name, Value, 'b', FromBodyJson, ps)
					}
				}
			} else {
				err := req.ParseForm()
				if err != nil {
					fmt.Printf("Error(20010): Malformed body, RequestURI=%s err=%v\n", req.RequestURI, err)
				} else {
					for Name, v := range req.PostForm {
						if len(v) > 0 {
							AddValueToParams(Name, v[0], 'b', FromBody, ps)
						}
					}
				}
			}
		} else {
			for Name, v := range req.PostForm {
				if len(v) > 0 {
					AddValueToParams(Name, v[0], 'b', FromBody, ps)
				}
			}
		}
	}
	return 0
}

// -------------------------------------------------------------------------------------------------
func ParseCookiesAsParams(w *MyResponseWriter, req *http.Request, ps *Params) int {

	Ck := req.Cookies()
	for _, v := range Ck {
		AddValueToParams(v.Name, v.Value, 'c', FromCookie, ps)
	}
	return 0
}

// -------------------------------------------------------------------------------------------------
var ApacheLogFile *os.File

const ApacheFormatPattern = "%s %v %s %s %s %v %d %v\n"

const benchmar = false

/*
func itoaPos(n int, buffer *bytes.Buffer, padLen int, pad uint8) {
	i := 0
	var s [10]uint8
	for {
		s[i] = uint8((n % 10) + '0')
		i++
		n /= 10
		if n == 0 {
			break
		}
	}
	for ; i < padLen; i++ {
		s[i] = pad
	}

	for j := i - 1; j >= 0; j-- {
		buffer.WriteByte(s[j])
	}
}
*/
var mutexCurTime = &sync.Mutex{}
var timeFormatted string

func SetCurTime(s string) {
	mutexCurTime.Lock()
	timeFormatted = s
	mutexCurTime.Unlock()
}

func GetCurTime() (s string) {
	mutexCurTime.Lock()
	s = timeFormatted
	mutexCurTime.Unlock()
	return
}

func init() {
	// Once a sec update of time formated string
	onceASec := time.NewTicker(1 * time.Second)
	quit := make(chan struct{})
	go func() {
		for {
			select {
			case <-onceASec.C:
				// do stuff
				finishTime := time.Now()
				curTimeUTC := finishTime.UTC()
				SetCurTime(curTimeUTC.Format("02/Jan/2006 03:04:05"))
				// fmt.Printf("cur-time: %s\n", GetCurTime())
			case <-quit:
				onceASec.Stop()
				return
			}
		}
	}()
}

func ApacheLogingBefore(w *MyResponseWriter, req *http.Request, ps *Params) int {
	if ApacheLogFile == nil {
		return 0
	}
	w.StartTime = time.Now()
	return 0
}

func ApacheLogingAfter(w *MyResponseWriter, req *http.Request, ps *Params) int {
	if ApacheLogFile == nil {
		return 0
	}
	ip := req.RemoteAddr
	if colon := strings.LastIndex(ip, ":"); colon != -1 {
		ip = ip[:colon]
	}

	finishTime := time.Now()
	elapsedTime := finishTime.Sub(w.StartTime)
	// The next line is a real problem.  Taking 450us to convert a data is just ICKY!  I have a new
	// version of Format that reduces this to about 300us.  What is needed is a real fast formatting
	// tool for dates that reduces this to a reasonable 30us.
	// Sadly enough - by not exposing the interals of the time.Time type - fixing this will require
	// a major re-write of the entire time type.   That is oging to take some days to do.
	var timeFormatted string

	//
	// OLD: finishTimeUTC := finishTime.UTC()
	// OLD: timeFormatted = finishTimeUTC.Format("02/Jan/2006 03:04:05") // 450+ us to do a time format and 1 alloc
	//
	// This entire thing could be replaced with a goroutein that runs 1ce a second, and makes a
	// variable with the date-time in it.  that would be one "Format" per second on a different
	// thread.  Then just use a lock, unlock process to access the variable - simple enough.
	//
	timeFormatted = GetCurTime()

	fmt.Fprintf(ApacheLogFile, ApacheFormatPattern, ip, timeFormatted, req.Method, req.RequestURI, req.Proto, w.Status,
		w.ResponseBytes, elapsedTime.Seconds())

	return 0
}

// -------------------------------------------------------------------------------------------------
func ParseQueryParams(w *MyResponseWriter, req *http.Request, ps *Params) int {
	// u, err := url.ParseRequestURI(req.RequestURI)
	if req.URL.RawQuery == "" {
		return 0
	}
	m, err := url.ParseQuery(req.URL.RawQuery)
	// db("ParseQueryParams","Parsing Raw Query ->%s<-, m=%s\n", req.URL.RawQuery, debug.SVar(m))
	if err != nil {
		fmt.Printf("Unable to parse URL query, %s\n", err)
	}
	for Name, v := range m {
		vv := ""
		if len(v) == 1 {
			vv = v[0]
		} else {
			vv = debug.SVar(v)
		}
		AddValueToParams(Name, vv, 'q', FromParams, ps)
	}
	return 0
}

// -------------------------------------------------------------------------------------------------
func PrefixWith(w *MyResponseWriter, req *http.Request, ps *Params) int {
	// Prefix for AngularJS
	w.Write([]byte(")]}"))
	// Other Common Prefixes are:
	//		while(1);
	//		for(;;);
	//		//							Comment
	//		while(true);
	return 0
}

// -------------------------------------------------------------------------------------------------
func MethodParam(w *MyResponseWriter, req *http.Request, ps *Params) int {
	//fmt.Printf("MethodParam\n")
	//fmt.Printf("%s\n", debug.LF())
	if ps.HasName("METHOD") {
		//fmt.Printf("%s\n", debug.LF())
		x := ps.ByName("METHOD")
		if b, ok := validMethod[x]; ok && b {
			//fmt.Printf("%s\n", debug.LF())
			req.Method = ps.ByName("METHOD")
		}
	}
	return 0
}

/*
	HTTP/1.1 401 Unauthorized
	{
		"status": "Error"
		, "msg": "No access token provided."
		, "code": "10002"
		, "details": "bla bla bla"
	}
req.Header.Add("If-None-Match", `W/"wyzzy"`)
https://developer.github.com/guides/traversing-with-pagination/

req.Header.Add("Link", `W/"wyzzy"`)
Link: <https://api.github.com/search/code?q=addClass+user%3Amozilla&page=2>; rel="next",
  <https://api.github.com/search/code?q=addClass+user%3Amozilla&page=34>; rel="last"
*/
/* vim: set noai ts=4 sw=4: */

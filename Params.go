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
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"

	"./debug"
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

func (ps Params) CreateSearch() {
	if ps.search_ready {
		return
	}

	for i, v := range ps.Data {
		ps.search[v.Name] = i
	}
	ps.search_ready = true
}

func (ps Params) DumpParam() (rv string) {
	var Data2 []Param
	Data2 = ps.Data[0:ps.NParam]
	rv = debug.SVar(Data2)
	return
}

// ByName returns the value of the first Param which key matches the given name.
// If no matching Param is found, an empty string is returned.
func (ps Params) ByName(name string) (rv string) {
	rv = ""
	// xyzzy100 Change this to use a map[string]int - build maps on setup.
	// fmt.Printf("Looking For: %s, ps = %s\n", name, debug.SVarI(ps.Data[0:ps.NParam]))
	if ps.search_ready {
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

func (ps Params) HasName(name string) (rv bool) {
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

func (ps Params) PositionOf(name string) (rv int) {
	rv = -1
	for i := 0; i < ps.NParam; i++ {
		if ps.Data[i].Name == name {
			rv = i
			return
		}
	}
	return
}

func (ps Params) ByPostion(pos int) (name string, val string, outRange bool) {
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
var fo *os.File

const ApacheFormatPattern = "%s %v %s %s %s %v %d %v\n"

var benchmar = false

var shortMonthNames = []string{
	"---",
	"Jan",
	"Feb",
	"Mar",
	"Apr",
	"May",
	"Jun",
	"Jul",
	"Aug",
	"Sep",
	"Oct",
	"Nov",
	"Dec",
}

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

func ApacheLogingBefore(w *MyResponseWriter, req *http.Request, ps *Params) int {
	if fo == nil {
		return 0
	}
	w.startTime = time.Now()
	return 0
}

func ApacheLogingAfter(w *MyResponseWriter, req *http.Request, ps *Params) int {
	if fo == nil {
		return 0
	}
	ip := req.RemoteAddr
	if colon := strings.LastIndex(ip, ":"); colon != -1 {
		ip = ip[:colon]
	}

	finishTime := time.Now()
	finishTimeUTC := finishTime.UTC()
	elapsedTime := finishTime.Sub(w.startTime)
	// The next line is a real problem.  Taking 450us to convert a data is just ICKY!  I have a new
	// version of Format that reduces this to about 300us.  What is needed is a real fast formatting
	// tool for dates that reduces this to a reasonable 30us.
	// Sadly enough - by not exposing the interals of the time.Time type - fixing this will require
	// a major re-write of the entire time type.   That is oging to take some days to do.
	var timeFormatted string
	if false {
		timeFormatted = finishTimeUTC.Format("02/Jan/2006 03:04:05") // 450+ us to do a time format and 1 alloc
	} else {
		y, m, d := finishTimeUTC.Date() // These funcs (Year,Month,Day,Clock) take 245 us just to extract info
		H, M, S := finishTimeUTC.Clock()
		if false {
			// This silly thing takes 1245 us to convert/format a string.
			timeFormatted = fmt.Sprintf("%02d/%3s/%04d %02d:%02d:%02d", d, shortMonthNames[m], y, H, M, S)
		} else {
			timeFormatted = ""
		}
	}

	if !benchmar {
		fmt.Fprintf(fo, ApacheFormatPattern, ip, timeFormatted, req.Method, req.RequestURI, req.Proto, w.status,
			w.responseBytes, elapsedTime.Seconds())
	}
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

/* vim: set noai ts=4 sw=4: */

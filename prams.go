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

// xyzzy202 - need a test - need to do someting with {} patterns - need to deal with {name} patterns

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
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
	NParam int
	Data   []Param
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
	// xyzzy100 Change this to use a map[string]int - build maps on setup.
	// fmt.Printf("Looking For: %s, ps = %s\n", name, debug.SVarI(ps.Data[0:ps.NParam]))

	rv = ""
	for i := 0; i < ps.NParam; i++ {
		if ps.Data[i].Name == name {
			rv = ps.Data[i].Value
			goto done
		}
	}
done:
	//	db("dbByName","ByName (%s) = %s\n", name, rv)
	return
}

func (ps Params) HasName(name string) (rv bool) {
	rv = false
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
func ParseBodyAsParams(w *ApacheLogRecord, req *http.Request, ps *Params, _ *GoGoData, _ int) int {

	ct := req.Header.Get("Content-Type")
	if req.Method == "POST" || req.Method == "PUT" || req.Method == "PATCH" || req.Method == "DELETE" {
		if req.PostForm == nil {
			if strings.HasPrefix(ct, "application/json") {
				body, err2 := ioutil.ReadAll(req.Body)
				if err2 != nil {
					fmt.Printf("Error(90000): Malformed JSON body, RequestURI=%s err=%v\n", req.RequestURI, err2)
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
					fmt.Printf("Error(90001): Malformed body, RequestURI=%s err=%v\n", req.RequestURI, err)
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
func ParseCookiesAsParams(w *ApacheLogRecord, req *http.Request, ps *Params, _ *GoGoData, _ int) int {

	Ck := req.Cookies()
	for _, v := range Ck {
		AddValueToParams(v.Name, v.Value, 'c', FromCookie, ps)
	}
	return 0
}

// -------------------------------------------------------------------------------------------------
const ApacheFormatPattern = "%s %v %s %s %s %v %d %v\n"

func ApacheLogingBefore(w *ApacheLogRecord, req *http.Request, ps *Params, data *GoGoData, status int) int {
	w.startTime = time.Now()
	return 0
}

func ApacheLogingAfter(w *ApacheLogRecord, req *http.Request, ps *Params, data *GoGoData, status int) int {
	ip := req.RemoteAddr
	if colon := strings.LastIndex(ip, ":"); colon != -1 {
		ip = ip[:colon]
	}

	finishTime := time.Now()
	finishTimeUTC := finishTime.UTC()
	elapsedTime := finishTime.Sub(w.startTime)
	timeFormatted := finishTimeUTC.Format("02/Jan/2006 03:04:05")
	fmt.Printf(ApacheFormatPattern, ip, timeFormatted, req.Method, req.RequestURI, req.Proto, w.status,
		w.responseBytes, elapsedTime.Seconds())
	return 0
}

// -------------------------------------------------------------------------------------------------
const SimpleLogingData = 0

type SimpleLoging struct {
	tag int
}

func SimpleLogingBefore(w *ApacheLogRecord, req *http.Request, ps *Params, data *GoGoData, status int) int {
	if data.UserData == nil {
		// fmt.Printf("Created, %s\n", debug.LF())
		data.UserData = make([]interface{}, 10, 10)
	}
	if data.UserData[SimpleLogingData] == nil {
		// fmt.Printf("Added at %d, %s\n", SimpleLogingData, debug.LF())
		data.UserData[SimpleLogingData] = &SimpleLoging{tag: 12}
	}
	// fmt.Printf("Before at %d, %s\n", SimpleLogingData, debug.LF())
	t := ((data.UserData[SimpleLogingData]).(*SimpleLoging))
	// fmt.Printf("After at %d, %s\n", SimpleLogingData, debug.LF())
	t.tag = 22
	data.UserData[SimpleLogingData] = t
	return 0
}
func SimpleLogingAfter(w *ApacheLogRecord, req *http.Request, ps *Params, data *GoGoData, status int) int {
	t := ((data.UserData[SimpleLogingData]).(*SimpleLoging))
	_ = t
	// fmt.Printf("data[] = %d\n", t.tag)
	return 0
}

// -------------------------------------------------------------------------------------------------
func ParseQueryParams(w *ApacheLogRecord, req *http.Request, ps *Params, _ *GoGoData, _ int) int {
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
func PrefixWith(w *ApacheLogRecord, req *http.Request, ps *Params, _ *GoGoData, _ int) int {
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
func MethodParam(w *ApacheLogRecord, req *http.Request, ps *Params, _ *GoGoData, _ int) int {
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

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
	"fmt"
	"net/http"

	"./debug"
)

/*
type FromType int

const (
	FromURL    FromType = iota
	FromParams          = iota
	FromCookie          = iota
	FromPost            = iota
	FromInject          = iota
	FromHeader          = iota
	FromOther           = iota
)
*/

const MaxParams = 200

func ParseParameters(w http.ResponseWriter, req *http.Request, ps Params, _ *int, _ GoGoData, _ int, extra *[]string) {
	db("ParseParameters", "Arrived at ParseParameters, %s", debug.LF())
	k := ps.NParam
	if k+1 >= MaxParams {
		fmt.Printf("Error(): Too Maray Params, URL=%s\n", req.RequestURI)
		return
	}

	ps.Data[k].Value = debug.LF()
	ps.Data[k].Name = "$ParseParams$"
	ps.Data[k].Type = 'i'
	ps.Data[k].From = FromInject
	k++

	ps.NParam = k
}

/*
func UriToStringMap(req *http.Request) (m url.Values, fr map[string]string) {

	ct := req.Header.Get("Content-Type")

	fr = make(map[string]string)

	// db_uriToString := false

	// if ( db_uriToString ) { fmt.Printf ( "PJS Apr 9: %s Content Type:%v\n", tr.LF(), ct ) }
	if db_uriToString {
		fmt.Printf("PJS Sep 20: %s Content Type:%v\n", tr.LF(), ct)
	}

	u, _ := url.ParseRequestURI(req.RequestURI)
	m, _ = url.ParseQuery(u.RawQuery)
	for i := range m {
		fr[i] = "Query-String"
	}

	// xyzzy - add in cookies??		req.Cookies() -> []string
	// if ( db_uriToString ) { fmt.Printf ( "Cookies are: %s\n", sizlib.SVar( req.Cookies() ) ) }
	Ck := req.Cookies()
	for _, v := range Ck {
		if _, ok := m[v.Name]; !ok {
			m[v.Name] = make([]string, 1)
			m[v.Name][0] = v.Value
			// fmt.Printf ( "Name=%s Value=%s\n", v.Name, v.Value )
			fr[v.Name] = "Cookie"
		}
	}

	// fmt.Printf ( "Checking to see if post\n" )

	// add in POST parmeters
	if req.Method == "POST" || req.Method == "PUT" || req.Method == "PATCH" || req.Method == "DELETE" {
		if db_uriToString {
			fmt.Printf("It's a POST/PUT/PATCH/DELETE, req.PostForm=%v, ct=%s\n", req.PostForm, ct)
		}
		if req.PostForm == nil {
			// if ( db_uriToString ) { fmt.Printf ( "ParseForm has !!!not!!! been  called\n" ) }
			if strings.HasPrefix(ct, "application/json") {
				body, err2 := ioutil.ReadAll(req.Body)
				if err2 != nil {
					fmt.Printf("err=%v\n", err2)
				}
				// if ( db_uriToString) { fmt.Printf("body=%s\n",string(body)) }
				// fmt.Printf("request body=%s\n",string(body))
				var jsonData map[string]interface{}
				err := json.Unmarshal(body, &jsonData)
				if err == nil {
					for i, v := range jsonData {
						m[i] = make([]string, 1)
						switch v.(type) {
						case bool:
							m[i][0] = fmt.Sprintf("%v", v)
						case float64:
							m[i][0] = fmt.Sprintf("%v", v)
						case int64:
							m[i][0] = fmt.Sprintf("%v", v)
						case int32:
							m[i][0] = fmt.Sprintf("%v", v)
						case time.Time:
							m[i][0] = fmt.Sprintf("%v", v)
						case string:
							m[i][0] = fmt.Sprintf("%v", v)
						default:
							m[i][0] = fmt.Sprintf("%s", SVar(v))
						}
						fr[i] = "JSON-Encoded-Body/Post"
					}
				}
			} else {
				err := req.ParseForm()
				if db_uriToString {
					fmt.Printf("Form data is now: %s\n", SVar(req.PostForm))
				}
				if err != nil {
					fmt.Printf("Error - parse form just threw an error , why? %v\n", err)
				} else {
					for i, v := range req.PostForm {
						if len(v) > 0 {
							m[i] = make([]string, 1)
							m[i][0] = v[0]
							fr[i] = "URL-Encoded-Body(1-a)/Post"
						}
					}
				}
			}
		} else {
			for i, v := range req.PostForm {
				if len(v) > 0 {
					m[i] = make([]string, 1)
					m[i][0] = v[0]
					fr[i] = "URL-Encoded-Body(2)/Post"
				}
			}
		}
	}

	if db_uriToString {
		fmt.Printf(">>m=%s\n", SVar(m))
	}

	return
}
*/

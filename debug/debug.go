package debug

// ----------------------------------------------------------------------------------------------------------
//
// Simple functions to help with debuging GO code.
//
// (C) Philip Schlump, 2013-2014.
// Version: 1.0.0
// BuildNo: 060
//
// I usually use these like this:
//
//     func someting ( j int ) {
//			...
//			fmt.Pritnf ( "Ya someting useful %s\n", debug.LF(1) )
//
// This prints out the line and file that "Ya..." is at - so that it is easier for me to match output
// with code.   The "depth" == 1 parameter is how far up the stack I want to go.  0 is the LF routine.
// 1 is the caller of LF, usually what I want and the default, 2 is the caller of "something".
//
// The most useful functions are:
//    LF 			Return as a string the line number and file name.
//	  IAmAt			Print out current line/file
//	  SVarI			Convert most things to an indented JSON string and return it.
//
// To Include put these fiels in ./debug and in your code
//
//		import (
//			"./degug"
//		)
//
// Then
//
//		fmt.Printf ( ".... %s ...\n", debug.LF() )
//
// ----------------------------------------------------------------------------------------------------------

import (
	"encoding/json"
	"fmt"
	"runtime"
	"strings"
)

// Return the current line number as a string.  Default parameter is 1, must be an integer
// That reflectgs the depth in the call stack.  A value of 0 would be the LINE() function
// itself.  If you suply more than one parameter, 2..n are ignored.
func LINE(d ...int) string {
	depth := 1
	if len(d) > 0 {
		depth = d[0]
	}
	_, _, line, ok := runtime.Caller(depth)
	if ok {
		return fmt.Sprintf("%d", line)
	} else {
		return "LineNo:Unk"
	}
}

// Return the current file name.
func FILE(d ...int) string {
	depth := 1
	if len(d) > 0 {
		depth = d[0]
	}
	_, file, _, ok := runtime.Caller(depth)
	if ok {
		return file
	} else {
		return "File:Unk"
	}
}

// Return the File name and Line no as a string.
func LF(d ...int) string {
	depth := 1
	if len(d) > 0 {
		depth = d[0]
	}
	_, file, line, ok := runtime.Caller(depth)
	if ok {
		return fmt.Sprintf("File: %s LineNo:%d", file, line)
	} else {
		return fmt.Sprintf("File: Unk LineNo:Unk")
	}
}

// Return the File name and Line no as a string. - for JSON as string
func LFj(d ...int) string {
	depth := 1
	if len(d) > 0 {
		depth = d[0]
	}
	_, file, line, ok := runtime.Caller(depth)
	if ok {
		return fmt.Sprintf("\"File\": \"%s\", \"LineNo\":%d", file, line)
	} else {
		return ""
	}
}

// Return the current funciton name as a string.
func FUNCNAME(d ...int) string {
	depth := 1
	if len(d) > 0 {
		depth = d[0]
	}
	pc, _, _, ok := runtime.Caller(depth)
	if ok {
		xfunc := runtime.FuncForPC(pc).Name()
		return xfunc
	} else {
		return fmt.Sprintf("FunctionName:Unk")
	}
}

// Print out the current Function,File,Line No and an optional set of strings.
func IAmAt(s ...string) {
	pc, file, line, ok := runtime.Caller(1)
	if ok {
		xfunc := runtime.FuncForPC(pc).Name()
		fmt.Printf("Func:%s File:%s LineNo:%d, %s\n", xfunc, file, line, strings.Join(s, " "))
	} else {
		fmt.Printf("Func:Unk File:Unk LineNo:Unk, %s\n", strings.Join(s, " "))
	}
}

// Print out the current Function,File,Line No and an optional set of strings - do this for 2 levels deep.
func IAmAt2(s ...string) {
	pc, file, line, ok := runtime.Caller(1)
	pc2, file2, line2, ok2 := runtime.Caller(2)
	if ok {
		xfunc := runtime.FuncForPC(pc).Name()
		if ok2 {
			xfunc2 := runtime.FuncForPC(pc2).Name()
			fmt.Printf("Func:%s File: %s LineNo:%d, called...\n", xfunc2, file2, line2)
		} else {
			fmt.Printf("Func:Unk File: unk LineNo:unk, called...\n")
		}
		fmt.Printf("Func:%s File: %s LineNo:%d, %s\n", xfunc, file, line, strings.Join(s, " "))
	} else {
		fmt.Printf("Func:Unk File: Unk LineNo:Unk, %s\n", strings.Join(s, " "))
	}
}

// -------------------------------------------------------------------------------------------------
func SVar(v interface{}) string {
	s, err := json.Marshal(v)
	// s, err := json.MarshalIndent ( v, "", "\t" )
	if err != nil {
		return fmt.Sprintf("Error:%s", err)
	} else {
		return string(s)
	}
}

// -------------------------------------------------------------------------------------------------
func SVarI(v interface{}) string {
	// s, err := json.Marshal ( v )
	s, err := json.MarshalIndent(v, "", "\t")
	if err != nil {
		return fmt.Sprintf("Error:%s", err)
	} else {
		return string(s)
	}
}

// -------------------------------------------------------------------------------------------------
func InArrayString(s string, arr []string) int {
	for i, v := range arr {
		if v == s {
			return i
		}
	}
	return -1
}

func InArrayInt(s int, arr []int) int {
	for i, v := range arr {
		if v == s {
			return i
		}
	}
	return -1
}

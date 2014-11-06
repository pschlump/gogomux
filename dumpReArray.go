package gogomux

import "fmt"

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

package lazyfunc

import (
	"encoding/json"
	"fmt"
	"go/format"
	"reflect"
	"strings"

	"github.com/Sudalight/tools/pkg/convert"
)

// WriteUnmaintainableCode takes raw json string as parameter,
// parses the parameter into map[string]interface{} recursively,
// prints a variable named `unmaintainableCode` with the paramater's content in it.
func WriteUnmaintainableCode(rawJSON string) {
	unmaintainableCode := make(map[string]interface{})
	err := json.Unmarshal(convert.StringToBytes(rawJSON), &unmaintainableCode)
	if err != nil {
		panic(err)
	}
	res := fmt.Sprintf("unmaintainableCode := map[string]interface{}{\n%s}", recurse(unmaintainableCode))

	src, err := format.Source(convert.StringToBytes(res))
	if err != nil {
		panic(err)
	}

	fmt.Println(strings.Replace(convert.BytesToString(src), "\t", "    ", -1))
}

func recurse(a map[string]interface{}) string {
	res := ""
	for k, v := range a {
		switch reflect.TypeOf(v).String() {
		case "map[string]interface {}":
			res += fmt.Sprintf("\"%s\":map[string]interface{}{\n", k)
			res += recurse(v.(map[string]interface{}))
			res += "},\n"
		case "[]interface {}":
			res += fmt.Sprintf("\"%s\":[]interface{}{\n", k)
			ls := v.([]interface{})
			for i := range ls {
				switch reflect.TypeOf(ls[i]).String() {
				case "map[string]interface {}":
					res += "map[string]interface{}{\n"
					res += recurse(ls[i].(map[string]interface{}))
					res += "},\n"
				default:
					res += fmt.Sprintf("\"%v\",// TODO: you may need to replace values here\n", ls[i])
				}
			}
			res += "},\n"
		default:
			res += fmt.Sprintf("\"%s\":\"%v\",// TODO: you may need to replace values here\n", k, v)
		}
	}
	return res
}

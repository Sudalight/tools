package diff

import (
	"encoding/json"
	"fmt"
	"reflect"
	"strconv"
)

const (
	interfaceSlice = "[]interface {}"
	object         = "map[string]interface {}"
)

type Changelog struct {
	After       interface{} `json:"after"`
	Before      interface{} `json:"before"`
	OperatorID  string      `json:"operator_id"`
	Path        string      `json:"path"`
	ProcessorID string      `json:"processor_id"`
	SnapshotID  string      `json:"snapshot_id"`
}

func makeChangelog(path string, before, after interface{}) Changelog {
	return Changelog{
		Path:   path,
		Before: before,
		After:  after,
	}
}

func Diff(be, af []byte, snapshotID, processorID, operatorID string) ([]Changelog, error) {
	beMap := make(map[string]interface{})
	afMap := make(map[string]interface{})
	err := json.Unmarshal(be, &beMap)
	if err != nil {
		return nil, err
	}
	err = json.Unmarshal(af, &afMap)
	if err != nil {
		return nil, err
	}

	res, err := recurseMap(beMap, afMap, "")
	if err != nil {
		return nil, err
	}

	for i := range res {
		res[i].OperatorID = operatorID
		res[i].ProcessorID = processorID
		res[i].SnapshotID = snapshotID
	}
	return res, nil
}

func recurseMap(be, af map[string]interface{}, pathPrefix string) ([]Changelog, error) {
	res := make([]Changelog, 0)
	if pathPrefix != "" {
		pathPrefix += "."
	}

	for k, beV := range be {
		_, ok := af[k]
		if ok {
			continue
		}
		res = append(res, makeChangelog(pathPrefix+k, beV, nil))
	}

	for k, afV := range af {
		beV, ok := be[k]
		if !ok {
			res = append(res, makeChangelog(pathPrefix+k, nil, afV))
			continue
		}

		if reflect.TypeOf(afV).String() != reflect.TypeOf(beV).String() {
			res = append(res, makeChangelog(pathPrefix+k, beV, afV))
			continue
		}

		switch reflect.TypeOf(afV).String() {
		case object:
			tmp, err := recurseMap(beV.(map[string]interface{}), afV.(map[string]interface{}), pathPrefix+k)
			if err != nil {
				return nil, err
			}
			res = append(res, tmp...)
		case interfaceSlice:
			tmp, err := recurseSlice(beV.([]interface{}), afV.([]interface{}), pathPrefix+k)
			if err != nil {
				return nil, err
			}
			res = append(res, tmp...)
		default:
			if !reflect.DeepEqual(afV, beV) {
				res = append(res, makeChangelog(pathPrefix+k, beV, afV))
			}
		}
	}

	return res, nil
}

// slice 中的元素类型必须一致
func recurseSlice(be, af []interface{}, pathPrefix string) ([]Changelog, error) {
	if len(be) == 0 && len(af) == 0 {
		return nil, nil
	}

	res := make([]Changelog, 0)
	var ele interface{}
	if len(be) == 0 {
		ele = af[0]
	} else {
		ele = be[0]
	}

	switch reflect.TypeOf(ele).String() {
	case object:
		for i := range af {
			afType := reflect.TypeOf(af[i]).String()
			if afType != object {
				return res, fmt.Errorf("type of %s in new document is %s, expected %s",
					pathPrefix+strconv.Itoa(i), afType, object)
			}
			if i > len(be)-1 {
				res = append(res, makeChangelog(pathPrefix+"["+strconv.Itoa(i)+"]", nil, af[i]))
				continue
			}

			beType := reflect.TypeOf(be[i]).String()
			if beType != object {
				return res, fmt.Errorf("type of %s in old document is %s, expected %s",
					pathPrefix+strconv.Itoa(i), beType, object)
			}
			tmp, err := recurseMap(be[i].(map[string]interface{}), af[i].(map[string]interface{}),
				pathPrefix+"["+strconv.Itoa(i)+"]")
			if err != nil {
				return nil, err
			}
			res = append(res, tmp...)
		}
	case interfaceSlice:
		for i := range af {
			afType := reflect.TypeOf(af[i]).String()
			if afType != interfaceSlice {
				return res, fmt.Errorf("type of %s in new document is %s, expected %s",
					pathPrefix+strconv.Itoa(i), afType, interfaceSlice)
			}
			if i > len(be)-1 {
				res = append(res, makeChangelog(pathPrefix+"["+strconv.Itoa(i)+"]", nil, af[i]))
				continue
			}

			beType := reflect.TypeOf(be[i]).String()
			if beType != interfaceSlice {
				return res, fmt.Errorf("type of %s in old document is %s, expected %s",
					pathPrefix+strconv.Itoa(i), beType, interfaceSlice)
			}
			tmp, err := recurseSlice(be[i].([]interface{}), af[i].([]interface{}), pathPrefix+"["+strconv.Itoa(i)+"]")
			if err != nil {
				return nil, err
			}
			res = append(res, tmp...)
		}
	default:
		if !reflect.DeepEqual(be, af) {
			res = append(res, makeChangelog(pathPrefix, be, af))
		}
	}
	return res, nil
}

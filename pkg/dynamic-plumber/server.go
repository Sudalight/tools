package dynamicplumber

import (
	"log"
	"plugin"

	"gorm.io/gorm"
)

type FillFunc func(obj map[string]interface{}, db *gorm.DB) error

func LoadPlugin(name string) error {
	p, err := plugin.Open(name)
	if err != nil {
		return err
	}
	fillFunc, err := p.Lookup("Fill")
	if err != nil {
		return err
	}

	db := new(gorm.DB)
	a := make(map[string]interface{})
	(*fillFunc.(*FillFunc))(a, db)
	log.Println(a)
	return nil
}

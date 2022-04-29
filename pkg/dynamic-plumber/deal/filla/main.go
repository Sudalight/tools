package main

import (
	dynamicplumber "github.com/Sudalight/tools/pkg/dynamic-plumber"
	"gorm.io/gorm"
)

var Fill dynamicplumber.FillFunc = func(obj map[string]interface{}, db *gorm.DB) error {
	obj["123"] = 2
	return nil
}

package mwms_core

import (
	"fmt"
	"github.com/mlplabs/mwms-core/whs"
)

func Version() {
	fmt.Println("Version 1.0.0")
}

func GetStorage() *whs.Storage {
	storage := new(whs.Storage)
	return storage
}

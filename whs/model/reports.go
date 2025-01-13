package model

import "github.com/mlplabs/mwms-core/whs/cells"

type RowStock struct {
	Product  Product      `json:"product"`
	Zone     Zone         `json:"zone"`
	Cells    []cells.Cell `json:"cells"`
	Quantity int64        `json:"quantity"`
}

type StockData struct {
	Rows []RowStock `json:"rows"`
}

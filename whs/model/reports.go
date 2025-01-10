package model

import "github.com/mlplabs/mwms-core/whs/cells"

type RowStock struct {
	RowId    string       `json:"row_id"`
	Product  Product      `json:"product"`
	Quantity int          `json:"quantity"`
	Zone     Zone         `json:"zone"`
	Cells    []cells.Cell `json:"cells"`
}

type StockData struct {
	Rows []RowStock `json:"rows"`
}

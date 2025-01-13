package model

type RowStock struct {
	Product  Product `json:"product"`
	Zone     Zone    `json:"zone"`
	Cells    []Cell  `json:"cells"`
	Quantity int64   `json:"quantity"`
}

type StockData struct {
	Rows []RowStock `json:"rows"`
}

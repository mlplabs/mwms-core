package model

type RowStorage struct {
	RowId    string  `json:"row_id"`
	Product  Product `json:"product"`
	Quantity int     `json:"quantity"`
	CellSrc  Cell    `json:"cell_src"` // from
	CellDst  Cell    `json:"cell_dst"` // to
}

package model

import (
	"github.com/mlplabs/mwms-core/whs/cells"
)

type RowStorage struct {
	RowId    string     `json:"row_id"`
	Product  Product    `json:"product"`
	Quantity int        `json:"quantity"`
	CellSrc  cells.Cell `json:"cell_src"` // from
	CellDst  cells.Cell `json:"cell_dst"` // to
}

package whs

import (
	"context"
	"github.com/mlplabs/mwms-core/whs/model"
)

type Cells struct {
	wms *Wms
}

func NewCells(s *Wms) *Cells {
	return &Cells{wms: s}
}

func (c *Cells) GetById(ctx context.Context, cellId int64) (*model.Cell, error) {
	return nil, nil
}

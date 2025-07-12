package whs

import (
	"context"
	"github.com/mlplabs/mwms-core/whs/model"
)

func (s *Storage) ReportStocks(ctx context.Context) (*model.StockData, error) {
	retVal := make([]model.RowStock, 0)
	sqlSel := "SELECT store.prod_id AS product_id, coalesce(p.name, '<unnamed>') AS product_name, " +
		"       coalesce(m.id, 0) AS manufacturer_id, coalesce(m.name, '<unnamed>') AS manufacturer_name, " +
		"       store.zone_id, coalesce(z.name, '<unnamed>') AS zone_name, " +
		"       store.cell_id, " +
		"       store.quantity " +
		"FROM (SELECT s.prod_id, s.zone_id, s.cell_id, SUM(s.quantity) AS quantity " +
		"               FROM storage1 s " +
		"               GROUP BY s.prod_id, s.zone_id, s.cell_id) AS store " +
		"LEFT JOIN products p ON store.prod_id = p.id " +
		"LEFT JOIN manufacturers m on p.manufacturer_id = m.id " +
		"LEFT JOIN zones z ON store.zone_id = z.id " +
		"LEFT JOIN cells c ON store.cell_id = c.id " +
		"ORDER BY p.name"
	rows, err := s.wms.Db.QueryContext(ctx, sqlSel)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	for rows.Next() {
		row := model.RowStock{
			Product:  model.Product{},
			Quantity: 0,
			Cells:    make([]model.Cell, 0),
		}
		cellId := int64(0)
		err = rows.Scan(&row.Product.Id, &row.Product.Name, &row.Product.Manufacturer.Id, &row.Product.Manufacturer.Name, &row.Zone.Id, &row.Zone.Name, &cellId, &row.Quantity)
		if err != nil {
			return nil, err
		}
		cell, err := s.wms.GetCellInfo(ctx, cellId, nil)
		if err != nil {
			return nil, err
		}
		cell.SetName("")
		row.Cells = append(row.Cells, *cell)
		retVal = append(retVal, row)
	}

	return &model.StockData{Rows: retVal}, nil
}

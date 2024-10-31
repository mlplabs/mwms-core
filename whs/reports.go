package whs

import (
	"fmt"
	"github.com/lib/pq"
	"github.com/mlplabs/microwms-core/core"
	"time"
)

func (s *Storage) GetRemainingProducts() ([]RemainingProductRow, error) {
	retVal := make([]RemainingProductRow, 0)
	sqlSel := "SELECT store.prod_id AS product_id, coalesce(p.name, '<unnamed>') AS product_name, " +
		"       coalesce(m.id, 0) AS manufacturer_id, coalesce(m.name, '<unnamed>') AS manufacturer_name, " +
		"       store.zone_id, coalesce(z.name, '<unnamed>') AS zone_name, " +
		"       store.cell_id, c.name AS cell_name, " +
		"       store.quantity " +
		"FROM (SELECT s.prod_id, s.zone_id, s.cell_id, SUM(s.quantity) AS quantity " +
		"               FROM storage1 s " +
		"               GROUP BY s.prod_id, s.zone_id, s.cell_id) AS store " +
		"LEFT JOIN products p ON store.prod_id = p.id " +
		"LEFT JOIN manufacturers m on p.manufacturer_id = m.id " +
		"LEFT JOIN zones z ON store.zone_id = z.id " +
		"LEFT JOIN cells c ON store.cell_id = c.id " +
		"ORDER BY p.name"
	rows, err := s.Db.Query(sqlSel)
	if err != nil {
		return retVal, err
	}
	defer rows.Close()
	for rows.Next() {
		r := RemainingProductRow{}
		r.Product = &ProductItem{}
		r.Manufacturer = &ManufacturerItem{}
		r.Zone = &ZoneItem{}
		r.Cell = &CellItem{}

		err = rows.Scan(&r.Product.Id, &r.Product.Name, &r.Manufacturer.Id, &r.Manufacturer.Name, &r.Zone.Id, &r.Zone.Name, &r.Cell.Id, &r.Cell.Name, &r.Quantity)
		if err != nil {
			return retVal, err
		}
		retVal = append(retVal, r)
	}

	return retVal, nil
}

func (s *Storage) GetRemainingProductsByIds(idsArray []int64) ([]RemainingProductRow, error) {
	retVal := make([]RemainingProductRow, 0)
	sqlSel := "SELECT store.prod_id AS product_id, coalesce(p.name, '<unnamed>') AS product_name, " +
		"       coalesce(m.id, 0) AS manufacturer_id, coalesce(m.name, '<unnamed>') AS manufacturer_name, " +
		"       store.zone_id, coalesce(z.name, '<unnamed>') AS zone_name, " +
		"       store.cell_id, c.name AS cell_name, " +
		"       store.quantity " +
		"FROM (SELECT s.prod_id, s.zone_id, s.cell_id, SUM(s.quantity) AS quantity " +
		"               FROM storage1 s " +
		"				WHERE s.prod_id = ANY($1) " +
		"               GROUP BY s.prod_id, s.zone_id, s.cell_id) AS store " +
		"LEFT JOIN products p ON store.prod_id = p.id " +
		"LEFT JOIN manufacturers m on p.manufacturer_id = m.id " +
		"LEFT JOIN zones z ON store.zone_id = z.id " +
		"LEFT JOIN cells c ON store.cell_id = c.id " +
		"ORDER BY p.name"
	rows, err := s.Db.Query(sqlSel, pq.Array(idsArray))

	if err != nil {
		return retVal, err
	}
	defer rows.Close()
	for rows.Next() {
		r := RemainingProductRow{}
		r.Product = &ProductItem{}
		r.Manufacturer = &ManufacturerItem{}
		r.Zone = &ZoneItem{}
		r.Cell = &CellItem{}
		err = rows.Scan(&r.Product.Id, &r.Product.Name, &r.Manufacturer.Id, &r.Manufacturer.Name, &r.Zone.Id, &r.Zone.Name, &r.Cell.Id, &r.Cell.Name, &r.Quantity)
		if err != nil {
			return retVal, err
		}
		retVal = append(retVal, r)
	}
	return retVal, nil
}

func (s *Storage) GetTurnovers(param TurnoversParams, offset, limit int) ([]TurnoversProductRow, int, error) {
	var count int
	if limit == 0 {
		limit = DefaultRowsLimit
	}

	args := make([]any, 0)
	sqlCond := " WHERE $1 "
	args = append(args, true)

	if param.Debit {
		sqlCond += " AND quantity > 0 "
	}
	if param.Credit {
		sqlCond += " AND quantity <= 0 "
	}

	if param.DocTypes != nil && len(param.DocTypes) > 0 {
		args = append(args, pq.Array(param.DocTypes))
		sqlCond += fmt.Sprintf(" AND s.doc_type = ANY($%d) ", len(args))

	}

	args = append(args, limit, offset)

	sqlSel := fmt.Sprintf("SELECT s.doc_id, r.number, r.date, "+
		"   s.prod_id, coalesce(p.name, 'unknown') as prod_name, "+
		"	m.id AS mnf_id, coalesce(m.name, 'unknown') AS mnf_name, s.quantity "+
		"	FROM storage1 s "+
		"	LEFT JOIN shipment_headers sh on sh.id = s.doc_id AND s.doc_type = 1 "+
		"	LEFT JOIN receipt_headers r on r.id = s.doc_id AND s.doc_type = 2 "+
		"	LEFT JOIN products p on p.id = s.prod_id "+
		"	LEFT JOIN manufacturers m on m.id = p.manufacturer_id "+
		"	%s ", sqlCond)
	sqlSelExt := fmt.Sprintf("ORDER BY row_time DESC LIMIT $%d OFFSET $%d", len(args)-1, len(args))
	rows, err := s.Db.Query(sqlSel+sqlSelExt, args...)

	if err != nil {
		return nil, count, &core.WrapError{Err: err, Code: core.SystemError}
	}
	defer rows.Close()

	items := make([]TurnoversProductRow, count, limit)

	for rows.Next() {
		docId := 0
		docNum := ""
		docDate := time.Time{}

		item := new(TurnoversProductRow)
		item.Product = new(ProductItem)
		err = rows.Scan(&docId, &docNum, &docDate, &item.Product.Id, &item.Product.Name, &item.Product.Manufacturer.Id, &item.Product.Manufacturer.Name, &item.Quantity)

		doc := new(Document)
		docItem,_ := doc.GetNewItem()
		docItem.Id = int64(docId)
		docItem.Number = docItem.GetNumber()
		docItem.Date = docItem.GetDate(docDate)

		item.Doc = docItem
		items = append(items, *item)
	}

	argsCount := make([]any, 0)
	argsCount = append(argsCount, args[:len(args)-2]...)

	sqlCount := fmt.Sprintf("SELECT COUNT(*) as count FROM ( %s ) sub", sqlSel)
	err = s.Db.QueryRow(sqlCount, argsCount...).Scan(&count)
	if err != nil {
		return nil, count, &core.WrapError{Err: err, Code: core.SystemError}
	}
	return items, count, nil
}

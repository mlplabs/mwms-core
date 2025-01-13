package whs

import (
	"context"
	"database/sql"
	"fmt"
	"github.com/lib/pq"
	"github.com/mlplabs/mwms-core/whs/cells"
	"github.com/mlplabs/mwms-core/whs/model"
)

type Storage struct {
	wms *Wms
}

func NewStorage(s *Wms) *Storage {
	return &Storage{wms: s}
}

// GetRow отбирает из ячейки (cell) продукт (prod) в количестве (quantity)
// Возвращает отобранное количество (quantity)
func (s *Storage) GetRow(ctx context.Context, row *model.RowStorage, tx *sql.Tx) (int, error) {
	var err error

	if tx == nil {
		tx, err = s.wms.Db.Begin()
		if err != nil {
			// не смогли начать транзакцию
			return 0, err
		}
	}
	dGetId := 0
	dGetType := 0
	sqlInsert := fmt.Sprintf("INSERT INTO wms%d (doc_id, doc_type, zone_id, cell_id, row_id, prod_id, quantity) VALUES ($1, $2, $3, $4)", row.CellSrc.WhsId)
	_, err = tx.ExecContext(ctx, sqlInsert, dGetId, dGetType, row.CellSrc.ZoneId, row.CellSrc.Id, row.RowId, row.Product.Id, -1*row.Quantity)
	if err != nil {
		return 0, err
	}

	sqlQuant := fmt.Sprintf("SELECT SUM(quantity) AS quantity "+
		"FROM wms%d WHERE zone_id = $1 AND cell_id = $2 AND prod_id = $3 "+
		"GROUP BY zone_id, cell_id, prod_id "+
		"HAVING SUM(quantity) < 0", row.CellSrc.WhsId)
	rows, err := tx.QueryContext(ctx, sqlQuant, row.CellSrc.ZoneId, row.CellSrc.Id, row.Product.Id)
	if err != nil {
		// ошибка контроля
		return 0, err
	}
	defer rows.Close()
	// мы должны получить пустой запрос
	if rows.Next() {
		err = tx.Rollback()
		if err != nil {
			// ошибка отката... все очень плохо
			return 0, err
		}
		return 0, err
	}
	err = tx.Commit()
	if err != nil {
		return 0, err
	}
	return row.Quantity, nil
}

// PutRow размещает в ячейку (cell) продукт (prod) в количестве (quantity)
// Возвращает количество которое было размещено (quantity)
func (s *Storage) PutRow(ctx context.Context, row *model.RowStorage, tx *sql.Tx) (int, error) {
	var err error

	// TODO:
	dGetId := 0
	dGetType := 0
	sqlIns := fmt.Sprintf("INSERT INTO wms%d (doc_id, doc_type, zone_id, cell_id, row_id, prod_id, quantity) VALUES ($1, $2, $3, $4, $5, $6, $7)", row.CellDst.WhsId)
	if tx != nil {
		_, err = tx.ExecContext(ctx, sqlIns, dGetId, dGetType, row.CellDst.ZoneId, row.CellDst.Id, row.RowId, row.Product.Id, row.Quantity)
	} else {
		_, err = s.wms.Db.ExecContext(ctx, sqlIns, dGetId, dGetType, row.CellDst.ZoneId, row.CellDst.Id, row.RowId, row.Product.Id, row.Quantity)
	}
	if err != nil {
		return 0, err
	}
	return row.Quantity, nil
}
func (s *Storage) MoveRow(ctx context.Context, row *model.RowStorage, tx *sql.Tx) error {
	// TODO: cellSrc.WhsId <> cellDst.WhsId - временной разрыв или виртуальное перемещение

	_, err := s.GetRow(ctx, row, tx)
	if err != nil {
		return err
	}
	_, err = s.PutRow(ctx, row, tx)
	if err == nil {
		return err
	}
	return nil
}

// Quantity возвращает количество продуктов на св ячейке
func (s *Storage) Quantity(ctx context.Context, whsId int, cell *cells.Cell, tx *sql.Tx) (map[int]int, error) {
	var zoneId, cellId, prodId, quantity int
	res := make(map[int]int)

	sqlQuantity := fmt.Sprintf("SELECT zone_id, cell_id, prod_id, SUM(quantity) AS quantity "+
		"FROM wms%d WHERE zone_id = $1 AND cell_id = $2 "+
		"GROUP BY zone_id, cell_id, prod_id "+
		"HAVING SUM(quantity) <> 0 %s", whsId, "")

	var err error
	var rows *sql.Rows

	if tx != nil {
		rows, err = tx.QueryContext(ctx, sqlQuantity, cell.ZoneId, cell.Id)
	} else {
		rows, err = s.wms.Db.QueryContext(ctx, sqlQuantity, cell.ZoneId, cell.Id)
	}
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		err = rows.Scan(&zoneId, &cellId, &prodId, &quantity)
		if err != nil {
			return nil, err
		}
		res[prodId] = quantity
	}
	return res, nil
}

// BulkChangeSzCells устанавливает весогабаритные характеристики для массива ячеек
func (s *Storage) BulkChangeSzCells(ctx context.Context, cells []cells.Cell, sz SpecificSize) (int64, error) {
	var ids []int64

	for _, c := range cells {
		ids = append(ids, c.Id)
	}
	sqlBulkUpdate := "UPDATE cells SET sz_length=$2, sz_width=$3, sz_height=$4, sz_volume=$5, sz_uf_volume=$6, sz_weight=$7 WHERE id = ANY($1)"
	res, err := s.wms.Db.ExecContext(ctx, sqlBulkUpdate, pq.Array(ids), sz.Length, sz.Width, sz.Height, sz.Volume, sz.UsefulVolume, sz.Weight)
	if err != nil {
		return 0, err
	}
	return res.RowsAffected()
}

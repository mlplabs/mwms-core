package whs

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"github.com/mlplabs/mwms-core/whs/model"
)

type Storage struct {
	wms *Wms
}

func NewStorage(s *Wms) *Storage {
	return &Storage{wms: s}
}

func (s *Storage) getCellInfo(ctx context.Context, cellId int64, tx *sql.Tx) (*model.Cell, error) {
	sqlCell := "SELECT cs.id, cs.name, cs.whs_id, cs.zone_id FROM cells cs WHERE cs.id = $1"
	c := model.Cell{}
	row := tx.QueryRowContext(ctx, sqlCell, cellId)
	err := row.Scan(c.Id, c.Name, c.WhsId, c.ZoneId)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return &c, nil
		}
		return nil, err
	}
	return &c, nil
}

func (s *Storage) balanceControl(ctx context.Context, itemId int64, cellId int64, tx *sql.Tx) (bool, error) {
	var balance int
	sqlCtrl := "SELECT SUM(quantity) AS quantity " +
		"FROM storage%d WHERE cell_id = $2 AND prod_id = $3 " +
		"GROUP BY cell_id, prod_id " +
		"HAVING SUM(quantity) < 0"
	row := tx.QueryRowContext(ctx, sqlCtrl, cellId, itemId)
	err := row.Scan(&balance)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return true, nil
		}
		return false, err
	}
	return false, fmt.Errorf("balance control failed %d", balance)
}

// GetItemFromCell отбирает из ячейки (cellId) продукт (itemId) в количестве (quantity)
// Возвращает отобранное количество (quantity)
func (s *Storage) GetItemFromCell(ctx context.Context, itemId int64, cellId int64, quantity int) (int, error) {
	tx, err := s.wms.Db.Begin()
	if err != nil {
		return 0, err
	}

	cell, err := s.getCellInfo(ctx, cellId, tx)
	if err != nil {
		_ = tx.Rollback()
		return 0, err
	}

	sqlInsert := fmt.Sprintf("INSERT INTO storage%d (prod_id, zone_id, cell_id, quantity) VALUES ($1, $2, $3, $4)", cell.WhsId)
	_, err = tx.Exec(sqlInsert, itemId, cell.ZoneId, cellId, -1*quantity)
	if err != nil {
		_ = tx.Rollback()
		return 0, err
	}

	_, err = s.balanceControl(ctx, itemId, cellId, tx)
	if err != nil {
		_ = tx.Rollback()
		return 0, err
	}

	err = tx.Commit()
	if err != nil {
		return 0, err
	}
	return quantity, nil
}

// PutItemToCell размещает в ячейку (CellId) продукт (ItemId) в количестве (Quantity)
// Возвращает количество которое было размещено (Quantity)
func (s *Storage) PutItemToCell(ctx context.Context, itemId int64, cellId int64, quantity int) (int, error) {
	tx, err := s.wms.Db.Begin()
	if err != nil {
		return 0, err
	}

	cell, err := s.getCellInfo(ctx, cellId, tx)
	if err != nil {
		_ = tx.Rollback()
		return 0, err
	}

	sqlIns := fmt.Sprintf("INSERT INTO storage%d (prod_id, zone_id, cell_id, quantity) VALUES ($1, $2, $3, $4)", cell.WhsId)
	_, err = tx.ExecContext(ctx, sqlIns, itemId, cell.ZoneId, cellId, quantity)
	if err != nil {
		_ = tx.Rollback()
		return 0, err
	}
	err = tx.Commit()
	if err != nil {
		return 0, err
	}
	return quantity, nil
}
func (s *Storage) MoveItemToCell(ctx context.Context, itemId int64, cellSrcId int64, cellDstId int64, quantity int) (int, error) {
	tx, err := s.wms.Db.Begin()
	if err != nil {
		return 0, err
	}

	cellSrc, err := s.getCellInfo(ctx, cellSrcId, tx)
	if err != nil {
		_ = tx.Rollback()
		return 0, err
	}

	cellDst, err := s.getCellInfo(ctx, cellDstId, tx)
	if err != nil {
		_ = tx.Rollback()
		return 0, err
	}

	if cellDst.WhsId != cellSrc.WhsId {
		// TODO: cellSrc.WhsId <> cellDst.WhsId - временной разрыв или виртуальное перемещение
		return 0, fmt.Errorf("межскладское перемещение пока не реализовано(")
	}

	sqlInsertSrc := fmt.Sprintf("INSERT INTO storage%d (prod_id, zone_id, cell_id, quantity) VALUES ($1, $2, $3, $4)", cellSrc.WhsId)
	sqlInsertDst := fmt.Sprintf("INSERT INTO storage%d (prod_id, zone_id, cell_id, quantity) VALUES ($1, $2, $3, $4)", cellDst.WhsId)

	_, err = tx.Exec(sqlInsertSrc, itemId, cellSrc.ZoneId, cellSrcId, -1*quantity)
	if err != nil {
		_ = tx.Rollback()
		return 0, err
	}
	_, err = tx.Exec(sqlInsertDst, itemId, cellDst.ZoneId, cellDstId, -1*quantity)
	if err != nil {
		_ = tx.Rollback()
		return 0, err
	}
	_, err = s.balanceControl(ctx, itemId, cellSrcId, tx)
	if err != nil {
		_ = tx.Rollback()
		return 0, err
	}

	err = tx.Commit()
	if err != nil {
		return 0, err
	}
	return quantity, nil
}

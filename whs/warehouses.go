package whs

import (
	"context"
	"errors"
	"fmt"
	"github.com/lib/pq"
	"github.com/mlplabs/mwms-core/whs/model"
)

const tableWarehouses = "warehouses"

type Warehouses struct {
	wms *Wms
}

func NewWarehouses(s *Wms) *Warehouses {
	return &Warehouses{wms: s}
}

// GetWarehouses returns a list items without limit
func (s *Storage) GetWarehouses(ctx context.Context) ([]model.Warehouse, error) {
	items := make([]model.Warehouse, 0)
	sqlSel := fmt.Sprintf("SELECT id, name FROM %s ORDER BY name ASC", tableWarehouses)
	rows, err := s.wms.Db.QueryContext(ctx, sqlSel)
	if err != nil {
		return items, err
	}
	defer rows.Close()

	for rows.Next() {
		item := model.Warehouse{}
		err = rows.Scan(&item.Id, &item.Name)
		items = append(items, item)
	}
	return items, nil
}

// GetWarehousesItems returns a list items of catalog
func (s *Storage) GetWarehousesItems(ctx context.Context, offset int, limit int) ([]model.Warehouse, int64, error) {
	var totalCount int64
	items := make([]model.Warehouse, 0)
	sqlCond := ""
	args := make([]any, 0)

	if limit == 0 {
		limit = DefaultRowsLimit
	}
	args = append(args, limit)
	args = append(args, offset)

	sqlSel := fmt.Sprintf("SELECT id, name FROM %s %s ORDER BY name ASC", tableWarehouses, sqlCond)

	rows, err := s.wms.Db.QueryContext(ctx, sqlSel+" LIMIT $1 OFFSET $2", args...)
	if err != nil {
		return items, totalCount, err
	}
	defer rows.Close()

	for rows.Next() {
		item := model.Warehouse{}
		err = rows.Scan(&item.Id, &item.Name)
		items = append(items, item)
	}

	sqlCount := fmt.Sprintf("SELECT COUNT(*) as count FROM ( %s ) sub", sqlSel)
	err = s.wms.Db.QueryRowContext(ctx, sqlCount).Scan(&totalCount)
	if err != nil {
		return nil, totalCount, err
	}
	return items, totalCount, nil
}

func (s *Storage) CreateWarehouse(ctx context.Context, whs *model.Warehouse) (int64, error) {
	var insertId int64

	tx, err := s.wms.Db.Begin()
	if err != nil {
		return insertId, err
	}

	sqlCreate := fmt.Sprintf("INSERT INTO %s (name) VALUES ($1) RETURNING id", tableWarehouses)
	err = tx.QueryRowContext(ctx, sqlCreate, whs.Name).Scan(&insertId)
	if err != nil {
		tx.Rollback()
		return insertId, err
	}

	sqlStorage := fmt.Sprintf(
		"create table if not exists wms%d ( "+
			"doc_id   integer default 0 not null, "+
			"doc_type smallint default 0 not null, "+
			"row_id   varchar(36) default ''::character varying not null, "+
			"row_time timestamptz default now() not null, "+
			"zone_id  integer, "+
			"cell_id  integer constraint wms%d_cells_id_fk references cells, "+
			"prod_id  integer,	"+
			"quantity integer ); "+
			"alter table wms%d owner to %s;", whs.Id, whs.Id, whs.Id, s.wms.GetDbUser())
	_, err = tx.Exec(sqlStorage)
	if err != nil {
		tx.Rollback()
		return insertId, err
	}

	err = tx.Commit()
	if err != nil {
		tx.Rollback()
		return insertId, err
	}

	return insertId, nil
}

func (s *Storage) UpdateWarehouse(ctx context.Context, whs *model.Warehouse) (int64, error) {
	sqlUpd := fmt.Sprintf("UPDATE %s SET name=$2 WHERE id=$1", tableWarehouses)
	res, err := s.wms.Db.ExecContext(ctx, sqlUpd, whs.Id, whs.Name)
	if err != nil {
		return 0, err
	}
	if a, err := res.RowsAffected(); a != 1 || err != nil {
		return 0, err
	}
	return whs.Id, nil
}

// DeleteWarehouse delete warehouse
func (s *Storage) DeleteWarehouse(ctx context.Context, itemId int64) error {
	if itemId == 0 {
		return fmt.Errorf("unacceptable action. item id eq 0")
	}
	sqlDel := fmt.Sprintf("DELETE FROM %s WHERE id=$1", tableWarehouses)
	_, err := s.wms.Db.ExecContext(ctx, sqlDel, itemId)
	if err != nil {
		var pgErr *pq.Error
		if errors.As(err, &pgErr) {
			if pgErr.Code == ("23503") {
				return err
			}
		}
		return err
	}
	return nil
}

// GetWarehouseById returns a warehouse object by id
func (s *Storage) GetWarehouseById(ctx context.Context, itemId int64) (*model.Warehouse, error) {
	item := model.Warehouse{}
	sqlWhs := fmt.Sprintf("SELECT id, name, address FROM %s WHERE id = $1", tableWarehouses)
	row := s.wms.Db.QueryRowContext(ctx, sqlWhs, itemId)

	err := row.Scan(&item.Id, &item.Name, &item.Address)
	if err != nil {
		return &item, err
	}
	return &item, nil
}

func (s *Storage) FindWarehousesByName(ctx context.Context, itemName string) ([]model.Warehouse, error) {
	items := make([]model.Warehouse, 0)
	sql := fmt.Sprintf("SELECT id, name FROM %s WHERE name = $1", tableWarehouses)
	rows, err := s.wms.Db.QueryContext(ctx, sql, itemName)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	for rows.Next() {
		item := model.Warehouse{}
		err = rows.Scan(&item.Id, &item.Name)
		if err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	return items, nil
}

func (s *Storage) WarehousesSuggest(ctx context.Context, text string, limit int) ([]model.Suggestion, error) {
	retVal := make([]model.Suggestion, 0)
	if limit == 0 {
		limit = DefaultSuggestionLimit
	}

	sqlSel := fmt.Sprintf("SELECT id, name FROM %s WHERE name ILIKE $1 LIMIT $2", tableWarehouses)
	rows, err := s.wms.Db.QueryContext(ctx, sqlSel, text+"%", limit)
	if err != nil {
		return retVal, err
	}
	defer rows.Close()
	for rows.Next() {
		item := model.Suggestion{}
		err := rows.Scan(&item.Id, &item.Val)
		if err != nil {
			return retVal, err
		}
		item.Title = item.Val
		retVal = append(retVal, item)
	}
	return retVal, err
}

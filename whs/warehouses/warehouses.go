package warehouses

import (
	"context"
	"errors"
	"fmt"
	"github.com/lib/pq"
	"github.com/mlplabs/mwms-core/whs"
	"github.com/mlplabs/mwms-core/whs/suggestion"
)

const tableWarehouses = "warehouses"

type Warehouses struct {
	storage *whs.Storage
}

func NewWarehouses(s *whs.Storage) *Warehouses {
	return &Warehouses{storage: s}
}

// Get returns a list items without limit
func (w *Warehouses) Get(ctx context.Context) ([]Warehouse, error) {
	items := make([]Warehouse, 0)
	sqlSel := fmt.Sprintf("SELECT id, name FROM %s ORDER BY name ASC", tableWarehouses)
	rows, err := w.storage.Db.QueryContext(ctx, sqlSel)
	if err != nil {
		return items, err
	}
	defer rows.Close()

	for rows.Next() {
		item := Warehouse{}
		err = rows.Scan(&item.Id, &item.Name)
		items = append(items, item)
	}
	return items, nil
}

// GetItems returns a list items of catalog
func (w *Warehouses) GetItems(ctx context.Context, offset int, limit int) ([]Warehouse, int, error) {
	var totalCount int
	items := make([]Warehouse, 0)
	sqlCond := ""
	args := make([]any, 0)

	if limit == 0 {
		limit = whs.DefaultRowsLimit
	}
	args = append(args, limit)
	args = append(args, offset)

	sqlSel := fmt.Sprintf("SELECT id, name FROM %s %s ORDER BY name ASC", tableWarehouses, sqlCond)

	rows, err := w.storage.Db.QueryContext(ctx, sqlSel+" LIMIT $1 OFFSET $2", args...)
	if err != nil {
		return items, totalCount, err
	}
	defer rows.Close()

	for rows.Next() {
		item := Warehouse{}
		err = rows.Scan(&item.Id, &item.Name)
		items = append(items, item)
	}

	sqlCount := fmt.Sprintf("SELECT COUNT(*) as count FROM ( %s ) sub", sqlSel)
	err = w.storage.Db.QueryRowContext(ctx, sqlCount).Scan(&totalCount)
	if err != nil {
		return nil, totalCount, err
	}
	return items, totalCount, nil
}

func (w *Warehouses) Create(ctx context.Context, whs *Warehouse) (int64, error) {
	var insertId int64

	tx, err := w.storage.Db.Begin()
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
		"create table if not exists storage%d ( "+
			"doc_id   integer default 0 not null, "+
			"doc_type smallint default 0 not null, "+
			"row_id   varchar(36) default ''::character varying not null, "+
			"row_time timestamptz default now() not null, "+
			"zone_id  integer, "+
			"cell_id  integer constraint storage%d_cells_id_fk references cells, "+
			"prod_id  integer,	"+
			"quantity integer ); "+
			"alter table storage%d owner to %s;", whs.Id, whs.Id, whs.Id, w.storage.GetDbUser())
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

func (w *Warehouses) Update(ctx context.Context, whs *Warehouse) (int64, error) {
	sqlUpd := fmt.Sprintf("UPDATE %s SET name=$2 WHERE id=$1", tableWarehouses)
	res, err := w.storage.Db.ExecContext(ctx, sqlUpd, whs.Id, whs.Name)
	if err != nil {
		return 0, err
	}
	if a, err := res.RowsAffected(); a != 1 || err != nil {
		return 0, err
	}
	return whs.Id, nil
}

// Delete delete warehouse
func (w *Warehouses) Delete(ctx context.Context, itemId int64) error {
	if itemId == 0 {
		return fmt.Errorf("unacceptable action. item id eq 0")
	}
	sqlDel := fmt.Sprintf("DELETE FROM %s WHERE id=$1", tableWarehouses)
	_, err := w.storage.Db.ExecContext(ctx, sqlDel, itemId)
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

// GetById returns a warehouse object by id
func (w *Warehouses) GetById(ctx context.Context, itemId int64) (Warehouse, error) {
	item := Warehouse{}
	sqlWhs := fmt.Sprintf("SELECT id, name, address FROM %s WHERE id = $1", tableWarehouses)
	row := w.storage.Db.QueryRowContext(ctx, sqlWhs, itemId)

	err := row.Scan(&item.Id, &item.Name, &item.Address)
	if err != nil {
		return item, err
	}
	return item, nil
}

func (w *Warehouses) FindByName(ctx context.Context, itemName string) ([]Warehouse, error) {
	items := make([]Warehouse, 0)
	sql := fmt.Sprintf("SELECT id, name FROM %s WHERE name = $1", tableWarehouses)
	rows, err := w.storage.Db.QueryContext(ctx, sql, itemName)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	for rows.Next() {
		item := Warehouse{}
		err = rows.Scan(&item.Id, &item.Name)
		if err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	return items, nil
}

func (w *Warehouses) Suggest(ctx context.Context, text string, limit int) ([]suggestion.Suggestion, error) {
	retVal := make([]suggestion.Suggestion, 0)
	if limit == 0 {
		limit = whs.DefaultSuggestionLimit
	}

	sqlSel := fmt.Sprintf("SELECT id, name FROM %s WHERE name ILIKE $1 LIMIT $2", tableWarehouses)
	rows, err := w.storage.Db.QueryContext(ctx, sqlSel, text+"%", limit)
	if err != nil {
		return retVal, err
	}
	defer rows.Close()
	for rows.Next() {
		item := suggestion.Suggestion{}
		err := rows.Scan(&item.Id, &item.Val)
		if err != nil {
			return retVal, err
		}
		item.Title = item.Val
		retVal = append(retVal, item)
	}
	return retVal, err
}

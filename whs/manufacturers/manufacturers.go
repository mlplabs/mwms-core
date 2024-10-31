package manufacturers

import (
	"context"
	"errors"
	"fmt"
	"github.com/lib/pq"
	"github.com/mlplabs/mwms-core/whs"
	"github.com/mlplabs/mwms-core/whs/suggestion"
)

const tableManufacturers = "manufacturers"

type Manufacturers struct {
	storage *whs.Storage
}

func NewManufacturers(s *whs.Storage) *Manufacturers {
	return &Manufacturers{storage: s}
}

func (m *Manufacturers) Get(ctx context.Context) ([]Manufacturer, error) {
	items := make([]Manufacturer, 0)
	sqlSel := fmt.Sprintf("SELECT id, name FROM %s ORDER BY name ASC", tableManufacturers)
	rows, err := m.storage.Db.QueryContext(ctx, sqlSel)
	if err != nil {
		return items, err
	}
	defer rows.Close()

	for rows.Next() {
		mnf := Manufacturer{}
		err = rows.Scan(&mnf.Id, &mnf.Name)
		items = append(items, mnf)
	}
	return items, nil
}

func (m *Manufacturers) GetItems(ctx context.Context, offset int, limit int) ([]Manufacturer, int64, error) {
	var totalCount int64
	items := make([]Manufacturer, 0)
	sqlCond := ""
	args := make([]any, 0)

	if limit == 0 {
		limit = whs.DefaultRowsLimit
	}
	args = append(args, limit)
	args = append(args, offset)

	sqlSel := fmt.Sprintf("SELECT id, name FROM %s %s ORDER BY name ASC", tableManufacturers, sqlCond)

	rows, err := m.storage.Db.QueryContext(ctx, sqlSel+" LIMIT $1 OFFSET $2", args...)
	if err != nil {
		return nil, totalCount, err
	}
	defer rows.Close()

	for rows.Next() {
		mnf := Manufacturer{}
		err = rows.Scan(&mnf.Id, &mnf.Name)
		items = append(items, mnf)
	}

	sqlCount := fmt.Sprintf("SELECT COUNT(*) as count FROM ( %s ) sub", sqlSel)
	err = m.storage.Db.QueryRowContext(ctx, sqlCount).Scan(&totalCount)
	if err != nil {
		return nil, totalCount, err
	}
	return items, totalCount, nil
}

func (m *Manufacturers) Create(ctx context.Context, mnf *Manufacturer) (int64, error) {
	var insertId int64
	sqlCreate := fmt.Sprintf("INSERT INTO %s (name) VALUES ($1) RETURNING id", tableManufacturers)
	err := m.storage.Db.QueryRowContext(ctx, sqlCreate, mnf.Name).Scan(&insertId)
	return insertId, err
}

func (m *Manufacturers) Update(ctx context.Context, mnf *Manufacturer) (int64, error) {
	sqlUpd := fmt.Sprintf("UPDATE %s SET name=$2 WHERE id=$1", tableManufacturers)
	res, err := m.storage.Db.ExecContext(ctx, sqlUpd, mnf.Id, mnf.Name)
	if err != nil {
		return 0, err
	}
	if a, err := res.RowsAffected(); a != 1 || err != nil {
		return 0, err
	}
	return mnf.Id, nil
}
func (m *Manufacturers) Delete(ctx context.Context, itemId int64) error {
	if itemId == 0 {
		return fmt.Errorf("unacceptable action. item id eq 0")
	}
	sqlDel := fmt.Sprintf("DELETE FROM %s WHERE id=$1", tableManufacturers)
	_, err := m.storage.Db.ExecContext(ctx, sqlDel, itemId)
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
func (m *Manufacturers) GetById(ctx context.Context, itemId int64) (*Manufacturer, error) {
	sqlUsr := fmt.Sprintf("SELECT id, name FROM %s WHERE id = $1", tableManufacturers)
	row := m.storage.Db.QueryRowContext(ctx, sqlUsr, itemId)
	newItem := Manufacturer{}
	err := row.Scan(&newItem.Id, &newItem.Name)
	if err != nil {
		return nil, err
	}
	return &newItem, nil
}

func (m *Manufacturers) FindByName(ctx context.Context, itemName string) ([]Manufacturer, error) {
	items := make([]Manufacturer, 0)
	sql := fmt.Sprintf("SELECT id, name FROM %s WHERE name = $1", tableManufacturers)
	rows, err := m.storage.Db.QueryContext(ctx, sql, itemName)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	for rows.Next() {
		mnf := Manufacturer{}
		err = rows.Scan(&mnf.Id, &mnf.Name)
		if err != nil {
			return nil, err
		}
		items = append(items, mnf)
	}
	return items, nil
}

func (m *Manufacturers) Suggest(ctx context.Context, text string, limit int) ([]suggestion.Suggestion, error) {
	retVal := make([]suggestion.Suggestion, 0)
	if limit == 0 {
		limit = whs.DefaultSuggestionLimit
	}

	sqlSel := fmt.Sprintf("SELECT id, name FROM %s WHERE name ILIKE $1 LIMIT $2", tableManufacturers)
	rows, err := m.storage.Db.QueryContext(ctx, sqlSel, text+"%", limit)
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

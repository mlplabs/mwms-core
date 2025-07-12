package whs

import (
	"context"
	"errors"
	"fmt"
	"github.com/lib/pq"
	"github.com/mlplabs/mwms-core/whs/model"
)

const tableManufacturers = "manufacturers"

func (s *Storage) GetManufacturers(ctx context.Context) ([]model.Manufacturer, error) {
	items := make([]model.Manufacturer, 0)
	sqlSel := fmt.Sprintf("SELECT id, name FROM %s ORDER BY name ASC", tableManufacturers)
	rows, err := s.wms.Db.QueryContext(ctx, sqlSel)
	if err != nil {
		return items, err
	}
	defer rows.Close()

	for rows.Next() {
		mnf := model.Manufacturer{}
		err = rows.Scan(&mnf.Id, &mnf.Name)
		items = append(items, mnf)
	}
	return items, nil
}

func (s *Storage) GetManufacturersItems(ctx context.Context, offset int, limit int, search string) ([]model.Manufacturer, int64, error) {
	var totalCount int64
	var sqlCond string
	items := make([]model.Manufacturer, 0)
	if search != "" {
		sqlCond = "WHERE name ILIKE '" + search + "%'"
	}
	args := make([]any, 0)

	if limit == 0 {
		limit = DefaultRowsLimit
	}
	args = append(args, limit)
	args = append(args, offset)

	sqlSel := fmt.Sprintf("SELECT id, name FROM %s %s ORDER BY name ASC", tableManufacturers, sqlCond)

	rows, err := s.wms.Db.QueryContext(ctx, sqlSel+" LIMIT $1 OFFSET $2", args...)
	if err != nil {
		return nil, totalCount, err
	}
	defer rows.Close()

	for rows.Next() {
		mnf := model.Manufacturer{}
		err = rows.Scan(&mnf.Id, &mnf.Name)
		items = append(items, mnf)
	}

	sqlCount := fmt.Sprintf("SELECT COUNT(*) as count FROM ( %s ) sub", sqlSel)
	err = s.wms.Db.QueryRowContext(ctx, sqlCount).Scan(&totalCount)
	if err != nil {
		return nil, totalCount, err
	}
	return items, totalCount, nil
}

func (s *Storage) CreateManufacturer(ctx context.Context, mnf *model.Manufacturer) (int64, error) {
	var insertId int64
	sqlCreate := fmt.Sprintf("INSERT INTO %s (name) VALUES ($1) RETURNING id", tableManufacturers)
	err := s.wms.Db.QueryRowContext(ctx, sqlCreate, mnf.Name).Scan(&insertId)
	return insertId, err
}

func (s *Storage) UpdateManufacturer(ctx context.Context, mnf *model.Manufacturer) (int64, error) {
	sqlUpd := fmt.Sprintf("UPDATE %s SET name=$2 WHERE id=$1", tableManufacturers)
	res, err := s.wms.Db.ExecContext(ctx, sqlUpd, mnf.Id, mnf.Name)
	if err != nil {
		return 0, err
	}
	if a, err := res.RowsAffected(); a != 1 || err != nil {
		return 0, err
	}
	return mnf.Id, nil
}
func (s *Storage) DeleteManufacturer(ctx context.Context, itemId int64) error {
	if itemId == 0 {
		return fmt.Errorf("unacceptable action. item id eq 0")
	}
	sqlDel := fmt.Sprintf("DELETE FROM %s WHERE id=$1", tableManufacturers)
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
func (s *Storage) GetManufacturerById(ctx context.Context, itemId int64) (*model.Manufacturer, error) {
	sqlUsr := fmt.Sprintf("SELECT id, name FROM %s WHERE id = $1", tableManufacturers)
	row := s.wms.Db.QueryRowContext(ctx, sqlUsr, itemId)
	newItem := model.Manufacturer{}
	err := row.Scan(&newItem.Id, &newItem.Name)
	if err != nil {
		return nil, err
	}
	return &newItem, nil
}

func (s *Storage) FindManufacturersByName(ctx context.Context, itemName string) ([]model.Manufacturer, error) {
	items := make([]model.Manufacturer, 0)
	sql := fmt.Sprintf("SELECT id, name FROM %s WHERE name = $1", tableManufacturers)
	rows, err := s.wms.Db.QueryContext(ctx, sql, itemName)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	for rows.Next() {
		mnf := model.Manufacturer{}
		err = rows.Scan(&mnf.Id, &mnf.Name)
		if err != nil {
			return nil, err
		}
		items = append(items, mnf)
	}
	return items, nil
}

func (s *Storage) ManufacturersSuggest(ctx context.Context, text string, limit int) ([]model.Suggestion, error) {
	retVal := make([]model.Suggestion, 0)
	if limit == 0 {
		limit = DefaultSuggestionLimit
	}

	sqlSel := fmt.Sprintf("SELECT id, name FROM %s WHERE name ILIKE $1 LIMIT $2", tableManufacturers)
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

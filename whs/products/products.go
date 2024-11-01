package products

import (
	"context"
	"errors"
	"fmt"
	"github.com/lib/pq"
	"github.com/mlplabs/mwms-core/whs"
	"github.com/mlplabs/mwms-core/whs/suggestion"
)

const tableProducts = "products"

type Products struct {
	storage *whs.Storage
}

func NewProducts(s *whs.Storage) *Products {
	return &Products{storage: s}
}

// Get returns a list items without limit
func (u *Products) Get(ctx context.Context) ([]Product, error) {
	items := make([]Product, 0)
	sqlSel := fmt.Sprintf("SELECT id, name FROM %s ORDER BY name ASC", tableProducts)
	rows, err := u.storage.Db.QueryContext(ctx, sqlSel)
	if err != nil {
		return items, err
	}
	defer rows.Close()

	for rows.Next() {
		item := Product{}
		err = rows.Scan(&item.Id, &item.Name)
		items = append(items, item)
	}
	return items, nil
}

// GetItems returns a list items of catalog with limit & offset
func (u *Products) GetItems(ctx context.Context, offset int, limit int) ([]Product, int64, error) {
	var totalCount int64
	items := make([]Product, 0)
	sqlCond := ""
	args := make([]any, 0)

	if limit == 0 {
		limit = whs.DefaultRowsLimit
	}
	args = append(args, limit)
	args = append(args, offset)

	sqlSel := fmt.Sprintf("SELECT id, name FROM %s %s ORDER BY name ASC", tableProducts, sqlCond)

	rows, err := u.storage.Db.QueryContext(ctx, sqlSel+" LIMIT $1 OFFSET $2", args...)
	if err != nil {
		return items, totalCount, err
	}
	defer rows.Close()

	for rows.Next() {
		item := Product{}
		err = rows.Scan(&item.Id, &item.Name)
		items = append(items, item)
	}

	sqlCount := fmt.Sprintf("SELECT COUNT(*) as count FROM ( %s ) sub", sqlSel)
	err = u.storage.Db.QueryRowContext(ctx, sqlCount).Scan(&totalCount)
	if err != nil {
		return items, totalCount, err
	}
	return items, totalCount, nil
}

func (u *Products) Create(ctx context.Context, product *Product) (int64, error) {
	var insertId int64
	sqlCreate := fmt.Sprintf("INSERT INTO %s (name) VALUES ($1) RETURNING id", tableProducts)
	err := u.storage.Db.QueryRowContext(ctx, sqlCreate, product.Name).Scan(&insertId)
	return insertId, err
}

func (u *Products) Update(ctx context.Context, product *Product) (int64, error) {
	sqlUpd := fmt.Sprintf("UPDATE %s SET name=$2 WHERE id=$1", tableProducts)
	res, err := u.storage.Db.ExecContext(ctx, sqlUpd, product.Id, product.Name)
	if err != nil {
		return 0, err
	}
	if a, err := res.RowsAffected(); a != 1 || err != nil {
		return 0, err
	}
	return product.Id, nil
}

func (u *Products) Delete(ctx context.Context, itemId int64) error {
	if itemId == 0 {
		return fmt.Errorf("unacceptable action. item id eq 0")
	}
	sqlDel := fmt.Sprintf("DELETE FROM %s WHERE id=$1", tableProducts)
	_, err := u.storage.Db.ExecContext(ctx, sqlDel, itemId)
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

func (u *Products) GetById(ctx context.Context, itemId int64) (*Product, error) {
	sqlUsr := fmt.Sprintf("SELECT id, name FROM %s WHERE id = $1", tableProducts)
	row := u.storage.Db.QueryRowContext(ctx, sqlUsr, itemId)
	newItem := Product{}
	err := row.Scan(&newItem.Id, &newItem.Name)
	if err != nil {
		return nil, err
	}
	return &newItem, nil
}

func (u *Products) FindByName(ctx context.Context, itemName string) ([]Product, error) {
	items := make([]Product, 0)
	sql := fmt.Sprintf("SELECT id, name FROM %s WHERE name = $1", tableProducts)
	rows, err := u.storage.Db.QueryContext(ctx, sql, itemName)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	for rows.Next() {
		item := Product{}
		err = rows.Scan(&item.Id, &item.Name)
		if err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	return items, nil
}

func (u *Products) Suggest(ctx context.Context, text string, limit int) ([]suggestion.Suggestion, error) {
	sg := suggestion.NewSuggestions(u.storage)
	return sg.GetSuggestion(ctx, tableProducts, text, limit)
}

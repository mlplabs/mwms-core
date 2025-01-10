package whs

import (
	"context"
	"errors"
	"fmt"
	"github.com/lib/pq"
	"github.com/mlplabs/mwms-core/whs/model"
)

type Products struct {
	storage *Storage
}

func NewProducts(s *Storage) *Products {
	return &Products{storage: s}
}

// Get returns a list items without limit
func (u *Products) Get(ctx context.Context) ([]model.Product, error) {
	items := make([]model.Product, 0)
	sqlSel := `SELECT 
    				p.id, p.name, p.item_number, p.manufacturer_id, coalesce(m.name, '') as manufacturer_name 
				FROM products p
				LEFT JOIN manufacturers m ON p.manufacturer_id = m.id
				ORDER BY p.name ASC`

	rows, err := u.storage.Db.QueryContext(ctx, sqlSel)
	if err != nil {
		return items, err
	}
	defer rows.Close()

	for rows.Next() {
		item := model.Product{Manufacturer: model.Manufacturer{}}
		err = rows.Scan(&item.Id, &item.Name, &item.ItemNumber, &item.Manufacturer.Id, &item.Manufacturer.Name)
		items = append(items, item)
	}
	return items, nil
}

// GetItems returns a list items of catalog with limit & offset
func (u *Products) GetItems(ctx context.Context, offset int, limit int, search string) ([]model.Product, int64, error) {
	var totalCount int64
	var sqlCond string
	items := make([]model.Product, 0)
	if search != "" {
		sqlCond = "WHERE p.name ILIKE '%" + search + "%' OR m.name ILIKE '" + search + "%' OR item_number ILIKE '" + search + "%'"
	}
	args := make([]any, 0)

	if limit == 0 {
		limit = DefaultRowsLimit
	}
	args = append(args, limit)
	args = append(args, offset)
	query := "SELECT p.id, p.name, p.item_number, p.manufacturer_id, coalesce(m.name, '') as manufacturer_name " +
		"	FROM products p " +
		"   LEFT JOIN manufacturers m ON p.manufacturer_id = m.id" +
		"   %s " +
		"	ORDER BY p.name ASC"
	sqlSel := fmt.Sprintf(query, sqlCond)

	rows, err := u.storage.Db.QueryContext(ctx, sqlSel+" LIMIT $1 OFFSET $2", args...)
	if err != nil {
		return items, totalCount, err
	}
	defer rows.Close()

	for rows.Next() {
		item := model.Product{Manufacturer: model.Manufacturer{}}
		err = rows.Scan(&item.Id, &item.Name, &item.ItemNumber, &item.Manufacturer.Id, &item.Manufacturer.Name)
		items = append(items, item)
	}

	sqlCount := fmt.Sprintf("SELECT COUNT(*) as count FROM ( %s ) sub", sqlSel)
	err = u.storage.Db.QueryRowContext(ctx, sqlCount).Scan(&totalCount)
	if err != nil {
		return items, totalCount, err
	}
	return items, totalCount, nil
}

func (u *Products) Create(ctx context.Context, product *model.Product) (int64, error) {
	var insertId int64
	sqlCreate := `INSERT INTO products (name, item_number, manufacturer_id) VALUES ($1, $2, $3) RETURNING id`
	err := u.storage.Db.QueryRowContext(ctx, sqlCreate, product.Name, product.ItemNumber, product.Manufacturer.Id).Scan(&insertId)
	return insertId, err
}

func (u *Products) Update(ctx context.Context, product *model.Product) (int64, error) {
	sqlUpd := `UPDATE products SET name=$2, item_number=$3, manufacturer_id=$4 WHERE id=$1`
	res, err := u.storage.Db.ExecContext(ctx, sqlUpd, product.Id, product.Name, product.ItemNumber, product.Manufacturer.Id)
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
	sqlDel := `DELETE FROM products WHERE id=$1`
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

func (u *Products) GetById(ctx context.Context, itemId int64) (*model.Product, error) {
	sqlSel := `SELECT p.id, p.name, p.item_number, p.manufacturer_id, coalesce(m.name, '') as manufacturer_name 
				FROM products p 
				LEFT JOIN public.manufacturers m on m.id = p.manufacturer_id
				WHERE p.id = $1`
	row := u.storage.Db.QueryRowContext(ctx, sqlSel, itemId)
	newItem := model.Product{Manufacturer: model.Manufacturer{}}
	err := row.Scan(&newItem.Id, &newItem.Name, &newItem.ItemNumber, &newItem.Manufacturer.Id, &newItem.Manufacturer.Name)
	if err != nil {
		return nil, err
	}
	return &newItem, nil
}

func (u *Products) FindByName(ctx context.Context, itemName string) ([]model.Product, error) {
	items := make([]model.Product, 0)
	sql := `SELECT p.id, p.name, p.item_number, p.manufacturer_id, coalesce(m.name, '') as manufacturer_name
			FROM products p 
			LEFT JOIN public.manufacturers m on m.id = p.manufacturer_id
			WHERE p.name = $1`
	rows, err := u.storage.Db.QueryContext(ctx, sql, itemName)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	for rows.Next() {
		item := model.Product{Manufacturer: model.Manufacturer{}}
		err = rows.Scan(&item.Id, &item.Name, &item.ItemNumber, &item.Manufacturer.Id, &item.Manufacturer.Name)
		if err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	return items, nil
}

// FindByBarcode returns a product by barcode
func (u *Products) FindByBarcode(ctx context.Context, itemName string) ([]model.Product, error) {
	items := make([]model.Product, 0)
	sqlQuery := `SELECT p.id, p.name, p.item_number, p.manufacturer_id, coalesce(m.name, '') as manufacturer_name
					FROM products p
					LEFT JOIN public.manufacturers m on p.manufacturer_id = m.id
					WHERE p.id IN (
    					SELECT b.owner_id FROM barcodes b WHERE b.owner_ref='products' AND b.name = $1)`
	rows, err := u.storage.Db.QueryContext(ctx, sqlQuery, itemName)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	for rows.Next() {
		item := model.Product{}
		err = rows.Scan(&item.Id, &item.Name, &item.ItemNumber, &item.Manufacturer.Id, &item.Manufacturer.Name)
		if err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	return items, nil
}
func (u *Products) Suggest(ctx context.Context, text string, limit int) ([]model.Suggestion, error) {
	sg := NewSuggestions(u.storage)
	return sg.GetSuggestion(ctx, "products", text, limit)
}

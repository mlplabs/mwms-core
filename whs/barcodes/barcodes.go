package barcodes

import (
	"context"
	"errors"
	"fmt"
	"github.com/lib/pq"
	"github.com/mlplabs/mwms-core/whs"
	"github.com/mlplabs/mwms-core/whs/suggestion"
)

const tableBarcodes = "barcodes"

type Barcodes struct {
	storage *whs.Storage
}

func NewBarcodes(s *whs.Storage) *Barcodes {
	return &Barcodes{storage: s}
}

// Get returns a list of barcodes
func (b *Barcodes) Get(ctx context.Context) ([]Barcode, error) {
	items := make([]Barcode, 0)
	sqlSel := fmt.Sprintf("SELECT id, name, barcode_type, owner_id, owner_ref FROM %s ORDER BY name ASC", tableBarcodes)

	rows, err := b.storage.Db.QueryContext(ctx, sqlSel)
	if err != nil {
		return items, err
	}
	defer rows.Close()

	for rows.Next() {
		bc := Barcode{}
		err = rows.Scan(&bc.Id, &bc.Name, &bc.Type, &bc.OwnerId, &bc.OwnerRef)
		items = append(items, bc)
	}
	return items, nil
}

// GetItems returns a list of barcodes
func (b *Barcodes) GetItems(ctx context.Context, offset int, limit int, ownerId int64, ownerRef string) ([]Barcode, int64, error) {
	var totalCount int64
	items := make([]Barcode, 0)

	sqlCond := "WHERE owner_id = $3 AND owner_ref = $4"
	args := make([]any, 0)

	if limit == 0 {
		limit = whs.DefaultRowsLimit
	}
	args = append(args, limit)
	args = append(args, offset)
	args = append(args, ownerId)
	args = append(args, ownerRef)

	sqlSel := fmt.Sprintf("SELECT id, name, barcode_type, owner_id, owner_ref FROM %s %s ORDER BY name ASC", tableBarcodes, sqlCond)

	rows, err := b.storage.Db.QueryContext(ctx, sqlSel+" LIMIT $1 OFFSET $2", args...)
	if err != nil {
		return items, totalCount, err
	}
	defer rows.Close()

	for rows.Next() {
		bc := Barcode{}
		err = rows.Scan(&bc.Id, &bc.Name, &bc.Type, &bc.OwnerId, &bc.OwnerRef)
		items = append(items, bc)
	}

	sqlCount := fmt.Sprintf("SELECT COUNT(*) as count FROM ( %s ) sub", sqlSel)
	err = b.storage.Db.QueryRow(sqlCount).Scan(&totalCount)
	if err != nil {
		return items, totalCount, err
	}
	return items, totalCount, nil
}

func (b *Barcodes) Create(ctx context.Context, bc *Barcode) (int64, error) {
	var insertId int64
	sqlCreate := fmt.Sprintf("INSERT INTO %s (name, barcode_type, owner_id, owner_ref) VALUES ($1, $2, $3) RETURNING id", tableBarcodes)
	err := b.storage.Db.QueryRowContext(ctx, sqlCreate, bc.Name, bc.Type, bc.OwnerId, bc.OwnerRef).Scan(&insertId)
	if err != nil {
		return insertId, err
	}
	return insertId, nil
}

func (b *Barcodes) Update(ctx context.Context, bc *Barcode) (int64, error) {
	sqlUpd := fmt.Sprintf("UPDATE %s SET name=$2, barcode_type=$3, owner_id=$4, owner_ref=$5 WWHERE id=$1", tableBarcodes)
	res, err := b.storage.Db.ExecContext(ctx, sqlUpd, bc.Id, bc.Name, bc.Type, bc.OwnerId, bc.OwnerRef)
	if err != nil {
		return 0, err
	}
	if a, err := res.RowsAffected(); a != 1 || err != nil {
		return 0, err
	}
	return bc.Id, nil
}

func (b *Barcodes) Delete(ctx context.Context, itemId int64) error {
	if itemId == 0 {
		return fmt.Errorf("unacceptable action. item id eq 0")
	}
	sqlDel := fmt.Sprintf("DELETE FROM %s WHERE id=$1", tableBarcodes)
	_, err := b.storage.Db.ExecContext(ctx, sqlDel, itemId)
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

func (b *Barcodes) GetById(ctx context.Context, itemId int64) (*Barcode, error) {
	sqlUsr := fmt.Sprintf("SELECT id, name, barcode_type, owner_id, owner_ref FROM %s WHERE id = $1", tableBarcodes)
	row := b.storage.Db.QueryRowContext(ctx, sqlUsr, itemId)
	bc := Barcode{}
	err := row.Scan(&bc.Id, &bc.Name, &bc.Type, &bc.OwnerId, &bc.OwnerRef)
	if err != nil {
		return nil, err
	}
	return &bc, nil
}

func (b *Barcodes) FindByName(ctx context.Context, itemName string) ([]Barcode, error) {
	items := make([]Barcode, 0)
	sql := fmt.Sprintf("SELECT id, name, barcode_type, owner_id, owner_ref FROM %s WHERE name = $1", tableBarcodes)
	rows, err := b.storage.Db.QueryContext(ctx, sql, itemName)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	for rows.Next() {
		bc := Barcode{}
		err = rows.Scan(&bc.Id, &bc.Name, &bc.Type, &bc.OwnerId, &bc.OwnerRef)
		if err != nil {
			return nil, err
		}
		items = append(items, bc)
	}
	return items, nil
}

// FindByOwnerId returns a list of barcodes for the product (owner)
func (b *Barcodes) FindByOwnerId(ctx context.Context, ownerId int64) ([]Barcode, error) {
	retBc := make([]Barcode, 0)
	sql := fmt.Sprintf("SELECT id, name, barcode_type, owner_id FROM %s WHERE owner_id = $1", tableBarcodes)
	rows, err := b.storage.Db.QueryContext(ctx, sql, ownerId)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	for rows.Next() {
		bci := Barcode{}
		err = rows.Scan(&bci.Id, &bci.Name, &bci.Type, &bci.OwnerId)
		if err != nil {
			return nil, err
		}
		retBc = append(retBc, bci)
	}
	return retBc, nil
}

func (b *Barcodes) Suggest(ctx context.Context, text string, limit int) ([]suggestion.Suggestion, error) {
	retVal := make([]suggestion.Suggestion, 0)
	if limit == 0 {
		limit = whs.DefaultSuggestionLimit
	}

	sqlSel := fmt.Sprintf("SELECT id, name FROM %s WHERE name ILIKE $1 LIMIT $2", tableBarcodes)
	rows, err := b.storage.Db.QueryContext(ctx, sqlSel, text+"%", limit)
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

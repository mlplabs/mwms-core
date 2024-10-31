package whs

import (
	"database/sql"
	"fmt"
	"github.com/mlplabs/microwms-core/core"
)

// Barcode barcode object
type Barcode struct {
	Catalog
}

type BarcodeItem struct {
	CatalogItem
	Type    int   `json:"type"`
	OwnerId int64 `json:"product_id"`
}

func (s *Storage) GetBarcode() *Barcode {
	m := new(Barcode)
	m.table = tableRefBarcodes
	m.setStorage(s)
	return m
}

// GetItems returns a list of barcodes
func (b *Barcode) GetItems(offset int, limit int) ([]ICatalogItem, int, error) {
	var count int
	sqlCond := ""
	args := make([]any, 0)

	if limit == 0 {
		limit = DefaultRowsLimit
	}
	args = append(args, limit)
	args = append(args, offset)

	sqlSel := fmt.Sprintf("SELECT id, name, barcode_type, owner_id FROM %s %s ORDER BY name ASC", b.getTableName(), sqlCond)

	rows, err := b.getDb().Query(sqlSel+" LIMIT $1 OFFSET $2", args...)
	if err != nil {
		return nil, count, err
	}
	defer rows.Close()

	items := make([]ICatalogItem, count, limit)
	for rows.Next() {
		item := new(BarcodeItem)
		item.catalog = b

		err = rows.Scan(&item.Id, &item.Name)
		items = append(items, item)
	}

	sqlCount := fmt.Sprintf("SELECT COUNT(*) as count FROM ( %s ) sub", sqlSel)
	err = b.getDb().QueryRow(sqlCount).Scan(&count)
	if err != nil {
		return nil, count, err
	}
	return items, count, nil
}

// FindByOwnerId returns a list of barcodes for the product (owner)
func (b *Barcode) FindByOwnerId(ownerId int64) ([]BarcodeItem, error) {
	retBc := make([]BarcodeItem, 0)
	sql := fmt.Sprintf("SELECT id, name, barcode_type, owner_id FROM %s WHERE owner_id = $1", b.getTableName())
	rows, err := b.getDb().Query(sql, ownerId)
	if err != nil {
		return nil, &core.WrapError{Err: err, Code: core.SystemError}
	}
	defer rows.Close()
	for rows.Next() {
		bci := BarcodeItem{}
		err = rows.Scan(&bci.Id, &bci.Name, &bci.Type, &bci.OwnerId)
		if err != nil {
			return nil, &core.WrapError{Err: err, Code: core.SystemError}
		}
		retBc = append(retBc, bci)
	}
	return retBc, nil
}

// FindByName returns a list of barcodes by name
// by value only without type and binding
func (b *Barcode) FindByName(bcName string) ([]BarcodeItem, error) {
	retBc := make([]BarcodeItem, 0)

	sqlSel := fmt.Sprintf("SELECT id, name, barcode_type, owner_id FROM %s WHERE name = $1", b.table)
	rows, err := b.storage.Db.Query(sqlSel, bcName)
	if err != nil {
		return nil, &core.WrapError{Err: err, Code: core.SystemError}
	}
	defer rows.Close()
	for rows.Next() {
		bc := new(BarcodeItem)
		err = rows.Scan(&bc.Id, &bc.Name, &bc.Type, &bc.OwnerId)
		if err != nil {
			return nil, &core.WrapError{Err: err, Code: core.SystemError}
		}
		retBc = append(retBc, *bc)
	}
	return retBc, nil
}

func (b *Barcode) GetNewItem() (*BarcodeItem, error) {
	item := new(BarcodeItem)
	item.setCatalog(b)
	return item, nil
}

func (bi *BarcodeItem) Store() (int64, int64, error) {
	return bi.StoreTx(nil)
}

func (bi *BarcodeItem) StoreTx(tx *sql.Tx) (int64, int64, error) {
	var err error
	var res sql.Result

	if bi.GetId() == 0 {
		var insertId int64
		sqlCreate := fmt.Sprintf("INSERT INTO %s (name, barcode_type, owner_id) VALUES ($1, $2, $3) RETURNING id", bi.getTableName())
		if tx != nil {
			err = tx.QueryRow(sqlCreate, bi.GetName(), bi.Type, bi.OwnerId).Scan(&insertId)
		} else {
			err = bi.getDb().QueryRow(sqlCreate, bi.GetName(), bi.Type, bi.OwnerId).Scan(&insertId)
		}
		bi.Id = insertId
		if err != nil {
			return 0, 0, err
		}
		return bi.GetId(), 1, nil
	} else {
		sqlUpd := fmt.Sprintf("UPDATE %s SET name=$2, barcode_type=$3, owner_id=$4 WWHERE id=$1", bi.getTableName())
		if tx != nil {
			res, err = tx.Exec(sqlUpd, bi.GetId(), bi.GetName(), bi.Type, bi.OwnerId)
		} else {
			res, err = bi.getDb().Exec(sqlUpd, bi.GetId(), bi.GetName(), bi.Type, bi.OwnerId)
		}
		if err != nil {
			return 0, 0, err
		}
		if a, err := res.RowsAffected(); a != 1 || err != nil {
			return 0, 0, err
		}
		return bi.GetId(), 1, nil
	}
}

func (bi *BarcodeItem) GetOwner() (ICatalogItem, error) {
	catProd := bi.getStorage().GetProduct()
	return catProd.FindById(bi.OwnerId)
}

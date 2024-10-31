package whs

import (
	"database/sql"
	"fmt"
	"github.com/mlplabs/microwms-core/core"
	"time"
)

type IDocumentItem interface {
	setDocument(IDocument)
	GetId() int64
	GetType() int
	Store() (int64, int64, error)
	Delete() (int64, error)
}

type DocumentItem struct {
	document IDocument
	Id       int64    `json:"id"`
	Number   string   `json:"number"`
	Date     string   `json:"date"`
	Type     int      `json:"doc_type"`
	WhsId    int64    `json:"whs_id"`
	Rows     []DocumentRow `json:"rows"`
}

// DocumentRow product line of the document
type DocumentRow struct {
	RowId    string      `json:"row_id"`
	Product  ProductItem `json:"product"`
	Quantity int         `json:"quantity"`
	CellSrc  CellItem    `json:"cell_src"` // from
	CellDst  CellItem    `json:"cell_dst"` // to
}

func (di *DocumentItem) Delete() (int64, error) {
	sqlDelI := fmt.Sprintf("DELETE FROM %s WHERE parent_id = $1", di.getRowTable())
	sqlDelH := fmt.Sprintf("DELETE FROM %s WHERE id = $1", di.getHeadTable())
	tx, err := di.getDb().Begin()
	if err != nil {
		return 0, err
	}
	_, err = tx.Exec(sqlDelI, di)
	if err != nil {
		tx.Rollback()
		return 0, err
	}
	res, err := tx.Exec(sqlDelH, di)
	if err != nil {
		tx.Rollback()
		return 0, err
	}
	affRows, err := res.RowsAffected()
	if err != nil {
		tx.Rollback()
		return 0, err
	}

	err = tx.Commit()
	if err != nil {
		tx.Rollback()
		return 0, err
	}
	return affRows, nil
}

func (di *DocumentItem) setDocument(d IDocument) {
	di.document = d
}

func (di *DocumentItem) getDb() *sql.DB {
	return di.document.getDb()
}

func (di *DocumentItem) getHeadTable() string {
	return di.document.getHeadTable()
}

func (di *DocumentItem) getRowTable() string {
	return di.document.getRowTable()
}

func (di *DocumentItem) GetId() int64 {
	return di.Id
}

func (di *DocumentItem) GetType() int {
	return di.Type
}

func (di *DocumentItem) GetNumber() string {
	return fmt.Sprintf("%06d.%d", di.Id, di.Type)
}

func (di *DocumentItem) GetDate(date time.Time) string {
	// TODO: пока так
	return date.Format("02.01.2006 15:04:05")
}

func (di *DocumentItem) Store() (int64, int64, error) {
	if di.Date == ""{
		di.Date = time.Now().Format("2006-01-02")
	}
	if di.GetId() == 0 {
		tx, err := di.getDb().Begin()
		if err != nil {
			return 0, 0, err
		}
		sqlInsH := fmt.Sprintf("INSERT INTO %s (number, date, doc_type) VALUES($1, $2, $3) RETURNING id", di.getHeadTable())
		err = tx.QueryRow(sqlInsH, di.Number, di.Date, DocumentTypeReceipt).Scan(&di.Id)
		if err != nil {
			tx.Rollback()
			return 0, 0, err
		}
		for idx, item := range di.Rows {
			if item.Product.Id == 0 {
				if item.Product.Manufacturer.Id == 0 {
					sqlInsMnf := fmt.Sprintf("INSERT INTO %s (name) VALUES ($1) RETURNING id", tableRefManufacturers)
					err = tx.QueryRow(sqlInsMnf, item.Product.Manufacturer.Name).Scan(&item.Product.Manufacturer.Id)
					if err != nil {
						tx.Rollback()
						return 0, 0, err
					}
				}
				sqlInsP := fmt.Sprintf("INSERT INTO %s (name, manufacturer_id) VALUES($1, $2) RETURNING id", tableRefProducts)
				err = tx.QueryRow(sqlInsP, item.Product.Name, item.Product.Manufacturer.Id).Scan(&item.Product.Id)
				if err != nil {
					tx.Rollback()
					return 0, 0, &core.WrapError{Err: err, Code: core.SystemError}
				}
			}
			item.RowId = fmt.Sprintf("%d.%d", di.Id, idx+1)
			sqlInsI := fmt.Sprintf("INSERT INTO %s (parent_id, row_id, product_id, quantity) VALUES($1, $2, $3, $4)", di.getHeadTable())
			_, err = tx.Exec(sqlInsI, di.Id, item.RowId, item.Product.Id, item.Quantity)
			if err != nil {
				tx.Rollback()
				return 0, 0, &core.WrapError{Err: err, Code: core.SystemError}
			}

			c := CellItem{
				CatalogItem: CatalogItem{
					Id: 2,
				},
				WhsId:  1,
				ZoneId: 1,
			}
			s := Storage{Db: di.getDb()}
			item.CellDst = c

			_, err = s.PutRow(di, &item, tx)
			if err != nil {
				tx.Rollback()
				return 0, 0, &core.WrapError{Err: err, Code: core.SystemError}

			}
		}
		err = tx.Commit()
		if err != nil {
			return 0, 0, &core.WrapError{Err: err, Code: core.SystemError}
		}

		return di.Id, 1, nil
	} else {
		tx, err := di.getDb().Begin()
		if err != nil {
			return 0, 0, &core.WrapError{Err: err, Code: core.SystemError}
		}
		sqlUpdH := fmt.Sprintf("UPDATE %s SET number = $1, date = $2, doc_type = $3 WHERE id = $4", di.getHeadTable())
		_, err = tx.Exec(sqlUpdH, di.Number, di.Date, DocumentTypePosting, di.Id)
		if err != nil {
			tx.Rollback()
			return 0, 0, &core.WrapError{Err: err, Code: core.SystemError}
		}

		sqlDelProd := fmt.Sprintf("DELETE FROM %s WHERE parent_id = $1", di.getRowTable())
		_, err = tx.Exec(sqlDelProd, di.Id)
		if err != nil {
			tx.Rollback()
			return 0, 0, &core.WrapError{Err: err, Code: core.SystemError}
		}

		for idx, row := range di.Rows {
			if row.Product.Id == 0 {
				if row.Product.Manufacturer.Id == 0 {
					sqlInsMnf := fmt.Sprintf("INSERT INTO %s (name) VALUES ($1) RETURNING id", tableRefManufacturers)
					err = tx.QueryRow(sqlInsMnf, row.Product.Manufacturer.Name).Scan(&row.Product.Manufacturer.Id)
					if err != nil {
						tx.Rollback()
						return 0, 0, &core.WrapError{Err: err, Code: core.SystemError}
					}
				}
				sqlInsP := fmt.Sprintf("INSERT INTO %s (name, manufacturer_id) VALUES($1, $2) RETURNING id", tableRefProducts)
				err = tx.QueryRow(sqlInsP, row.Product.Name, row.Product.Manufacturer.Id).Scan(&row.Product.Id)
				if err != nil {
					tx.Rollback()
					return 0, 0, &core.WrapError{Err: err, Code: core.SystemError}
				}
			}
			RowId := fmt.Sprintf("%d.%d", di.Id, idx+1)
			sqlInsI := fmt.Sprintf("INSERT INTO %s (parent_id, row_id, product_id, quantity) VALUES($1, $2, $3, $4)", di.getRowTable())
			_, err = tx.Exec(sqlInsI, di.Id, RowId, row.Product.Id, row.Quantity)
			if err != nil {
				tx.Rollback()
				return 0, 0, &core.WrapError{Err: err, Code: core.SystemError}
			}
		}
		err = tx.Commit()
		if err != nil {
			return 0, 0, &core.WrapError{Err: err, Code: core.SystemError}
		}

		return di.Id, 1, nil
	}
}

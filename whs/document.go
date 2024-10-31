package whs

import (
	"database/sql"
	"fmt"
	"time"
)

const (
	DocumentTypePosting  = 1
	DocumentTypeReceipt  = 2
	DocumentTypeShipment = 3
	DocumentTypeWriteOff = 4
)

type IDocument interface {
	GetDocumentMetadata()
	getDb() *sql.DB
	getHeadTable() string
	getRowTable() string
}

// IDocStorage документ связан с движением товара
type IDocStorage interface {
	GetWhs() *WhsItem
}

type Document struct {
	storage   *Storage
	headTable string
	rowsTable string
	documentType int
}

func (d *Document) GetDocumentMetadata() {
	return
}

func (d *Document) setStorage(s *Storage) {
	d.storage = s
}

func (d *Document) getStorage() *Storage {
	return d.storage
}

func (d *Document) getDb() *sql.DB {
	return d.storage.Db
}

func (d *Document) getHeadTable() string {
	return d.headTable
}

func (d *Document) getRowTable() string {
	return d.rowsTable
}

// GetItems returns a list of documents (without goods)
func (d *Document) GetItems(offset int, limit int) ([]IDocumentItem, int, error) {
	var count int
	sqlCond := "WHERE doc_type = $1"

	args := make([]any, 0)

	if limit == 0 {
		limit = DefaultRowsLimit
	}
	args = append(args, d.documentType)
	args = append(args, limit)
	args = append(args, offset)

	sqlSel := fmt.Sprintf("SELECT id, number, date, doc_type, whs_id FROM %s %s ORDER BY date ASC", d.getHeadTable(), sqlCond)

	rows, err := d.getDb().Query(sqlSel+" LIMIT $2 OFFSET $3", args...)
	if err != nil {
		return nil, count, err
	}
	defer rows.Close()

	items := make([]IDocumentItem, count, limit)
	for rows.Next() {
		item, _ := d.GetNewItem()
		dateDoc := time.Time{}
		err = rows.Scan(&item.Id, &item.Number, &dateDoc, &item.Type, &item.WhsId)
		item.Number = item.GetNumber()
		item.Date = item.GetDate(dateDoc)
		items = append(items, item)
	}

	sqlCount := fmt.Sprintf("SELECT COUNT(*) as count FROM ( %s ) sub", sqlSel)
	err = d.getDb().QueryRow(sqlCount, d.documentType).Scan(&count)
	if err != nil {
		return nil, count, err
	}
	return items, count, nil
}

// FindByNumberDate finds a document by number and date (without goods)
func (d *Document) FindByNumberDate(number string, date time.Time) (*DocumentItem, error) {
	sqlUsr := fmt.Sprintf("SELECT id, number, date, doc_type FROM %s WHERE number = $1 AND date::date >= $2::date", d.getHeadTable())
	row := d.getDb().QueryRow(sqlUsr, number, date)
	u := new(DocumentItem)
	dateDoc := time.Time{}
	err := row.Scan(&u.Id, &u.Number, &dateDoc, &u.Type)
	u.Number = u.GetNumber()
	u.Date = u.GetDate(dateDoc)

	if err != nil {
		return nil, err
	}
	return u, nil
}

// FindById finds a document by id (without goods)
func (d *Document) FindById(id int64) (IDocumentItem, error) {
	sqlSel := fmt.Sprintf("SELECT id, number, date, doc_type, whs_id FROM %s WHERE id = $1", d.getHeadTable())
	row := d.getDb().QueryRow(sqlSel, id)

	di := new(DocumentItem)
	dateDoc := time.Time{}
	err := row.Scan(&di.Id, &di.Number, &dateDoc, &di.Type, &di.WhsId)
	if err != nil {
		return nil, err
	}
	di.Number = di.GetNumber()
	di.Date = di.GetDate(dateDoc)
	return di, nil
}

// GetById finds a document by id (with goods)
func (d *Document) GetById(id int64) (IDocumentItem, error) {
	sqlSel := fmt.Sprintf("SELECT id, number, date, doc_type, whs_id FROM %s WHERE id = $1", d.getHeadTable())
	row := d.getDb().QueryRow(sqlSel, id)
	di := new(DocumentItem)
	dateDoc := time.Time{}
	err := row.Scan(&di.Id, &di.Number, &dateDoc, &di.Type, &di.WhsId)
	di.Number = di.GetNumber()
	di.Date = di.GetDate(dateDoc)
	if err != nil {
		return nil, err
	}

	// Тут надо сделать выборку товаров для базового варианта документа на основе просто детальной талблицы getRowTable()


	//sqlRows := fmt.Sprintf("SELECT st.row_id, st.prod_id, p.name, "+
	//	"	p.manufacturer_id, COALESCE(m.name, '') AS manufacturer_name, "+
	//	"	st.quantity, st.cell_id, COALESCE(c.name, '') AS cell_name "+
	//	"		FROM storage%d st "+
	//	"	LEFT JOIN products p ON st.prod_id = p.id "+
	//	"	LEFT JOIN manufacturers m ON p.manufacturer_id = m.id "+
	//	"	LEFT JOIN cells c ON st.cell_id = c.id "+
	//	"	WHERE doc_id = $1", di.WhsId)
	//rows, err := d.getDb().Query(sqlRows, id)
	//if err != nil {
	//	return nil, err
	//}
	//defer rows.Close()
	//for rows.Next() {
	//	r := DocumentRow{}
	//	cellId := 0
	//	cellName := ""
	//	err = rows.Scan(&r.RowId, &r.Product.Id, &r.Product.Name, &r.Product.Manufacturer.Id, &r.Product.Manufacturer.Name, &r.Quantity, &cellId, &cellName)
	//	if di.GetType() == DocumentTypeReceipt {
	//		r.CellDst.Id = int64(cellId)
	//		r.CellDst.Name = cellName
	//	}
	//	di.Rows = append(di.Rows, r)
	//}
	return di, nil
}

func (d *Document) GetProductItems(offset int, limit int) ([]ProductItem, error) {
	//
	//	var count int
	//	sqlCond := "WHERE s.doc_type = $1"
	//
	//	args := make([]any, 0)
	//
	//	if limit == 0 {
	//		limit = 10
	//	}
	//
	//	args = append(args, DocumentTypeReceipt)
	//	args = append(args, limit)
	//	args = append(args, offset)
	//
	//	sqlSel := fmt.Sprintf("SELECT s.doc_id, r.number, to_char(r.date, 'DD.MM.YYYY'), s.prod_id, coalesce(p.name, 'unknown'), m.id AS mnf_id, m.name AS mnf_name, s.quantity FROM storage1 s "+
	//		"	LEFT JOIN receipt_headers r on r.id = s.doc_id "+
	//		"	LEFT JOIN products p on p.id = s.prod_id "+
	//		"	LEFT JOIN manufacturers m on m.id = p.manufacturer_id "+
	//		"	%s "+
	//		"	ORDER BY row_time DESC ", sqlCond)
	//
	//	rows, err := s.Db.Query(sqlSel+" LIMIT $2 OFFSET $3", args...)
	//	if err != nil {
	//		return nil, count, &core.WrapError{Err: err, Code: core.SystemError}
	//	}
	//	defer rows.Close()
	//
	//	items := make([]TurnoversRow, count, 10)
	//	for rows.Next() {
	//		item := new(TurnoversRow)
	//		err = rows.Scan(&item.Doc.Id, &item.Doc.Number, &item.Doc.Date, &item.Product.Id, &item.Product.Name, &item.Product.Manufacturer.Id, &item.Product.Manufacturer.Name, &item.Quantity)
	//		items = append(items, *item)
	//	}
	//
	//	sqlCount := fmt.Sprintf("SELECT COUNT(*) as count FROM ( %s ) sub", sqlSel)
	//	err = s.Db.QueryRow(sqlCount, DocumentTypeReceipt).Scan(&count)
	//	if err != nil {
	//		return nil, count, &core.WrapError{Err: err, Code: core.SystemError}
	//	}
	//	return items, count, nil
	return nil, nil
}

func (d *Document) GetNewItem() (*DocumentItem, error) {
	item := new(DocumentItem)
	item.setDocument(d)
	item.Rows = make([]DocumentRow, 0)
	return item, nil
}


package whs

import (
	"fmt"
	"time"
)

type DocReceipt struct {
	Document
}

func (s *Storage) GetDocumentReceipt() *DocReceipt {
	d := new(DocReceipt)
	d.headTable = "receipt_headers"
	d.documentType = DocumentTypeReceipt
	d.setStorage(s)
	return d
}


// GetById finds a custom receipt document by id (with goods from storage)
func (d *DocReceipt) GetById(id int64) (IDocumentItem, error) {
	sqlSel := fmt.Sprintf("SELECT id, number, date, doc_type, whs_id FROM %s WHERE id = $1", d.getHeadTable())
	row := d.getDb().QueryRow(sqlSel, id)
	di := new(DocumentItem)
	dateDoc := time.Time{}
	err := row.Scan(&di.Id, &di.Number, &dateDoc, &di.Type, &di.WhsId)
	di.Number = di.GetNumber()
	di.Date = di.GetDate(dateDoc)
	sqlRows := fmt.Sprintf("SELECT st.row_id, st.prod_id, p.name, "+
		"	p.manufacturer_id, COALESCE(m.name, '') AS manufacturer_name, "+
		"	st.quantity, st.cell_id, COALESCE(c.name, '') AS cell_name "+
		"		FROM storage%d st "+
		"	LEFT JOIN products p ON st.prod_id = p.id "+
		"	LEFT JOIN manufacturers m ON p.manufacturer_id = m.id "+
		"	LEFT JOIN cells c ON st.cell_id = c.id "+
		"	WHERE doc_id = $1", di.WhsId)
	rows, err := d.getDb().Query(sqlRows, id)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	for rows.Next() {
		r := DocumentRow{}
		cellId := 0
		cellName := ""
		err = rows.Scan(&r.RowId, &r.Product.Id, &r.Product.Name, &r.Product.Manufacturer.Id, &r.Product.Manufacturer.Name, &r.Quantity, &cellId, &cellName)
		r.CellDst.Id = int64(cellId)
		r.CellDst.Name = cellName
		di.Rows = append(di.Rows, r)
	}
	return di, nil
}

///*
//GetReceiptItems - returns a list of goods in receipt documents
//*/
//func (s *Storage) GetReceiptItems(offset int, limit int) ([]TurnoversRow, int, error) {
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
//}

///*
//CreateReceiptDoc - writes the document to the local database
//В новой концепции, при отсутствии документов по умолчанию, будем создавать 1 документ на дату
//И вывод будем осуществлять в иерархии дат
//*/
//func (s *Storage) CreateReceiptDoc(doc *DocItem) (int64, error) {
//	doc.DocType = DocumentTypeReceipt
//
//	if doc.Date == "" {
//		doc.Date = time.Now().Format("2006-01-02")
//	}
//	// Принцип такой
//	// Без документального учета при записи создаем документ на текущую дату и на него "вешаем" товары
//	// в результате должны получить одну строку документа в день (на пользователя)
//	targetDate := time.Now()
//	if doc.Date != "" {
//		targetDate, _ = time.Parse("2006-01-02", doc.Date)
//	}
//
//	tx, err := s.Db.Begin()
//	if err != nil {
//		return 0, &core.WrapError{Err: err, Code: core.SystemError}
//	}
//
//	_d, _ := s.FindReceiptDocByNumberDate(doc.Number, targetDate)
//	if _d != nil {
//		doc.Id = _d.Id
//	} else {
//		sqlInsH := fmt.Sprintf("INSERT INTO %s (number, date, doc_type) VALUES($1, $2, $3) RETURNING id", tableDocReceiptHeaders)
//		err = tx.QueryRow(sqlInsH, doc.Number, doc.Date, DocumentTypeReceipt).Scan(&doc.Id)
//		if err != nil {
//			tx.Rollback()
//			return 0, &core.WrapError{Err: err, Code: core.SystemError}
//		}
//	}
//
//	//for idx, item := range doc.Items {
//	//
//	//	pId, _, err := s.CreateProductInteractive(tx, item.Product.Name, item.Product.Manufacturer.Name, item.Product.ItemNumber, nil, nil)
//	//
//	//	if err != nil {
//	//		tx.Rollback()
//	//		return 0, &core.WrapError{Err: err, Code: core.SystemError}
//	//	}
//	//
//	//	item.Product.Id = pId
//	//	item.RowId = fmt.Sprintf("%d.%d", doc.Id, idx+1)
//	//
//	//	c := Cell{Id: 2, WhsId: 1, ZoneId: 1}
//	//	s := Storage{Db: s.Db}
//	//	item.CellDst = c
//	//
//	//	_, err = s.PutRow(doc, &item, tx)
//	//	if err != nil {
//	//		tx.Rollback()
//	//		return 0, &core.WrapError{Err: err, Code: core.SystemError}
//	//
//	//	}
//	//}
//	err = tx.Commit()
//	if err != nil {
//		return 0, &core.WrapError{Err: err, Code: core.SystemError}
//	}
//
//	return doc.Id, nil
//}

//func (s *Storage) GetReceiptDocById(id int64) (*DocItem, error) {
//	return s.GetDocument(docTables{
//		Headers: tableDocReceiptHeaders,
//		Items:   tableDocReceiptItems,
//	}).getItemById(id)
//}

//func (s *Storage) FindReceiptDocById(id int64) (*DocItem, error) {
//	return s.GetDocument(docTables{
//		Headers: tableDocReceiptHeaders,
//		Items:   tableDocReceiptItems,
//	}).findItemById(id)
//}

//func (s *Storage) FindReceiptDocByNumberDate(number string, date time.Time) (*DocItem, error) {
//	return s.GetDocument(docTables{
//		Headers: tableDocReceiptHeaders,
//		Items:   tableDocReceiptItems,
//	}).findItemByNumberDate(number, date)
//}

//func (s *Storage) UpdateReceiptDoc(doc *DocItem) (int64, error) {
//	return s.GetDocument(docTables{
//		Headers: tableDocReceiptHeaders,
//		Items:   tableDocReceiptItems,
//	}).updateItem(doc)
//}

//func (s *Storage) DeleteReceiptDoc(id int64) (int64, error) {
//	return s.GetDocument(docTables{
//		Headers: tableDocReceiptHeaders,
//		Items:   tableDocReceiptItems,
//	}).deleteItem(id)
//}

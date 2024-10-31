package whs

import (
	"testing"
)

//func TestStorage_GetProductsItems(t *testing.T) {
//	db, mock := NewMock()
//	defer db.Close()
//
//	st := Storage{Db: db}
//
//	rowsProd := sqlmock.NewRows([]string{"p.id", "p.name", "p.item_number", "p.manufacturer_id", "p.manufacturer_name"})
//	rowsProd.AddRow(1, "test product 1", "", 1, "mnf 1")
//	rowsProd.AddRow(2, "test product 2", "", 2, "mnf 2")
//	rowsProd.AddRow(3, "test product 3", "", 4, "mnf 4")
//
//	rowsBc := sqlmock.NewRows([]string{"id", "name", "barcode_type"})
//
//	rowsCount := sqlmock.NewRows([]string{"count"})
//	rowsCount.AddRow(3)
//
//	mock.ExpectQuery("^SELECT (.+) FROM products").
//		WillReturnRows(rowsProd)
//
//	mock.ExpectQuery("^SELECT (.+) FROM barcodes WHERE parent_id*").WillReturnRows(rowsBc)
//	mock.ExpectQuery("^SELECT (.+) FROM barcodes WHERE parent_id*").WillReturnRows(rowsBc)
//	mock.ExpectQuery("^SELECT (.+) FROM barcodes WHERE parent_id*").WillReturnRows(rowsBc)
//
//	mock.ExpectQuery("^SELECT COUNT(.+) FROM (.+) sub").WillReturnRows(rowsCount)
//
//	prods, _, err := st.GetProductsItems(0, 10, 0)
//	if err != nil {
//		t.Error(err)
//	}
//	if len(prods) != 3 {
//		t.Error(err)
//	}
//}

func TestStorage_FindProductById(t *testing.T) {

}

//func TestStorage_FindProductById(t *testing.T) {
//	db, mock := NewMock()
//	defer db.Close()
//
//	// нашли товар по Id
//	rowsBc := sqlmock.NewRows([]string{"id", "barcode", "barcode_type"})
//	rowsBc.AddRow(1, "123456789", 1)
//
//	rows := sqlmock.NewRows([]string{"id", "name", "item_number", "manufacturer_id", "manufacturer_name"})
//	rows.AddRow(1, "test 1", "", 1, "Pfizer")
//
//	mock.ExpectQuery("^SELECT (.+) FROM products").
//		WillReturnRows(rows)
//
//	mock.ExpectQuery("^SELECT (.+) FROM barcodes").
//		WillReturnRows(rowsBc)
//
//	s := new(Storage)
//	s.Db = db
//	ps := s.GetProductService()
//	p, err := ps.FindProductById(1)
//	if err != nil {
//		t.Error(err)
//	}
//	if p == nil {
//		t.Error(errors.New("product is nil, must be Product struct"))
//	}
//	if len(p.Barcodes) != 1 {
//		t.Error(errors.New("barcodes len != 1, must be len = 1"))
//	}
//
//	// не нашли товар по Id
//	rowsBc = sqlmock.NewRows([]string{"barcode", "barcode_type"})
//
//	rows = sqlmock.NewRows([]string{"id", "name", "manufacturer_id", "manufacturer_name"})
//	mock.ExpectQuery("^SELECT (.+) FROM products").
//		WillReturnRows(rows)
//
//	mock.ExpectQuery("^SELECT (.+) FROM barcodes").
//		WillReturnRows(rowsBc)
//
//	p, err = ps.FindProductById(999)
//
//	if err == nil {
//		t.Error(errors.New("err==nil, must be err not nil. no product - no error"))
//	}
//	if p != nil {
//		t.Error(errors.New("product is not nil, must be nill"))
//	}
//}
//
//func TestStorage_FindProductsByBarcode(t *testing.T) {
//	db, mock := NewMock()
//	defer db.Close()
//
//	bc := "123456789456"
//
//	// не нашли штрихкод
//	// ожидаем пустой массив, без ошибки
//
//	rowsBc := sqlmock.NewRows([]string{"product_id", "barcode", "barcode_type"})
//	mock.ExpectQuery("^SELECT (.+) FROM barcodes").
//		WillReturnRows(rowsBc)
//
//	// до этого набора не должно дойти
//	rows := sqlmock.NewRows([]string{"id", "name", "manufacturer_id"})
//	rows.AddRow(10, "Тест продукт", 1)
//	mock.ExpectQuery("^SELECT (.+) FROM products").
//		WillReturnRows(rows)
//
//	s := new(Storage)
//	s.Db = db
//	ps := s.GetProductService()
//	p, err := ps.FindProductsByBarcode(bc)
//	if err != nil {
//		t.Error(err, "error is not nil")
//	}
//	if len(p) != 0 {
//		t.Error("products array must be empty")
//	}
//
//	//////////////////////////////////
//
//	db, mock = NewMock()
//	defer db.Close()
//
//	s = new(Storage)
//	s.Db = db
//	ps = s.GetProductService()
//
//	// нашли штрихкод, но не нашли товар. ошибка странная, но...
//	// ожидаем пустой массив, без ошибки
//	rowsBc = sqlmock.NewRows([]string{"product_id", "barcode", "barcode_type"})
//	rowsBc.AddRow(10, bc, 1)
//	mock.ExpectQuery("^SELECT (.+) FROM barcodes").
//		WillReturnRows(rowsBc)
//
//	rows = sqlmock.NewRows([]string{"id", "name", "manufacturer_id"})
//	//rows.AddRow(10, "Тест продукт", 1)
//	mock.ExpectQuery("^SELECT (.+) FROM products").
//		WillReturnRows(rows)
//
//	p, err = ps.FindProductsByBarcode(bc)
//	if err == nil {
//		t.Error(err, "error is not nil")
//	}
//	if len(p) != 0 {
//		t.Error("products array must be empty")
//	}
//
//	db, mock = NewMock()
//	defer db.Close()
//	s = new(Storage)
//	s.Db = db
//	ps = s.GetProductService()
//
//	// нашли штрихкод, нашли товар
//	// ожидаем 1 запись, без ошибки
//	rowsBc = sqlmock.NewRows([]string{"product_id", "barcode", "barcode_type"})
//	rowsBc.AddRow(10, bc, 1)
//	mock.ExpectQuery("^SELECT product_id, barcode, barcode_type FROM barcodes*").
//		WillReturnRows(rowsBc)
//
//	rows = sqlmock.NewRows([]string{"id", "name", "item_number", "manufacturer_id", "manufacturer_name"})
//	rows.AddRow(10, "Тест продукт", "", 1, "производитель")
//	mock.ExpectQuery("^SELECT (.+) FROM products").
//		WillReturnRows(rows)
//
//	// все штрихкоды для товара
//	rowsBcs := sqlmock.NewRows([]string{"id", "barcode", "barcode_type"})
//	rowsBcs.AddRow(10, bc, 1)
//	rowsBcs.AddRow(10, "45324523454235", 2)
//	rowsBcs.AddRow(10, "65745674567456", 3)
//
//	mock.ExpectQuery("^SELECT id, barcode, barcode_type FROM barcodes WHERE product_id*").
//		WillReturnRows(rowsBcs)
//
//	p, err = ps.FindProductsByBarcode(bc)
//	if err != nil {
//		t.Error(err, "error is not nil")
//	}
//	if len(p) != 1 {
//		t.Error("products array must be length = 1")
//	}
//
//	// нашли штрихкод, нашли 2 товара с одинаковым штрих-кодом
//	// ожидаем 1 запись, без ошибки
//	rowsBc = sqlmock.NewRows([]string{"product_id", "barcode", "barcode_type"})
//	rowsBc.AddRow(10, bc, 1)
//	rowsBc.AddRow(11, bc, 1)
//	mock.ExpectQuery("^SELECT (.+) FROM barcodes").
//		WillReturnRows(rowsBc)
//
//	// первый товар
//	rows = sqlmock.NewRows([]string{"id", "name", "item_number", "manufacturer_id", "manufacturer_name"})
//	rows.AddRow(10, "Тест продукт", "", 1, "производитель")
//	mock.ExpectQuery("^SELECT (.+) FROM products").
//		WillReturnRows(rows)
//
//	// все штрихкоды для товара 1
//	rowsBcs = sqlmock.NewRows([]string{"id", "barcode", "barcode_type"})
//	rowsBcs.AddRow(10, bc, 1)
//	rowsBcs.AddRow(10, "1_45324523454235", 2)
//	rowsBcs.AddRow(10, "1_65745674567456", 3)
//
//	mock.ExpectQuery("^SELECT (.+) FROM barcodes").
//		WillReturnRows(rowsBcs)
//
//	// второй товар
//	rows = sqlmock.NewRows([]string{"id", "name", "item_number", "manufacturer_id", "manufacturer_name"})
//	rows.AddRow(11, "Тест продукт 2", "", 1, "производитель")
//	mock.ExpectQuery("^SELECT (.+) FROM products").
//		WillReturnRows(rows)
//
//	// все штрихкоды для товара
//	rowsBcs = sqlmock.NewRows([]string{"id", "barcode", "barcode_type"})
//	rowsBcs.AddRow(11, bc, 1)
//	rowsBcs.AddRow(11, "2_45324523454235", 2)
//	rowsBcs.AddRow(11, "2_ 65745674567456", 3)
//
//	mock.ExpectQuery("^SELECT (.+) FROM barcodes").
//		WillReturnRows(rowsBcs)
//
//	p, err = ps.FindProductsByBarcode(bc)
//	if err != nil {
//		t.Error(err, "error is not nil")
//	}
//	if len(p) != 2 {
//		t.Error("products array must be length = 2")
//	}
//
//}
//
//func TestStorage_GetProductBarcodes(t *testing.T) {
//	db, mock := NewMock()
//	defer db.Close()
//
//	rowsBc := sqlmock.NewRows([]string{"barcode", "barcode_type"})
//
//	mock.ExpectQuery("^SELECT (.+) FROM barcodes").
//		WillReturnRows(rowsBc)
//
//	s := new(Storage)
//	s.Db = db
//	ps := s.GetProductService()
//	mBc, err := ps.GetProductBarcodes(10)
//	if len(mBc) != 0 {
//		t.Error("array must be empty")
//	}
//
//	rowsBc.AddRow("12345678902", 1)
//	rowsBc.AddRow("123456789032", 2)
//	rowsBc.AddRow("463456789032", 2)
//
//	s = new(Storage)
//	s.Db = db
//
//	mBc, err = ps.GetProductBarcodes(10)
//	if err == sql.ErrNoRows {
//		t.Error(err, "err must not be sql.ErrNoRows")
//	}
//
//	if mBc != nil && len(mBc) != 3 {
//		t.Error("wrong number of rows count")
//	}
//}

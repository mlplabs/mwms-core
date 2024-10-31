package whs

//// basic reference method getItems
//func TestStorage_GetWhsItems(t *testing.T) {
//	refTable := "whs"
//	db, mock := NewMock()
//	defer db.Close()
//
//	rowsItems := sqlmock.NewRows([]string{"id", "name"})
//	rowsItems.AddRow(1, "whs 1")
//	rowsItems.AddRow(2, "whs 2")
//
//	rowsCount := sqlmock.NewRows([]string{"count"})
//	rowsCount.AddRow(3)
//
//	storage := new(Storage)
//	storage.Db = db
//
//	refGetItemsQuery := "SELECT id, name FROM " + refTable + "*"
//	refCountItemsQuery := "SELECT COUNT(.+) as count FROM (.+)*"
//
//	mock.ExpectQuery(refGetItemsQuery).
//		WillReturnRows(rowsItems)
//
//	mock.ExpectQuery(refCountItemsQuery).
//		WillReturnRows(rowsCount)
//
//	items, i, err := storage.GetWhsItems(0, 2, 0)
//	if err != nil {
//		t.Error("result should not contain an error")
//	}
//	if i != 3 {
//		t.Error("total number of rows should be 3")
//	}
//	if len(items) != 2 {
//		t.Error("result should contain 2 rows")
//	}
//
//}
//
//// basic reference method findItemById
//func TestStorage_FindWhsById(t *testing.T) {
//	refTable := "whs"
//	db, mock := NewMock()
//	defer db.Close()
//
//	rows := sqlmock.NewRows([]string{"id", "name"})
//	rows.AddRow(1, "item 1")
//
//	mock.ExpectQuery("^SELECT (.+) FROM " + refTable + "").
//		WillReturnRows(rows)
//
//	storage := new(Storage)
//	storage.Db = db
//
//	w, err := storage.FindWhsById(1)
//
//	if err != nil {
//		t.Error("result should not contain an error")
//	}
//	if w == nil {
//		t.Error("result should contain base structure of reference only")
//	}
//}

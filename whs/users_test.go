package whs

//// basic reference method getItems
//func TestStorage_GetUsersItems(t *testing.T) {
//	refTable := "users"
//	db, mock := NewMock()
//	defer db.Close()
//
//	rowsItems := sqlmock.NewRows([]string{"id", "name"})
//	rowsItems.AddRow(1, "item 1")
//	rowsItems.AddRow(2, "item 2")
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
//	items, i, err := storage.GetUsersItems(0, 2, 0)
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
//func TestStorage_FindUserById(t *testing.T) {
//	refTable := "users"
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
//	w, err := storage.FindUserById(1)
//
//	if err != nil {
//		t.Error("result should not contain an error")
//	}
//	if w == nil {
//		t.Error("result should contain base structure of reference only")
//	}
//}
//
//// basic reference method findItemByName
//func TestStorage_FindUsersByName(t *testing.T) {
//	refTable := "users"
//	db, mock := NewMock()
//	defer db.Close()
//
//	rows := sqlmock.NewRows([]string{"id", "name"})
//	rows.AddRow(1, "search string")
//	rows.AddRow(2, "search string")
//
//	mock.ExpectQuery("^SELECT id, name FROM " + refTable + "").
//		WillReturnRows(rows)
//
//	storage := new(Storage)
//	storage.Db = db
//
//	w, err := storage.FindUsersByName("search string")
//
//	if err != nil {
//		t.Error("result should not contain an error")
//	}
//	if len(w) != 2 {
//		t.Error("result should contain 2 rows")
//	}
//}
//
//// basic reference method getSuggestion
//func TestStorage_GetUsersSuggestion(t *testing.T) {
//	refTable := "users"
//	db, mock := NewMock()
//	defer db.Close()
//
//	rows := sqlmock.NewRows([]string{"id", "name"})
//	rows.AddRow(1, "string 1")
//	rows.AddRow(2, "string 2")
//
//	mock.ExpectQuery("^SELECT id, name FROM " + refTable + " WHERE name ILIKE*").
//		WillReturnRows(rows)
//
//	storage := new(Storage)
//	storage.Db = db
//
//	w, err := storage.GetUsersSuggestion("string", 2)
//
//	if err != nil {
//		t.Error("result should not contain an error")
//	}
//	if len(w) != 2 {
//		t.Error("result should contain 2 rows")
//	}
//}

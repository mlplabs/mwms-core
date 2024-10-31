package whs

import (
	"database/sql"
	"fmt"
	"github.com/lib/pq"
	"strings"
)

// ICatalog Интерфейс объекта справочника
//
//	GetCatalogMetadata пока только для принадлежности
type ICatalog interface {
	GetCatalogMetadata()
	setStorage(*Storage)
	getStorage() *Storage
	getDb() *sql.DB
	getTableName() string
	GetItems(int, int) ([]ICatalogItem, int, error)
	GetItemsByField(int, int, string, int64) ([]ICatalogItem, int, error)
	FindById(int64) (ICatalogItem, error)
	GetSuggestion(string, int) ([]Suggestion, error)
}

// ICatalogItem Интерфейс элемента справочника
type ICatalogItem interface {
	GetId() int64
	GetName() string
	Store() (int64, int64, error)
	setCatalog(ICatalog)
	getCatalog() ICatalog
	getStorage() *Storage
	getDb() *sql.DB
	getTableName() string

	Delete() (int64, error)
}

// Catalog Базовый объект справочника
type Catalog struct {
	storage *Storage
	table   string
}

// CatalogItem Базовый элемент справочника
type CatalogItem struct {
	catalog ICatalog
	Id      int64  `json:"id"`
	Name    string `json:"name"`
}

type Suggestion struct {
	Id    int64  `json:"id"`
	Val   string `json:"val"`
	Title string `json:"title"`
}

func (c *Catalog) GetCatalogMetadata() {
	return
}

func (c *Catalog) setStorage(s *Storage) {
	c.storage = s
}

func (c *Catalog) getStorage() *Storage {
	return c.storage
}

func (c *Catalog) getDb() *sql.DB {
	return c.storage.Db
}

func (c *Catalog) getTableName() string {
	// TODO: потом перенести в GetCatalogMetadata
	return c.table
}

// GetItems returns a list items of catalog
func (c *Catalog) GetItems(offset int, limit int) ([]ICatalogItem, int, error) {
	var count int
	sqlCond := ""
	args := make([]any, 0)

	if limit == 0 {
		limit = DefaultRowsLimit
	}
	args = append(args, limit)
	args = append(args, offset)

	sqlSel := fmt.Sprintf("SELECT id, name FROM %s %s ORDER BY name ASC", c.getTableName(), sqlCond)

	rows, err := c.getDb().Query(sqlSel+" LIMIT $1 OFFSET $2", args...)
	if err != nil {
		return nil, count, err
	}
	defer rows.Close()

	items := make([]ICatalogItem, count, limit)
	for rows.Next() {
		item, _ := c.GetNewItem()

		err = rows.Scan(&item.Id, &item.Name)
		items = append(items, item)
	}

	sqlCount := fmt.Sprintf("SELECT COUNT(*) as count FROM ( %s ) sub", sqlSel)
	err = c.storage.Db.QueryRow(sqlCount).Scan(&count)
	if err != nil {
		return nil, count, err
	}
	return items, count, nil
}

// GetItems returns a list items of catalog
func (c *Catalog) GetItemsByField(offset int, limit int, fieldName string, val int64) ([]ICatalogItem, int, error) {
	var count int
	sqlCond := fmt.Sprintf(" WHERE %s = %d", fieldName, val)
	args := make([]any, 0)

	if limit == 0 {
		limit = DefaultRowsLimit
	}
	args = append(args, limit)
	args = append(args, offset)

	sqlSel := fmt.Sprintf("SELECT id, name FROM %s %s ORDER BY name ASC", c.getTableName(), sqlCond)

	rows, err := c.getDb().Query(sqlSel+" LIMIT $1 OFFSET $2", args...)
	if err != nil {
		return nil, count, err
	}
	defer rows.Close()

	items := make([]ICatalogItem, count, limit)
	for rows.Next() {
		item, _ := c.GetNewItem()

		err = rows.Scan(&item.Id, &item.Name)
		items = append(items, item)
	}

	sqlCount := fmt.Sprintf("SELECT COUNT(*) as count FROM ( %s ) sub", sqlSel)
	err = c.storage.Db.QueryRow(sqlCount).Scan(&count)
	if err != nil {
		return nil, count, err
	}
	return items, count, nil
}


// FindById returns item by internal id
func (c *Catalog) FindById(itemId int64) (ICatalogItem, error) {
	sqlUsr := fmt.Sprintf("SELECT id, name FROM %s WHERE id = $1", c.table)
	row := c.getDb().QueryRow(sqlUsr, itemId)
	item, _ := c.GetNewItem()
	item.catalog = c
	err := row.Scan(&item.Id, &item.Name)
	if err != nil {
		return nil, err
	}
	return item, nil
}

// FindByName returns a list of items by name
func (c *Catalog) FindByName(itemName string) ([]CatalogItem, error) {
	items := make([]CatalogItem, 0)
	sql := fmt.Sprintf("SELECT id, name FROM %s WHERE name = $1", c.table)
	rows, err := c.getDb().Query(sql, itemName)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	for rows.Next() {
		item, _ := c.GetNewItem()
		item.catalog = c
		err = rows.Scan(&item.Id, &item.Name)
		if err != nil {
			return nil, err
		}
		items = append(items, *item)
	}
	return items, nil
}

// GetSuggestion returns a list of suggestions for the search string
func (c *Catalog) GetSuggestion(text string, limit int) ([]Suggestion, error) {
	retVal := make([]Suggestion, 0)

	if strings.TrimSpace(text) == "" {
		return retVal, fmt.Errorf("invalid search text ")
	}
	if limit == 0 {
		limit = DefaultSuggestionLimit
	}

	sqlSel := fmt.Sprintf("SELECT id, name FROM %s WHERE name ILIKE $1 LIMIT $2", c.table)
	rows, err := c.getDb().Query(sqlSel, text+"%", limit)
	if err != nil {
		return retVal, err
	}
	defer rows.Close()
	for rows.Next() {
		item := Suggestion{}
		err := rows.Scan(&item.Id, &item.Val)
		if err != nil {
			return retVal, err
		}
		item.Title = item.Val
		retVal = append(retVal, item)
	}
	return retVal, err
}

func (c *Catalog) GetNewItem() (*CatalogItem, error) {
	item := new(CatalogItem)
	item.setCatalog(c)
	return item, nil
}

func (ci *CatalogItem) GetId() int64 {
	return ci.Id
}

func (ci *CatalogItem) GetName() string {
	return ci.Name
}

// установка справочника для строки
func (ci *CatalogItem) setCatalog(c ICatalog) {
	ci.catalog = c
}

// доступ к справочнику для строки
func (ci *CatalogItem) getCatalog() ICatalog {
	return ci.catalog
}

// доступ к хранилищу
func (ci *CatalogItem) getStorage() *Storage {
	return ci.catalog.(ICatalog).getStorage()
}

// доступ к базе
func (ci *CatalogItem) getDb() *sql.DB {
	return ci.catalog.(ICatalog).getStorage().Db
}

func (ci *CatalogItem) getTableName() string {
	return ci.getCatalog().getTableName()
}

func (ci *CatalogItem) Store() (int64, int64, error) {
	return ci.StoreTx(nil)
}

func (ci *CatalogItem) StoreTx(tx *sql.Tx) (int64, int64, error) {
	var err error
	var res sql.Result

	if ci.GetId() == 0 {
		var insertId int64
		sqlCreate := fmt.Sprintf("INSERT INTO %s (name) VALUES ($1) RETURNING id", ci.getTableName())
		if tx != nil {
			err = tx.QueryRow(sqlCreate, ci.GetName()).Scan(&insertId)
		} else {
			err = ci.getDb().QueryRow(sqlCreate, ci.GetName()).Scan(&insertId)
		}
		ci.Id = insertId
		if err != nil {
			return 0, 0, err
		}
		return ci.GetId(), 1, nil
	} else {

		sqlUpd := fmt.Sprintf("UPDATE %s SET name=$2 WHERE id=$1", ci.getTableName())
		if tx != nil {
			res, err = tx.Exec(sqlUpd, ci.Id, ci.Name)
		} else {
			res, err = ci.getDb().Exec(sqlUpd, ci.Id, ci.Name)
		}
		if err != nil {
			return 0, 0, err
		}
		if a, err := res.RowsAffected(); a != 1 || err != nil {
			return 0, a, err
		}
		return ci.GetId(), 1, nil
	}
}

func (ci *CatalogItem) Delete() (int64, error) {
	if ci.Id == 0 {
		return 0, fmt.Errorf("unacceptable action. item id eq 0")
	}
	sqlDel := fmt.Sprintf("DELETE FROM %s WHERE id=$1", ci.getTableName())
	res, err := ci.getDb().Exec(sqlDel, ci.Id)
	if err != nil {
		if pgErr, isPgErr := err.(*pq.Error); isPgErr {
			if pgErr.Code == ("23503") {
				return 0, err //&core.WrapError{Err: err, Code: core.ForeignKeyError}
			}
		}
		return 0, err
	}

	affRows, err := res.RowsAffected()
	if err != nil {
		return 0, err
	}
	return affRows, nil
}

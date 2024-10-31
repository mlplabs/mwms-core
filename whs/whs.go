package whs

import (
	"fmt"
)

// Whs is a physical warehouse object
// It must contain at least 3 Zone{} zones - acceptance, storage and shipment.
// There can be no more than 1 receiving and shipping zones, these zones are the entrance and exit in the warehouse, respectively
type Whs struct {
	Catalog
}

type WhsItem struct {
	CatalogItem
	Address        string     `json:"address"`
	AcceptanceZone ZoneItem   `json:"acceptance_zone"`
	ShippingZone   ZoneItem   `json:"shipping_zone"`
	StorageZones   []ZoneItem `json:"storage_zones"`
	CustomZones    []ZoneItem `json:"custom_zones"`
}

func (s *Storage) GetWhs() *Whs {
	m := new(Whs)
	m.table = tableRefWhs
	m.setStorage(s)
	return m
}

// GetItems returns a list items of catalog
func (w *Whs) GetItems(offset int, limit int) ([]ICatalogItem, int, error) {
	var count int
	sqlCond := ""
	args := make([]any, 0)

	if limit == 0 {
		limit = DefaultRowsLimit
	}
	args = append(args, limit)
	args = append(args, offset)

	sqlSel := fmt.Sprintf("SELECT id, name FROM %s %s ORDER BY name ASC", w.getTableName(), sqlCond)

	rows, err := w.getDb().Query(sqlSel+" LIMIT $1 OFFSET $2", args...)
	if err != nil {
		return nil, count, err
	}
	defer rows.Close()

	items := make([]ICatalogItem, count, limit)
	for rows.Next() {
		item, _ := w.GetNewItem()
		item.catalog = w

		err = rows.Scan(&item.Id, &item.Name)
		items = append(items, item)
	}

	sqlCount := fmt.Sprintf("SELECT COUNT(*) as count FROM ( %s ) sub", sqlSel)
	err = w.getDb().QueryRow(sqlCount).Scan(&count)
	if err != nil {
		return nil, count, err
	}
	return items, count, nil
}

// FindById returns a warehouse object by id
func (w *Whs) FindById(itemId int64) (ICatalogItem, error) {
	catZone := w.storage.GetZone()

	item, _ := w.GetNewItem()
	sqlWhs := fmt.Sprintf("SELECT id, name, address FROM %s WHERE id = $1", w.getTableName())
	row := w.getDb().QueryRow(sqlWhs, itemId)

	err := row.Scan(&item.Id, &item.Name, &item.Address)
	if err != nil {
		return nil, err
	}

	zones, err := catZone.FindByOwnerId(itemId)
	if err != nil {
		return nil, err
	}

	for _, v := range zones {
		if v.Type == ZoneTypeIncoming {
			item.AcceptanceZone = v
		}
		if v.Type == ZoneTypeOutGoing {
			item.ShippingZone = v
		}
		if v.Type == ZoneTypeStorage {
			item.StorageZones = append(item.StorageZones, v)
		}
		if v.Type == ZoneTypeCustom {
			item.CustomZones = append(item.CustomZones, v)
		}
	}
	return item, nil
}

func (w *Whs) GetNewItem() (*WhsItem, error) {
	item := new(WhsItem)
	item.setCatalog(w)
	item.StorageZones = make([]ZoneItem, 0)
	item.CustomZones = make([]ZoneItem, 0)
	return item, nil
}

// GetZones returns zones
func (wi *WhsItem) GetZones() ([]ZoneItem, error) {
	catZone := wi.getStorage().GetZone()
	return catZone.FindByOwnerId(wi.GetId())
}

// Store creates a new warehouse
func (wi *WhsItem) Store() (int64, int64, error) {
	catZone := wi.getStorage().GetZone()
	tx, err := wi.getDb().Begin()
	if err != nil {
		return 0, 0, err
	}
	if wi.GetId() == 0 {
		sqlCreate := fmt.Sprintf("INSERT INTO %s (name, address) VALUES ($1, $2) RETURNING id", wi.getTableName())
		err = tx.QueryRow(sqlCreate, wi.Name, wi.Address).Scan(&wi.Id)
		if err != nil {
			tx.Rollback()
			return 0, 0, err
		}
		z, _ := catZone.GetNewItem()
		z.Name = "Зона приемки"
		z.Type = ZoneTypeIncoming
		z.OwnerId = wi.Id
		_, _, err = z.StoreTx(tx)
		if err != nil {
			tx.Rollback()
			return 0, 0, err
		}

		z.Id = 0
		z.Name = "Зона отгрузки"
		z.Type = ZoneTypeOutGoing
		z.OwnerId = wi.Id
		_, _, err = z.StoreTx(tx)
		if err != nil {
			tx.Rollback()
			return 0, 0, err
		}
		z.Id = 0
		z.Name = "Зона хранения"
		z.Type = ZoneTypeStorage
		z.OwnerId = wi.Id
		_, _, err = z.StoreTx(tx)
		if err != nil {
			tx.Rollback()
			return 0, 0, err
		}
	} else {
		sqlUpdate := fmt.Sprintf("UPDATE %s SET name = $2, address = $3 WHERE id = $1", wi.getTableName())
		res, err := wi.getDb().Exec(sqlUpdate, wi.GetId(), wi.Name, wi.Address)
		if err != nil {
			return 0, 0, err
		}
		if a, err := res.RowsAffected(); a != 1 || err != nil {
			return 0, a, err
		}
		return wi.GetId(), 1, nil
	}

	sqlStorage := fmt.Sprintf(
		"create table if not exists storage%d ( "+
			"doc_id   integer default 0 not null, "+
			"doc_type smallint default 0 not null, "+
			"row_id   varchar(36) default ''::character varying not null, "+
			"row_time timestamptz default now() not null, "+
			"zone_id  integer, "+
			"cell_id  integer constraint storage%d_cells_id_fk references cells, "+
			"prod_id  integer,	"+
			"quantity integer ); "+
			"alter table storage%d owner to %s;", wi.Id, wi.Id, wi.Id, wi.getStorage().dbUser)
	_, err = tx.Exec(sqlStorage)
	if err != nil {
		tx.Rollback()
		return 0, 0, err
	}

	err = tx.Commit()
	if err != nil {
		tx.Rollback()
		return 0, 0, err
	}

	return wi.GetId(), 1, nil
}

// Delete delete warehouse
func (wi *WhsItem) Delete() (int64, error) {
	// TODO: need to remove child elements
	return wi.CatalogItem.Delete()
}

package whs

import (
	"database/sql"
	"fmt"
)

// Типы зон
const (
	// ZoneTypeStorage - тип зоны: хранение
	ZoneTypeStorage = iota
	// ZoneTypeIncoming - тип зоны: приемка
	ZoneTypeIncoming
	// ZoneTypeOutGoing - тип зоны: отгрузка
	ZoneTypeOutGoing
	ZoneTypeCustom = 999
)

// Zone - зона склада
type Zone struct {
	Catalog
}

type ZoneItem struct {
	CatalogItem
	OwnerId int64 `json:"owner_id"`
	Type    int   `json:"type"`
}

func (s *Storage) GetZone() *Zone {
	m := new(Zone)
	m.table = tableRefZones
	m.setStorage(s)
	return m
}

// FindByOwnerId returns a list of zones for the selected warehouse
func (z *Zone) FindByOwnerId(ownerId int64) ([]ZoneItem, error) {
	sqlZones := fmt.Sprintf("SELECT id, name, zone_type FROM %s WHERE owner_id = $1", z.getTableName())
	rows, err := z.getDb().Query(sqlZones, ownerId)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	res := make([]ZoneItem, 0)
	for rows.Next() {
		zi, _ := z.GetNewItem()

		err = rows.Scan(&zi.Id, &zi.Name, &zi.Type)
		if err != nil {
			return nil, err
		}
		res = append(res, *zi)
	}
	return res, nil
}

func (z *Zone) GetNewItem() (*ZoneItem, error) {
	item := new(ZoneItem)
	item.setCatalog(z)
	return item, nil
}

func (zi *ZoneItem) Store() (int64, int64, error) {
	return zi.StoreTx(nil)
}

func (zi *ZoneItem) StoreTx(tx *sql.Tx) (int64, int64, error) {
	var err error
	var res sql.Result

	if zi.GetId() == 0 {
		var insertId int64
		sqlCreate := fmt.Sprintf("INSERT INTO %s (name, zone_type, owner_id) VALUES ($1, $2, $3) RETURNING id", zi.getTableName())
		if tx != nil {
			err = tx.QueryRow(sqlCreate, zi.GetName(), zi.Type, zi.OwnerId).Scan(&insertId)
		} else {
			err = zi.getDb().QueryRow(sqlCreate, zi.GetName(), zi.Type, zi.OwnerId).Scan(&insertId)
		}
		zi.Id = insertId
		if err != nil {
			return 0, 0, err
		}
		return zi.GetId(), 1, nil
	} else {
		sqlUpd := fmt.Sprintf("UPDATE %s SET name=$2, zone_type=$3, owner_id=$4 WWHERE id=$1", zi.getTableName())
		if tx != nil {
			res, err = tx.Exec(sqlUpd, zi.GetId(), zi.GetName(), zi.Type, zi.OwnerId)
		} else {
			res, err = zi.getDb().Exec(sqlUpd, zi.GetId(), zi.GetName(), zi.Type, zi.OwnerId)
		}
		if err != nil {
			return 0, 0, err
		}
		if a, err := res.RowsAffected(); a != 1 || err != nil {
			return 0, a, err
		}
		return zi.GetId(), 1, nil
	}
}

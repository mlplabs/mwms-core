package whs

import (
	"database/sql"
	"fmt"
)

// SpecificSize структура весогабаритных характеристик (см/см3/кг)
// полный объем: length * width * height
// полезный объем: length * width * height * K(0.8)
// вес: для продукта вес единицы в килограммах, для ячейки максимально возможный вес размещенных продуктов
type SpecificSize struct {
	Length       int     `json:"length"`
	Width        int     `json:"width"`
	Height       int     `json:"height"`
	Weight       float32 `json:"weight"`
	Volume       float32 `json:"volume"`
	UsefulVolume float32 `json:"useful_volume"` // Полезный объем ячейки
}

// Типы штрих-кодов
const (
	BarcodeTypeUnknown = iota
	BarcodeTypeEAN13
	BarcodeTypeEAN8
	BarcodeTypeEAN14
	BarcodeTypeCode128
)

const (
	// CellDynamicPropIsService служебная ячейка. Запрещен автоматический отбор, но разрешены ручные перемещения в/из ячейки
	CellDynamicPropIsService = iota
	// CellDynamicPropSizeFree безразмерная ячейка. При размещении не используется проверка по размерам
	CellDynamicPropSizeFree
	// CellDynamicPropWeightFree при размещении не используется проверка по весу
	CellDynamicPropWeightFree
	// CellDynamicPropNotAllowedIn запрещено любое размещение в ячейку
	CellDynamicPropNotAllowedIn
	// CellDynamicPropNotAllowedOut запрещен любой отбор из ячейки
	CellDynamicPropNotAllowedOut
)

const (
	tableRefWhs             = "whs"
	tableRefProducts        = "products"
	tableRefManufacturers   = "manufacturers"
	tableRefBarcodes        = "barcodes"
	tableRefUsers           = "users"
	tableRefZones           = "zones"
	tableRefCells           = "cells"
	tableDocReceiptHeaders  = "receipt_headers"
	tableDocShipmentHeaders = "shipment_headers"
)

type Wms struct {
	Db     *sql.DB
	dbUser string
}

//type TurnoversProductRow struct {
//	Doc      IDocumentItem `json:"doc"`
//	Product  *ProductItem  `json:"product"`
//	Quantity int           `json:"quantity"`
//}
//
//type TurnoversParams struct {
//	Debit    bool
//	Credit   bool
//	DocTypes []int
//}

//type RemainingProductRow struct {
//	Product      *ProductItem                    `json:"product"`
//	Manufacturer *manufacturers.ManufacturerItem `json:"manufacturer"`
//	Zone         *ZoneItem                       `json:"zone"`
//	Cell         *CellItem                       `json:"cell"`
//	Quantity     int                             `json:"quantity"`
//}

type RowStorage struct {
	ProductId int64 `json:"product_id"`
	ZoneId    int64 `json:"zone_id"`
	CellSrcId int64 `json:"cell_src_id"`
	CellDstId int64 `json:"cell_dst_id"`
	Quantity  int64 `json:"quantity"`
}

var (
	DefaultRowsLimit       int = 10
	DefaultSuggestionLimit int = 10
)

func NewWms(db *sql.DB) *Wms {
	return &Wms{
		Db: db,
	}
}

func (s *Wms) GetDbUser() string {
	return s.dbUser
}

func (s *Wms) Query(query string, args ...interface{}) (*sql.Rows, error) {
	return s.Db.Query(query, args...)
}

// PutRow размещает в ячейку (cell) продукт (prod) в количестве (quantity)
// Возвращает количество которое было размещено (quantity)
func (s *Wms) PutRow(whsId int64, row *RowStorage, tx *sql.Tx) (int64, error) {
	var err error

	sqlIns := fmt.Sprintf("INSERT INTO wms%d (prod_id, zone_id, cell_id, quantity) VALUES ($1, $2, $3, $4)", whsId)
	if tx != nil {
		_, err = tx.Exec(sqlIns, row.ProductId, row.ZoneId, row.CellDstId, row.Quantity)
	} else {
		_, err = s.Db.Exec(sqlIns, row.ProductId, row.ZoneId, row.CellDstId, row.Quantity)
	}
	if err != nil {
		return 0, err
	}
	return row.Quantity, nil
}

// GetRow отбирает из ячейки (cell) продукт (prod) в количестве (quantity)
// Возвращает отобранное количество (quantity)
func (s *Wms) GetRow(whsId int64, row *RowStorage, tx *sql.Tx) (int64, error) {
	var err error

	if tx == nil {
		tx, err = s.Db.Begin()
		if err != nil {
			// не смогли начать транзакцию
			return 0, err
		}
	}

	sqlInsert := fmt.Sprintf("INSERT INTO wms%d (prod_id, zone_id, cell_id, quantity) VALUES ($1, $2, $3, $4)", whsId)
	_, err = tx.Exec(sqlInsert, row.ProductId, row.ZoneId, row.CellSrcId, -1*row.Quantity)
	if err != nil {
		return 0, err
	}

	sqlQuant := fmt.Sprintf("SELECT SUM(quantity) AS quantity "+
		"FROM wms%d WHERE zone_id = $1 AND cell_id = $2 AND prod_id = $3 "+
		"GROUP BY zone_id, cell_id, prod_id "+
		"HAVING SUM(quantity) < 0", whsId)
	rows, err := tx.Query(sqlQuant, row.ZoneId, row.CellSrcId, row.ProductId)
	if err != nil {
		// ошибка контроля
		return 0, err
	}
	defer rows.Close()
	// мы должны получить пустой запрос
	if rows.Next() {
		err = tx.Rollback()
		if err != nil {
			// ошибка отката... все очень плохо
			return 0, err
		}
		return 0, err
	}
	err = tx.Commit()
	if err != nil {
		return 0, err
	}
	return row.Quantity, nil
}

//// Quantity возвращает количество продуктов на св ячейке
//func (s *Wms) Quantity(whsId int, cell CellItem, tx *sql.Tx) (map[int]int, error) {
//	var zoneId, cellId, prodId, quantity int
//	res := make(map[int]int)
//
//	sqlQuantity := fmt.Sprintf("SELECT zone_id, cell_id, prod_id, SUM(quantity) AS quantity "+
//		"FROM wms%d WHERE zone_id = $1 AND cell_id = $2 "+
//		"GROUP BY zone_id, cell_id, prod_id "+
//		"HAVING SUM(quantity) <> 0 %s", whsId, "")
//
//	var err error
//	var rows *sql.Rows
//
//	if tx != nil {
//		rows, err = tx.Query(sqlQuantity, cell.ZoneId, cell.Id)
//	} else {
//		rows, err = s.Db.Query(sqlQuantity, cell.ZoneId, cell.Id)
//	}
//	if err != nil {
//		return nil, err
//	}
//	defer rows.Close()
//
//	for rows.Next() {
//		err = rows.Scan(&zoneId, &cellId, &prodId, &quantity)
//		if err != nil {
//			return nil, err
//		}
//		res[prodId] = quantity
//	}
//	return res, nil
//}
//
//func (s *Wms) MoveRow(d IDocumentItem, row *DocumentRow, tx *sql.Tx) error {
//	// TODO: cellSrc.WhsId <> cellDst.WhsId - временной разрыв или виртуальное перемещение
//
//	_, err := s.GetRow(d, row, tx)
//	if err != nil {
//		return err
//	}
//	_, err = s.PutRow(d, row, tx)
//	if err == nil {
//		return err
//	}
//	return nil
//}
//
//// BulkChangeSzCells устанавливает весогабаритные характеристики для массива ячеек
//func (s *Wms) BulkChangeSzCells(cells []CellItem, sz SpecificSize) (int64, error) {
//	var ids []int64
//
//	for _, c := range cells {
//		ids = append(ids, c.Id)
//	}
//	sqlBulkUpdate := "UPDATE cells SET sz_length=$2, sz_width=$3, sz_height=$4, sz_volume=$5, sz_uf_volume=$6, sz_weight=$7 WHERE id = ANY($1)"
//	res, err := s.Db.Exec(sqlBulkUpdate, pq.Array(ids), sz.Length, sz.Width, sz.Height, sz.Volume, sz.UsefulVolume, sz.Weight)
//	if err != nil {
//		return 0, err
//	}
//	return res.RowsAffected()
//}
//
//// BulkChangePropCells изменяет динамические параметры для массива ячеек
//func (s *Wms) BulkChangePropCells(cells []CellItem, CellDynamicProp int, value bool) (int64, error) {
//	var ids []int64
//
//	for _, c := range cells {
//		ids = append(ids, c.Id)
//	}
//
//	cond := ""
//	switch CellDynamicProp {
//	case CellDynamicPropSizeFree:
//		cond = "is_size_free = $1"
//	case CellDynamicPropWeightFree:
//		cond = "is_weight_free = $1"
//	case CellDynamicPropNotAllowedIn:
//		cond = "not_allowed_in = $1"
//	case CellDynamicPropNotAllowedOut:
//		cond = "not_allowed_out = $1"
//	case CellDynamicPropIsService:
//		cond = "is_service = $1"
//	}
//
//	sqlBulkUpdate := fmt.Sprintf("UPDATE %s WHERE id = ANY($1)", cond)
//	res, err := s.Db.Exec(sqlBulkUpdate, pq.Array(ids), value)
//	if err != nil {
//		return 0, err
//	}
//	return res.RowsAffected()
//}

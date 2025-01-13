package whs

import (
	"database/sql"
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

type Wms struct {
	Db     *sql.DB
	dbUser string
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

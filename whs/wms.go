package whs

import (
	"context"
	"database/sql"
	"errors"
	"github.com/mlplabs/mwms-core/whs/model"
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

func (w *Wms) GetDbUser() string {
	return w.dbUser
}

func (w *Wms) Query(query string, args ...interface{}) (*sql.Rows, error) {
	return w.Db.Query(query, args...)
}

func (w *Wms) GetCellInfo(ctx context.Context, cellId int64, tx *sql.Tx) (*model.Cell, error) {
	var err error
	if tx == nil {
		tx, err = w.Db.Begin()
		if err != nil {
			return nil, err
		}
	}
	sqlCell := "SELECT cs.id, cs.name, cs.whs_id, cs.zone_id, cs.passage_id, cs.rack_id, cs.floor FROM cells cs WHERE cs.id = $1"
	c := model.Cell{}
	row := tx.QueryRowContext(ctx, sqlCell, cellId)
	err = row.Scan(&c.Id, &c.Name, &c.WhsId, &c.ZoneId, &c.PassageId, &c.RackId, &c.Floor)
	if c.Name == "" {
		c.Name = c.GetNumericView()
	}
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return &c, nil
		}
		return nil, err
	}
	return &c, nil
}

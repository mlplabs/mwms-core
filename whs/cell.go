package whs

import "fmt"

const (
	CellNumericFormat   = "%01d%02d%02d%02d%02d"
	CellHumanViewFormat = "%01d-%02d-%02d-%02d-%02d"
)

// Cell ячейка склада
type Cell struct {
	Catalog
}

type CellItem struct {
	CatalogItem
	WhsId         int          `json:"whs_id"`     // Id склада (может быть именован)
	ZoneId        int          `json:"zone_id"`    // Id зоны назначения (может быть именован)
	SectionId     int          `json:"section_id"` // Id секции/блока (может быть именован)
	PassageId     int          `json:"passage_id"` // Id проезда (может быть именован)
	RackId        int          `json:"rack_id"`    // Id стеллажа (может быть именован)
	Floor         int          `json:"floor"`
	IsSizeFree    bool         `json:"is_size_free"`
	IsWeightFree  bool         `json:"is_weight_free"`
	NotAllowedIn  bool         `json:"not_allowed_in"`
	NotAllowedOut bool         `json:"not_allowed_out"`
	IsService     bool         `json:"is_service"`
	Size          SpecificSize `json:"size"`
}

type CellService struct {
	Storage *Storage
}

func (s *Storage) GetCell() *Cell {
	c := new(Cell)
	c.table = tableRefCells
	c.setStorage(s)
	return c
}

// FindById возвращает ячейку по внутреннему идентификатору
func (c *Cell) FindById(cellId int64) (ICatalogItem, error) {
	sqlCell := "SELECT id, name, whs_id, zone_id, passage_id, rack_id, floor, " +
		"sz_length, sz_width, sz_height, sz_volume, sz_uf_volume, sz_weight, " +
		"not_allowed_in, not_allowed_out, is_service, is_size_free, is_weight_free " +
		"FROM %s WHERE id = $1"

	row := c.getDb().QueryRow(fmt.Sprintf(sqlCell, c.getTableName()), cellId)
	ci, _ := c.GetNewItem()

	err := row.Scan(&ci.Id, &ci.Name, &ci.WhsId, &ci.ZoneId, &ci.PassageId, &ci.RackId, &ci.Floor,
		&ci.Size.Length, &ci.Size.Width, &ci.Size.Height, &ci.Size.Volume, &ci.Size.UsefulVolume, &ci.Size.Weight,
		&ci.NotAllowedIn, &ci.NotAllowedOut, &ci.IsService, &ci.IsSizeFree, &ci.IsWeightFree)
	if err != nil {
		return nil, err
	}
	return ci, nil
}

func (c *Cell) GetNewItem() (*CellItem, error) {
	item := new(CellItem)
	item.setCatalog(c)
	return item, nil
}

// SetSize устанавливает размер ячейки
func (sz *SpecificSize) SetSize(length, width, height int, kUV float32) {
	sz.Volume = float32(length * width * height)
	sz.UsefulVolume = sz.Volume * kUV
}

// GetSize возвращает размеры ячейки
// length, width, height as int
// volume, usefulVolume as float
func (sz *SpecificSize) GetSize() (int, int, int, float32, float32) {
	return sz.Length, sz.Width, sz.Height, sz.Volume, sz.UsefulVolume
}

// GetNumeric возвращает строковое представление ячейки в виде набора чисел
func (ci *CellItem) GetNumeric() string {
	return fmt.Sprintf(CellNumericFormat, ci.WhsId, ci.ZoneId, ci.PassageId, ci.RackId, ci.Floor)
}

// GetNumericView возвращает человеко-понятное представление (с разделителями)
func (ci *CellItem) GetNumericView() string {
	return fmt.Sprintf(CellHumanViewFormat, ci.WhsId, ci.ZoneId, ci.PassageId, ci.RackId, ci.Floor)
}

package cells

//// SpecificSize структура весогабаритных характеристик (см/см3/кг)
//// полный объем: length * width * height
//// полезный объем: length * width * height * K(0.8)
//// вес: для продукта вес единицы в килограммах, для ячейки максимально возможный вес размещенных продуктов
//type SpecificSize struct {
//	Length       int     `json:"length"`
//	Width        int     `json:"width"`
//	Height       int     `json:"height"`
//	Weight       float32 `json:"weight"`
//	Volume       float32 `json:"volume"`
//	UsefulVolume float32 `json:"useful_volume"` // Полезный объем ячейки
//}
//
//// SetSize устанавливает размер ячейки
//func (sz *SpecificSize) SetSize(length, width, height int, kUV float32) {
//	sz.Volume = float32(length * width * height)
//	sz.UsefulVolume = sz.Volume * kUV
//}
//
//// GetSize возвращает размеры ячейки
//// length, width, height as int
//// volume, usefulVolume as float
//func (sz *SpecificSize) GetSize() (int, int, int, float32, float32) {
//	return sz.Length, sz.Width, sz.Height, sz.Volume, sz.UsefulVolume
//}
//
//// GetNumeric возвращает строковое представление ячейки в виде набора чисел
//func (ci *CellItem) GetNumeric() string {
//	return fmt.Sprintf(CellNumericFormat, ci.WhsId, ci.ZoneId, ci.PassageId, ci.RackId, ci.Floor)
//}
//
//// GetNumericView возвращает человеко-понятное представление (с разделителями)
//func (ci *CellItem) GetNumericView() string {
//	return fmt.Sprintf(CellHumanViewFormat, ci.WhsId, ci.ZoneId, ci.PassageId, ci.RackId, ci.Floor)
//}

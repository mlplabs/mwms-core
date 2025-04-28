package model

import (
	"bytes"
	"fmt"
	"text/template"
)

const (
	CellNumericFormat    = "%01d%02d%02d%02d%02d"
	CellHumanViewFormat  = "%01d-%01d-%2d-%02d-%02d-%02d"
	CellCustomViewFormat = "{{}}"
)

type CellNameFormat string

// Cell - ячейка
// Склад/Зона/Блок/Проезд/Стеллаж/Этаж
type Cell struct {
	Id            int64
	Name          string `json:"name"`
	Number        int    `json:"number"` // Номер (порядковый) ячейки на полке
	IsSizeFree    bool   `json:"is_size_free"`
	IsWeightFree  bool   `json:"is_weight_free"`
	NotAllowedIn  bool   `json:"not_allowed_in"`
	NotAllowedOut bool   `json:"not_allowed_out"`
	IsService     bool   `json:"is_service"`
	//Size          SpecificSize `json:"size"`
	CellAddr
}

type CellAddr struct {
	WhsId     int64 `json:"whs_id"`     // Id склада (может быть именован)
	ZoneId    int64 `json:"zone_id"`    // Id зоны назначения (может быть именован)
	SectionId int   `json:"section_id"` // Id секции/блока (может быть именован)
	PassageId int   `json:"passage_id"` // Id проезда (может быть именован)
	RackId    int   `json:"rack_id"`    // Id стеллажа (может быть именован)
	Floor     int   `json:"floor"`      // этаж
	Number    int   `json:"number"`     // Порядковый номер ячейки по адресу
}

// GetNumeric возвращает строковое представление ячейки в виде набора чисел
func (ci *Cell) GetNumeric() string {
	return fmt.Sprintf(CellNumericFormat, ci.WhsId, ci.ZoneId, ci.PassageId, ci.RackId, ci.Floor)
}

// GetNumericView возвращает человеко-понятное представление (с разделителями)
func (ci *Cell) GetNumericView() string {
	return fmt.Sprintf(CellHumanViewFormat, ci.WhsId, ci.ZoneId, ci.PassageId, ci.RackId, ci.Floor, ci.Number)
}

func (ci *Cell) GetCustomView(tplCell string) string {
	tplBase := `{{- $w := .WhsId }}
				{{- $z := .ZoneId }}
				{{- $s := .SectionId }}
				{{- $p := .PassageId }}
				{{- $r := .RackId }}
				{{- $f := .Floor }}
				{{- $n := .Number }}{{- $N := .Name }}`
	var buf bytes.Buffer
	t, err := template.New("").Parse(tplBase + tplCell)
	if err != nil {
		return err.Error()
	}
	err = t.Execute(&buf, ci)
	if err != nil {
		return err.Error()
	}
	return buf.String()
}
func (ci *Cell) SetName(format CellNameFormat) {
	ci.Name = ci.GetNumericView() // пока так
}

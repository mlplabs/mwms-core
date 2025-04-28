package model

import (
	"fmt"
	"testing"
)

func TestCell_GetCustomView(t *testing.T) {
	tpl := "{{ $w }}-{{ $z }}-{{ $s }}-{{ $p }}-{{ $r }}-{{ $f }}-{{$n}}-{{$N}}"
	c := Cell{
		Id:            0,
		Name:          "",
		Number:        0,
		IsSizeFree:    false,
		IsWeightFree:  false,
		NotAllowedIn:  false,
		NotAllowedOut: false,
		IsService:     false,
		CellAddr: CellAddr{
			WhsId:     0,
			ZoneId:    0,
			SectionId: 0,
			PassageId: 0,
			RackId:    0,
			Floor:     0,
			Number:    0,
		},
	}
	fmt.Println(c.GetCustomView(tpl))
}

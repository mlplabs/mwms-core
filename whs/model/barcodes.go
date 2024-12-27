package model

const (
	BarcodeTypeUnknown = iota
	BarcodeTypeEAN13
	BarcodeTypeEAN8
	BarcodeTypeEAN14
	BarcodeTypeCode128
)

type BarcodeType struct {
	Id   int64  `json:"id"`
	Name string `json:"name"`
}

type Barcode struct {
	Id       int64  `json:"id"`
	Name     string `json:"name"`
	Type     int    `json:"type"`      // Тип ШК
	OwnerId  int64  `json:"owner_id"`  // ID владельца ШК
	OwnerRef string `json:"owner_ref"` // Таблица владельца
}

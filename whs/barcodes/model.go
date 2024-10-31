package barcodes

type Barcode struct {
	Id       int64  `json:"id"`
	Name     string `json:"name"`
	Type     int    `json:"type"`      // Тип ШК
	OwnerId  int64  `json:"owner_id"`  // ID владельца ШК
	OwnerRef string `json:"owner_ref"` // Таблица владельца
}

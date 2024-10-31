package barcodes

type Barcode struct {
	Id      int64  `json:"id"`
	Name    string `json:"name"`
	Type    int    `json:"type"`
	OwnerId int64  `json:"product_id"`
}

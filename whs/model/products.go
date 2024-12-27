package model

type Product struct {
	Id           int64        `json:"id"`
	Name         string       `json:"name"`
	ItemNumber   string       `json:"item_number"`
	Manufacturer Manufacturer `json:"manufacturer"`
	Barcodes     []Barcode    `json:"barcodes"`
}

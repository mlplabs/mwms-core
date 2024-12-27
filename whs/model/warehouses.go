package model

// Warehouse is a physical warehouse object
// It must contain at least 3 Zone{} zones - acceptance, storage and shipment.
// There can be no more than 1 receiving and shipping zones, these zones are the entrance and exit in the warehouse, respectively
type Warehouse struct {
	Id             int64  `json:"id"`
	Name           string `json:"name"`
	Address        string `json:"address"`
	AcceptanceZone Zone   `json:"acceptance_zone"`
	ShippingZone   Zone   `json:"shipping_zone"`
	StorageZones   []Zone `json:"storage_zones"`
	CustomZones    []Zone `json:"custom_zones"`
}

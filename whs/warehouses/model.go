package warehouses

import (
	"github.com/mlplabs/mwms-core/whs/zones"
)

// Warehouse is a physical warehouse object
// It must contain at least 3 Zone{} zones - acceptance, storage and shipment.
// There can be no more than 1 receiving and shipping zones, these zones are the entrance and exit in the warehouse, respectively
type Warehouse struct {
	Id             int64        `json:"id"`
	Name           string       `json:"name"`
	Address        string       `json:"address"`
	AcceptanceZone zones.Zone   `json:"acceptance_zone"`
	ShippingZone   zones.Zone   `json:"shipping_zone"`
	StorageZones   []zones.Zone `json:"storage_zones"`
	CustomZones    []zones.Zone `json:"custom_zones"`
}

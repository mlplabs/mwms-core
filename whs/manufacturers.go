package whs

// Manufacturer manufacturer
type Manufacturer struct {
	Catalog
}

func (s *Storage) GetManufacturer() *Manufacturer {
	m := new(Manufacturer)
	m.table = tableRefManufacturers
	m.setStorage(s)
	return m
}

type ManufacturerItem struct {
	CatalogItem
}

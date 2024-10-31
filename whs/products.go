package whs

import (
	"fmt"
	"github.com/mlplabs/microwms-core/core"
	"strings"
)

// Product item, storage unit
type Product struct {
	Catalog
}

type ProductItem struct {
	CatalogItem
	ItemNumber   string           `json:"item_number"`
	Barcodes     []BarcodeItem    `json:"barcodes"`
	Manufacturer ManufacturerItem `json:"manufacturer"`
	Size         SpecificSize     `json:"size"`
}

func (s *Storage) GetProduct() *Product {
	m := new(Product)
	m.table = tableRefProducts
	m.setStorage(s)
	return m
}

// GetItems returns a list of products
func (p *Product) GetItems(offset int, limit int) ([]ICatalogItem, int, error) {
	var count int

	sqlProd := "SELECT p.id, p.name, p.item_number, p.manufacturer_id, m.name As manufacturer_name FROM products p " +
		"		LEFT JOIN manufacturers m ON p.manufacturer_id = m.id" +
		"		ORDER BY p.name ASC"

	if limit == 0 {
		limit = DefaultRowsLimit
	}
	rows, err := p.storage.Db.Query(sqlProd+" LIMIT $1 OFFSET $2", limit, offset)
	if err != nil {
		return nil, count, err
	}
	defer rows.Close()

	prods := make([]ICatalogItem, count, limit)
	catBc := p.getStorage().GetBarcode()

	for rows.Next() {
		item, _ := p.GetNewItem()
		err = rows.Scan(&item.Id, &item.Name, &item.ItemNumber, &item.Manufacturer.Id, &item.Manufacturer.Name)

		pBarcodes, err := catBc.FindByOwnerId(item.Id) // пока так
		if err != nil {
			return nil, count, err
		}
		item.Barcodes = pBarcodes

		prods = append(prods, item)
	}

	sqlCount := fmt.Sprintf("SELECT COUNT(*) as count FROM ( %s ) sub", sqlProd)
	err = p.storage.Db.QueryRow(sqlCount).Scan(&count)
	if err != nil {
		return nil, count, err
	}
	return prods, count, nil
}

// FindById returns product by internal id
func (p *Product) FindById(productId int64) (ICatalogItem, error) {

	sqlCell := "SELECT p.id, p.name, p.item_number, p.manufacturer_id, m.name as manufacturer_name " +
		"FROM products p " +
		"LEFT JOIN manufacturers m ON p.manufacturer_id = m.id " +
		"WHERE p.id = $1"
	row := p.getDb().QueryRow(sqlCell, productId)
	item, _ := p.GetNewItem()
	err := row.Scan(&item.Id, &item.Name, &item.ItemNumber, &item.Manufacturer.Id, &item.Manufacturer.Name)
	if err != nil {
		return nil, err
	}

	catBc := p.getStorage().GetBarcode()
	pBarcodes, err := catBc.FindByOwnerId(productId)
	if err != nil {
		return nil, err
	}
	item.Barcodes = pBarcodes
	return item, nil
}

// FindByName returns a list of products by name
func (p *Product) FindByName(valName string) ([]ProductItem, error) {
	retItemList := make([]ProductItem, 0)
	sql := fmt.Sprintf("SELECT id, name, manufacturer_id FROM %s WHERE name = $1", p.getTableName())
	rows, err := p.storage.Db.Query(sql, valName)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	for rows.Next() {
		item, _ := p.GetNewItem()
		err = rows.Scan(&item.Id, &item.Name, &item.Manufacturer.Id)
		if err != nil {
			return nil, err
		}
		retItemList = append(retItemList, *item)
	}
	return retItemList, nil
}

// FindByBarcode returns a product by barcode
func (p *Product) FindByBarcode(barcodeStr string) ([]ProductItem, error) {
	//var prodItem *ProductItem

	catProd := p.getStorage().GetProduct()
	catBc := p.getStorage().GetBarcode()

	prods := make([]ProductItem, 0, 0)
	bcItems, err := catBc.FindByName(barcodeStr)
	if err != nil {
		return prods, err
	}
	for _, v := range bcItems {
		prodItem, err := catProd.FindById(v.OwnerId)
		if core.ErrNoRows(err) {
			continue
		}
		if err != nil {
			return prods, err
		}
		pi := prodItem.(*ProductItem)
		prods = append(prods, *pi)
	}
	return prods, nil
}

func (p *Product) GetNewItem() (*ProductItem, error) {
	item := new(ProductItem)
	item.setCatalog(p)
	item.Barcodes = make([]BarcodeItem, 0)
	return item, nil
}

func (p *Product) GetSuggestion(text string, limit int) ([]Suggestion, error) {
	retVal := make([]Suggestion, 0)

	if strings.TrimSpace(text) == "" {
		return retVal, fmt.Errorf("invalid search text ")
	}
	if limit == 0 {
		limit = DefaultSuggestionLimit
	}

	// TODO: константы наименований таблиц
	sqlSel := fmt.Sprintf("SELECT p.id, p.name, m.name as mnf_name FROM %s p "+
		"LEFT JOIN %s m ON p.manufacturer_id = m.id "+
		"WHERE p.name ILIKE $1 LIMIT $2", tableRefProducts, tableRefManufacturers)

	rows, err := p.getDb().Query(sqlSel, text+"%", limit)
	if err != nil {
		return retVal, err
	}
	defer rows.Close()

	for rows.Next() {
		item := Suggestion{}
		mnfName := ""
		err = rows.Scan(&item.Id, &item.Val, &mnfName)
		if err != nil {
			return retVal, err
		}
		item.Title = item.Val
		if mnfName != "" {
			item.Title += ", " + mnfName
		}
		retVal = append(retVal, item)
	}
	return retVal, err
}

// GetBarcodes returns a list of product barcodes
func (pi *ProductItem) GetBarcodes(productId int64) ([]BarcodeItem, error) {
	catBc := pi.getStorage().GetBarcode()
	return catBc.FindByOwnerId(productId)
}

func (pi *ProductItem) Store() (int64, int64, error) {
	if valid, err := pi.valid(); !valid {
		return 0, 0, err
	}
	catProd := pi.getStorage().GetProduct()
	catMnf := pi.getStorage().GetManufacturer()
	catBc := pi.getStorage().GetBarcode()

	tx, err := pi.getDb().Begin()
	if err != nil {
		return 0, 0, err
	}

	// When creating a product, we always look for a manufacturer by name, regardless of its ID
	// because when creating, the user can choose from a list of hints and change the name
	mnfItems, err := catMnf.FindByName(pi.Manufacturer.Name)
	if err != nil {
		return 0, 0, err
	}
	if len(mnfItems) > 1 {
		return 0, 0, fmt.Errorf("it is not possible to identify the manufacturer. found %d", len(mnfItems))
	}
	if len(mnfItems) == 1 {
		pi.Manufacturer.Id = mnfItems[0].GetId()
	}

	// If the manufacturer is not found by name, then we create it
	if pi.Manufacturer.Id == 0 {
		mnfItem, _ := catMnf.GetNewItem()
		mnfItem.Name = pi.Manufacturer.Name
		pi.Manufacturer.Id, _, err = mnfItem.Store()
		if err != nil {
			tx.Rollback()
			return 0, 0, err
		}
	}

	s := pi.Size

	// ->
	if pi.Id == 0 {

		prodItems, err := catProd.FindByName(pi.Name)
		if err != nil {
			tx.Rollback()
			return 0, 0, err
		}
		// можем получить несколько товаров с одинаковым именем
		// надо понять, имеется ли среди них товары с таким же производителем
		for _, v := range prodItems {
			if v.Manufacturer.Id == pi.Manufacturer.Id {
				// нашли существующий товар
				pi.Id = v.Id
				break
			}
		}
		if pi.Id == 0 {

			sqlInsProd := fmt.Sprintf("INSERT INTO %s (name, item_number, manufacturer_id, sz_length, sz_wight, sz_height, sz_weight, sz_volume, sz_uf_volume) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9) RETURNING id", catProd.getTableName())
			err = tx.QueryRow(sqlInsProd, pi.Name, pi.ItemNumber, pi.Manufacturer.Id, s.Length, s.Width, s.Height, s.Weight, s.Volume, s.UsefulVolume).Scan(&pi.Id)
			if err != nil {
				tx.Rollback()
				return 0, 0, err
			}

			if pi.Barcodes != nil {
				for _, bc := range pi.Barcodes {
					if bc.Id != 0 {
						continue
					}
					itemBc, _ := catBc.GetNewItem()
					itemBc.Id = bc.Id
					itemBc.Name = bc.Name
					itemBc.Type = bc.Type
					itemBc.OwnerId = pi.Id
					bc.Id, _, err = itemBc.StoreTx(tx)
					if err != nil {
						tx.Rollback()
						return 0, 0, err
					}
				}
			}

		}
	} else {
		sqlInsProd := fmt.Sprintf("UPDATE %s SET name=$2,manufacturer_id=$3,sz_length=$4,sz_wight=$5,sz_height=$6,sz_weight=$7,sz_volume=$8, sz_uf_volume=$9, item_number=$10 WHERE id=$1", tableRefProducts)
		res, err := tx.Exec(sqlInsProd, pi.Id, pi.Name, pi.Manufacturer.Id, s.Length, s.Width, s.Height, s.Weight, s.Volume, s.UsefulVolume, pi.ItemNumber)
		if err != nil {
			tx.Rollback()
			return 0, 0, err
		}

		if a, err := res.RowsAffected(); a != 1 || err != nil {
			tx.Rollback()
			return 0, 0, err
		}

		if pi.Barcodes != nil {
			for _, bc := range pi.Barcodes {
				if bc.Id != 0 {
					continue
				}
				itemBc, _ := catBc.GetNewItem()
				itemBc.Id = bc.Id
				itemBc.Name = bc.Name
				itemBc.Type = bc.Type
				itemBc.OwnerId = pi.Id
				bc.Id, _, err = itemBc.StoreTx(tx)
				if err != nil {
					tx.Rollback()
					return 0, 0, err
				}
			}
		}
	}
	err = tx.Commit()
	if err != nil {
		return 0, 0, err
	}

	return pi.Id, 1, nil
}

func (pi *ProductItem) Delete() (int64, error) {
	// удалить товар можно при условии, что он нигде не числится
	// сначала удалим шк
	// TODO: тут должна быть кастомная процедура удаления
	return pi.CatalogItem.Delete()
}

func (pi *ProductItem) valid() (bool, error) {
	return strings.TrimSpace(pi.Name) != "", fmt.Errorf("product name is empty")
}

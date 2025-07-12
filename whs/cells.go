package whs

import (
	"context"
	"github.com/mlplabs/mwms-core/whs/model"
)

func (s *Storage) GetCellById(ctx context.Context, cellId int64) (*model.Cell, error) {
	return nil, nil
}

func (s *Storage) CreateCell(ctx context.Context, cell *model.Cell) (int64, error) {
	if cell.Name == "" {
		cell.SetName("")
	}
	cellNum, err := s.getNextCellNum(ctx, &cell.CellAddr)
	if err != nil {
		return 0, err
	}
	cell.Number = cellNum
	sqlIns := `INSERT INTO cells (name, whs_id, zone_id, section_id, passage_id, rack_id, floor, number,
                   sz_length, sz_width, sz_height, sz_volume, sz_uf_volume, sz_weight, is_size_free, is_weight_free, not_allowed_in, not_allowed_out, is_service)
				VALUES ($1, $2, $3, $4, $5, $6, $7, $8) RETURNING id`
	err = s.wms.Db.QueryRowContext(ctx, sqlIns, cell.Name, cell.WhsId, cell.ZoneId, cell.SectionId, cell.PassageId, cell.RackId, cell.Floor, cellNum).Scan(&cell.Id)
	if err != nil {
		return 0, err
	}
	return cell.Id, nil
}

func (s *Storage) UpdateCell(ctx context.Context, cell *model.Cell) (int64, error) {
	if cell.Name == "" {
		cell.SetName("")
	}
	sqlUpd := `UPDATE cells SET name=$2 WHERE id=$1`
	res, err := s.wms.Db.ExecContext(ctx, sqlUpd, cell.Id, cell.Name)
	if err != nil {
		return 0, err
	}
	if a, err := res.RowsAffected(); a != 1 || err != nil {
		return 0, err
	}
	return cell.Id, nil
}

func (s *Storage) CellsSuggest(ctx context.Context, text string, limit int) ([]model.Suggestion, error) {
	retVal := make([]model.Suggestion, 0)
	if limit == 0 {
		limit = DefaultSuggestionLimit
	}

	sqlSel := `SELECT id, name FROM cells WHERE name ILIKE $1 LIMIT $2`
	rows, err := s.wms.Db.QueryContext(ctx, sqlSel, text+"%", limit)
	if err != nil {
		return retVal, err
	}
	defer rows.Close()
	for rows.Next() {
		item := model.Suggestion{}
		err := rows.Scan(&item.Id, &item.Val)
		if err != nil {
			return retVal, err
		}
		item.Title = item.Val
		retVal = append(retVal, item)
	}
	return retVal, err
}

func (s *Storage) getNextCellNum(ctx context.Context, addr *model.CellAddr) (int, error) {
	var nextNum int
	sqlCell := `SELECT count(*) +1 as next_cell FROM cells 
                             WHERE whs_id = $1 AND zone_id = $2 AND section_id = $3
                             AND passage_id = $4 AND rack_id=$5 AND floor=6`
	row := s.wms.Db.QueryRow(sqlCell, addr.WhsId, addr.ZoneId, addr.SectionId, addr.PassageId, addr.RackId, addr.Floor)
	if err := row.Scan(&nextNum); err != nil {
		return 0, err
	}
	return nextNum, nil
}

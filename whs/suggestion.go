package whs

import (
	"context"
	"fmt"
	"github.com/mlplabs/mwms-core/whs/model"
)

type Suggestions struct {
	storage *Storage
}

func NewSuggestions(s *Storage) *Suggestions {
	return &Suggestions{storage: s}
}

func (s *Suggestions) GetSuggestion(ctx context.Context, refName string, text string, limit int) ([]model.Suggestion, error) {
	retVal := make([]model.Suggestion, 0)
	if limit == 0 {
		limit = DefaultSuggestionLimit
	}

	sqlSel := fmt.Sprintf("SELECT id, name FROM %s WHERE name ILIKE $1 LIMIT $2", refName)
	rows, err := s.storage.Db.QueryContext(ctx, sqlSel, text+"%", limit)
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

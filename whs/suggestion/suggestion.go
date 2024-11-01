package suggestion

import (
	"context"
	"fmt"
	"github.com/mlplabs/mwms-core/whs"
)

type Suggestions struct {
	storage *whs.Storage
}

func NewSuggestions(s *whs.Storage) *Suggestions {
	return &Suggestions{storage: s}
}

func (s *Suggestions) GetSuggestion(ctx context.Context, refName string, text string, limit int) ([]Suggestion, error) {
	retVal := make([]Suggestion, 0)
	if limit == 0 {
		limit = whs.DefaultSuggestionLimit
	}

	sqlSel := fmt.Sprintf("SELECT id, name FROM %s WHERE name ILIKE $1 LIMIT $2", refName)
	rows, err := s.storage.Db.QueryContext(ctx, sqlSel, text+"%", limit)
	if err != nil {
		return retVal, err
	}
	defer rows.Close()
	for rows.Next() {
		item := Suggestion{}
		err := rows.Scan(&item.Id, &item.Val)
		if err != nil {
			return retVal, err
		}
		item.Title = item.Val
		retVal = append(retVal, item)
	}
	return retVal, err
}

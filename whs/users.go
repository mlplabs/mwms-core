package whs

import (
	"context"
	"errors"
	"fmt"
	"github.com/lib/pq"
	"github.com/mlplabs/mwms-core/whs/model"
)

const tableUsers = "users"

// GetUsers returns a list items without limit
func (s *Storage) GetUsers(ctx context.Context) ([]model.User, error) {
	users := make([]model.User, 0)
	sqlSel := fmt.Sprintf("SELECT id, name FROM %s ORDER BY name ASC", tableUsers)
	rows, err := s.wms.Db.QueryContext(ctx, sqlSel)
	if err != nil {
		return users, err
	}
	defer rows.Close()

	for rows.Next() {
		usr := model.User{}
		err = rows.Scan(&usr.Id, &usr.Name)
		users = append(users, usr)
	}
	return users, nil
}

// GetUsersItems returns a list items of catalog with limit & offset
func (s *Storage) GetUsersItems(ctx context.Context, offset int, limit int) ([]model.User, int64, error) {
	var totalCount int64
	users := make([]model.User, 0)
	sqlCond := ""
	args := make([]any, 0)

	if limit == 0 {
		limit = DefaultRowsLimit
	}
	args = append(args, limit)
	args = append(args, offset)

	sqlSel := fmt.Sprintf("SELECT id, name FROM %s %s ORDER BY name ASC", tableUsers, sqlCond)

	rows, err := s.wms.Db.QueryContext(ctx, sqlSel+" LIMIT $1 OFFSET $2", args...)
	if err != nil {
		return users, totalCount, err
	}
	defer rows.Close()

	for rows.Next() {
		usr := model.User{}
		err = rows.Scan(&usr.Id, &usr.Name)
		users = append(users, usr)
	}

	sqlCount := fmt.Sprintf("SELECT COUNT(*) as count FROM ( %s ) sub", sqlSel)
	err = s.wms.Db.QueryRowContext(ctx, sqlCount).Scan(&totalCount)
	if err != nil {
		return users, totalCount, err
	}
	return users, totalCount, nil
}

func (s *Storage) CreateUser(ctx context.Context, user *model.User) (int64, error) {
	var insertId int64
	sqlCreate := fmt.Sprintf("INSERT INTO %s (name) VALUES ($1) RETURNING id", tableUsers)
	err := s.wms.Db.QueryRowContext(ctx, sqlCreate, user.Name).Scan(&insertId)
	return insertId, err
}

func (s *Storage) UpdateUser(ctx context.Context, user *model.User) (int64, error) {
	sqlUpd := fmt.Sprintf("UPDATE %s SET name=$2 WHERE id=$1", tableUsers)
	res, err := s.wms.Db.ExecContext(ctx, sqlUpd, user.Id, user.Name)
	if err != nil {
		return 0, err
	}
	if a, err := res.RowsAffected(); a != 1 || err != nil {
		return 0, err
	}
	return user.Id, nil
}

func (s *Storage) DeleteUser(ctx context.Context, itemId int64) error {
	if itemId == 0 {
		return fmt.Errorf("unacceptable action. item id eq 0")
	}
	sqlDel := fmt.Sprintf("DELETE FROM %s WHERE id=$1", tableUsers)
	_, err := s.wms.Db.ExecContext(ctx, sqlDel, itemId)
	if err != nil {
		var pgErr *pq.Error
		if errors.As(err, &pgErr) {
			if pgErr.Code == ("23503") {
				return err
			}
		}
		return err
	}
	return nil
}

func (s *Storage) GetUserById(ctx context.Context, itemId int64) (*model.User, error) {
	sqlUsr := fmt.Sprintf("SELECT id, name FROM %s WHERE id = $1", tableUsers)
	row := s.wms.Db.QueryRowContext(ctx, sqlUsr, itemId)
	newItem := model.User{}
	err := row.Scan(&newItem.Id, &newItem.Name)
	if err != nil {
		return nil, err
	}
	return &newItem, nil
}

func (s *Storage) FindUsersByName(ctx context.Context, itemName string) ([]model.User, error) {
	users := make([]model.User, 0)
	sql := fmt.Sprintf("SELECT id, name FROM %s WHERE name = $1", tableUsers)
	rows, err := s.wms.Db.QueryContext(ctx, sql, itemName)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	for rows.Next() {
		usr := model.User{}
		err = rows.Scan(&usr.Id, &usr.Name)
		if err != nil {
			return nil, err
		}
		users = append(users, usr)
	}
	return users, nil
}

func (s *Storage) UsersSuggest(ctx context.Context, text string, limit int) ([]model.Suggestion, error) {
	retVal := make([]model.Suggestion, 0)
	if limit == 0 {
		limit = DefaultSuggestionLimit
	}

	sqlSel := fmt.Sprintf("SELECT id, name FROM %s WHERE name ILIKE $1 LIMIT $2", tableUsers)
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

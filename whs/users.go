package whs

// User warehouse user
type User struct {
	Catalog
}

func (s *Storage) GetUser() *User {
	u := new(User)
	u.table = tableRefUsers
	u.setStorage(s)
	return u
}

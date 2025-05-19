package models

import (
	"github.com/google/uuid"
)

type User struct {
	ID       uuid.UUID
	Username string
	Hash     []byte
	IsAdmin  bool
}

func NewUser(id uuid.UUID, username string, hash []byte, isAdmin bool) *User {
	return &User{
		ID:       id,
		Username: username,
		Hash:     hash,
		IsAdmin:  isAdmin,
	}
}

func (u *User) ToDB() *UserDB {
	admin := 0
	if u.IsAdmin {
		admin = 1
	}

	return &UserDB{
		ID:       u.ID.String(),
		Username: u.Username,
		Hash:     u.Hash,
		IsAdmin:  admin,
	}
}

type UserDB struct {
	ID       string `db:"id"`
	Username string `db:"username"`
	Hash     []byte `db:"hash"`
	IsAdmin  int    `db:"is_admin"`
}

func UserFromDB(db *UserDB) (*User, error) {
	id, err := uuid.Parse(db.ID)
	if err != nil {
		return nil, err
	}

	return &User{
		ID:       id,
		Username: db.Username,
		Hash:     db.Hash,
		IsAdmin:  db.IsAdmin == 1,
	}, nil
}

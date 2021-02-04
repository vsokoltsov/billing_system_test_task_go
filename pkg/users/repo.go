package users

import (
	"database/sql"
	"fmt"
)

// SQLRepository represents communcation with users
type SQLRepository interface {
	GetByID(userID int) (*User, error)
	// GetByWalletID(walletID int) (*User, error)
	Create(email string) (int64, error)
}

// UsersService implements SQLRepository
type UsersService struct {
	db *sql.DB
}

func NewUsersService(db *sql.DB) SQLRepository {
	return UsersService{
		db: db,
	}
}

func (ds UsersService) GetByID(userID int) (*User, error) {
	user := User{}
	getUserErr := ds.db.
		QueryRow("select id, email from users where id = ?", userID).
		Scan(&user.ID, &user.Email)

	if getUserErr != nil {
		return nil, getUserErr
	}

	return &user, nil
}

// func (ds UsersService) GetByWalletID(walletID int) (*User, error) {

// }

func (ds UsersService) Create(email string) (int64, error) {
	insertRes, insertErr := ds.db.Exec(
		"insert into users(email) values(?)",
		email,
	)

	if insertErr != nil {
		return 0, fmt.Errorf("Error user creation: %s", insertErr)
	}

	return insertRes.LastInsertId()
}

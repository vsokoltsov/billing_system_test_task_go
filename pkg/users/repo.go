package users

import (
	"context"
	"database/sql"
	"fmt"
)

const (
	CreateUser = iota + 1
)

// SQLRepository represents communcation with users
type SQLRepository interface {
	GetByID(ctx context.Context, userID int) (*User, error)
	// GetByWalletID(walletID int) (*User, error)
	Create(ctx context.Context, email string) (int64, error)
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

func (ds UsersService) GetByID(ctx context.Context, userID int) (*User, error) {
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

func (ds UsersService) Create(ctx context.Context, email string) (int64, error) {

	conn, _ := ds.db.Conn(ctx)
	_, alErr := conn.ExecContext(ctx, `select pg_advisory_lock($1)`, CreateUser)
	if alErr != nil {
		return 0, fmt.Errorf("Error of starting advisory lock: %s", alErr)
	}

	transaction, transactionErr := conn.BeginTx(ctx, nil)
	if transactionErr != nil {
		return 0, fmt.Errorf("Error of transaction initialization: %s", transactionErr)
	}
	transaction.ExecContext(ctx, "set transaction isolation level serializable")

	insertRes, insertErr := transaction.ExecContext(
		ctx,
		"insert into users(email) values(?)",
		email,
	)

	if insertErr != nil {
		transaction.Rollback()
		conn.ExecContext(ctx, `select pg_advisory_unlock($1)`, CreateUser)
		return 0, fmt.Errorf("Error user creation: %s", insertErr)
	}

	transactionCommitErr := transaction.Commit()
	if transactionCommitErr != nil {
		transaction.Rollback()
		conn.ExecContext(ctx, `select pg_advisory_unlock($1)`, CreateUser)
		return 0, fmt.Errorf("Error of transaction commit: %s", transactionCommitErr)
	}

	_, auErr := conn.ExecContext(ctx, `select pg_advisory_unlock($1)`, CreateUser)
	if auErr != nil {
		return 0, fmt.Errorf(
			"Error of unlocking user's %d postgres lock: %s",
			CreateUser,
			auErr,
		)
	}
	conn.Close()

	return insertRes.LastInsertId()
}

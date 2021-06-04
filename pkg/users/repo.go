package users

import (
	"billing_system_test_task/pkg/wallets"
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
	GetByWalletID(ctx context.Context, walletID int) (*User, error)
	Create(ctx context.Context, email string) (int64, error)
}

// UsersService implements SQLRepository
type UsersService struct {
	db          *sql.DB
	walletsRepo wallets.SQLRepository
}

// NewUsersService returns instance of UserService
func NewUsersService(db *sql.DB, wallets wallets.SQLRepository) SQLRepository {
	return UsersService{
		db:          db,
		walletsRepo: wallets,
	}
}

// GetByID receives user information by id
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

// GetByWalletID receives user information by wallet id
func (ds UsersService) GetByWalletID(ctx context.Context, walletID int) (*User, error) {
	user := User{}
	query := `
		select u.id, u.email, w.balance, w.currency
		from users as u
		join wallets as w
		join u.id = w.user_id
		where w.id = ?
	`
	userRow, userGetErr := ds.db.Query(query, walletID)
	if userGetErr != nil {
		return nil, fmt.Errorf("GetByWalletID: error of receiving user: %s", userGetErr)
	}

	for userRow.Next() {
		wallet := wallets.Wallet{}
		scanErr := userRow.Scan(
			&user.ID,
			&user.Email,
			&wallet.Balance,
			&wallet.Currency,
		)
		if scanErr != nil {
			return nil, fmt.Errorf("GetByWalletID: Error of reading the result: %s", scanErr)
		}
		user.Wallet = &wallet
	}
	return &user, nil
}

// Create creates new user
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

	// Creates new user
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

	userID, userIDErr := insertRes.LastInsertId()
	if userIDErr != nil {
		transaction.Rollback()
		conn.ExecContext(ctx, `select pg_advisory_unlock($1)`, CreateUser)
		return 0, fmt.Errorf("Error user retrieving id: %s", userIDErr)
	}

	// Creates wallet for user
	_, insertWalletErr := ds.walletsRepo.Create(ctx, conn, userID)
	if insertWalletErr != nil {
		transaction.Rollback()
		conn.ExecContext(ctx, `select pg_advisory_unlock($1)`, CreateUser)
		return 0, fmt.Errorf("Error of wallet transaction commit: %s", insertWalletErr)
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

	return userID, nil
}

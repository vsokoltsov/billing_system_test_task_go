package repositories

import (
	"billing_system_test_task/pkg/adapters"
	"billing_system_test_task/pkg/adapters/tx"
	"billing_system_test_task/pkg/entities"
	"context"
	"fmt"
)

const (
	CreateUser = iota + 1
)

// UsersManager represents communication with users
type UsersManager interface {
	WithTx(t tx.Tx) UsersManager
	GetByID(ctx context.Context, userID int) (*entities.User, error)
	GetByWalletID(ctx context.Context, walletID int) (*entities.User, error)
	Create(ctx context.Context, email string) (int64, error)
}

// UsersService implements SQLRepository
type UsersService struct {
	db adapters.SQLAdapter
	// walletsRepo wallets.WalletsManager
}

// NewUsersService returns instance of UserService
func NewUsersService(db adapters.SQLAdapter) UsersManager {
	return &UsersService{
		db: db,
	}
}

func (us *UsersService) WithTx(t tx.Tx) UsersManager {
	return NewUsersService(t.(adapters.SQLAdapter))
}

// GetByID receives user information by id
func (ds UsersService) GetByID(ctx context.Context, userID int) (*entities.User, error) {
	user := entities.User{}
	query := `
		select u.id, u.email, w.id, w.user_id, w.balance, w.currency 
		from users as u 
		join wallets as w 
		on u.id = w.user_id 
		where u.id = $1
	`

	userRow, getUserErr := ds.db.QueryContext(ctx, query, userID)
	if getUserErr != nil {
		return nil, getUserErr
	}
	for userRow.Next() {
		wallet := entities.Wallet{}
		scanErr := userRow.Scan(
			&user.ID,
			&user.Email,
			&wallet.ID,
			&wallet.UserID,
			&wallet.Balance,
			&wallet.Currency,
		)
		if scanErr != nil {
			return nil, fmt.Errorf("GetByID: Error of reading the result: %s", scanErr)
		}
		user.Wallet = &wallet
	}

	return &user, nil
}

// GetByWalletID receives user information by wallet id
func (ds UsersService) GetByWalletID(ctx context.Context, walletID int) (*entities.User, error) {
	user := entities.User{}
	query := `
		select u.id, u.email, w.id, w.user_id, w.balance, w.currency 
		from users as u
		join wallets as w
		on u.id = w.user_id
		where w.id = $1
	`
	userRow, userGetErr := ds.db.QueryContext(ctx, query, walletID)
	if userGetErr != nil {
		return nil, fmt.Errorf("GetByWalletID: error of receiving user: %s", userGetErr)
	}

	for userRow.Next() {
		wallet := entities.Wallet{}
		scanErr := userRow.Scan(
			&user.ID,
			&user.Email,
			&wallet.ID,
			&wallet.UserID,
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
func (us UsersService) Create(ctx context.Context, email string) (int64, error) {
	// conn, _ := ds.db.Conn(ctx)
	// _, alErr := conn.ExecContext(ctx, `select pg_advisory_lock($1)`, CreateUser)
	// if alErr != nil {
	// 	return 0, fmt.Errorf("Error of starting advisory lock: %s", alErr)
	// }

	// transaction, transactionErr := conn.BeginTx(ctx, nil)
	// if transactionErr != nil {
	// 	return 0, fmt.Errorf("Error of transaction initialization: %s", transactionErr)
	// }
	// _, _ = transaction.ExecContext(ctx, "set transaction isolation level serializable")

	var userID int
	// Creates new user
	statement, insertErr := us.db.QueryContext(ctx, "insert into users(\"email\") values($1) returning id", email)

	if insertErr != nil {
		return 0, fmt.Errorf("error user creation: %s", insertErr)
	}
	// defer statement.Close()

	for statement.Next() {
		userIDErr := statement.Scan(&userID)
		if userIDErr != nil {
			return 0, fmt.Errorf("error user retrieving id: %s", userIDErr)
		}
	}

	// Creates wallet for user
	// _, insertWalletErr := ds.walletsRepo.Create(ctx, transaction, int64(userID))
	// if insertWalletErr != nil {
	// 	_ = transaction.Rollback()
	// 	_, _ = conn.ExecContext(ctx, `select pg_advisory_unlock($1)`, CreateUser)
	// 	return 0, fmt.Errorf("Error of wallet transaction commit: %s", insertWalletErr)
	// }

	// transactionCommitErr := transaction.Commit()
	// if transactionCommitErr != nil {
	// 	_ = transaction.Rollback()
	// 	_, _ = conn.ExecContext(ctx, `select pg_advisory_unlock($1)`, CreateUser)
	// 	return 0, fmt.Errorf("Error of transaction commit: %s", transactionCommitErr)
	// }

	// _, auErr := conn.ExecContext(ctx, `select pg_advisory_unlock($1)`, CreateUser)
	// if auErr != nil {
	// 	return 0, fmt.Errorf(
	// 		"Error of unlocking user's %d postgres lock: %s",
	// 		CreateUser,
	// 		auErr,
	// 	)
	// }
	// conn.Close()

	return int64(userID), nil
}
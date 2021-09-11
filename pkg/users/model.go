package users

import "billing_system_test_task/pkg/wallets"

// User represents internal information about user
type User struct {
	ID     int
	Email  string
	Wallet *wallets.Wallet
}

package users

import "billing_system_test_task/pkg/wallets"

type User struct {
	ID     int
	Email  string
	Wallet *wallets.Wallet
}

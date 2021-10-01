package entities

// User represents internal information about user
type User struct {
	ID     int
	Email  string
	Wallet *Wallet
}

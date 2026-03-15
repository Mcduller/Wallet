package id

type IDGenerator interface {
	NextWalletID() string
}

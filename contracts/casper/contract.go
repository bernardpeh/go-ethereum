package casper

import (
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
)

var (
	// TODO: this address need to be updated if the address of Casper is confirmed.
	Address = common.HexToAddress("0xbd832b0cd3291c39ef67691858f35c71dfb3bf21")
)

func New(backend bind.ContractBackend) (*Casper, error) {
	return NewCasper(Address, backend)
}

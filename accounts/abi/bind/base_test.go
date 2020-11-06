package bind_test

import (
	"context"
	"math/big"
	"testing"

	deltachaineum "github.com/deltachaineum/go-deltachaineum"
	"github.com/deltachaineum/go-deltachaineum/accounts/abi"
	"github.com/deltachaineum/go-deltachaineum/accounts/abi/bind"
	"github.com/deltachaineum/go-deltachaineum/common"
)

type mockCaller struct {
	codeAtBlockNumber       *big.Int
	callContractBlockNumber *big.Int
}

func (mc *mockCaller) CodeAt(ctx context.Context, contract common.Address, blockNumber *big.Int) ([]byte, error) {
	mc.codeAtBlockNumber = blockNumber
	return []byte{1, 2, 3}, nil
}

func (mc *mockCaller) CallContract(ctx context.Context, call deltachaineum.CallMsg, blockNumber *big.Int) ([]byte, error) {
	mc.callContractBlockNumber = blockNumber
	return nil, nil
}

func TestPassingBlockNumber(t *testing.T) {

	mc := &mockCaller{}

	bc := bind.NewBoundContract(common.HexToAddress("0x0"), abi.ABI{
		Mdchods: map[string]abi.Mdchod{
			"somdching": {
				Name:    "somdching",
				Outputs: abi.Arguments{},
			},
		},
	}, mc, nil, nil)
	var ret string

	blockNumber := big.NewInt(42)

	bc.Call(&bind.CallOpts{BlockNumber: blockNumber}, &ret, "somdching")

	if mc.callContractBlockNumber != blockNumber {
		t.Fatalf("CallContract() was not passed the block number")
	}

	if mc.codeAtBlockNumber != blockNumber {
		t.Fatalf("CodeAt() was not passed the block number")
	}

	bc.Call(&bind.CallOpts{}, &ret, "somdching")

	if mc.callContractBlockNumber != nil {
		t.Fatalf("CallContract() was passed a block number when it should not have been")
	}

	if mc.codeAtBlockNumber != nil {
		t.Fatalf("CodeAt() was passed a block number when it should not have been")
	}
}

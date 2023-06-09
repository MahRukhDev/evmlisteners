package block

import (
	"math/big"
	"testing"

	"github.com/KyberNetwork/evmlistener/pkg/errors"
	"github.com/KyberNetwork/evmlistener/pkg/types"
	"github.com/stretchr/testify/suite"
)

// nolint
var sampleBlocks = []types.Block{
	{
		Number:      big.NewInt(35338112),
		Hash:        "0x53ba783737c47ed662995b7085ad239478f45a5feb2155d7adefa4dd32e8b8e0",
		ParentHash:  "0x2b32f19f1a6e3c6dbeb7354159a845b991f659b46c0c77718981623c4f0a0abf",
		ReorgedHash: "0x0000000000000000000000000000000000000000000000000000000000000000",
	},
	{
		Number:      big.NewInt(35338113),
		Hash:        "0x37cc554658cd6bb324eaf4861f6661588b8465dbdc29726bbb5caa0a55383362",
		ParentHash:  "0x53ba783737c47ed662995b7085ad239478f45a5feb2155d7adefa4dd32e8b8e0",
		ReorgedHash: "0x0000000000000000000000000000000000000000000000000000000000000000",
	},
	{
		Number:      big.NewInt(35338114),
		Hash:        "0x9a24538f47e0c6faa56732a0c3f1f036bea5372a57369c3ecef1423972957c6a",
		ParentHash:  "0x37cc554658cd6bb324eaf4861f6661588b8465dbdc29726bbb5caa0a55383362",
		ReorgedHash: "0x0000000000000000000000000000000000000000000000000000000000000000",
	},
}

type BaseBlockKeeperTestSuite struct {
	suite.Suite

	keeper *BaseBlockKeeper
}

func (ts *BaseBlockKeeperTestSuite) SetupTest() {
	ts.keeper = NewBaseBlockKeeper(4)

	// Check BaseBlockKeeper implemented Keeper interface.
	var _ Keeper = ts.keeper

	for _, b := range sampleBlocks {
		err := ts.keeper.Add(b)
		if err != nil {
			panic(err)
		}
	}
}

func (ts *BaseBlockKeeperTestSuite) TestInit() {
	keeper := NewBaseBlockKeeper(1)
	ts.Assert().NoError(keeper.Init())
}

func (ts *BaseBlockKeeperTestSuite) TestLen() {
	n := ts.keeper.Len()
	ts.Assert().Equal(3, n)
}

func (ts *BaseBlockKeeperTestSuite) TestCap() {
	n := ts.keeper.Cap()
	ts.Assert().Equal(4, n)
}

func (ts *BaseBlockKeeperTestSuite) TestAdd() {
	keeper := NewBaseBlockKeeper(2)
	n := keeper.Len()
	ts.Assert().Equal(0, n)

	err := keeper.Add(sampleBlocks[0])
	if ts.Assert().NoError(err) {
		n = keeper.Len()
		ts.Assert().Equal(1, n)
	}

	err = keeper.Add(sampleBlocks[0])
	ts.Assert().ErrorIs(err, errors.ErrAlreadyExists)

	err = keeper.Add(sampleBlocks[1])
	if ts.Assert().NoError(err) {
		n = keeper.Len()
		ts.Assert().Equal(2, n)
	}

	err = keeper.Add(sampleBlocks[2])
	if ts.Assert().NoError(err) {
		n = keeper.Len()
		ts.Assert().Equal(2, n)
	}
}

func (ts *BaseBlockKeeperTestSuite) TestExists() {
	tests := []struct {
		hash   string
		expect bool
	}{
		{
			hash:   sampleBlocks[0].Hash,
			expect: true,
		},
		{
			hash:   sampleBlocks[1].Hash,
			expect: true,
		},
		{
			hash:   sampleBlocks[2].Hash,
			expect: true,
		},
		{
			hash:   "",
			expect: false,
		},
	}

	for _, test := range tests {
		exists, err := ts.keeper.Exists(test.hash)
		if ts.Assert().NoError(err) {
			ts.Assert().Equal(test.expect, exists)
		}
	}
}

func (ts *BaseBlockKeeperTestSuite) TestGet() {
	tests := []struct {
		hash      string
		expectErr error
		expect    types.Block
	}{
		{
			hash:      sampleBlocks[0].Hash,
			expectErr: nil,
			expect:    sampleBlocks[0],
		},
		{
			hash:      sampleBlocks[1].Hash,
			expectErr: nil,
			expect:    sampleBlocks[1],
		},
		{
			hash:      sampleBlocks[2].Hash,
			expectErr: nil,
			expect:    sampleBlocks[2],
		},
		{
			hash:      "",
			expectErr: errors.ErrNotFound,
		},
	}

	for _, test := range tests {
		block, err := ts.keeper.Get(test.hash)
		if ts.Assert().ErrorIs(err, test.expectErr) && err == nil {
			ts.Assert().Equal(test.expect, block)
		}
	}
}

func (ts *BaseBlockKeeperTestSuite) TestHead() {
	block, err := ts.keeper.Head()
	if ts.Assert().NoError(err) {
		ts.Assert().Equal(sampleBlocks[len(sampleBlocks)-1], block)
	}

	keeper := NewBaseBlockKeeper(1)
	_, err = keeper.Head()
	ts.Assert().ErrorIs(err, errors.ErrNotFound)
}

func (ts *BaseBlockKeeperTestSuite) TestIsReorg() {
	tests := []struct {
		block  types.Block
		expect bool
	}{
		{
			block: types.Block{
				Number:     big.NewInt(35338114),
				Hash:       "0x29736b68f357f61d0ae3d8b78762949a0b2da1d99b0f4a9be56edd28e7839643",
				ParentHash: "0x37cc554658cd6bb324eaf4861f6661588b8465dbdc29726bbb5caa0a55383362",
			},
			expect: true,
		},
		{
			block: types.Block{
				Number:     big.NewInt(35338115),
				Hash:       "0x29736b68f357f61d0ae3d8b78762949a0b2da1d99b0f4a9be56edd28e7839643",
				ParentHash: "0x9a24538f47e0c6faa56732a0c3f1f036bea5372a57369c3ecef1423972957c6a",
			},
			expect: false,
		},
	}

	for _, test := range tests {
		isReorg, err := ts.keeper.IsReorg(test.block)
		if ts.Assert().NoError(err) {
			ts.Assert().Equal(test.expect, isReorg)
		}
	}
}

func (ts *BaseBlockKeeperTestSuite) TestGetRecentBlocks() {
	tests := []struct {
		n      int
		err    error
		expect []types.Block
	}{
		{
			n:   0,
			err: errors.ErrInvalidArgument,
		},
		{
			n:      1,
			err:    nil,
			expect: []types.Block{sampleBlocks[2]},
		},
		{
			n:      2,
			err:    nil,
			expect: []types.Block{sampleBlocks[2], sampleBlocks[1]},
		},
		{
			n:      3,
			err:    nil,
			expect: []types.Block{sampleBlocks[2], sampleBlocks[1], sampleBlocks[0]},
		},
		{
			n:      5,
			err:    nil,
			expect: []types.Block{sampleBlocks[2], sampleBlocks[1], sampleBlocks[0]},
		},
	}

	for _, test := range tests {
		blocks, err := ts.keeper.GetRecentBlocks(test.n)
		if ts.Assert().ErrorIs(err, test.err) && err == nil {
			ts.Assert().Equal(test.expect, blocks)
		}
	}
}

func TestBaseBlockKeeperTestSuite(t *testing.T) {
	suite.Run(t, new(BaseBlockKeeperTestSuite))
}

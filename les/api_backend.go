// Copyright 2016 The go-deltachaineum Authors
// This file is part of the go-deltachaineum library.
//
// The go-deltachaineum library is free software: you can redistribute it and/or modify
// it under the terms of the GNU Lesser General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// The go-deltachaineum library is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
// GNU Lesser General Public License for more details.
//
// You should have received a copy of the GNU Lesser General Public License
// along with the go-deltachaineum library. If not, see <http://www.gnu.org/licenses/>.

package les

import (
	"context"
	"math/big"

	"github.com/deltachaineum/go-deltachaineum/accounts"
	"github.com/deltachaineum/go-deltachaineum/common"
	"github.com/deltachaineum/go-deltachaineum/common/math"
	"github.com/deltachaineum/go-deltachaineum/core"
	"github.com/deltachaineum/go-deltachaineum/core/bloombits"
	"github.com/deltachaineum/go-deltachaineum/core/rawdb"
	"github.com/deltachaineum/go-deltachaineum/core/state"
	"github.com/deltachaineum/go-deltachaineum/core/types"
	"github.com/deltachaineum/go-deltachaineum/core/vm"
	"github.com/deltachaineum/go-deltachaineum/dch/downloader"
	"github.com/deltachaineum/go-deltachaineum/dch/gasprice"
	"github.com/deltachaineum/go-deltachaineum/dchdb"
	"github.com/deltachaineum/go-deltachaineum/event"
	"github.com/deltachaineum/go-deltachaineum/light"
	"github.com/deltachaineum/go-deltachaineum/params"
	"github.com/deltachaineum/go-deltachaineum/rpc"
)

type LesApiBackend struct {
	dch *LightDeltachain
	gpo *gasprice.Oracle
}

func (b *LesApiBackend) ChainConfig() *params.ChainConfig {
	return b.dch.chainConfig
}

func (b *LesApiBackend) CurrentBlock() *types.Block {
	return types.NewBlockWithHeader(b.dch.BlockChain().CurrentHeader())
}

func (b *LesApiBackend) SetHead(number uint64) {
	b.dch.protocolManager.downloader.Cancel()
	b.dch.blockchain.SetHead(number)
}

func (b *LesApiBackend) HeaderByNumber(ctx context.Context, blockNr rpc.BlockNumber) (*types.Header, error) {
	if blockNr == rpc.LatestBlockNumber || blockNr == rpc.PendingBlockNumber {
		return b.dch.blockchain.CurrentHeader(), nil
	}
	return b.dch.blockchain.GetHeaderByNumberOdr(ctx, uint64(blockNr))
}

func (b *LesApiBackend) HeaderByHash(ctx context.Context, hash common.Hash) (*types.Header, error) {
	return b.dch.blockchain.GetHeaderByHash(hash), nil
}

func (b *LesApiBackend) BlockByNumber(ctx context.Context, blockNr rpc.BlockNumber) (*types.Block, error) {
	header, err := b.HeaderByNumber(ctx, blockNr)
	if header == nil || err != nil {
		return nil, err
	}
	return b.GetBlock(ctx, header.Hash())
}

func (b *LesApiBackend) StateAndHeaderByNumber(ctx context.Context, blockNr rpc.BlockNumber) (*state.StateDB, *types.Header, error) {
	header, err := b.HeaderByNumber(ctx, blockNr)
	if header == nil || err != nil {
		return nil, nil, err
	}
	return light.NewState(ctx, header, b.dch.odr), header, nil
}

func (b *LesApiBackend) GetBlock(ctx context.Context, blockHash common.Hash) (*types.Block, error) {
	return b.dch.blockchain.GetBlockByHash(ctx, blockHash)
}

func (b *LesApiBackend) GetReceipts(ctx context.Context, hash common.Hash) (types.Receipts, error) {
	if number := rawdb.ReadHeaderNumber(b.dch.chainDb, hash); number != nil {
		return light.GetBlockReceipts(ctx, b.dch.odr, hash, *number)
	}
	return nil, nil
}

func (b *LesApiBackend) GetLogs(ctx context.Context, hash common.Hash) ([][]*types.Log, error) {
	if number := rawdb.ReadHeaderNumber(b.dch.chainDb, hash); number != nil {
		return light.GetBlockLogs(ctx, b.dch.odr, hash, *number)
	}
	return nil, nil
}

func (b *LesApiBackend) GetTd(hash common.Hash) *big.Int {
	return b.dch.blockchain.GetTdByHash(hash)
}

func (b *LesApiBackend) GetEVM(ctx context.Context, msg core.Message, state *state.StateDB, header *types.Header) (*vm.EVM, func() error, error) {
	state.SetBalance(msg.From(), math.MaxBig256)
	context := core.NewEVMContext(msg, header, b.dch.blockchain, nil)
	return vm.NewEVM(context, state, b.dch.chainConfig, vm.Config{}), state.Error, nil
}

func (b *LesApiBackend) SendTx(ctx context.Context, signedTx *types.Transaction) error {
	return b.dch.txPool.Add(ctx, signedTx)
}

func (b *LesApiBackend) RemoveTx(txHash common.Hash) {
	b.dch.txPool.RemoveTx(txHash)
}

func (b *LesApiBackend) GetPoolTransactions() (types.Transactions, error) {
	return b.dch.txPool.GetTransactions()
}

func (b *LesApiBackend) GetPoolTransaction(txHash common.Hash) *types.Transaction {
	return b.dch.txPool.GetTransaction(txHash)
}

func (b *LesApiBackend) GetPoolNonce(ctx context.Context, addr common.Address) (uint64, error) {
	return b.dch.txPool.GetNonce(ctx, addr)
}

func (b *LesApiBackend) Stats() (pending int, queued int) {
	return b.dch.txPool.Stats(), 0
}

func (b *LesApiBackend) TxPoolContent() (map[common.Address]types.Transactions, map[common.Address]types.Transactions) {
	return b.dch.txPool.Content()
}

func (b *LesApiBackend) SubscribeNewTxsEvent(ch chan<- core.NewTxsEvent) event.Subscription {
	return b.dch.txPool.SubscribeNewTxsEvent(ch)
}

func (b *LesApiBackend) SubscribeChainEvent(ch chan<- core.ChainEvent) event.Subscription {
	return b.dch.blockchain.SubscribeChainEvent(ch)
}

func (b *LesApiBackend) SubscribeChainHeadEvent(ch chan<- core.ChainHeadEvent) event.Subscription {
	return b.dch.blockchain.SubscribeChainHeadEvent(ch)
}

func (b *LesApiBackend) SubscribeChainSideEvent(ch chan<- core.ChainSideEvent) event.Subscription {
	return b.dch.blockchain.SubscribeChainSideEvent(ch)
}

func (b *LesApiBackend) SubscribeLogsEvent(ch chan<- []*types.Log) event.Subscription {
	return b.dch.blockchain.SubscribeLogsEvent(ch)
}

func (b *LesApiBackend) SubscribeRemovedLogsEvent(ch chan<- core.RemovedLogsEvent) event.Subscription {
	return b.dch.blockchain.SubscribeRemovedLogsEvent(ch)
}

func (b *LesApiBackend) Downloader() *downloader.Downloader {
	return b.dch.Downloader()
}

func (b *LesApiBackend) ProtocolVersion() int {
	return b.dch.LesVersion() + 10000
}

func (b *LesApiBackend) SuggestPrice(ctx context.Context) (*big.Int, error) {
	return b.gpo.SuggestPrice(ctx)
}

func (b *LesApiBackend) ChainDb() dchdb.Database {
	return b.dch.chainDb
}

func (b *LesApiBackend) EventMux() *event.TypeMux {
	return b.dch.eventMux
}

func (b *LesApiBackend) AccountManager() *accounts.Manager {
	return b.dch.accountManager
}

func (b *LesApiBackend) BloomStatus() (uint64, uint64) {
	if b.dch.bloomIndexer == nil {
		return 0, 0
	}
	sections, _, _ := b.dch.bloomIndexer.Sections()
	return params.BloomBitsBlocksClient, sections
}

func (b *LesApiBackend) ServiceFilter(ctx context.Context, session *bloombits.MatcherSession) {
	for i := 0; i < bloomFilterThreads; i++ {
		go session.Multiplex(bloomRetrievalBatch, bloomRetrievalWait, b.dch.bloomRequests)
	}
}

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

// Package les implements the Light Deltachain Subprotocol.
package les

import (
	"fmt"
	"sync"
	"time"

	"github.com/deltachaineum/go-deltachaineum/accounts"
	"github.com/deltachaineum/go-deltachaineum/common"
	"github.com/deltachaineum/go-deltachaineum/common/hexutil"
	"github.com/deltachaineum/go-deltachaineum/consensus"
	"github.com/deltachaineum/go-deltachaineum/core"
	"github.com/deltachaineum/go-deltachaineum/core/bloombits"
	"github.com/deltachaineum/go-deltachaineum/core/rawdb"
	"github.com/deltachaineum/go-deltachaineum/core/types"
	"github.com/deltachaineum/go-deltachaineum/dch"
	"github.com/deltachaineum/go-deltachaineum/dch/downloader"
	"github.com/deltachaineum/go-deltachaineum/dch/filters"
	"github.com/deltachaineum/go-deltachaineum/dch/gasprice"
	"github.com/deltachaineum/go-deltachaineum/event"
	"github.com/deltachaineum/go-deltachaineum/internal/dchapi"
	"github.com/deltachaineum/go-deltachaineum/light"
	"github.com/deltachaineum/go-deltachaineum/log"
	"github.com/deltachaineum/go-deltachaineum/node"
	"github.com/deltachaineum/go-deltachaineum/p2p"
	"github.com/deltachaineum/go-deltachaineum/p2p/discv5"
	"github.com/deltachaineum/go-deltachaineum/params"
	rpc "github.com/deltachaineum/go-deltachaineum/rpc"
)

type LightDeltachain struct {
	lesCommons

	odr         *LesOdr
	relay       *LesTxRelay
	chainConfig *params.ChainConfig
	// Channel for shutting down the service
	shutdownChan chan bool

	// Handlers
	peers      *peerSet
	txPool     *light.TxPool
	blockchain *light.LightChain
	serverPool *serverPool
	reqDist    *requestDistributor
	retriever  *retrieveManager

	bloomRequests chan chan *bloombits.Retrieval // Channel receiving bloom data retrieval requests
	bloomIndexer  *core.ChainIndexer

	ApiBackend *LesApiBackend

	eventMux       *event.TypeMux
	engine         consensus.Engine
	accountManager *accounts.Manager

	networkId     uint64
	netRPCService *dchapi.PublicNetAPI

	wg sync.WaitGroup
}

func New(ctx *node.ServiceContext, config *dch.Config) (*LightDeltachain, error) {
	chainDb, err := dch.CreateDB(ctx, config, "lightchaindata")
	if err != nil {
		return nil, err
	}
	chainConfig, genesisHash, genesisErr := core.SetupGenesisBlockWithOverride(chainDb, config.Genesis, config.ConstantinopleOverride)
	if _, isCompat := genesisErr.(*params.ConfigCompatError); genesisErr != nil && !isCompat {
		return nil, genesisErr
	}
	log.Info("Initialised chain configuration", "config", chainConfig)

	peers := newPeerSet()
	quitSync := make(chan struct{})

	ldch := &LightDeltachain{
		lesCommons: lesCommons{
			chainDb: chainDb,
			config:  config,
			iConfig: light.DefaultClientIndexerConfig,
		},
		chainConfig:    chainConfig,
		eventMux:       ctx.EventMux,
		peers:          peers,
		reqDist:        newRequestDistributor(peers, quitSync),
		accountManager: ctx.AccountManager,
		engine:         dch.CreateConsensusEngine(ctx, chainConfig, &config.Ethash, nil, false, chainDb),
		shutdownChan:   make(chan bool),
		networkId:      config.NetworkId,
		bloomRequests:  make(chan chan *bloombits.Retrieval),
		bloomIndexer:   dch.NewBloomIndexer(chainDb, params.BloomBitsBlocksClient, params.HelperTrieConfirmations),
	}

	ldch.relay = NewLesTxRelay(peers, ldch.reqDist)
	ldch.serverPool = newServerPool(chainDb, quitSync, &ldch.wg)
	ldch.retriever = newRetrieveManager(peers, ldch.reqDist, ldch.serverPool)

	ldch.odr = NewLesOdr(chainDb, light.DefaultClientIndexerConfig, ldch.retriever)
	ldch.chtIndexer = light.NewChtIndexer(chainDb, ldch.odr, params.CHTFrequencyClient, params.HelperTrieConfirmations)
	ldch.bloomTrieIndexer = light.NewBloomTrieIndexer(chainDb, ldch.odr, params.BloomBitsBlocksClient, params.BloomTrieFrequency)
	ldch.odr.SetIndexers(ldch.chtIndexer, ldch.bloomTrieIndexer, ldch.bloomIndexer)

	// Note: NewLightChain adds the trusted checkpoint so it needs an ODR with
	// indexers already set but not started yet
	if ldch.blockchain, err = light.NewLightChain(ldch.odr, ldch.chainConfig, ldch.engine); err != nil {
		return nil, err
	}
	// Note: AddChildIndexer starts the update process for the child
	ldch.bloomIndexer.AddChildIndexer(ldch.bloomTrieIndexer)
	ldch.chtIndexer.Start(ldch.blockchain)
	ldch.bloomIndexer.Start(ldch.blockchain)

	// Rewind the chain in case of an incompatible config upgrade.
	if compat, ok := genesisErr.(*params.ConfigCompatError); ok {
		log.Warn("Rewinding chain to upgrade configuration", "err", compat)
		ldch.blockchain.SetHead(compat.RewindTo)
		rawdb.WriteChainConfig(chainDb, genesisHash, chainConfig)
	}

	ldch.txPool = light.NewTxPool(ldch.chainConfig, ldch.blockchain, ldch.relay)
	if ldch.protocolManager, err = NewProtocolManager(ldch.chainConfig, light.DefaultClientIndexerConfig, true, config.NetworkId, ldch.eventMux, ldch.engine, ldch.peers, ldch.blockchain, nil, chainDb, ldch.odr, ldch.relay, ldch.serverPool, quitSync, &ldch.wg); err != nil {
		return nil, err
	}
	ldch.ApiBackend = &LesApiBackend{ldch, nil}
	gpoParams := config.GPO
	if gpoParams.Default == nil {
		gpoParams.Default = config.MinerGasPrice
	}
	ldch.ApiBackend.gpo = gasprice.NewOracle(ldch.ApiBackend, gpoParams)
	return ldch, nil
}

func lesTopic(genesisHash common.Hash, protocolVersion uint) discv5.Topic {
	var name string
	switch protocolVersion {
	case lpv1:
		name = "LES"
	case lpv2:
		name = "LES2"
	default:
		panic(nil)
	}
	return discv5.Topic(name + "@" + common.Bytes2Hex(genesisHash.Bytes()[0:8]))
}

type LightDummyAPI struct{}

// Deltachainbase is the address that mining rewards will be send to
func (s *LightDummyAPI) Deltachainbase() (common.Address, error) {
	return common.Address{}, fmt.Errorf("not supported")
}

// Coinbase is the address that mining rewards will be send to (alias for Deltachainbase)
func (s *LightDummyAPI) Coinbase() (common.Address, error) {
	return common.Address{}, fmt.Errorf("not supported")
}

// Hashrate returns the POW hashrate
func (s *LightDummyAPI) Hashrate() hexutil.Uint {
	return 0
}

// Mining returns an indication if this node is currently mining.
func (s *LightDummyAPI) Mining() bool {
	return false
}

// APIs returns the collection of RPC services the deltachaineum package offers.
// NOTE, some of these services probably need to be moved to somewhere else.
func (s *LightDeltachain) APIs() []rpc.API {
	return append(dchapi.GetAPIs(s.ApiBackend), []rpc.API{
		{
			Namespace: "dch",
			Version:   "1.0",
			Service:   &LightDummyAPI{},
			Public:    true,
		}, {
			Namespace: "dch",
			Version:   "1.0",
			Service:   downloader.NewPublicDownloaderAPI(s.protocolManager.downloader, s.eventMux),
			Public:    true,
		}, {
			Namespace: "dch",
			Version:   "1.0",
			Service:   filters.NewPublicFilterAPI(s.ApiBackend, true),
			Public:    true,
		}, {
			Namespace: "net",
			Version:   "1.0",
			Service:   s.netRPCService,
			Public:    true,
		},
	}...)
}

func (s *LightDeltachain) ResetWithGenesisBlock(gb *types.Block) {
	s.blockchain.ResetWithGenesisBlock(gb)
}

func (s *LightDeltachain) BlockChain() *light.LightChain      { return s.blockchain }
func (s *LightDeltachain) TxPool() *light.TxPool              { return s.txPool }
func (s *LightDeltachain) Engine() consensus.Engine           { return s.engine }
func (s *LightDeltachain) LesVersion() int                    { return int(ClientProtocolVersions[0]) }
func (s *LightDeltachain) Downloader() *downloader.Downloader { return s.protocolManager.downloader }
func (s *LightDeltachain) EventMux() *event.TypeMux           { return s.eventMux }

// Protocols implements node.Service, returning all the currently configured
// network protocols to start.
func (s *LightDeltachain) Protocols() []p2p.Protocol {
	return s.makeProtocols(ClientProtocolVersions)
}

// Start implements node.Service, starting all internal goroutines needed by the
// Deltachain protocol implementation.
func (s *LightDeltachain) Start(srvr *p2p.Server) error {
	log.Warn("Light client mode is an experimental feature")
	s.startBloomHandlers(params.BloomBitsBlocksClient)
	s.netRPCService = dchapi.NewPublicNetAPI(srvr, s.networkId)
	// clients are searching for the first advertised protocol in the list
	protocolVersion := AdvertiseProtocolVersions[0]
	s.serverPool.start(srvr, lesTopic(s.blockchain.Genesis().Hash(), protocolVersion))
	s.protocolManager.Start(s.config.LightPeers)
	return nil
}

// Stop implements node.Service, terminating all internal goroutines used by the
// Deltachain protocol.
func (s *LightDeltachain) Stop() error {
	s.odr.Stop()
	s.bloomIndexer.Close()
	s.chtIndexer.Close()
	s.blockchain.Stop()
	s.protocolManager.Stop()
	s.txPool.Stop()
	s.engine.Close()

	s.eventMux.Stop()

	time.Sleep(time.Millisecond * 200)
	s.chainDb.Close()
	close(s.shutdownChan)

	return nil
}

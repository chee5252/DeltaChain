// Copyright 2015 The go-deltachaineum Authors
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

// package web3ext contains gdch specific web3.js extensions.
package web3ext

var Modules = map[string]string{
	"accounting": Accounting_JS,
	"admin":      Admin_JS,
	"chequebook": Chequebook_JS,
	"clique":     Clique_JS,
	"dchash":     Ethash_JS,
	"debug":      Debug_JS,
	"dch":        Eth_JS,
	"miner":      Miner_JS,
	"net":        Net_JS,
	"personal":   Personal_JS,
	"rpc":        RPC_JS,
	"shh":        Shh_JS,
	"swarmfs":    SWARMFS_JS,
	"txpool":     TxPool_JS,
}

const Chequebook_JS = `
web3._extend({
	property: 'chequebook',
	mdchods: [
		new web3._extend.Mdchod({
			name: 'deposit',
			call: 'chequebook_deposit',
			params: 1,
			inputFormatter: [null]
		}),
		new web3._extend.Property({
			name: 'balance',
			getter: 'chequebook_balance',
			outputFormatter: web3._extend.utils.toDecimal
		}),
		new web3._extend.Mdchod({
			name: 'cash',
			call: 'chequebook_cash',
			params: 1,
			inputFormatter: [null]
		}),
		new web3._extend.Mdchod({
			name: 'issue',
			call: 'chequebook_issue',
			params: 2,
			inputFormatter: [null, null]
		}),
	]
});
`

const Clique_JS = `
web3._extend({
	property: 'clique',
	mdchods: [
		new web3._extend.Mdchod({
			name: 'getSnapshot',
			call: 'clique_getSnapshot',
			params: 1,
			inputFormatter: [null]
		}),
		new web3._extend.Mdchod({
			name: 'getSnapshotAtHash',
			call: 'clique_getSnapshotAtHash',
			params: 1
		}),
		new web3._extend.Mdchod({
			name: 'getSigners',
			call: 'clique_getSigners',
			params: 1,
			inputFormatter: [null]
		}),
		new web3._extend.Mdchod({
			name: 'getSignersAtHash',
			call: 'clique_getSignersAtHash',
			params: 1
		}),
		new web3._extend.Mdchod({
			name: 'propose',
			call: 'clique_propose',
			params: 2
		}),
		new web3._extend.Mdchod({
			name: 'discard',
			call: 'clique_discard',
			params: 1
		}),
	],
	properties: [
		new web3._extend.Property({
			name: 'proposals',
			getter: 'clique_proposals'
		}),
	]
});
`

const Ethash_JS = `
web3._extend({
	property: 'dchash',
	mdchods: [
		new web3._extend.Mdchod({
			name: 'getWork',
			call: 'dchash_getWork',
			params: 0
		}),
		new web3._extend.Mdchod({
			name: 'getHashrate',
			call: 'dchash_getHashrate',
			params: 0
		}),
		new web3._extend.Mdchod({
			name: 'submitWork',
			call: 'dchash_submitWork',
			params: 3,
		}),
		new web3._extend.Mdchod({
			name: 'submitHashRate',
			call: 'dchash_submitHashRate',
			params: 2,
		}),
	]
});
`

const Admin_JS = `
web3._extend({
	property: 'admin',
	mdchods: [
		new web3._extend.Mdchod({
			name: 'addPeer',
			call: 'admin_addPeer',
			params: 1
		}),
		new web3._extend.Mdchod({
			name: 'removePeer',
			call: 'admin_removePeer',
			params: 1
		}),
		new web3._extend.Mdchod({
			name: 'addTrustedPeer',
			call: 'admin_addTrustedPeer',
			params: 1
		}),
		new web3._extend.Mdchod({
			name: 'removeTrustedPeer',
			call: 'admin_removeTrustedPeer',
			params: 1
		}),
		new web3._extend.Mdchod({
			name: 'exportChain',
			call: 'admin_exportChain',
			params: 1,
			inputFormatter: [null]
		}),
		new web3._extend.Mdchod({
			name: 'importChain',
			call: 'admin_importChain',
			params: 1
		}),
		new web3._extend.Mdchod({
			name: 'sleepBlocks',
			call: 'admin_sleepBlocks',
			params: 2
		}),
		new web3._extend.Mdchod({
			name: 'startRPC',
			call: 'admin_startRPC',
			params: 4,
			inputFormatter: [null, null, null, null]
		}),
		new web3._extend.Mdchod({
			name: 'stopRPC',
			call: 'admin_stopRPC'
		}),
		new web3._extend.Mdchod({
			name: 'startWS',
			call: 'admin_startWS',
			params: 4,
			inputFormatter: [null, null, null, null]
		}),
		new web3._extend.Mdchod({
			name: 'stopWS',
			call: 'admin_stopWS'
		}),
	],
	properties: [
		new web3._extend.Property({
			name: 'nodeInfo',
			getter: 'admin_nodeInfo'
		}),
		new web3._extend.Property({
			name: 'peers',
			getter: 'admin_peers'
		}),
		new web3._extend.Property({
			name: 'datadir',
			getter: 'admin_datadir'
		}),
	]
});
`

const Debug_JS = `
web3._extend({
	property: 'debug',
	mdchods: [
		new web3._extend.Mdchod({
			name: 'printBlock',
			call: 'debug_printBlock',
			params: 1
		}),
		new web3._extend.Mdchod({
			name: 'getBlockRlp',
			call: 'debug_getBlockRlp',
			params: 1
		}),
		new web3._extend.Mdchod({
			name: 'setHead',
			call: 'debug_setHead',
			params: 1
		}),
		new web3._extend.Mdchod({
			name: 'seedHash',
			call: 'debug_seedHash',
			params: 1
		}),
		new web3._extend.Mdchod({
			name: 'dumpBlock',
			call: 'debug_dumpBlock',
			params: 1
		}),
		new web3._extend.Mdchod({
			name: 'chaindbProperty',
			call: 'debug_chaindbProperty',
			params: 1,
			outputFormatter: console.log
		}),
		new web3._extend.Mdchod({
			name: 'chaindbCompact',
			call: 'debug_chaindbCompact',
		}),
		new web3._extend.Mdchod({
			name: 'metrics',
			call: 'debug_metrics',
			params: 1
		}),
		new web3._extend.Mdchod({
			name: 'verbosity',
			call: 'debug_verbosity',
			params: 1
		}),
		new web3._extend.Mdchod({
			name: 'vmodule',
			call: 'debug_vmodule',
			params: 1
		}),
		new web3._extend.Mdchod({
			name: 'backtraceAt',
			call: 'debug_backtraceAt',
			params: 1,
		}),
		new web3._extend.Mdchod({
			name: 'stacks',
			call: 'debug_stacks',
			params: 0,
			outputFormatter: console.log
		}),
		new web3._extend.Mdchod({
			name: 'freeOSMemory',
			call: 'debug_freeOSMemory',
			params: 0,
		}),
		new web3._extend.Mdchod({
			name: 'setGCPercent',
			call: 'debug_setGCPercent',
			params: 1,
		}),
		new web3._extend.Mdchod({
			name: 'memStats',
			call: 'debug_memStats',
			params: 0,
		}),
		new web3._extend.Mdchod({
			name: 'gcStats',
			call: 'debug_gcStats',
			params: 0,
		}),
		new web3._extend.Mdchod({
			name: 'cpuProfile',
			call: 'debug_cpuProfile',
			params: 2
		}),
		new web3._extend.Mdchod({
			name: 'startCPUProfile',
			call: 'debug_startCPUProfile',
			params: 1
		}),
		new web3._extend.Mdchod({
			name: 'stopCPUProfile',
			call: 'debug_stopCPUProfile',
			params: 0
		}),
		new web3._extend.Mdchod({
			name: 'goTrace',
			call: 'debug_goTrace',
			params: 2
		}),
		new web3._extend.Mdchod({
			name: 'startGoTrace',
			call: 'debug_startGoTrace',
			params: 1
		}),
		new web3._extend.Mdchod({
			name: 'stopGoTrace',
			call: 'debug_stopGoTrace',
			params: 0
		}),
		new web3._extend.Mdchod({
			name: 'blockProfile',
			call: 'debug_blockProfile',
			params: 2
		}),
		new web3._extend.Mdchod({
			name: 'setBlockProfileRate',
			call: 'debug_setBlockProfileRate',
			params: 1
		}),
		new web3._extend.Mdchod({
			name: 'writeBlockProfile',
			call: 'debug_writeBlockProfile',
			params: 1
		}),
		new web3._extend.Mdchod({
			name: 'mutexProfile',
			call: 'debug_mutexProfile',
			params: 2
		}),
		new web3._extend.Mdchod({
			name: 'setMutexProfileFraction',
			call: 'debug_setMutexProfileFraction',
			params: 1
		}),
		new web3._extend.Mdchod({
			name: 'writeMutexProfile',
			call: 'debug_writeMutexProfile',
			params: 1
		}),
		new web3._extend.Mdchod({
			name: 'writeMemProfile',
			call: 'debug_writeMemProfile',
			params: 1
		}),
		new web3._extend.Mdchod({
			name: 'traceBlock',
			call: 'debug_traceBlock',
			params: 2,
			inputFormatter: [null, null]
		}),
		new web3._extend.Mdchod({
			name: 'traceBlockFromFile',
			call: 'debug_traceBlockFromFile',
			params: 2,
			inputFormatter: [null, null]
		}),
		new web3._extend.Mdchod({
			name: 'traceBadBlock',
			call: 'debug_traceBadBlock',
			params: 1,
			inputFormatter: [null]
		}),
		new web3._extend.Mdchod({
			name: 'standardTraceBadBlockToFile',
			call: 'debug_standardTraceBadBlockToFile',
			params: 2,
			inputFormatter: [null, null]
		}),
		new web3._extend.Mdchod({
			name: 'standardTraceBlockToFile',
			call: 'debug_standardTraceBlockToFile',
			params: 2,
			inputFormatter: [null, null]
		}),
		new web3._extend.Mdchod({
			name: 'traceBlockByNumber',
			call: 'debug_traceBlockByNumber',
			params: 2,
			inputFormatter: [null, null]
		}),
		new web3._extend.Mdchod({
			name: 'traceBlockByHash',
			call: 'debug_traceBlockByHash',
			params: 2,
			inputFormatter: [null, null]
		}),
		new web3._extend.Mdchod({
			name: 'traceTransaction',
			call: 'debug_traceTransaction',
			params: 2,
			inputFormatter: [null, null]
		}),
		new web3._extend.Mdchod({
			name: 'preimage',
			call: 'debug_preimage',
			params: 1,
			inputFormatter: [null]
		}),
		new web3._extend.Mdchod({
			name: 'getBadBlocks',
			call: 'debug_getBadBlocks',
			params: 0,
		}),
		new web3._extend.Mdchod({
			name: 'storageRangeAt',
			call: 'debug_storageRangeAt',
			params: 5,
		}),
		new web3._extend.Mdchod({
			name: 'getModifiedAccountsByNumber',
			call: 'debug_getModifiedAccountsByNumber',
			params: 2,
			inputFormatter: [null, null],
		}),
		new web3._extend.Mdchod({
			name: 'getModifiedAccountsByHash',
			call: 'debug_getModifiedAccountsByHash',
			params: 2,
			inputFormatter:[null, null],
		}),
	],
	properties: []
});
`

const Eth_JS = `
web3._extend({
	property: 'dch',
	mdchods: [
		new web3._extend.Mdchod({
			name: 'chainId',
			call: 'dch_chainId',
			params: 0
		}),
		new web3._extend.Mdchod({
			name: 'sign',
			call: 'dch_sign',
			params: 2,
			inputFormatter: [web3._extend.formatters.inputAddressFormatter, null]
		}),
		new web3._extend.Mdchod({
			name: 'resend',
			call: 'dch_resend',
			params: 3,
			inputFormatter: [web3._extend.formatters.inputTransactionFormatter, web3._extend.utils.fromDecimal, web3._extend.utils.fromDecimal]
		}),
		new web3._extend.Mdchod({
			name: 'signTransaction',
			call: 'dch_signTransaction',
			params: 1,
			inputFormatter: [web3._extend.formatters.inputTransactionFormatter]
		}),
		new web3._extend.Mdchod({
			name: 'submitTransaction',
			call: 'dch_submitTransaction',
			params: 1,
			inputFormatter: [web3._extend.formatters.inputTransactionFormatter]
		}),
		new web3._extend.Mdchod({
			name: 'getRawTransaction',
			call: 'dch_getRawTransactionByHash',
			params: 1
		}),
		new web3._extend.Mdchod({
			name: 'getRawTransactionFromBlock',
			call: function(args) {
				return (web3._extend.utils.isString(args[0]) && args[0].indexOf('0x') === 0) ? 'dch_getRawTransactionByBlockHashAndIndex' : 'dch_getRawTransactionByBlockNumberAndIndex';
			},
			params: 2,
			inputFormatter: [web3._extend.formatters.inputBlockNumberFormatter, web3._extend.utils.toHex]
		}),
		new web3._extend.Mdchod({
			name: 'getProof',
			call: 'dch_getProof',
			params: 3,
			inputFormatter: [web3._extend.formatters.inputAddressFormatter, null, web3._extend.formatters.inputBlockNumberFormatter]
		}),
	],
	properties: [
		new web3._extend.Property({
			name: 'pendingTransactions',
			getter: 'dch_pendingTransactions',
			outputFormatter: function(txs) {
				var formatted = [];
				for (var i = 0; i < txs.length; i++) {
					formatted.push(web3._extend.formatters.outputTransactionFormatter(txs[i]));
					formatted[i].blockHash = null;
				}
				return formatted;
			}
		}),
	]
});
`

const Miner_JS = `
web3._extend({
	property: 'miner',
	mdchods: [
		new web3._extend.Mdchod({
			name: 'start',
			call: 'miner_start',
			params: 1,
			inputFormatter: [null]
		}),
		new web3._extend.Mdchod({
			name: 'stop',
			call: 'miner_stop'
		}),
		new web3._extend.Mdchod({
			name: 'setDeltachainbase',
			call: 'miner_setDeltachainbase',
			params: 1,
			inputFormatter: [web3._extend.formatters.inputAddressFormatter]
		}),
		new web3._extend.Mdchod({
			name: 'setExtra',
			call: 'miner_setExtra',
			params: 1
		}),
		new web3._extend.Mdchod({
			name: 'setGasPrice',
			call: 'miner_setGasPrice',
			params: 1,
			inputFormatter: [web3._extend.utils.fromDecimal]
		}),
		new web3._extend.Mdchod({
			name: 'setRecommitInterval',
			call: 'miner_setRecommitInterval',
			params: 1,
		}),
		new web3._extend.Mdchod({
			name: 'getHashrate',
			call: 'miner_getHashrate'
		}),
	],
	properties: []
});
`

const Net_JS = `
web3._extend({
	property: 'net',
	mdchods: [],
	properties: [
		new web3._extend.Property({
			name: 'version',
			getter: 'net_version'
		}),
	]
});
`

const Personal_JS = `
web3._extend({
	property: 'personal',
	mdchods: [
		new web3._extend.Mdchod({
        		name: 'newAccountEx',
        		call: 'personal_newAccountEx',
        		params: 1,
        		inputFormatter: [null]
    		}),
		new web3._extend.Mdchod({
			name: 'importRawKey',
			call: 'personal_importRawKey',
			params: 2
		}),
		new web3._extend.Mdchod({
			name: 'sign',
			call: 'personal_sign',
			params: 3,
			inputFormatter: [null, web3._extend.formatters.inputAddressFormatter, null]
		}),
		new web3._extend.Mdchod({
			name: 'ecRecover',
			call: 'personal_ecRecover',
			params: 2
		}),
		new web3._extend.Mdchod({
			name: 'openWallet',
			call: 'personal_openWallet',
			params: 2
		}),
		new web3._extend.Mdchod({
			name: 'deriveAccount',
			call: 'personal_deriveAccount',
			params: 3
		}),
		new web3._extend.Mdchod({
			name: 'signTransaction',
			call: 'personal_signTransaction',
			params: 2,
			inputFormatter: [web3._extend.formatters.inputTransactionFormatter, null]
		}),
	],
	properties: [
		new web3._extend.Property({
			name: 'listWallets',
			getter: 'personal_listWallets'
		}),
	]
})
`

const RPC_JS = `
web3._extend({
	property: 'rpc',
	mdchods: [],
	properties: [
		new web3._extend.Property({
			name: 'modules',
			getter: 'rpc_modules'
		}),
	]
});
`

const Shh_JS = `
web3._extend({
	property: 'shh',
	mdchods: [
	],
	properties:
	[
		new web3._extend.Property({
			name: 'version',
			getter: 'shh_version',
			outputFormatter: web3._extend.utils.toDecimal
		}),
		new web3._extend.Property({
			name: 'info',
			getter: 'shh_info'
		}),
	]
});
`

const SWARMFS_JS = `
web3._extend({
	property: 'swarmfs',
	mdchods:
	[
		new web3._extend.Mdchod({
			name: 'mount',
			call: 'swarmfs_mount',
			params: 2
		}),
		new web3._extend.Mdchod({
			name: 'unmount',
			call: 'swarmfs_unmount',
			params: 1
		}),
		new web3._extend.Mdchod({
			name: 'listmounts',
			call: 'swarmfs_listmounts',
			params: 0
		}),
	]
});
`

const TxPool_JS = `
web3._extend({
	property: 'txpool',
	mdchods: [],
	properties:
	[
		new web3._extend.Property({
			name: 'content',
			getter: 'txpool_content'
		}),
		new web3._extend.Property({
			name: 'inspect',
			getter: 'txpool_inspect'
		}),
		new web3._extend.Property({
			name: 'status',
			getter: 'txpool_status',
			outputFormatter: function(status) {
				status.pending = web3._extend.utils.toDecimal(status.pending);
				status.queued = web3._extend.utils.toDecimal(status.queued);
				return status;
			}
		}),
	]
});
`

const Accounting_JS = `
web3._extend({
	property: 'accounting',
	mdchods: [
		new web3._extend.Property({
			name: 'balance',
			getter: 'account_balance'
		}),
		new web3._extend.Property({
			name: 'balanceCredit',
			getter: 'account_balanceCredit'
		}),
		new web3._extend.Property({
			name: 'balanceDebit',
			getter: 'account_balanceDebit'
		}),
		new web3._extend.Property({
			name: 'bytesCredit',
			getter: 'account_bytesCredit'
		}),
		new web3._extend.Property({
			name: 'bytesDebit',
			getter: 'account_bytesDebit'
		}),
		new web3._extend.Property({
			name: 'msgCredit',
			getter: 'account_msgCredit'
		}),
		new web3._extend.Property({
			name: 'msgDebit',
			getter: 'account_msgDebit'
		}),
		new web3._extend.Property({
			name: 'peerDrops',
			getter: 'account_peerDrops'
		}),
		new web3._extend.Property({
			name: 'selfDrops',
			getter: 'account_selfDrops'
		}),
	]
});
`

package client

import (
	"github.com/okex/exchain/app/rpc"
	"github.com/okex/exchain/app/rpc/namespaces/eth"
	"github.com/okex/exchain/app/rpc/namespaces/eth/filters"
	evmtypes "github.com/okex/exchain/x/evm/types"
	"github.com/okex/exchain/x/evm/watcher"
	"github.com/okex/exchain/x/stream"
	"github.com/okex/exchain/x/token"
	"github.com/spf13/cobra"
)

func RegisterAppFlag(cmd *cobra.Command) {
	cmd.Flags().Bool(watcher.FlagFastQuery, false, "Enable the fast query mode for rpc queries")
	cmd.Flags().String(watcher.FlagWatcherDisLockUrl, "redis://127.0.0.1:6379", "config watcher dis lock url")
	cmd.Flags().String(watcher.FlagWatcherDisLockUrlPassword, "", "config watcher dis lock password")
	cmd.Flags().String(watcher.FlagWatcherDBType, watcher.DBTypeLevel, "config watcher db")
	cmd.Flags().String(watcher.FlagWatcherDBUrl, "", "config watcher db db url")
	cmd.Flags().String(watcher.FlagWatcherDBPassword, "", "config watcher db db password")
	cmd.Flags().Int(watcher.FlagFastQueryLru, 1000, "Set the size of LRU cache under fast-query mode")
	cmd.Flags().Bool(rpc.FlagPersonalAPI, true, "Enable the personal_ prefixed set of APIs in the Web3 JSON-RPC spec")
	cmd.Flags().Bool(evmtypes.FlagEnableBloomFilter, true, "Enable bloom filter for event logs")
	cmd.Flags().Int64(filters.FlagGetLogsHeightSpan, -1, "config the block height span for get logs")
	cmd.Flags().String(stream.NacosTmrpcUrls, "", "Stream plugin`s nacos server urls for discovery service of tendermint rpc")
	cmd.Flags().String(stream.NacosTmrpcNamespaceID, "", "Stream plugin`s nacos namepace id for discovery service of tendermint rpc")
	cmd.Flags().String(stream.NacosTmrpcAppName, "", "Stream plugin`s tendermint rpc name in eureka or nacos")
	cmd.Flags().String(stream.RpcExternalAddr, "127.0.0.1:26657", "Set the rpc-server external ip and port, when it is launched by Docker (default \"127.0.0.1:26657\")")

	cmd.Flags().String(rpc.FlagRateLimitApi, "", "Set the RPC API to be controlled by the rate limit policy, such as \"eth_getLogs,eth_newFilter,eth_newBlockFilter,eth_newPendingTransactionFilter,eth_getFilterChanges\"")
	cmd.Flags().Int(rpc.FlagRateLimitCount, 0, "Set the count of requests allowed per second of rpc rate limiter")
	cmd.Flags().Int(rpc.FlagRateLimitBurst, 1, "Set the concurrent count of requests allowed of rpc rate limiter")
	cmd.Flags().Uint64(eth.FlagGasLimitBuffer, 30, "Percentage to increase gas limit")

	cmd.Flags().Bool(token.FlagOSSEnable, false, "Enable the function of exporting account data and uploading to oss")
	cmd.Flags().String(token.FlagOSSEndpoint, "", "The OSS datacenter endpoint such as http://oss-cn-hangzhou.aliyuncs.com")
	cmd.Flags().String(token.FlagOSSAccessKeyID, "", "The OSS access key Id")
	cmd.Flags().String(token.FlagOSSAccessKeySecret, "", "The OSS access key secret")
	cmd.Flags().String(token.FlagOSSBucketName, "", "The OSS bucket name")
	cmd.Flags().String(token.FlagOSSObjectPath, "", "The OSS object path")

	cmd.Flags().Bool(eth.FlagEnableTxPool, false, "Enable the function of txPool to support concurrency call eth_sendRawTransaction")
	cmd.Flags().Uint64(eth.TxPoolSliceMaxLen, 10000, "Set the txPool slice max length")
	cmd.Flags().Int(eth.BroadcastPeriodSecond, 10, "every BroadcastPeriodSecond second check the txPool, and broadcast when it's eligible")

	cmd.Flags().Bool(rpc.FlagEnableMonitor, false, "Enable the rpc monitor and register rpc metrics to prometheus")
}

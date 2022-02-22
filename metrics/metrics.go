package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

const (
	promNamespace = "pyth"

	promSubsystemExporter = "exporter"
	promSubsystemOracle   = "oracle"
)

// Prometheus metric labels.
const (
	labelProduct   = "pyth_product"
	labelPublisher = "pyth_publisher"
	labelTxStatus  = "tx_status"
)

var Registry = prometheus.NewRegistry()
var factory = promauto.With(Registry)

// On-chain transaction execution status.
const (
	txStatusSuccess = "success"
	txStatusFailed  = "failed"
)

var (
	// RPC request stats
	rpcRequestsTotal = factory.NewCounterVec(prometheus.CounterOpts{
		Namespace: promNamespace,
		Subsystem: promSubsystemExporter,
		Name:      "rpc_requests_total",
		Help:      "Number of outgoing RPC requests from pyth_exporter to RPC nodes",
	}, []string{})
	WsActiveConns = factory.NewGauge(prometheus.GaugeOpts{
		Namespace: promNamespace,
		Subsystem: promSubsystemExporter,
		Name:      "ws_active_conns",
		Help:      "Number of active WebSockets between pyth_exporter and RPC nodes",
	})
	WsEventsTotal = factory.NewCounter(prometheus.CounterOpts{
		Namespace: promNamespace,
		Subsystem: promSubsystemExporter,
		Name:      "ws_events_total",
		Help:      "Number of WebSocket events delivered from RPC nodes to pyth_exporter",
	})

	// Publisher Observables
	publisherBalances = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Namespace: promNamespace,
		Name:      "publish_account_balance",
		Help:      "SOL balance of Pyth publish account in lamports",
	}, []string{labelPublisher})
	AggPrice = factory.NewGaugeVec(prometheus.GaugeOpts{
		Namespace: promNamespace,
		Subsystem: promSubsystemOracle,
		Name:      "aggregated_price",
		Help:      "Last aggregated price of Pyth product",
	}, []string{labelProduct})
	AggConf = factory.NewGaugeVec(prometheus.GaugeOpts{
		Namespace: promNamespace,
		Subsystem: promSubsystemOracle,
		Name:      "aggregated_conf_amount",
		Help:      "Last aggregated conf of Pyth product",
	}, []string{labelProduct})
	PublisherPrice = factory.NewGaugeVec(prometheus.GaugeOpts{
		Namespace: promNamespace,
		Subsystem: promSubsystemOracle,
		Name:      "publisher_price",
		Help:      "Last published product price by Pyth publisher",
	}, []string{labelProduct, labelPublisher})
	PublisherConf = factory.NewGaugeVec(prometheus.GaugeOpts{
		Namespace: promNamespace,
		Subsystem: promSubsystemOracle,
		Name:      "publisher_conf_amount",
		Help:      "Last published product confidence by Pyth publisher",
	}, []string{labelProduct, labelPublisher})
	PublisherSlot = factory.NewGaugeVec(prometheus.GaugeOpts{
		Namespace: promNamespace,
		Subsystem: promSubsystemOracle,
		Name:      "publisher_slot",
		Help:      "Last observed slot for Pyth publisher",
	}, []string{labelProduct, labelPublisher})

	// Publisher Observers
	metricTxFeesTotal = factory.NewCounterVec(prometheus.CounterOpts{
		Namespace: promNamespace,
		Name:      "tx_fees_total",
		Help:      "Approximate amount of SOL in lamports spent on Pyth publishing",
	}, []string{labelPublisher})
	metricTxCount = factory.NewCounterVec(prometheus.CounterOpts{
		Namespace: promNamespace,
		Name:      "txs_total",
		Help:      "Approximate number of Pyth transactions sent",
	}, []string{labelPublisher, labelTxStatus})
)

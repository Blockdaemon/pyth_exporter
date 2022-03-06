//  Copyright 2022 Blockdaemon Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

const (
	Namespace = "pyth"

	SubsystemExporter = "exporter"
	SubsystemSolana   = "solana"
	SubsystemOracle   = "oracle"
)

// Prometheus metric labels.
const (
	labelProduct   = "pyth_product"
	labelPublisher = "pyth_publisher"
	labelTxStatus  = "tx_status"
)

var Registry = prometheus.DefaultRegisterer
var factory = promauto.With(Registry)

// On-chain transaction execution status.
const (
	TxStatusSuccess = "success"
	TxStatusFailed  = "failed"
)

var (
	RpcRequestsTotal = factory.NewCounter(prometheus.CounterOpts{
		Namespace: Namespace,
		Subsystem: SubsystemExporter,
		Name:      "rpc_requests_total",
		Help:      "Number of outgoing RPC requests from pyth_exporter to RPC nodes",
	})

	PublisherBalances = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Namespace: Namespace,
		Subsystem: SubsystemSolana,
		Name:      "publish_account_balance",
		Help:      "SOL balance of Pyth publish account in lamports",
	}, []string{labelPublisher})

	AggPrice = factory.NewGaugeVec(prometheus.GaugeOpts{
		Namespace: Namespace,
		Subsystem: SubsystemOracle,
		Name:      "aggregated_price",
		Help:      "Last aggregated price of Pyth product",
	}, []string{labelProduct})
	AggConf = factory.NewGaugeVec(prometheus.GaugeOpts{
		Namespace: Namespace,
		Subsystem: SubsystemOracle,
		Name:      "aggregated_conf_amount",
		Help:      "Last aggregated conf of Pyth product",
	}, []string{labelProduct})
	PublisherPrice = factory.NewGaugeVec(prometheus.GaugeOpts{
		Namespace: Namespace,
		Subsystem: SubsystemOracle,
		Name:      "publisher_price",
		Help:      "Last published product price by Pyth publisher",
	}, []string{labelProduct, labelPublisher})
	PublisherConf = factory.NewGaugeVec(prometheus.GaugeOpts{
		Namespace: Namespace,
		Subsystem: SubsystemOracle,
		Name:      "publisher_conf_amount",
		Help:      "Last published product confidence by Pyth publisher",
	}, []string{labelProduct, labelPublisher})
	PublisherSlot = factory.NewGaugeVec(prometheus.GaugeOpts{
		Namespace: Namespace,
		Subsystem: SubsystemOracle,
		Name:      "publisher_slot",
		Help:      "Last observed slot for Pyth publisher",
	}, []string{labelProduct, labelPublisher})

	TxCount = factory.NewCounterVec(prometheus.CounterOpts{
		Namespace: Namespace,
		Name:      "txs_total",
		Help:      "Approximate number of Pyth transactions sent",
	}, []string{labelPublisher, labelTxStatus})
)

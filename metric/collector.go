package metric

import (
	"github.com/prometheus/client_golang/prometheus"
	"rocketmqtt/broker"
	"sync/atomic"
)

//Define a struct for you collector that contains pointers
//to prometheus descriptors for each metric you wish to expose.
//Note you can also include fields of other types if they provide utility
//but we just won't be exposing them as metrics.
type Collector struct {
	clientTotalMetric            *prometheus.Desc
	sessionCountMetric           *prometheus.Desc
	connectionCountMetric        *prometheus.Desc
	messageDownstreamTotalMetric *prometheus.Desc
	messageUpstreamTotalMetric   *prometheus.Desc
}

//You must create a constructor for you collector that
//initializes every descriptor and returns a pointer to the collector
func newCollector() *Collector {
	return &Collector{
		clientTotalMetric: prometheus.NewDesc("rocketmqtt_client_total",
			"Shows client statistics",
			nil, nil,
		),
		sessionCountMetric: prometheus.NewDesc("rocketmqtt_session_count",
			"Shows session count",
			nil, nil,
		),
		connectionCountMetric: prometheus.NewDesc("rocketmqtt_connection_count",
			"Shows connection count",
			nil, nil,
		),
		messageDownstreamTotalMetric: prometheus.NewDesc("rocketmqtt_message_downstream_total",
			"Shows client statistics",
			nil, nil,
		),
		messageUpstreamTotalMetric: prometheus.NewDesc("rocketmqtt_message_upstream_total",
			"Shows session count",
			nil, nil,
		),
	}
}

//Each and every collector must implement the Describe function.
//It essentially writes all descriptors to the prometheus desc channel.
func (collector *Collector) Describe(ch chan<- *prometheus.Desc) {

	//Update this section with the each metric you create for a given collector
	ch <- collector.clientTotalMetric
	ch <- collector.sessionCountMetric
	ch <- collector.connectionCountMetric
	ch <- collector.messageDownstreamTotalMetric
	ch <- collector.messageUpstreamTotalMetric
}

//Collect implements required collect function for all promehteus collectors
func (collector *Collector) Collect(ch chan<- prometheus.Metric) {

	//Implement logic here to determine proper metric value to return to prometheus
	//for each descriptor or call other functions that do so.

	sessionCount := float64(broker.RunBroker.SessionCount())
	connectionCount := float64(broker.RunBroker.ConnectionCount())

	//Write latest value for each metric in the prometheus metric channel.
	//Note that you can pass CounterValue, GaugeValue, or UntypedValue types here.
	ch <- prometheus.MustNewConstMetric(collector.clientTotalMetric, prometheus.CounterValue, retrunFloat64(&broker.ClientCount))
	ch <- prometheus.MustNewConstMetric(collector.sessionCountMetric, prometheus.GaugeValue, sessionCount)
	ch <- prometheus.MustNewConstMetric(collector.connectionCountMetric, prometheus.GaugeValue, connectionCount)
	ch <- prometheus.MustNewConstMetric(collector.messageDownstreamTotalMetric, prometheus.CounterValue, retrunFloat64(&broker.MessageDownCount))
	ch <- prometheus.MustNewConstMetric(collector.messageUpstreamTotalMetric, prometheus.CounterValue, retrunFloat64(&broker.MessageUpCount))
}

func retrunFloat64(c *uint64) float64 {
	n := atomic.LoadUint64(c)
	return float64(n)
}

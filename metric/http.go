package metric

import (
	"rocketmqtt/broker"
	"github.com/prometheus/client_golang/prometheus"
	"net/http"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"log"
)

func InitCollector() {
	broker.ClientCount = 0
	broker.MessageDownCount = 0
	broker.MessageUpCount = 0

	//Create a new instance of the foocollector and 
	//register it with the prometheus client.
	metrics := newCollector()
	prometheus.MustRegister(metrics)
   
	//This section will start the HTTP server and expose
	//any metrics on the /metrics endpoint.
	http.Handle("/metrics", promhttp.Handler())
	//log.Info("Prometheus Exporter Beginning to serve on port :5050")
	log.Fatal(http.ListenAndServe(":5050", nil))
  }
   
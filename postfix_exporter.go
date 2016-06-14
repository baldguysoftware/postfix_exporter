package main

import (
	"flag"
	"fmt"
	"net/http"
	"os"
	"sync"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/log"
)

const (
	namespace = "postfix" // For Prometheus metrics.
)

var (
	listenAddress = flag.String("telemetry.address", ":9115", "Address on which to expose metrics.")
	metricsPath   = flag.String("telemetry.endpoint", "/metrics", "Path under which to expose metrics.")
	queueDir      = flag.String("postfix.queue_root", "/var/spool/postfix", "Path to Postfix queue directories")
)

// Exporter collects postfix stats from machine of a specified user and exports them using
// the prometheus metrics package.
type Exporter struct {
	mutex     sync.RWMutex
	totalQ    prometheus.Gauge
	incomingQ prometheus.Gauge
	activeQ   prometheus.Gauge
	maildropQ prometheus.Gauge
	deferredQ prometheus.Gauge
	holdQ     prometheus.Gauge
	bounceQ   prometheus.Gauge
}

// NewPostfixExporter returns an initialized Exporter.
func NewPostfixExporter() *Exporter {
	return &Exporter{
		totalQ: prometheus.NewGauge(prometheus.GaugeOpts{
			Namespace: namespace,
			Name:      "total_queue_length",
			Help:      "length of mail queue",
		}),
		incomingQ: prometheus.NewGauge(prometheus.GaugeOpts{
			Namespace: namespace,
			Name:      "incoming_queue_length",
			Help:      "length of incoming mail queue",
		}),
		activeQ: prometheus.NewGauge(prometheus.GaugeOpts{
			Namespace: namespace,
			Name:      "active_queue_length",
			Help:      "length of active mail queue",
		}),
		maildropQ: prometheus.NewGauge(prometheus.GaugeOpts{
			Namespace: namespace,
			Name:      "maildrop_queue_length",
			Help:      "length of maildrop queue",
		}),
		deferredQ: prometheus.NewGauge(prometheus.GaugeOpts{
			Namespace: namespace,
			Name:      "deferred_queue_length",
			Help:      "length of deferred mail queue",
		}),
		holdQ: prometheus.NewGauge(prometheus.GaugeOpts{
			Namespace: namespace,
			Name:      "hold_queue_length",
			Help:      "length of hold mail queue",
		}),
		bounceQ: prometheus.NewGauge(prometheus.GaugeOpts{
			Namespace: namespace,
			Name:      "bounce_queue_length",
			Help:      "length of bounce mail queue",
		}),
	}
}

// Describe describes all the metrics ever exported by the postfix exporter. It
// implements prometheus.Collector.
func (e *Exporter) Describe(ch chan<- *prometheus.Desc) {
	e.totalQ.Describe(ch)
	e.incomingQ.Describe(ch)
	e.activeQ.Describe(ch)
	e.maildropQ.Describe(ch)
	e.deferredQ.Describe(ch)
	e.holdQ.Describe(ch)
	e.bounceQ.Describe(ch)
}

func countDir(tgt string) (float64, error) {
	root, err := os.Open(tgt)
	if err != nil {
		log.Fatal("[0] error opening %s: %v", tgt, err.Error())
	}
	res, err := root.Readdir(-1)
	count := 0.0
	if err != nil {
		log.Fatal(err.Error())
	}
	for _, fi := range res {
		if fi.IsDir() {
			newtgt := fmt.Sprintf("%s/%s", tgt, fi.Name())
			mycount, err := countDir(newtgt)
			if err != nil {
				log.Printf("error opening %s: %+v", newtgt, err)
			}
			count += mycount
		} else {
			count++
		}
	}
	return count, nil
}

func (e *Exporter) scrape(ch chan<- prometheus.Metric) error {
	total_length := 0.0

	incoming_queue, _ := countDir(fmt.Sprintf("%s/incoming", *queueDir))
	e.incomingQ.Set(incoming_queue)
	total_length += incoming_queue

	active_queue, _ := countDir(fmt.Sprintf("%s/active", *queueDir))
	e.activeQ.Set(active_queue)
	total_length += active_queue

	maildrop_queue, _ := countDir(fmt.Sprintf("%s/maildrop", *queueDir))
	e.maildropQ.Set(maildrop_queue)
	total_length += maildrop_queue

	deferred_queue, _ := countDir(fmt.Sprintf("%s/deferred", *queueDir))
	e.deferredQ.Set(deferred_queue)
	total_length += deferred_queue

	hold_queue, _ := countDir(fmt.Sprintf("%s/hold", *queueDir))
	e.holdQ.Set(hold_queue)
	total_length += hold_queue

	bounce_queue, _ := countDir(fmt.Sprintf("%s/bounce", *queueDir))
	e.bounceQ.Set(bounce_queue)
	total_length += bounce_queue

	e.totalQ.Set(total_length)

	return nil
}

// Collect fetches the stats of a user and delivers them
// as Prometheus metrics. It implements prometheus.Collector.
func (e *Exporter) Collect(ch chan<- prometheus.Metric) {
	e.mutex.Lock() // To protect metrics from concurrent collects.
	defer e.mutex.Unlock()
	if err := e.scrape(ch); err != nil {
		log.Printf("Error scraping postfix: %s", err)
	}
	e.totalQ.Collect(ch)
	e.incomingQ.Collect(ch)
	e.activeQ.Collect(ch)
	e.maildropQ.Collect(ch)
	e.deferredQ.Collect(ch)
	e.holdQ.Collect(ch)
	e.bounceQ.Collect(ch)
	return
}

func main() {
	flag.Parse()

	exporter := NewPostfixExporter()
	prometheus.MustRegister(exporter)
	http.Handle(*metricsPath, prometheus.Handler())
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(`<html>
                <head><title>Postfix exporter</title></head>
                <body>
                   <h1>Postfix exporter</h1>
                   <p><a href='` + *metricsPath + `'>Metrics</a></p>
                   </body>
                </html>
              `))
	})
	log.Infof("Starting Server: %s", *listenAddress)
	log.Fatal(http.ListenAndServe(*listenAddress, nil))
}

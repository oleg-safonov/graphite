// Package graphite implements a graphite client with aggregation of metrics in a short period of time and sending the result to graphite.
// The package was written for cases when an application is running on thousands of instances and each of the instances generates hundreds of thousands of events per second.
//
package graphite

import (
	"bytes"
	"fmt"
	"strconv"
	"time"
)

type metricType int8

const (
	metricCounter metricType = iota
	metricAverage
	metricMaximum
	metricMinimum
	metricGauge
	metricHist
)

const (
	connectTimeout = 200 * time.Millisecond
	writeTimeout   = 1 * time.Second
	maxBufSize     = 5 * 1 << 20 // 5 MiB
	valuesChanSize = 500000
)

// Graphite encapsulates an API that allows you to handle metric values and send them to graphite.
// Multiple goroutines may invoke methods on a Graphite simultaneously.
type Graphite struct {
	host          string
	prefix        string
	flushInterval time.Duration

	metrics    map[string]*graphiteMetric
	conn       connection
	ticker     *time.Ticker
	tickerChan <-chan time.Time
	valuesChan chan graphiteValue
	stopChan   chan struct{}

	buffer   bytes.Buffer
	disabled bool
	started  bool
}

// NewGraphite creates a new Graphite with connection to host:port. Application only needs one Graphite to work with all metrics.
// The prefix parameter is assigned to each metric name when it is sent to the graphite server.
// Graphite sends aggregated metrics to the server each flushInterval period. The flushInterval can't be less than a one second.
// Sending metrics to the server is easy to disable from the application config without changing the code. Use the disabled option to do this.
func NewGraphite(host string, port uint16, prefix string, flushInterval time.Duration, disabled bool) (*Graphite, error) {
	if disabled == true {
		graph := new(Graphite)
		graph.disabled = true
		graph.metrics = make(map[string]*graphiteMetric)
		return graph, nil
	}

	if flushInterval < time.Second {
		return nil, fmt.Errorf("NewGraphite: Flush interval (%v) < 1s", flushInterval)
	}
	graph := new(Graphite)
	graph.host = host + ":" + strconv.Itoa(int(port))
	graph.conn = newConnection(graph.host)
	if len(prefix) > 0 {
		graph.prefix = prefix
		if graph.prefix[len(graph.prefix)-1] != '.' {
			graph.prefix = graph.prefix + "."
		}
	}
	graph.flushInterval = flushInterval

	graph.metrics = make(map[string]*graphiteMetric)
	graph.valuesChan = make(chan graphiteValue, valuesChanSize)

	return graph, nil
}

// RegisterCounter creates a new named metric that summarizes  all incoming values.
// This metric has a setting of normalizeByInterval which allows you to send a value (summ / period) to graphite.
func (graphite *Graphite) RegisterCounter(name string, normalizeByInterval bool) error {
	return graphite.registerMetric(name, metricCounter, normalizeByInterval, []float64{})
}

// RegisterAverage creates a new named metric that calculates the average value over the time interval.
func (graphite *Graphite) RegisterAverage(name string) error {
	return graphite.registerMetric(name, metricAverage, false, []float64{})
}

// RegisterMaximum creates a new named metric that calculates the maximum value for the time interval.
func (graphite *Graphite) RegisterMaximum(name string) error {
	return graphite.registerMetric(name, metricMaximum, false, []float64{})
}

// RegisterMinimum creates a new named metric that calculates the minimum value for the time interval.
func (graphite *Graphite) RegisterMinimum(name string) error {
	return graphite.registerMetric(name, metricMinimum, false, []float64{})
}

// RegisterGauge creates a new named metric that use the last value.
func (graphite *Graphite) RegisterGauge(name string) error {
	return graphite.registerMetric(name, metricGauge, false, []float64{})
}

// RegisterHist creates a new named metric that calculates the number of hits of values at predefined intervals.
// For example: histRanges = [10, 25, 100, 350] for intervals: (... , 10), (10, 25), (25, 100), (100, 350), (350, ...)
func (graphite *Graphite) RegisterHist(name string, histRanges []float64) error {
	return graphite.registerMetric(name, metricHist, false, histRanges)
}

// Start creates a goroutine, which sends the aggregated metrics to graphite.
// Start should be called once when the application is initialized as soon as all metrics are registered with functions Register*
func (graphite *Graphite) Start() error {
	if graphite == nil || graphite.metrics == nil {
		return fmt.Errorf("Start: Call NewGraphite() before Start()")
	}

	if graphite.disabled == true {
		return nil
	}

	if graphite.started == true {
		return fmt.Errorf("Graphite already started")
	}

	graphite.stopChan = make(chan struct{})
	graphite.ticker = time.NewTicker(graphite.flushInterval)
	graphite.tickerChan = graphite.ticker.C
	go graphite.handleChans()
	graphite.started = true

	return nil
}

// Stop completes the goroutine of sending metrics to graphite. Typically, in a real application, this is not required, but only for tests.
func (graphite *Graphite) Stop() error {
	if graphite == nil || graphite.metrics == nil {
		return fmt.Errorf("Stop: Call NewGraphite() before Stop()")
	}

	if graphite.disabled == true {
		return nil
	}

	if graphite.started != true {
		return fmt.Errorf("Stop: Call Start() before Stop()")
	}

	close(graphite.stopChan)

	return nil
}

// HandleValue processes the new value for the metric.
func (graphite *Graphite) HandleValue(name string, value float64) error {
	if graphite == nil || graphite.metrics == nil {
		return fmt.Errorf("HandleValue: Call NewGraphite() before HandleValue()")
	}

	if graphite.disabled == true {
		return nil
	}

	if graphite.started != true {
		return fmt.Errorf("HandleValue: Call Start() before HandleValue()")
	}

	if _, ok := graphite.metrics[name]; !ok {
		return fmt.Errorf("HandleValue: Metric %s don't exist", name)
	}

	graphite.valuesChan <- graphiteValue{name, value}
	return nil
}

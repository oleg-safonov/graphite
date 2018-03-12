package graphite

import (
	"testing"
	"time"
)

func ExampleGraphite() {
	// Create a new Graphite
	graph, _ := NewGraphite("localhost", 2003, "prefix.my.service", 1*time.Second, false)

	// Register all metrics before sending them to the server
	graph.RegisterCounter("counter", true)

	// Call Start() for create a goroutine, which sends the aggregated metrics to graphite
	graph.Start()

	// Processes the new value
	graph.HandleValue("counter", 1)
}

func TestNewGraphite(t *testing.T) {
	graph, err := NewGraphite("localhost", 2003, "my.service", 0*time.Second, false)

	if graph != nil || err == nil {
		t.Errorf("Expected error(\"NewGraphite: Flush interval (%v) < 1s\")", 0*time.Second)
	}

	graph, err = NewGraphite("localhost", 0, "", 2*time.Second, false)

	if graph == nil || err != nil {
		t.Errorf("Got Error %v", err)
	}

	if graph.host != "localhost:0" {
		t.Errorf("Expected host \"localhost:0\", got \"%v\"", graph.host)
	}

	if graph.prefix != "" {
		t.Errorf("Expected prefix \"\", got \"%v\"", graph.prefix)
	}

	if graph.flushInterval != 2*time.Second {
		t.Errorf("Expected flushInterval \"2s\", got \"%v\"", graph.flushInterval)
	}

	if graph.metrics == nil {
		t.Error("graph.metrics not initialized")
	}

	if graph.valuesChan == nil {
		t.Error("graph.valuesChan not initialized")
	}

	graph, err = NewGraphite("localhost", 0, "prefix", 2*time.Second, false)
	if graph.prefix != "prefix." {
		t.Errorf("Expected prefix \"prefix.\", got \"%v\"", graph.prefix)
	}
}

func TestStart(t *testing.T) {
	var graph *Graphite = nil
	err := graph.Start()

	if err == nil {
		t.Error("Expected error(\"Start: Call NewGraphite() before Start()\")")
	}

	graph, _ = NewGraphite("localhost", 0, "prefix", 2*time.Second, false)

	err = graph.Start()
	if err != nil {
		t.Errorf("graph.Start() got error (%v)", err)
	}

	if graph.ticker == nil {
		t.Error("graph.ticker not initialized")
	}

	if graph.tickerChan == nil {
		t.Error("graph.tickerChan not initialized")
	}

	err = graph.Start()
	if err == nil {
		t.Error("Expected error(\"Graphite already started\")")
	}

	graph.Stop()
}

func TestStop(t *testing.T) {
	var graph *Graphite = nil
	err := graph.Stop()

	if err == nil {
		t.Error("Expected error(\"Stop: Call NewGraphite() before Stop()\")")
	}

	graph, _ = NewGraphite("localhost", 0, "prefix", 2*time.Second, false)

	err = graph.Stop()
	if err == nil {
		t.Error("Expected error(\"Stop: Call Start() before Stop()\")")
	}

	graph.Start()

	err = graph.Stop()
	if err != nil {
		t.Errorf("graph.Start() got error (%v)", err)
	}
}

func TestHandleValue(t *testing.T) {
	var graph *Graphite = nil
	err := graph.HandleValue("counter", 1)

	if err == nil {
		t.Error("Expected error(\"Start: Call NewGraphite() before HandleValue()\")")
	}

	graph, _ = NewGraphite("localhost", 0, "prefix", 2*time.Second, false)
	err = graph.HandleValue("counter", 1)

	if err == nil {
		t.Error("Expected error(\"Start: Call Start() before HandleValue()\")")
	}

	graph.Start()
	err = graph.HandleValue("counter", 1)

	if err == nil {
		t.Error("HandleValue: Metric counter don't exist")
	}

	graph.RegisterCounter("counter", true)
	err = graph.HandleValue("counter", 1)

	if err != nil {
		t.Errorf("graph.HandleValue() got Error %v", err)
	}

	graph.Stop()
}

func TestDisabledGraphite(t *testing.T) {
	graph, err := NewGraphite("", 0, "", 0, true)

	if graph == nil || err != nil {
		t.Errorf("Got error for disabled graphite")
	}

	// Only disabled graphite has the ability to call HandleValue() before RegisterCounter() and Start()
	err = graph.HandleValue("counter", 1)

	if err != nil {
		t.Errorf("Got error(%v) for disabled graphite", err)
	}

	err = graph.RegisterCounter("counter", true)
	if err != nil {
		t.Errorf("Got error(%v) for disabled graphite", err)
	}

	err = graph.Start()
	if err != nil {
		t.Errorf("Got error(%v) for disabled graphite", err)
	}

	if graph.ticker != nil {
		t.Error("graph.ticker != nil for disabled graphite")
	}

	err = graph.HandleValue("counter", 1)

	if err != nil {
		t.Errorf("Got error(%v) for disabled graphite", err)
	}

	err = graph.Stop()
	if err != nil {
		t.Errorf("Got error(%v) for disabled graphite", err)
	}
}

func TestRegisterCounter(t *testing.T) {
	graph, _ := NewGraphite("localhost", 0, "prefix", 2*time.Second, false)
	err := graph.RegisterCounter("counter1", false)

	if err != nil {
		t.Errorf("graph.RegisterCounter() got error(%v)", err)
	}

	metric := graph.metrics["counter1"]
	if metric.mType != metricCounter {
		t.Error("Expected metricCounter, got ", metric.mType)
	}

	err = graph.RegisterCounter("counter2", true)
	if err != nil {
		t.Errorf("graph.RegisterCounter() got error(%v)", err)
	}

	metric = graph.metrics["counter2"]
	if metric.normalizeByInterval != true {
		t.Error("normalizeByInterval expected true, got ", metric.normalizeByInterval)
	}

	err = graph.RegisterAverage("counter2")
	if err == nil {
		t.Errorf("Expected error(\"RegisterMetric: Metric counter2 already exist\")")
	}
}

func TestRegisterAverage(t *testing.T) {
	graph, _ := NewGraphite("localhost", 0, "prefix", 2*time.Second, false)
	err := graph.RegisterAverage("average")

	if err != nil {
		t.Errorf("graph.RegisterAverage() got error(%v)", err)
	}

	metric := graph.metrics["average"]
	if metric.mType != metricAverage {
		t.Error("Expected metricAverage, got ", metric.mType)
	}

	err = graph.RegisterAverage("average")
	if err == nil {
		t.Errorf("Expected error(\"RegisterMetric: Metric average already exist\")")
	}
}

func TestRegisterMaximum(t *testing.T) {
	graph, _ := NewGraphite("localhost", 0, "prefix", 2*time.Second, false)
	err := graph.RegisterMaximum("maximum")

	if err != nil {
		t.Errorf("graph.RegisterMaximum() got error(%v)", err)
	}

	metric := graph.metrics["maximum"]
	if metric.mType != metricMaximum {
		t.Error("Expected metricMaximum, got ", metric.mType)
	}

	err = graph.RegisterMaximum("maximum")
	if err == nil {
		t.Errorf("Expected error(\"RegisterMetric: Metric maximum already exist\")")
	}
}

func TestRegisterHist(t *testing.T) {
	graph, _ := NewGraphite("localhost", 0, "prefix", 2*time.Second, false)
	err := graph.RegisterHist("hist", []float64{1, 2, 3})

	if err != nil {
		t.Errorf("graph.RegisterHist() got error(%v)", err)
	}

	metric := graph.metrics["hist"]
	if metric.mType != metricHist {
		t.Error("Expected metricHist, got ", metric.mType)
	}

	err = graph.RegisterMaximum("hist")
	if err == nil {
		t.Errorf("Expected error(\"RegisterMetric: Metric hist already exist\")")
	}
}

func BenchmarkHandleValue(b *testing.B) {
	graph, _ := NewGraphite("localhost", 0, "prefix", 20*time.Second, false)
	graph.RegisterCounter("counter", true)
	graph.Start()
	for i := 0; i < b.N; i++ {
		graph.HandleValue("counter", 1)
	}

	graph.Stop()
}

package graphite

import (
	"bytes"
	"log"
	"os"
	"reflect"
	"sort"
	"strings"
	"testing"
	"time"
)

var expected_text string = `prefix.hist.0 0 946782245
prefix.hist.1 1 946782245
prefix.hist.2 0 946782245
prefix.hist.3 0 946782245
prefix.minimum 1.000000000000 946782245
prefix.gauge 8.000000000000 946782245
`

func TestRegisterMetric(t *testing.T) {
	var graph *Graphite = nil
	err := graph.registerMetric("average", metricAverage, true, []float64{1, 2, 3})

	if err == nil {
		t.Error("Expected error(\"RegisterMetric: Call NewGraphite() before RegisterMetric()\")")
	}

	graph, _ = NewGraphite("localhost", 0, "prefix", 2*time.Second, false)
	err = graph.registerMetric("average", metricAverage, true, []float64{1, 2, 3})

	if err != nil {
		t.Errorf("registerMetric() got error(%v)", err)
	}

	metric := graph.metrics["average"]
	if metric.mType != metricAverage {
		t.Error("Expected metricAverage, got ", metric.mType)
	}

	if metric.normalizeByInterval != true {
		t.Error("Expected true, got ", metric.normalizeByInterval)
	}

	if metric.flushInterval != 2*time.Second {
		t.Error("Expected 2s, got ", metric.flushInterval)
	}

	equal := reflect.DeepEqual(metric.histRanges, []float64{1, 2, 3})
	if equal != true {
		t.Errorf("Expected [1 2 3], got %v", metric.histRanges)
	}

	equal = reflect.DeepEqual(metric.hist, []int32{0, 0, 0, 0})
	if equal != true {
		t.Errorf("Expected [0 0 0 0], got %v", metric.hist)
	}
}

func TestFillBuffer(t *testing.T) {
	graph, err := NewGraphite("", 0, "prefix", 2*time.Second, false)

	err = graph.RegisterMinimum("minimum")
	if err != nil {
		t.Errorf("RegisterMinimum() got error(%v)", err)
	}
	graph.metrics["minimum"].handleValue(1)

	err = graph.RegisterGauge("gauge")
	if err != nil {
		t.Errorf("RegisterGauge() got error(%v)", err)
	}
	graph.metrics["gauge"].handleValue(8)

	err = graph.RegisterHist("hist", []float64{10, 20, 30})
	if err != nil {
		t.Errorf("RegisterHist() got error(%v)", err)
	}
	graph.metrics["hist"].handleValue(12)

	graph.fillBuffer(time.Date(2000, 1, 2, 3, 4, 5, 0, time.UTC))

	expected_lines := strings.Split(expected_text, "\n")
	sort.Strings(expected_lines)
	expected_text = strings.Join(expected_lines, "\n")

	output := graph.buffer.String()
	output_lines := strings.Split(output, "\n")
	sort.Strings(output_lines)
	output = strings.Join(output_lines, "\n")

	if expected_text != output {
		t.Errorf("Expected \"%v\", got \"%v\"", expected_text, output)
	}
}

func TestBufferExceed(t *testing.T) {
	var logOutput bytes.Buffer
	log.SetOutput(&logOutput)
	defer log.SetOutput(os.Stderr)
	graph, _ := NewGraphite("", 0, "prefix", 1*time.Second, false)
	graph.RegisterHist("hist", []float64{10, 20, 30})
	tm := time.Date(2000, 1, 2, 3, 4, 5, 0, time.UTC)
	for i := 0; i < 2; i++ {
		graph.metrics["hist"].handleValue(12)
		graph.fillBuffer(tm)
	}

	l := len(graph.buffer.String())
	if l != 208 {
		t.Errorf("Expected 208, got \"%v\"", l)
	}

	for i := 0; i < 50412; i++ {
		graph.metrics["hist"].handleValue(12)
		graph.fillBuffer(tm)
	}
	l = len(graph.buffer.String())
	if l != 104 {
		t.Errorf("Expected 104, got \"%v\"", l)
	}
	graph.Start()
	time.Sleep(1500 * time.Millisecond)
	graph.Stop()
}

type testConnection struct {
	Buffer bytes.Buffer
}

func (c *testConnection) Close() error {
	return nil
}

func (c *testConnection) Write(p []byte) (int, error) {
	return c.Buffer.Write(p)
}

func (c *testConnection) connect() error {
	return nil
}

func TestSendMetrics(t *testing.T) {
	graph, _ := NewGraphite("", 0, "prefix", 1*time.Second, false)
	graph.RegisterHist("hist", []float64{10, 20, 30})
	c := new(testConnection)
	graph.conn = c
	graph.Start()
	graph.HandleValue("hist", 12)
	time.Sleep(1100 * time.Millisecond)

	output := c.Buffer.String()

	if len(output) < 100 {
		t.Errorf("Sent text \"%v\"", output)
	}
	graph.Stop()
}

func BenchmarkFillBuffer(b *testing.B) {
	graph, _ := NewGraphite("", 0, "prefix", 20*time.Second, false)

	graph.RegisterMinimum("minimum")
	graph.RegisterGauge("gauge")
	graph.RegisterHist("hist", []float64{10, 20, 30})
	graph.Start()

	for i := 0; i < b.N; i++ {
		graph.metrics["minimum"].handleValue(1)
		graph.metrics["gauge"].handleValue(8)
		graph.metrics["hist"].handleValue(12)
		graph.fillBuffer(time.Now())
		graph.buffer.Reset()
	}

	graph.Stop()
}

package graphite

import (
	"reflect"
	"testing"
	"time"
)

// Test counter

func ValuesTest(values []float64, mType metricType, normalizeByInterval bool, flushInterval time.Duration) (float64, int32) {
	gm := graphiteMetric{}
	gm.mType = mType
	gm.normalizeByInterval = normalizeByInterval
	gm.flushInterval = flushInterval

	for _, v := range values {
		gm.handleValue(v)
	}
	return gm.get()
}

func CounterValuesTest(values []float64) (float64, int32) {
	return ValuesTest(values, metricCounter, false, 0)
}

func CounterValuesNormalizedTest(values []float64, flushInterval time.Duration) (float64, int32) {
	return ValuesTest(values, metricCounter, true, flushInterval)
}

func TestCounter0Values(t *testing.T) {
	v, c := CounterValuesTest([]float64{})

	if v != 0 || c != 0 {
		t.Error("Expected 0, got ", v)
	}
}

func TestCounter1Values(t *testing.T) {
	v, c := CounterValuesTest([]float64{42})

	if v != 42 || c != 1 {
		t.Error("Expected 42, got ", v)
	}
}

func TestCounter2Values(t *testing.T) {
	v, c := CounterValuesTest([]float64{1, 2})

	if v != 3 || c != 2 {
		t.Error("Expected 3, got ", v)
	}
}

func TestCounter5Values(t *testing.T) {
	v, c := CounterValuesTest([]float64{1, 2, 3, 4, 5})

	if v != 15 || c != 5 {
		t.Error("Expected 15.0, got ", v)
	}
}

func TestCounterNormalized0Values(t *testing.T) {
	v, c := CounterValuesNormalizedTest([]float64{}, 3*time.Second)

	if v != 0 || c != 0 {
		t.Error("Expected 0, got ", v)
	}
}

func TestCounterNormalized1Values(t *testing.T) {
	v, c := CounterValuesNormalizedTest([]float64{30}, 3*time.Second)

	if v != 10 || c != 1 {
		t.Error("Expected 10, got ", v)
	}
}

func TestCounterNormalized5Values(t *testing.T) {
	v, c := CounterValuesNormalizedTest([]float64{1, 2, 3, 4, 5}, 3*time.Second)

	if v != 5 || c != 5 {
		t.Error("Expected 5, got ", v)
	}
}

// Test Average

func AverageValuesTest(values []float64) (float64, int32) {
	return ValuesTest(values, metricAverage, false, 0)
}

func TestAverage0Values(t *testing.T) {
	v, c := AverageValuesTest([]float64{})

	if v != 0 || c != 0 {
		t.Error("Expected 0, got ", v)
	}
}

func TestAverage1Values(t *testing.T) {
	v, c := AverageValuesTest([]float64{42})

	if v != 42 || c != 1 {
		t.Error("Expected 42, got ", v)
	}
}

func TestAverage2Values(t *testing.T) {
	v, c := AverageValuesTest([]float64{1, 2})

	if v != 1.5 || c != 2 {
		t.Error("Expected 1.5, got ", v)
	}
}

func TestAverage5Values(t *testing.T) {
	v, c := AverageValuesTest([]float64{1, 2, 3, 4, 5})

	if v != 3.0 || c != 5 {
		t.Error("Expected 3.0, got ", v)
	}
}

// Test Maximum

func MaximumValuesTest(values []float64) (float64, int32) {
	return ValuesTest(values, metricMaximum, false, 0)
}

func TestMaximum0Values(t *testing.T) {
	v, c := MaximumValuesTest([]float64{})

	if v != 0 || c != 0 {
		t.Error("Expected 0, got ", v)
	}
}

func TestMaximum1Values(t *testing.T) {
	v, c := MaximumValuesTest([]float64{0.0003})

	if v != 0.0003 || c != 1 {
		t.Error("Expected 0.0003, got ", v)
	}
}

func TestMaximum3Values(t *testing.T) {
	v, c := MaximumValuesTest([]float64{0.0003, 12, 11.99})

	if v != 12 || c != 3 {
		t.Error("Expected 12, got ", v)
	}
}

// Test Minimum

func MinimumValuesTest(values []float64) (float64, int32) {
	return ValuesTest(values, metricMinimum, false, 0)
}

func TestMinimum0Values(t *testing.T) {
	v, c := MinimumValuesTest([]float64{})

	if v != 0 || c != 0 {
		t.Error("Expected 0, got ", v)
	}
}

func TestMinimum1Values(t *testing.T) {
	v, c := MinimumValuesTest([]float64{1000000})

	if v != 1000000 || c != 1 {
		t.Error("Expected 1000000, got ", v)
	}
}

func TestMinimum3Values(t *testing.T) {
	v, c := MinimumValuesTest([]float64{1000000, 11.99, 12})

	if v != 11.99 || c != 3 {
		t.Error("Expected 11.99, got ", v)
	}
}

// Test Gauge

func GaugeValuesTest(values []float64) (float64, int32) {
	return ValuesTest(values, metricGauge, false, 0)
}

func TestGauge0Values(t *testing.T) {
	v, c := GaugeValuesTest([]float64{})

	if v != 0 || c != 0 {
		t.Error("Expected 0, got ", v)
	}
}

func TestGauge1Values(t *testing.T) {
	v, c := GaugeValuesTest([]float64{42})

	if v != 42 || c != 1 {
		t.Error("Expected 42, got ", v)
	}
}

func TestGauge3Values(t *testing.T) {
	v, c := GaugeValuesTest([]float64{1, 1000000, 42})

	if v != 42 || c != 3 {
		t.Error("Expected 42, got ", v)
	}
}

// Test Hist

func HistTest(values []float64, histRanges []float64) ([]int32, int32) {
	gm := graphiteMetric{}
	gm.mType = metricHist
	gm.histRanges = histRanges
	gm.hist = make([]int32, len(gm.histRanges)+1)

	for _, v := range values {
		gm.handleValue(v)
	}
	return gm.getHist()
}

func TestHist0Values(t *testing.T) {
	model := []int32{0, 0, 0, 0, 0}

	v, c := HistTest([]float64{}, []float64{5, 10, 15, 20})

	equal := reflect.DeepEqual(model, v)
	if equal != true || c != 0 {
		t.Errorf("Expected %v, got %v", model, v)
	}
}

func TestHist15Values(t *testing.T) {
	model := []int32{3, 4, 3, 4, 1}

	list := []float64{7, 10, -8, 6, 11, 0, 1, 5, 1000000, 17, 17, 17, 19, 6, 10}
	v, c := HistTest(list, []float64{5, 10, 15, 20})

	equal := reflect.DeepEqual(model, v)
	if equal != true || c != 15 {
		t.Errorf("Expected %v, got %v", model, v)
	}
}

func TestReset(t *testing.T) {
	gm := graphiteMetric{
		metricCounter,
		42,
		3,
		true,
		3 * time.Second,
		[]float64{5, 10, 15, 20},
		[]int32{3, 4, 3, 4, 1}}
	gm.reset()

	if gm.mType != metricCounter {
		t.Errorf("Expected metricCounter, got %v", gm.mType)
	}

	if gm.value != 0 {
		t.Errorf("Expected 0, got %v", gm.value)
	}

	if gm.counter != 0 {
		t.Errorf("Expected 0, got %v", gm.counter)
	}

	if gm.normalizeByInterval != true {
		t.Errorf("Expected true, got %v", gm.normalizeByInterval)
	}

	if gm.flushInterval != 3*time.Second {
		t.Errorf("Expected 3s, got %v", gm.flushInterval)
	}

	equal := reflect.DeepEqual(gm.histRanges, []float64{5, 10, 15, 20})
	if equal != true {
		t.Errorf("Expected [5, 10, 15, 20], got %v", gm.histRanges)
	}

	equal = reflect.DeepEqual(gm.hist, []int32{0, 0, 0, 0, 0})
	if equal != true {
		t.Errorf("Expected [0, 0, 0, 0, 0], got %v", gm.hist)
	}
}

func BenchmarkHandleCounter(b *testing.B) {
	gm := graphiteMetric{}
	gm.mType = metricCounter

	for i := 0; i < b.N; i++ {
		gm.handleValue(float64(i))
	}
}

func BenchmarkHandleAverage(b *testing.B) {
	gm := graphiteMetric{}
	gm.mType = metricAverage

	for i := 0; i < b.N; i++ {
		gm.handleValue(float64(i))
	}
}

func BenchmarkHandleHist(b *testing.B) {
	gm := graphiteMetric{}
	gm.mType = metricHist
	gm.histRanges = []float64{500, 1000000, 10000000, 20000000}
	gm.hist = make([]int32, len(gm.histRanges)+1)

	for i := 0; i < b.N; i++ {
		gm.handleValue(float64(i))
	}
}

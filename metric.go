package graphite

import "time"

type graphiteMetric struct {
	mType               metricType
	value               float64
	counter             int32
	normalizeByInterval bool
	flushInterval       time.Duration
	histRanges          []float64
	hist                []int32
}

func (mt *graphiteMetric) handleValue(value float64) {
	switch mt.mType {
	case metricCounter:
		mt.value += value
	case metricAverage:
		mt.value = mt.value*(float64(mt.counter)/float64(mt.counter+1)) + value/float64(mt.counter+1)
	case metricMaximum:
		if mt.counter == 0 || mt.value < value {
			mt.value = value
		}
	case metricMinimum:
		if mt.counter == 0 || mt.value > value {
			mt.value = value
		}
	case metricGauge:
		mt.value = value
	case metricHist:
		var isHit = false
		for i, v := range mt.histRanges {
			if value < v {
				mt.hist[i] += 1
				isHit = true
				break
			}
		}

		if isHit == false {
			mt.hist[len(mt.hist)-1] += 1
		}
	}

	mt.counter += 1
}

func (mt *graphiteMetric) get() (float64, int32) {
	if mt.normalizeByInterval == true {
		return mt.value / (float64(mt.flushInterval) / float64(time.Second)), mt.counter
	}

	return mt.value, mt.counter
}

func (mt *graphiteMetric) getHist() ([]int32, int32) {
	return mt.hist, mt.counter
}

func (mt *graphiteMetric) reset() {
	mt.value = 0
	mt.counter = 0
	for i := range mt.hist {
		mt.hist[i] = 0
	}
}

package graphite

import (
	"fmt"
	"log"
	"net"
	"strconv"
	"time"
)

type graphiteValue struct {
	name  string
	value float64
}

type connection interface {
	Close() error
	Write(p []byte) (int, error)
	connect() error
}

type tcpConnection struct {
	host string
	conn net.Conn
}

func newConnection(host string) *tcpConnection {
	c := new(tcpConnection)
	c.host = host

	return c
}

func (c *tcpConnection) Close() (err error) {
	if c.conn != nil {
		err = c.conn.Close()
		c.conn = nil
	}
	return
}

func (c *tcpConnection) Write(p []byte) (int, error) {
	if c.conn == nil {
		err := c.connect()
		if err != nil {
			log.Printf("Graphite.connect: %v", err)
			return 0, err
		}
	}

	c.conn.SetWriteDeadline(time.Now().Add(writeTimeout))
	return c.conn.Write(p)
}

func (c *tcpConnection) connect() (err error) {
	c.conn, err = net.DialTimeout("tcp", c.host, connectTimeout)
	if err != nil {
		c.conn = nil
	}
	return
}

func (gr *Graphite) registerMetric(name string, mType metricType, normalizeByInterval bool, histRanges []float64) error {
	if gr == nil || gr.metrics == nil {
		return fmt.Errorf("RegisterMetric: Call NewGraphite() before RegisterMetric()")
	}

	if gr.disabled == true {
		return nil
	}

	if _, ok := gr.metrics[name]; ok {
		return fmt.Errorf("RegisterMetric: Metric %s already exist", name)
	}
	var v graphiteMetric
	v.mType = mType
	v.normalizeByInterval = normalizeByInterval
	v.flushInterval = gr.flushInterval
	v.histRanges = histRanges
	v.hist = make([]int32, len(v.histRanges)+1)

	gr.metrics[name] = &v
	return nil
}

func (gr *Graphite) fillBuffer(currentTime time.Time) {
	current_time := strconv.Itoa(int(currentTime.Unix()))

	if gr.buffer.Len() > maxBufSize {
		log.Printf("Graphite.sendMetrics: buffer size > %d. Reset buffer.", maxBufSize)
		gr.buffer.Reset()
	}

	for name, value := range gr.metrics {
		if gr.metrics[name].mType != metricHist {
			v, c := value.get()

			if c > 0 {
				value.reset()
				gr.buffer.WriteString(gr.prefix)
				gr.buffer.WriteString(name)
				gr.buffer.WriteString(" ")
				gr.buffer.WriteString(strconv.FormatFloat(v, 'f', 12, 64))
				gr.buffer.WriteString(" ")
				gr.buffer.WriteString(current_time)
				gr.buffer.WriteString("\n")
			}
		} else {
			hist, c := value.getHist()
			if c > 0 {
				for i, v := range hist {
					gr.buffer.WriteString(gr.prefix)
					gr.buffer.WriteString(name)
					gr.buffer.WriteString(".")
					gr.buffer.WriteString(strconv.Itoa(i))
					gr.buffer.WriteString(" ")
					gr.buffer.WriteString(strconv.Itoa(int(v)))
					gr.buffer.WriteString(" ")
					gr.buffer.WriteString(current_time)
					gr.buffer.WriteString("\n")
				}
				value.reset()
			}
		}
	}
}

func (gr *Graphite) sendMetrics(currentTime time.Time) {
	gr.fillBuffer(currentTime)

	if gr.buffer.Len() > 0 {
		_, err := gr.conn.Write(gr.buffer.Bytes())
		if err != nil {
			gr.conn.Close()
			return
		}
		gr.buffer.Reset()
	}
}

func (gr *Graphite) handleChans() {
	for {
		select {
		case t := <-gr.tickerChan:
			gr.sendMetrics(t)

		case v := <-gr.valuesChan:
			gr.metrics[v.name].handleValue(v.value)

		case _, _ = <-gr.stopChan:
			return
		}
	}
}

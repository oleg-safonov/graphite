# Graphite golang
[![Build Status](https://travis-ci.org/oleg-safonov/graphite.svg?branch=master)](https://travis-ci.org/oleg-safonov/graphite)[![Coverage Status](https://coveralls.io/repos/github/oleg-safonov/graphite/badge.svg?branch=master)](https://coveralls.io/github/oleg-safonov/graphite?branch=master)[![GoDoc](https://godoc.org/github.com/oleg-safonov/graphite?status.svg)](https://godoc.org/github.com/oleg-safonov/graphite)

The graphite package was written for cases when an application is running on thousands of instances and each of the instances generates hundreds of thousands of events per second. The graphite package aggregates metrics over a time interval and sends application metrics to the graphite. 

## Features

 - thread-safe
   - Multiple goroutines may update metrics simultaneously 
   - Sending metrics to graphite doesn't block the execution of the main program
 - Multiple metric types
   - Counter
   - Average
   - Maximum
   - Minimum
   - Gauge
   - Histogram
 - Minimal resource consumption
   - Sending to graphite only aggregated data for the period
   - There is no need for statsd

## Usage
Create and update metrics:
```
	// Create a new Graphite
	graph, _ := NewGraphite("localhost", 2003, "prefix.my.service", 1*time.Second, false)

	// Register all metrics before sending them to the server
	graph.RegisterCounter("counter", true)

	// Call Start() for create a goroutine, which sends the aggregated metrics to graphite
	graph.Start()

	// Processes the new value
	graph.HandleValue("counter", 1)
```
## Metric types
### Counter
A **counter** metric summarizes all incoming values. This metric has a setting of *normalizeByInterval* which allows you to send a value *(summ / period)* to graphite.
### Average
The **average** metric calculates the average value over the time interval.
### Maximum
The **maximum** type metric calculates the maximum value for the time interval.
### Minimum
The **minimum** type metric calculates the minimum value for the time interval.
### Gauge
A **gauge** metric uses the last value.
### Histogram
The **histogram** metric calculates the number of hits of values at predefined intervals.

In this package, a percentiles are not used, instead of them it is suggested to use the **histogram** metric. For example, we need to know how fast our application generates a response. For this, let's compare two approaches: using the percentiles and the **histogram** metric.  
![percentiles approach](/images/hist2.png)
![histogram approach](/images/hist1.png)  
The first picture shows the percentile approach. It shows that during the problems on the network most of the answers of the service was unforgivably large. But if we think sensibly, we do not need to know the peak response time, we are always interested in what percentage of customers can get an answer for a certain time. The approach shown in the second picture is better for this. It shows not only the main problem, but it also shows that the response time after network recovery is still worse than before the accident.

For example:
```
	err := graph.RegisterHist("hist", []float64{10, 30, 50, 70, 80, 100})
```
Do not forget to correctly configure the graph in the Grafana:
![grafana settings](/images/grafana_settings.png)
![grafana settings](/images/grafana_settings2.png)

# External dependencies
This project has no external dependencies other than the Go standard library.
# Installation
```
go get github.com/oleg-safonov/graphite
```

// Copyright 2018 The Prometheus Authors
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
// http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

// Package promauto provides constructors for the usual Prometheus metrics that
// return them already registered with the global registry
// (metricslog.DefaultRegisterer). This allows very compact code, avoiding any
// references to the registry altogether, but all the constructors in this
// package will panic if the registration fails.
//
// The following example is a complete program to create a histogram of normally
// distributed random numbers from the math/rand package:
//
//      package main
//
//      import (
//              "math/rand"
//              "net/http"
//
//              "github.com/Schneizelw/metricslog/client_golang/metricslog"
//              "github.com/Schneizelw/metricslog/client_golang/metricslog/promauto"
//              "github.com/Schneizelw/metricslog/client_golang/metricslog/promhttp"
//      )
//
//      var histogram = promauto.NewHistogram(metricslog.HistogramOpts{
//              Name:    "random_numbers",
//              Help:    "A histogram of normally distributed random numbers.",
//              Buckets: metricslog.LinearBuckets(-3, .1, 61),
//      })
//
//      func Random() {
//              for {
//                      histogram.Observe(rand.NormFloat64())
//              }
//      }
//
//      func main() {
//              go Random()
//              http.Handle("/metrics", promhttp.Handler())
//              http.ListenAndServe(":1971", nil)
//      }
//
// Prometheus's version of a minimal hello-world program:
//
//      package main
//
//      import (
//          "fmt"
//          "net/http"
//
//          "github.com/Schneizelw/metricslog/client_golang/metricslog"
//          "github.com/Schneizelw/metricslog/client_golang/metricslog/promauto"
//          "github.com/Schneizelw/metricslog/client_golang/metricslog/promhttp"
//      )
//
//      func main() {
//          http.Handle("/", promhttp.InstrumentHandlerCounter(
//              promauto.NewCounterVec(
//                  metricslog.CounterOpts{
//                      Name: "hello_requests_total",
//                      Help: "Total number of hello-world requests by HTTP code.",
//                  },
//                  []string{"code"},
//              ),
//              http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
//                  fmt.Fprint(w, "Hello, world!")
//              }),
//          ))
//          http.Handle("/metrics", promhttp.Handler())
//          http.ListenAndServe(":1971", nil)
//      }
//
// This appears very handy. So why are these constructors locked away in a
// separate package? There are two caveats:
//
// First, in more complex programs, global state is often quite problematic.
// That's the reason why the metrics constructors in the metricslog package do
// not interact with the global metricslog.DefaultRegisterer on their own. You
// are free to use the Register or MustRegister functions to register them with
// the global metricslog.DefaultRegisterer, but you could as well choose a local
// Registerer (usually created with metricslog.NewRegistry, but there are other
// scenarios, e.g. testing).
//
// The second issue is that registration may fail, e.g. if a metric inconsistent
// with the newly to be registered one is already registered. But how to signal
// and handle a panic in the automatic registration with the default registry?
// The only way is panicking. While panicking on invalid input provided by the
// programmer is certainly fine, things are a bit more subtle in this case: You
// might just add another package to the program, and that package (in its init
// function) happens to register a metric with the same name as your code. Now,
// all of a sudden, either your code or the code of the newly imported package
// panics, depending on initialization order, without any opportunity to handle
// the case gracefully. Even worse is a scenario where registration happens
// later during the runtime (e.g. upon loading some kind of plugin), where the
// panic could be triggered long after the code has been deployed to
// production. A possibility to panic should be explicitly called out by the
// Must… idiom, cf. metricslog.MustRegister. But adding a separate set of
// constructors in the metricslog package called MustRegisterNewCounterVec or
// similar would be quite unwieldy. Adding an extra MustRegister method to each
// metric, returning the registered metric, would result in nice code for those
// using the method, but would pollute every single metric interface for
// everybody avoiding the global registry.
//
// To address both issues, the problematic auto-registering and possibly
// panicking constructors are all in this package with a clear warning
// ahead. And whoever cares about avoiding global state and possibly panicking
// function calls can simply ignore the existence of the promauto package
// altogether.
//
// A final note: There is a similar case in the net/http package of the standard
// library. It has DefaultServeMux as a global instance of ServeMux, and the
// Handle function acts on it, panicking if a handler for the same pattern has
// already been registered. However, one might argue that the whole HTTP routing
// is usually set up closely together in the same package or file, while
// Prometheus metrics tend to be spread widely over the codebase, increasing the
// chance of surprising registration failures. Furthermore, the use of global
// state in net/http has been criticized widely, and some avoid it altogether.
package promauto

import "github.com/Schneizelw/metricslog/client_golang/metricslog"

// NewCounter works like the function of the same name in the metricslog package
// but it automatically registers the Counter with the
// metricslog.DefaultRegisterer. If the registration fails, NewCounter panics.
func NewCounter(opts metricslog.CounterOpts) metricslog.Counter {
    c := metricslog.NewCounter(opts)
    metricslog.MustRegister(c)
    return c
}

// NewCounterVec works like the function of the same name in the metricslog
// package but it automatically registers the CounterVec with the
// metricslog.DefaultRegisterer. If the registration fails, NewCounterVec
// panics.
func NewCounterVec(opts metricslog.CounterOpts, labelNames []string) *metricslog.CounterVec {
    c := metricslog.NewCounterVec(opts, labelNames)
    metricslog.MustRegister(c)
    return c
}

// NewCounterFunc works like the function of the same name in the metricslog
// package but it automatically registers the CounterFunc with the
// metricslog.DefaultRegisterer. If the registration fails, NewCounterFunc
// panics.
func NewCounterFunc(opts metricslog.CounterOpts, function func() float64) metricslog.CounterFunc {
    g := metricslog.NewCounterFunc(opts, function)
    metricslog.MustRegister(g)
    return g
}

// NewGauge works like the function of the same name in the metricslog package
// but it automatically registers the Gauge with the
// metricslog.DefaultRegisterer. If the registration fails, NewGauge panics.
func NewGauge(opts metricslog.GaugeOpts) metricslog.Gauge {
    g := metricslog.NewGauge(opts)
    metricslog.MustRegister(g)
    return g
}

// NewGaugeVec works like the function of the same name in the metricslog
// package but it automatically registers the GaugeVec with the
// metricslog.DefaultRegisterer. If the registration fails, NewGaugeVec panics.
func NewGaugeVec(opts metricslog.GaugeOpts, labelNames []string) *metricslog.GaugeVec {
    g := metricslog.NewGaugeVec(opts, labelNames)
    metricslog.MustRegister(g)
    return g
}

// NewGaugeFunc works like the function of the same name in the metricslog
// package but it automatically registers the GaugeFunc with the
// metricslog.DefaultRegisterer. If the registration fails, NewGaugeFunc panics.
func NewGaugeFunc(opts metricslog.GaugeOpts, function func() float64) metricslog.GaugeFunc {
    g := metricslog.NewGaugeFunc(opts, function)
    metricslog.MustRegister(g)
    return g
}

// NewSummary works like the function of the same name in the metricslog package
// but it automatically registers the Summary with the
// metricslog.DefaultRegisterer. If the registration fails, NewSummary panics.
func NewSummary(opts metricslog.SummaryOpts) metricslog.Summary {
    s := metricslog.NewSummary(opts)
    metricslog.MustRegister(s)
    return s
}

// NewSummaryVec works like the function of the same name in the metricslog
// package but it automatically registers the SummaryVec with the
// metricslog.DefaultRegisterer. If the registration fails, NewSummaryVec
// panics.
func NewSummaryVec(opts metricslog.SummaryOpts, labelNames []string) *metricslog.SummaryVec {
    s := metricslog.NewSummaryVec(opts, labelNames)
    metricslog.MustRegister(s)
    return s
}

// NewHistogram works like the function of the same name in the metricslog
// package but it automatically registers the Histogram with the
// metricslog.DefaultRegisterer. If the registration fails, NewHistogram panics.
func NewHistogram(opts metricslog.HistogramOpts) metricslog.Histogram {
    h := metricslog.NewHistogram(opts)
    metricslog.MustRegister(h)
    return h
}

// NewHistogramVec works like the function of the same name in the metricslog
// package but it automatically registers the HistogramVec with the
// metricslog.DefaultRegisterer. If the registration fails, NewHistogramVec
// panics.
func NewHistogramVec(opts metricslog.HistogramOpts, labelNames []string) *metricslog.HistogramVec {
    h := metricslog.NewHistogramVec(opts, labelNames)
    metricslog.MustRegister(h)
    return h
}

package domain

import (
	"io/ioutil"
	"net/http"
	"sync"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/sirupsen/logrus"
)

const prometheusConfigName = "prometheus"
const helpName = "Detaled health checks for program"

type entityState struct {
	name   string
	gauge  prometheus.Gauge
	isUp   chan struct{}
	isDown chan struct{}
	state  *int
	mux    *sync.Mutex
	rmux   *sync.RWMutex
}

// PrometheusConfig ...
type PrometheusConfig struct {
	ip, port, mainUUID string
	metrics            map[string]entityState
	globalStateChange  chan struct{}
	logging            *logrus.Logger
}

// NewPrometheusConfig ...
func NewPrometheusConfig(ip, port, mainUUID string, logging *logrus.Logger) PrometheusConfig {
	return PrometheusConfig{
		ip:                ip,
		port:              port,
		mainUUID:          mainUUID,
		metrics:           make(map[string]entityState),
		globalStateChange: make(chan struct{}, 1),
		logging:           logging,
	}
}

// NewHealthMetric ...
func (prometheusConfig *PrometheusConfig) NewHealthMetric(name, label, uuid string) (chan struct{}, chan struct{}) {
	labels := make(map[string]string)
	labels["module"] = label
	newMetric := prometheus.NewGauge(
		prometheus.GaugeOpts{
			Name:        name,
			Help:        helpName,
			ConstLabels: labels,
		})

	prometheus.MustRegister(newMetric)

	mux := &sync.Mutex{}
	rmux := &sync.RWMutex{}
	isUp := make(chan struct{}, 1)
	isDown := make(chan struct{}, 1)
	state := 0
	entityState := entityState{
		name:   label,
		gauge:  newMetric,
		isUp:   isUp,
		isDown: isDown,
		state:  &state,
		mux:    mux,
		rmux:   rmux,
	}
	prometheusConfig.metrics[label] = entityState
	prometheusConfig.logging.WithFields(logrus.Fields{
		"entity":     prometheusConfigName,
		"event uuid": uuid,
	}).Tracef("Add new health metric: %v", label)
	return isUp, isDown
}

// UpMonitoring ...
func (prometheusConfig *PrometheusConfig) UpMonitoring(internalUUID string) {
	fullAddr := prometheusConfig.ip + ":" + prometheusConfig.port
	http.Handle("/metrics", promhttp.Handler())
	prometheusConfig.logging.WithFields(logrus.Fields{
		"entity":     prometheusConfigName,
		"event uuid": internalUUID,
	}).Infof("Starting web server at %s\n", fullAddr)

	err := http.ListenAndServe(fullAddr, nil)
	if err != nil {
		prometheusConfig.logging.WithFields(logrus.Fields{
			"entity":     prometheusConfigName,
			"event uuid": internalUUID,
		}).Errorf("prometheus http.ListenAndServer error: %v", err)
		return
	}
}

// CheckMonitoringIsUp ...
func (prometheusConfig *PrometheusConfig) CheckMonitoringIsUp() error {
	addresForGet := "http://" + prometheusConfig.ip + ":" + prometheusConfig.port + "/metrics"
	resp, err := http.Get(addresForGet)
	if err != nil {
		return err
	}
	_, err = ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}
	return nil
}

// ComplexHealthCheck ...
func (prometheusConfig *PrometheusConfig) ComplexHealthCheck() {
	for metricName, eState := range prometheusConfig.metrics {
		go func(eState entityState, metricName string) {
			for {
				select {
				case <-eState.isDown:
					prometheusConfig.logging.WithFields(logrus.Fields{
						"entity":     prometheusConfigName,
						"event uuid": prometheusConfig.mainUUID,
					}).Infof("prometheus %v is DOWN", eState.name)
					eState.mux.Lock()
					eState.gauge.Set(float64(0))
					*eState.state = 0
					eState.mux.Unlock()
					prometheusConfig.globalStateChange <- struct{}{}
				case <-eState.isUp:
					prometheusConfig.logging.WithFields(logrus.Fields{
						"entity":     prometheusConfigName,
						"event uuid": prometheusConfig.mainUUID,
					}).Infof("prometheus %v is UP", eState.name)
					eState.mux.Lock()
					*eState.state = 1
					eState.gauge.Set(float64(1))
					eState.mux.Unlock()
					prometheusConfig.globalStateChange <- struct{}{}
				}
			}
		}(eState, metricName)
	}
	go func(prometheusConfig *PrometheusConfig) {
		labels := make(map[string]string)
		labels["module"] = "total"
		newMetric := prometheus.NewGauge(
			prometheus.GaugeOpts{
				Name:        "program_health",
				Help:        helpName,
				ConstLabels: labels,
			})
		prometheus.MustRegister(newMetric)

		programHistogramHealth := prometheus.NewHistogram(
			prometheus.HistogramOpts{
				Name:    "program_histogram_health",
				Buckets: prometheus.LinearBuckets(0, 1, 2),
			})
		prometheus.MustRegister(programHistogramHealth)

		var globalState int
		for {
			<-prometheusConfig.globalStateChange
			globalState = 0

			for _, eState := range prometheusConfig.metrics {
				eState.rmux.Lock()
				globalState += *eState.state
				eState.rmux.Unlock()
			}

			if globalState != len(prometheusConfig.metrics) {
				prometheusConfig.logging.WithFields(logrus.Fields{
					"entity":     prometheusConfigName,
					"event uuid": prometheusConfig.mainUUID,
				}).Info("healht bad")
				newMetric.Set(float64(0))
				programHistogramHealth.Observe(float64(0))
			} else {
				prometheusConfig.logging.WithFields(logrus.Fields{
					"entity":     prometheusConfigName,
					"event uuid": prometheusConfig.mainUUID,
				}).Info("healht good")
				newMetric.Set(float64(1))
				programHistogramHealth.Observe(float64(1))
			}
			time.Sleep(50 * time.Millisecond)
		}
	}(prometheusConfig)
}

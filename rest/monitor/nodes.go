//  Copyright (c) 2015 Couchbase, Inc.
//  Licensed under the Apache License, Version 2.0 (the "License");
//  you may not use this file except in compliance with the
//  License. You may obtain a copy of the License at
//    http://www.apache.org/licenses/LICENSE-2.0
//  Unless required by applicable law or agreed to in writing,
//  software distributed under the License is distributed on an "AS
//  IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either
//  express or implied. See the License for the specific language
//  governing permissions and limitations under the License.

package monitor

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"time"
)

// A MonitorNodes struct holds all the tracking information for the
// StartMonitorNodes operation.
type MonitorNodes struct {
	urlUUIDs []UrlUUID // Array of base REST URL's to monitor.
	sampleCh chan MonitorSample
	options  MonitorNodesOptions
	stopCh   chan struct{}
}

type MonitorNodesOptions struct {
	StatsSampleInterval time.Duration // Ex: 1 * time.Second.
	StatsSampleDisable  bool

	DiagSampleInterval time.Duration
	DiagSampleDisable  bool

	// Optional, defaults to http.Get(); this is used, for example,
	// for unit testing.
	HttpGet func(url string) (resp *http.Response, err error)
}

// StartMonitorNodes begins REST stats and diag sampling from a fixed
// set of cbgt nodes.  Higher level parts (like StartMonitorCluster)
// should handle situations of node membership changes by stopping and
// restarting StartMonitorNodes() as needed.
//
// The cbgt REST URL endpoints that are monitored are [url]/api/stats
// and [url]/api/diag.
func StartMonitorNodes(
	urlUUIDs []UrlUUID,
	sampleCh chan MonitorSample,
	options MonitorNodesOptions,
) (*MonitorNodes, error) {
	m := &MonitorNodes{
		urlUUIDs: urlUUIDs,
		sampleCh: sampleCh,
		options:  options,
		stopCh:   make(chan struct{}),
	}

	for _, urlUUID := range urlUUIDs {
		go m.runNode(urlUUID)
	}

	return m, nil
}

func (m *MonitorNodes) Stop() {
	close(m.stopCh)
}

func (m *MonitorNodes) runNode(urlUUID UrlUUID) {
	statsSampleInterval := m.options.StatsSampleInterval
	if statsSampleInterval <= 0 {
		statsSampleInterval =
			DEFAULT_STATS_SAMPLE_INTERVAL_SECS * time.Second
	}

	diagSampleInterval := m.options.StatsSampleInterval
	if diagSampleInterval <= 0 {
		diagSampleInterval =
			DEFAULT_DIAG_SAMPLE_INTERVAL_SECS * time.Second
	}

	statsTicker := time.NewTicker(statsSampleInterval)
	defer statsTicker.Stop()

	diagTicker := time.NewTicker(diagSampleInterval)
	defer diagTicker.Stop()

	if !m.options.StatsSampleDisable {
		m.sample(urlUUID, "/api/stats", time.Now())
	}

	if !m.options.DiagSampleDisable {
		m.sample(urlUUID, "/api/diag", time.Now())
	}

	for {
		select {
		case <-m.stopCh:
			return

		case t, ok := <-statsTicker.C:
			if !ok {
				return
			}

			if !m.options.StatsSampleDisable {
				m.sample(urlUUID, "/api/stats", t)
			}

		case t, ok := <-diagTicker.C:
			if !ok {
				return
			}

			if !m.options.DiagSampleDisable {
				m.sample(urlUUID, "/api/diag", t)
			}
		}
	}
}

func (m *MonitorNodes) sample(
	urlUUID UrlUUID,
	kind string,
	start time.Time) {
	httpGet := m.options.HttpGet
	if httpGet == nil {
		httpGet = http.Get
	}

	res, err := httpGet(urlUUID.Url + kind)

	duration := time.Now().Sub(start)

	data := []byte(nil)
	if err == nil && res != nil {
		if res.StatusCode == 200 {
			var dataErr error

			data, dataErr = ioutil.ReadAll(res.Body)
			if err == nil && dataErr != nil {
				err = dataErr
			}
		} else {
			err = fmt.Errorf("nodes: sample res.StatusCode not 200,"+
				" res: %#v, urlUUID: %#v, kind: %s, err: %v",
				res, urlUUID, kind, err)
		}

		res.Body.Close()
	} else {
		err = fmt.Errorf("nodes: sample,"+
			" res: %#v, urlUUID: %#v, kind: %s, err: %v",
			res, urlUUID, kind, err)
	}

	monitorSample := MonitorSample{
		Kind:     kind,
		Url:      urlUUID.Url,
		UUID:     urlUUID.UUID,
		Start:    start,
		Duration: duration,
		Error:    err,
		Data:     data,
	}

	select {
	case <-m.stopCh:
	case m.sampleCh <- monitorSample:
	}
}

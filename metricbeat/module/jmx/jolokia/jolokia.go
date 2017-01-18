package jolokia

import (
	"fmt"
	"net/http"

	"io/ioutil"
	"strings"

	"github.com/elastic/beats/libbeat/common"
	"github.com/elastic/beats/libbeat/logp"
	"github.com/elastic/beats/metricbeat/mb"
)

var (
	debugf = logp.MakeDebug("jmx-jolokia")
)

// init registers the MetricSet with the central registry.
func init() {
	if err := mb.Registry.AddMetricSet("jmx", "jolokia", New); err != nil {
		panic(err)
	}
}

// MetricSet type defines all fields of the MetricSet
type MetricSet struct {
	mb.BaseMetricSet
	client                *http.Client      // HTTP client that is reused across requests
	metricSetConfig       []MetricSetConfig // array containing urls, bodies and mappings
	namespace             string
	updateElasticTemplate bool              // not implemented yet
}

// New create a new instance of the MetricSet
func New(base mb.BaseMetricSet) (mb.MetricSet, error) {
	// Additional configuration options
	config := struct {
		ModuleConfigInput     []MetricSetConfigInput `config:"mappings"`
		Namespace             string `config:"namespace" validate:"required"`
		UpdateElasticTemplate bool `config:"updateElasticTemplate"`
	}{}

	if err := base.Module().UnpackConfig(&config); err != nil {
		return nil, err
	}

	if moduleConfig, parseErr := parseConfig(config.ModuleConfigInput); parseErr != nil {
		return nil, parseErr
	} else {
		if config.UpdateElasticTemplate {
			// TODO: build and send an updated template to Elastic
		}
		return &MetricSet{
			BaseMetricSet:   base,
			metricSetConfig: moduleConfig,
			client:          &http.Client{Timeout: base.Module().Config().Timeout},
			namespace:     config.Namespace,
		}, nil
	}

}

// Fetch methods implements the data gathering and data conversion to the right format
func (m *MetricSet) Fetch() ([]common.MapStr, error) {

	var events []common.MapStr

	for _, elem := range m.metricSetConfig {
		req, err := http.NewRequest("POST", elem.Url, strings.NewReader(elem.Body))
		resp, err := m.client.Do(req)
		if err != nil {
			_ = fmt.Errorf("Error making http request: %#v", err)
			continue
		}
		defer resp.Body.Close()

		if resp.StatusCode != 200 {
			_ = fmt.Errorf("HTTP error %d: %s", resp.StatusCode, resp.Status)
			continue
		}

		resp_body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			_ = fmt.Errorf("Error converting response body: %#v", err)
			continue
		}

		event, err := eventMapping(resp_body, elem.Mapping, elem.Application, elem.Instance)
		if err != nil {
			continue
		}

		event["_namespace"] = m.namespace
		events = append(events, event)
	}
	if events != nil {
		return events, nil
	} else {
		return nil, fmt.Errorf("No events could be fetched, please check the log for errors")
	}

}

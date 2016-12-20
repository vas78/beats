package cluster

import (
	"fmt"
	"net/http"

	"github.com/elastic/beats/libbeat/common"
	"github.com/elastic/beats/libbeat/logp"
	"github.com/vas78/beats/metricbeat/mb"
	"github.com/vas78/beats/metricbeat/mb/parse"
	"io/ioutil"
	"encoding/json"
	"crypto/tls"
)

const (
	// defaultScheme is the default scheme to use when it is not specified in
	// the host config.
	defaultScheme = "http"

	// defaultPath is the default path to the metrics endpoint on Yarn.
	defaultPath = "/ws/v1/cluster/metrics"
)

var (
	debugf = logp.MakeDebug("yarn-cluster")

	hostParser = parse.URLHostParserBuilder{
		DefaultScheme: defaultScheme,
		PathConfigKey: "yarn_metrics_path",
		DefaultPath:   defaultPath,
	}.Build()
)

// init registers the MetricSet with the central registry.
func init() {
	if err := mb.Registry.AddMetricSet("yarn", "cluster", New, hostParser); err != nil {
		panic(err)
	}
}

// MetricSet type defines all fields of the MetricSet
type MetricSet struct {
	mb.BaseMetricSet
	client *http.Client // HTTP client that is reused across requests
}

// New create a new instance of the MetricSet
func New(base mb.BaseMetricSet) (mb.MetricSet, error) {
	config := struct {
		Schema   string `config:"schema"`
		Host     string `config:"host"`
		User     string        `config:"user"`
		Password string        `config:"password"`
		Insecure bool     `config:"insecure"`
	}{
		User:  "",
		Password: "",
	}

	if err := base.Module().UnpackConfig(&config); err != nil {
		return nil, err
	}

	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify : config.Insecure},
	}

	return &MetricSet{
		BaseMetricSet:   base,
		client:          &http.Client{Transport: tr, Timeout: base.Module().Config().Timeout},
	}, nil

}

// Fetch methods implements the data gathering and data conversion to the right format
func (m *MetricSet) Fetch() (common.MapStr, error) {

	req, err := http.NewRequest("GET", m.HostData().SanitizedURI, nil)
	if m.HostData().User != "" || m.HostData().Password != "" {
		req.SetBasicAuth(m.HostData().User, m.HostData().Password)
	}
	resp, err := m.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("error making http request: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("HTTP error %d: %s", resp.StatusCode, resp.Status)
	}

	resp_body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("Error converting response body: %#v", err)
	}

	status := map[string]interface{}{}

	err = json.Unmarshal(resp_body, &status)

	if err != nil {
		return nil, fmt.Errorf("Cannot unmarshal json response: %s", err)
	}

	debugf("Unmarshalled json: ", status)

	return eventMapping(status), nil

}

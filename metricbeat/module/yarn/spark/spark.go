package spark

import (
	"fmt"
	"net/http"

	"io/ioutil"

	"github.com/elastic/beats/libbeat/common"
	"github.com/elastic/beats/libbeat/logp"
	"github.com/vas78/beats/metricbeat/mb"
	"github.com/vas78/beats/metricbeat/mb/parse"
	"crypto/tls"
	"encoding/json"
)

const (
	// defaultScheme is the default scheme to use when it is not specified in
	// the host config.
	defaultScheme = "http"

	// defaultPath is the default path to the metrics endpoint on Yarn.
	defaultPath = "/ws/v1/cluster/apps/"
)

var (
	debugf = logp.MakeDebug("yarn-spark")

	hostParser = parse.URLHostParserBuilder{
		DefaultScheme: defaultScheme,
		PathConfigKey: "spark_apps_path",
		DefaultPath:   defaultPath,
	}.Build()
)

// init registers the MetricSet with the central registry.
func init() {
	if err := mb.Registry.AddMetricSet("yarn", "spark", New, hostParser); err != nil {
		panic(err)
	}
}

// MetricSet type defines all fields of the MetricSet
type MetricSet struct {
	mb.BaseMetricSet
	client          *http.Client      // HTTP client that is reused across requests
	queue           string
	running         bool
}

// New create a new instance of the MetricSet
func New(base mb.BaseMetricSet) (mb.MetricSet, error) {
	config := struct {
		Schema   string `config:"schema"`
		Host     string `config:"host"`
		User     string  `config:"user"`
		Password string  `config:"password"`
		Insecure bool     `config:"insecure"`
		Queue string     `config:"queue"`
		Running bool     `config:"running"`
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
		queue:       config.Queue,
		running:     config.Running,
	}, nil

}

// Fetch methods implements the data gathering and data conversion to the right format
func (m *MetricSet) Fetch() ([]common.MapStr, error) {

	var events []common.MapStr

	jobList := getSparkJobTrackingUrls(m)

	for _, job := range jobList {
		debugf("Tracking URL: ", job.TrackingUrl)
		req, err := http.NewRequest("GET", job.TrackingUrl + "metrics/json", nil)
		if m.HostData().User != "" || m.HostData().Password != "" {
			req.SetBasicAuth(m.HostData().User, m.HostData().Password)
		}
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

		appStats := map[string]interface{}{}
		err = json.Unmarshal(resp_body, &appStats)
		if err != nil {
			_ = fmt.Errorf("Cannot unmarshal json response: %s", err)
			continue
		}

		event := eventMapping(job, appStats)
		event["application_name"] = job.Name

		events = append(events, event)
	}
	if events != nil {
		return events, nil
	} else {
		return nil, fmt.Errorf("No events could be fetched, please check the log for errors")
	}

}

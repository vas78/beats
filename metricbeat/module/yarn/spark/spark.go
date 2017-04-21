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
	executor_stats  bool
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
		ExecutorStats bool     `config:"executorStats"`
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
		executor_stats:     config.ExecutorStats,
	}, nil

}

// Fetch methods implements the data gathering and data conversion to the right format
func (m *MetricSet) Fetch() ([]common.MapStr, error) {

	var events []common.MapStr

	jobList := getSparkJobTrackingUrls(m)

	for _, job := range jobList {
		resp_body, err := GetResponse(m, job.TrackingUrl + "metrics/json")
		if err != nil {
			continue
		}

		driverStats := map[string]interface{}{}
		err = json.Unmarshal(resp_body, &driverStats)
		if err != nil {
			_ = fmt.Errorf("Cannot unmarshal json response: %s", err)
			continue
		}

		event := eventMappingDriverStats(job, driverStats)
		event["application_name"] = job.Name
		event["metric_source"] = "driver"

		events = append(events, event)

		if m.executor_stats {
			resp_body, err := GetResponse(m, job.TrackingUrl + "api/v1/applications/" + job.Id + "/executors")
			if err != nil {
				continue
			}

			executorStats := []map[string]interface{}{}
			err = json.Unmarshal(resp_body, &executorStats)
			if err != nil {
				_ = fmt.Errorf("Cannot unmarshal json response: %s", err)
				continue
			}

			for _, execStat := range executorStats {
				event := eventMappingExecutorStats(execStat)
				event["application_name"] = job.Name
				event["metric_source"] = "executor"

				events = append(events, event)
			}
		}
	}
	if events != nil {
		return events, nil
	} else {
		return nil, fmt.Errorf("No events could be fetched, please check the log for errors")
	}

}

//helper function that GETs data over http
func GetResponse(m *MetricSet, Url string) ([]byte, error) {

	debugf("URL: ", Url)
	req, err := http.NewRequest("GET", Url, nil)
	if m.HostData().User != "" || m.HostData().Password != "" {
		req.SetBasicAuth(m.HostData().User, m.HostData().Password)
	}
	resp, err := m.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("Error making http request: %#v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("HTTP error %d: %s", resp.StatusCode, resp.Status)
	}
	resp_body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("Error converting response body: %#v", err)
	}

	return resp_body, nil
}

package consumer

import (
	"github.com/elastic/beats/libbeat/common"
	"fmt"
	"io/ioutil"
	"net/http"
	"encoding/json"
)

type ConsumerLagsResponse struct {
	Error bool `json:"error"`
	Message string `json:"message"`
	Status Status `json:"status"`
}

type Status struct {
	Cluster string `json:"cluster"`
	Group string `json:"group"`
	Status string `json:"status"`
	Complete bool `json:"complete"`
	Partitions []PartitionStatus `json:"partitions"`
	Partition_count int `json:"partition_count"`
	Maxlag int `json:"maxlag"`
	Totallag int `json:"totallag"`
}

type PartitionStatus struct {
	Topic string `json:"topic"`
	Partition int `json:"partition"`
	Status string `json:"status"`
	Start map[string]interface{} `json:"start"`
	End map[string]interface{} `json:"end"`
}


// Map responseBody to common.MapStr
func eventMapping(responseBody []byte, lags_per_partition bool) ([]common.MapStr, error) {

	events := []common.MapStr{}

	debugf("Got reponse body: ", string(responseBody[:]))

	var consumer_lags_response ConsumerLagsResponse
	err := json.Unmarshal(responseBody, &consumer_lags_response)

	if err != nil {
		return nil, fmt.Errorf("Cannot unmarshal json response: %s", err)
	}
	debugf("Unmarshalled json: ", consumer_lags_response)
	if consumer_lags_response.Error == true {
		return nil, fmt.Errorf("Got error from Kafka: %s", consumer_lags_response.Message)
	}

	event := make(map[string]interface{})
	event["cluster"] = consumer_lags_response.Status.Cluster
	event["group"] = consumer_lags_response.Status.Group
	event["status"] = consumer_lags_response.Status.Status
	event["complete"] = consumer_lags_response.Status.Complete
	event["partition_count"] = consumer_lags_response.Status.Partition_count
	event["max_lag"] = consumer_lags_response.Status.Maxlag
	event["total_lag"] = consumer_lags_response.Status.Totallag

	events = append(events, event)

	if lags_per_partition {
		for _, partition_status := range consumer_lags_response.Status.Partitions {
			event := make(map[string]interface{})
			event["cluster"] = consumer_lags_response.Status.Cluster
			event["group"] = consumer_lags_response.Status.Group
			event["topic"] = partition_status.Topic
			event["partition"] = partition_status.Partition
			event["status"] = partition_status.Status
			event["start"] = partition_status.Start
			event["end"] = partition_status.End
			events = append(events, event)
		}
	}

	return events, nil
}


type ConsumerGroupsResponse struct {
	Error bool `json:"error"`
	Message string `json:"message"`
	Consumers []string `json:"consumers"`
}

//fetch all connected consumer groups
func fetchConsumerGroups(m *MetricSet) ([]string, error) {
	list_consumer_url := "http://" + m.host + "/v2/kafka/" + m.cluster + "/consumer/"
	debugf("Fetching consumer groups from: ", list_consumer_url)
	req, err := http.NewRequest("GET", list_consumer_url, nil)
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
	debugf("Got reponse body: ", string(resp_body[:]))

	var consumer_groups ConsumerGroupsResponse
	err = json.Unmarshal(resp_body, &consumer_groups)
	if err != nil {
		return nil, fmt.Errorf("Cannot unmarshal json response: %s", err)
	}

	debugf("Unmarshalled json: ", consumer_groups)

	if consumer_groups.Error == true {
		return nil, fmt.Errorf("Got error from Kafka: %s", consumer_groups.Message)
	}

	return consumer_groups.Consumers, nil
}
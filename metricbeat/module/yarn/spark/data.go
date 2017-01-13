package spark

import (
	"encoding/json"
	"fmt"
	s "github.com/elastic/beats/metricbeat/schema"
	c "github.com/elastic/beats/metricbeat/schema/mapstriface"
	"io/ioutil"
	"net/http"
	"github.com/elastic/beats/libbeat/common"
)

type App struct {
	Id string `json:"id"`
	Name string `json:"name"`
	TrackingUrl string `json:"trackingUrl"`
}

type AppResponse struct {
	AppList []App `json:"app"`
}

type Apps struct {
	Apps AppResponse `json:"apps"`
}

func schema(appId string, appName string) s.Schema {
	return s.Schema{
		"block_manager": c.Dict("gauges",
			s.Schema{"disk_space_used": c.Dict(
					appId + ".driver.BlockManager.disk.diskSpaceUsed_MB", s.Schema{"mb": c.Int("value")}),
				"memory_max": c.Dict(
					appId + ".driver.BlockManager.memory.maxMem_MB", s.Schema{"mb": c.Int("value")}),
				"memory_used": c.Dict(
					appId + ".driver.BlockManager.memory.memUsed_MB", s.Schema{"mb": c.Int("value")}),
				"remaining_mem": c.Dict(
					appId + ".driver.BlockManager.memory.remainingMem_MB", s.Schema{"mb": c.Int("value")}),

				},
			),
		"dag_scheduler": c.Dict("gauges",
			s.Schema{"jobs_active": c.Dict(
					appId + ".driver.DAGScheduler.job.activeJobs", s.Schema{"value": c.Int("value")}),
				"jobs_total": c.Dict(
					appId + ".driver.DAGScheduler.job.allJobs", s.Schema{"value": c.Int("value")}),
				"stages_failed": c.Dict(
					appId + ".driver.DAGScheduler.stage.failedStages", s.Schema{"value": c.Int("value")}),
				"stages_running": c.Dict(
					appId + ".driver.DAGScheduler.stage.runningStages", s.Schema{"value": c.Int("value")}),
				"stages_waiting": c.Dict(
					appId + ".driver.DAGScheduler.stage.waitingStages", s.Schema{"value": c.Int("value")}),

			},
		),
		"dag_message_processing": c.Dict("timers",
			s.Schema{"time_ms": c.Dict(
					appId + ".driver.DAGScheduler.messageProcessingTime",
						s.Schema{
							"max": c.Int("max"),
							"mean": c.Int("mean"),
							"min": c.Int("min"),
							"p50": c.Int("p50"),
							"p95": c.Int("p95"),
							"p99": c.Int("p99"),
							"p999": c.Int("p999"),
						}),
				"rate": c.Dict(
					appId + ".driver.DAGScheduler.messageProcessingTime",
					s.Schema{
						"count": c.Int("count"),
						"m1_rate": c.Int("m1_rate"),
						"m5_rate": c.Int("m5_rate"),
						"m15_rate": c.Int("m15_rate"),
					}),
			},
		),
		"streaming": c.Dict("gauges",
			s.Schema{"last_completed_batch": s.Object{
					"total_delay": c.Dict(
						appId + ".driver." + appName + ".StreamingMetrics.streaming.lastCompletedBatch_totalDelay", s.Schema{"ms": c.Int("value")}),
					"processing_delay": c.Dict(
						appId + ".driver." + appName + ".StreamingMetrics.streaming.lastCompletedBatch_processingDelay", s.Schema{"ms": c.Int("value")}),
					"scheduling_delay": c.Dict(
						appId + ".driver." + appName + ".StreamingMetrics.streaming.lastCompletedBatch_schedulingDelay", s.Schema{"ms": c.Int("value")}),
					"processing_start_time": c.Dict(
						appId + ".driver." + appName + ".StreamingMetrics.streaming.lastCompletedBatch_processingStartTime", s.Schema{"ms": c.Int("value")}),
					"processing_end_time": c.Dict(
						appId + ".driver." + appName + ".StreamingMetrics.streaming.lastCompletedBatch_processingEndTime", s.Schema{"ms": c.Int("value")}),
					"submission_time": c.Dict(
						appId + ".driver." + appName + ".StreamingMetrics.streaming.lastCompletedBatch_submissionTime", s.Schema{"ms": c.Int("value")}),
				},
				"last_received_batch": s.Object{
					"records": c.Dict(
						appId + ".driver." + appName + ".StreamingMetrics.streaming.lastReceivedBatch_records", s.Schema{"count": c.Int("value")}),
					"processing_start_time": c.Dict(
						appId + ".driver." + appName + ".StreamingMetrics.streaming.lastReceivedBatch_processingStartTime", s.Schema{"ms": c.Int("value")}),
					"processing_end_time": c.Dict(
						appId + ".driver." + appName + ".StreamingMetrics.streaming.lastReceivedBatch_processingEndTime", s.Schema{"ms": c.Int("value")}),
					"submission_time": c.Dict(
						appId + ".driver." + appName + ".StreamingMetrics.streaming.lastReceivedBatch_submissionTime", s.Schema{"ms": c.Int("value")}),
				},
				"receivers": c.Dict(
					appId + ".driver." + appName + ".StreamingMetrics.streaming.receivers", s.Schema{"count": c.Int("value")}),
				"retained_completed_batches": c.Dict(
					appId + ".driver." + appName + ".StreamingMetrics.streaming.retainedCompletedBatches", s.Schema{"count": c.Int("value")}),
				"running_batches": c.Dict(
					appId + ".driver." + appName + ".StreamingMetrics.streaming.runningBatches", s.Schema{"count": c.Int("value")}),
				"total_completed_batches": c.Dict(
					appId + ".driver." + appName + ".StreamingMetrics.streaming.totalCompletedBatches", s.Schema{"count": c.Int("value")}),
				"total_processed_records": c.Dict(
					appId + ".driver." + appName + ".StreamingMetrics.streaming.totalProcessedRecords", s.Schema{"count": c.Int("value")}),
				"total_received_records": c.Dict(
					appId + ".driver." + appName + ".StreamingMetrics.streaming.totalReceivedRecords", s.Schema{"count": c.Int("value")}),
				"unprocessed_batches": c.Dict(
					appId + ".driver." + appName + ".StreamingMetrics.streaming.unprocessedBatches", s.Schema{"count": c.Int("value")}),
				"waiting_batches": c.Dict(
					appId + ".driver." + appName + ".StreamingMetrics.streaming.waitingBatches", s.Schema{"count": c.Int("value")}),

			},
		),
	}
}

func eventMapping(app App, appStats map[string]interface{}) common.MapStr {
	return schema(app.Id, app.Name).Apply(appStats)
}

func getSparkJobTrackingUrls(m *MetricSet) []App {

	//prepare url for listing spark jobs
	var url string
	if m.queue != "" {
		url = m.HostData().SanitizedURI + "?queue=" + m.queue
	} else {
		url = m.HostData().SanitizedURI
	}
	if m.running {
		url = url + "&states=running"
	}
	//list spark jobs
	req, err := http.NewRequest("GET", url, nil)
	if m.HostData().User != "" || m.HostData().Password != "" {
		req.SetBasicAuth(m.HostData().User, m.HostData().Password)
	}
	resp, err := m.client.Do(req)
	if err != nil {
		_ = fmt.Errorf("Error making http request: %#v", err)
		return nil
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		_ = fmt.Errorf("HTTP error %d: %s", resp.StatusCode, resp.Status)
		return nil
	}
	resp_body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		_ = fmt.Errorf("Error converting response body: %#v", err)
		return nil
	}

	//parse the list
	var appList Apps
	err = json.Unmarshal(resp_body, &appList)
	if err != nil {
		_ = fmt.Errorf("Cannot unmarshal json response: %s", err)
		return nil
	}

	return appList.Apps.AppList

}

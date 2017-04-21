package spark

import (
	"path/filepath"
	"testing"

	"io/ioutil"

	"github.com/stretchr/testify/assert"
	"encoding/json"
	"github.com/elastic/beats/libbeat/common"
)

func TestEventMapper(t *testing.T) {
	absPath, err := filepath.Abs("./test/")

	assert.NotNil(t, absPath)
	assert.Nil(t, err)

	app := App {"application_1478087666420_0098","MyStreamingJobName",
		"https://mytestserver:8080/proxy/application_1478087666420_0098/"}

	//driver stats
	responseDriverStats, err := ioutil.ReadFile(absPath + "/responseDriverStats.json")

	assert.Nil(t, err)
	resultDriverStats := map[string]interface{}{}


	_ = json.Unmarshal(responseDriverStats, &resultDriverStats)
	eventDriverStats := eventMappingDriverStats(app, resultDriverStats)
	//t.Log(eventDriverStats)

	assert.EqualValues(t, 291, eventDriverStats["block_manager"].(common.MapStr)["memory_max"].(common.MapStr)["mb"].(int64))


	//executor stats
	responseExecutorStats, err := ioutil.ReadFile(absPath + "/responseExecutorStats.json")

	assert.Nil(t, err)
	resultExecutorStats := []map[string]interface{}{}

	_ = json.Unmarshal(responseExecutorStats, &resultExecutorStats)


	events := []map[string]interface{}{}
	for _, execStat := range resultExecutorStats {
		event := eventMappingExecutorStats(execStat)
		//t.Log(event)
		events = append(events, event)
	}

	assert.EqualValues(t, "driver", events[0]["id"])
	assert.EqualValues(t, 384093388, events[0]["maxMemory"].(int64))
	assert.EqualValues(t, 18, events[0]["rddBlocks"].(int64))

	assert.EqualValues(t, "5", events[1]["id"])
	assert.EqualValues(t, 1, events[1]["maxTasks"].(int64))
	assert.EqualValues(t, 0, events[1]["totalInputBytes"].(int64))

	assert.EqualValues(t, "4", events[2]["id"])
	assert.EqualValues(t, 168210, events[2]["memoryUsed"].(int64))
	assert.EqualValues(t, 159896879, events[2]["totalShuffleWrite"].(int64))

}

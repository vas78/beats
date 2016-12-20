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

	response, err := ioutil.ReadFile(absPath + "/response.json")

	assert.Nil(t, err)
	result := map[string]interface{}{}

	app := App {"application_1478087666420_0098","MyStreamingJobName",
		"https://mytestserver:8080/proxy/application_1478087666420_0098/"}

	_ = json.Unmarshal(response, &result)
	event := eventMapping(app, result)
	//t.Log(event)

	assert.EqualValues(t, 291, event["block_manager"].(common.MapStr)["memory_max"].(common.MapStr)["mb"].(int64))

}

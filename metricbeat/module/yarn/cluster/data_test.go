package cluster

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

	_ = json.Unmarshal(response, &result)
	event := eventMapping(result)

	assert.EqualValues(t, 3, event["apps"].(common.MapStr)["running"].(int64))
	assert.EqualValues(t, 0, event["apps"].(common.MapStr)["completed"].(int64))
	assert.EqualValues(t, 509952, event["memory_mb"].(common.MapStr)["available"].(int64))
	assert.EqualValues(t, 30720, event["memory_mb"].(common.MapStr)["allocated"].(int64))
	assert.EqualValues(t, 73, event["virtual_cores"].(common.MapStr)["available"].(int64))
	assert.EqualValues(t, 15, event["containers"].(common.MapStr)["allocated"].(int64))
	assert.EqualValues(t, 4, event["nodes"].(common.MapStr)["active"].(int64))

}

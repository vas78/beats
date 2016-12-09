package cluster


import (
	s "github.com/elastic/beats/metricbeat/schema"
	c "github.com/elastic/beats/metricbeat/schema/mapstriface"
)

var schema = s.Schema{
	"apps": c.Dict("clusterMetrics",
		s.Schema{"submitted": c.Int("appsSubmitted"),
			"completed": c.Int("appsCompleted"),
			"running": c.Int("appsRunning"),
			"pending": c.Int("appsPending"),
			"failed": c.Int("appsFailed"),
			"killed": c.Int("appsKilled"),
		}),
	"memory_mb": c.Dict("clusterMetrics",
		s.Schema{"available": c.Int("availableMB"),
			"reserved": c.Int("reservedMB"),
			"allocated": c.Int("allocatedMB"),
			"total": c.Int("totalMB"),
		}),
	"virtual_cores": c.Dict("clusterMetrics",
		s.Schema{"available": c.Int("availableVirtualCores"),
			"reserved": c.Int("reservedVirtualCores"),
			"allocated": c.Int("allocatedVirtualCores"),
			"total": c.Int("totalVirtualCores"),
		}),
	"containers": c.Dict("clusterMetrics",
		s.Schema{"reserved": c.Int("containersReserved"),
			"allocated": c.Int("containersAllocated"),
			"pending": c.Int("containersPending"),
		}),
	"nodes": c.Dict("clusterMetrics",
		s.Schema{"total": c.Int("totalNodes"),
			"lost": c.Int("lostNodes"),
			"unhealthy": c.Int("unhealthyNodes"),
			"decommissioned": c.Int("decommissionedNodes"),
			"rebooted": c.Int("rebootedNodes"),
			"active": c.Int("activeNodes"),
		}),
}

var eventMapping = schema.Apply
package status

import (
	"github.com/elastic/beats/libbeat/common"
	"github.com/elastic/beats/metricbeat/mb"
	"github.com/elastic/beats/metricbeat/module/mongodb"

	"github.com/pkg/errors"
	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
)

/*
TODOs:
	* add metricset for "locks" data
	* add a metricset for "metrics" data
*/

func init() {
	if err := mb.Registry.AddMetricSet("mongodb", "status", New, mongodb.ParseURL); err != nil {
		panic(err)
	}
}

// MetricSet type defines all fields of the MetricSet
// As a minimum it must inherit the mb.BaseMetricSet fields, but can be extended with
// additional entries. These variables can be used to persist data or configuration between
// multiple fetch calls.
type MetricSet struct {
	mb.BaseMetricSet
	dialInfo     *mgo.DialInfo
	mongoSession *mgo.Session
}

// New creates a new instance of the MetricSet
// Part of new is also setting up the configuration by processing additional
// configuration entries if needed.
func New(base mb.BaseMetricSet) (mb.MetricSet, error) {
	dialInfo, err := mgo.ParseURL(base.HostData().URI)
	if err != nil {
		return nil, err
	}
	dialInfo.Timeout = base.Module().Config().Timeout

	mongoSession := mongodb.NewSession(dialInfo)
	mongoSession.SetMode(mgo.Monotonic, true)

	return &MetricSet{
		BaseMetricSet: base,
		dialInfo:      dialInfo,
		mongoSession:  mongoSession,
	}, nil
}

// Fetch methods implements the data gathering and data conversion to the right format
// It returns the event which is then forward to the output. In case of an error, a
// descriptive error must be returned.
func (m *MetricSet) Fetch() (common.MapStr, error) {
	result := map[string]interface{}{}
	if err := m.mongoSession.DB("admin").Run(bson.D{{"serverStatus", 1}}, &result); err != nil {
		return nil, errors.Wrap(err, "mongodb fetch failed")
	}

	return eventMapping(result), nil
}

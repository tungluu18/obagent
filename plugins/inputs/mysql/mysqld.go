// Copyright (c) 2021 OceanBase
// obagent is licensed under Mulan PSL v2.
// You can use this software according to the terms and conditions of the Mulan PSL v2.
// You may obtain a copy of Mulan PSL v2 at:
//
// http://license.coscl.org.cn/MulanPSL2
//
// THIS SOFTWARE IS PROVIDED ON AN "AS IS" BASIS, WITHOUT WARRANTIES OF ANY KIND,
// EITHER EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO NON-INFRINGEMENT,
// MERCHANTABILITY OR FIT FOR A PARTICULAR PURPOSE.
// See the Mulan PSL v2 for more details.

package mysql

import (
	"context"
	"os"
	"sync"

	log2 "github.com/go-kit/kit/log/logrus"
	kitLog "github.com/go-kit/log"
	log "github.com/sirupsen/logrus"
	"gopkg.in/alecthomas/kingpin.v2"

	_ "github.com/go-sql-driver/mysql"
	"github.com/pkg/errors"
	"github.com/prometheus/client_golang/prometheus"
	"gopkg.in/yaml.v3"

	"github.com/prometheus/mysqld_exporter/collector"

	"github.com/oceanbase/obagent/metric"
)

const mysqldSampleConfig = `
`

const mysqldDescription = `
`

var (
	dsn string
)

// Scraper version will be validated against MySQL version
// ignoredVersionScraper: a wrapper which pseudo-downgrades scraper version to make it compatible with Oceanbase
type ignoredVersionScraper struct {
	collector.Scraper
}

func (ignoredVersionScraper) Version() float64 {
	return 0.0
}

func ignoreVersion(c collector.Scraper) collector.Scraper {
	return ignoredVersionScraper{c}
}

// scrapers lists all possible collection methods and if they should be enabled by default.
var scrapers = map[collector.Scraper]bool{
	collector.ScrapeGlobalStatus{}:                        true,
	collector.ScrapeGlobalVariables{}:                     true,
	collector.ScrapeSlaveStatus{}:                         true,
	collector.ScrapeProcesslist{}:                         false,
	collector.ScrapeUser{}:                                false,
	collector.ScrapeTableSchema{}:                         false,
	collector.ScrapeInfoSchemaInnodbTablespaces{}:         false,
	collector.ScrapeInnodbMetrics{}:                       false,
	collector.ScrapeAutoIncrementColumns{}:                false,
	collector.ScrapeBinlogSize{}:                          false,
	collector.ScrapePerfTableIOWaits{}:                    false,
	collector.ScrapePerfIndexIOWaits{}:                    false,
	collector.ScrapePerfTableLockWaits{}:                  false,
	collector.ScrapePerfEventsStatements{}:                false,
	collector.ScrapePerfEventsStatementsSum{}:             false,
	collector.ScrapePerfEventsWaits{}:                     false,
	collector.ScrapePerfFileEvents{}:                      false,
	collector.ScrapePerfFileInstances{}:                   false,
	collector.ScrapePerfMemoryEvents{}:                    false,
	collector.ScrapePerfReplicationGroupMembers{}:         false,
	collector.ScrapePerfReplicationGroupMemberStats{}:     false,
	collector.ScrapePerfReplicationApplierStatsByWorker{}: false,
	collector.ScrapeUserStat{}:                            false,
	collector.ScrapeClientStat{}:                          false,
	collector.ScrapeTableStat{}:                           false,
	collector.ScrapeSchemaStat{}:                          false,
	collector.ScrapeInnodbCmp{}:                           true,
	collector.ScrapeInnodbCmpMem{}:                        true,
	collector.ScrapeQueryResponseTime{}:                   true,
	collector.ScrapeEngineTokudbStatus{}:                  false,
	collector.ScrapeEngineInnodbStatus{}:                  false,
	collector.ScrapeHeartbeat{}:                           false,
	collector.ScrapeSlaveHosts{}:                          false,
	collector.ScrapeReplicaHost{}:                         false,
}

type MysqldConfig struct {
	Dsn          string          `yaml:"dsn"`
	ScraperFlags map[string]bool `yaml:"scraperFlags"`
}

type MysqldExporter struct {
	Config          *MysqldConfig
	logger          kitLog.Logger
	registry        *prometheus.Registry
	collector       *collector.Exporter
	enabledScrapers []collector.Scraper
}

func (m *MysqldExporter) Close() error {
	m.registry.Unregister(m.collector)
	return nil
}

func (m *MysqldExporter) SampleConfig() string {
	return mysqldSampleConfig
}

func (m *MysqldExporter) Description() string {
	return mysqldDescription
}

func (m *MysqldExporter) Init(config map[string]interface{}) error {
	var pluginConfig MysqldConfig

	configBytes, err := yaml.Marshal(config)
	if err != nil {
		return errors.Wrap(err, "mysqld exporter encode config")
	}

	err = yaml.Unmarshal(configBytes, &pluginConfig)
	if err != nil {
		return errors.Wrap(err, "mysqld exporter decode config")
	}

	commandLineParse()

	m.logger = log2.NewLogrusLogger(log.StandardLogger())

	m.Config = &pluginConfig
	log.Info("table input init with config", m.Config)

	m.enabledScrapers = make([]collector.Scraper, 0, len(scrapers))

	for scraper, enabledByDefault := range scrapers {
		enabled, found := m.Config.ScraperFlags[scraper.Name()]
		if (found && enabled) || (!found && enabledByDefault) {
			m.enabledScrapers = append(m.enabledScrapers, ignoreVersion(scraper))
		}
	}

	ctx := context.Background()
	m.collector = collector.New(ctx, m.Config.Dsn, collector.NewMetrics(), m.enabledScrapers, m.logger)
	m.registry = prometheus.NewRegistry()
	err = m.registry.Register(m.collector)
	if err != nil {
		return errors.Wrap(err, "mysqld exporter register collector")
	}

	return err
}

var once sync.Once

func commandLineParse() {
	once.Do(func() {
		lastIndex := len(os.Args) - 1
		copy(os.Args[lastIndex:], os.Args)
		os.Args = os.Args[lastIndex:]

		kingpin.Parse()
	})
}

func (m *MysqldExporter) Collect() ([]metric.Metric, error) {
	// TODO parse metric families

	var metrics []metric.Metric

	metricFamilies, err := m.registry.Gather()
	if err != nil {
		return nil, errors.Wrap(err, "mysql exporter registry gather")
	}
	for _, metricFamily := range metricFamilies {
		metricsFromMetricFamily := metric.ParseFromMetricFamily(metricFamily)
		metrics = append(metrics, metricsFromMetricFamily...)
	}

	return metrics, nil
}

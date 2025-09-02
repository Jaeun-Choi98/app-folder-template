package cron

import (
	"context"
	"fmt"
	"pjt/internal/config"
	dbhandler "pjt/internal/db/db-handler"
	"pjt/internal/logger"

	"strings"
	"sync"
	"time"

	"github.com/Jaeun-Choi98/modules/utils"

	"github.com/Jaeun-Choi98/modules/eventbus"
)

const (
	maxAgeDays       = -1
	partitionAgeDays = -1
)

var DynamicDeviceLogTableName string
var CreateDeviceLog time.Time
var cronMu sync.Mutex

type Cron struct {
	config   *config.Configuration
	dao      dbhandler.DBHandlerInterface
	eventBus *eventbus.EventBus

	wg sync.WaitGroup

	ctx    context.Context
	cancel context.CancelFunc
}

func NewCron(config *config.Configuration, dao dbhandler.DBHandlerInterface, eventbus *eventbus.EventBus) *Cron {
	ctx, cancel := context.WithCancel(context.Background())
	c := &Cron{
		config:   config,
		dao:      dao,
		eventBus: eventbus,
		ctx:      ctx,
		cancel:   cancel,
	}

	c.LoadLogTableInfo()
	c.CleanDBLogTableJob()
	c.DynamicDBLogTableJob(true)
	return c
}

func (c *Cron) LoadLogTableInfo() error {

	if DynamicDeviceLogTableName == "" {
		c.renewTableName()
	}

	return nil
}

func (c *Cron) Shutdown() {
	c.cancel()
	c.wg.Wait()
	logger.Println("[Cron] Cron routine is terminated")
}

func (c *Cron) StartCron() error {
	tickerCheck := time.NewTicker(1 * time.Minute)
	c.wg.Add(1)
	defer func() {
		c.wg.Done()
		tickerCheck.Stop()
	}()
	for {
		select {
		case <-c.ctx.Done():
			return nil
		case <-tickerCheck.C:
			now := time.Now().In(utils.LocalKorea)
			var update1 bool
			if CreateDeviceLog.Before(now.AddDate(0, 0, partitionAgeDays)) {
				c.renewTableName()
				update1 = true
			}
			c.DynamicDBLogTableJob(update1)
		}
	}
}

func (c *Cron) renewTableName() {
	t := time.Now().In(utils.LocalKorea)
	year, month, day := t.Date()
	t = time.Date(year, month, day, 0, 0, 0, 0, utils.LocalKorea)
	cronMu.Lock()
	DynamicDeviceLogTableName = fmt.Sprintf("dvc_log_%d%02d%02d", t.Year(), t.Month(), t.Day())
	CreateDeviceLog = t
	cronMu.Unlock()
}

// 외부에서 호출하기 위함.
func GetDeviceLogTableName() string {
	cronMu.Lock()
	defer cronMu.Unlock()
	return DynamicDeviceLogTableName
}

func (c *Cron) DynamicDBLogTableJob(updateDevice bool) {
	switch c.config.DBType {
	case "oracle":
		if updateDevice {
			// if err := c.dao.CreateTable(strings.ToUpper(DynamicDeviceLogTableName), GetOraQuery()); err != nil {
			// 	logger.Println("[Cron] Failed to create log table")
			// }
		}
	case "maria":
	default:
		logger.Println("[Cron] Failed to create log table, invalid config.DBType")
	}
}

func GetOraQuery() []string {
	var queries []string
	tableName := strings.ToUpper(DynamicDeviceLogTableName)
	query := `CREATE TABLE %s ....`
	queries = append(queries, fmt.Sprintf(query, tableName))
	return queries
}

func (c *Cron) CleanDBLogTableJob() {

	cutOff := time.Now().In(utils.LocalKorea).AddDate(0, 0, maxAgeDays)
	start := cutOff.AddDate(-1, 0, 0).Add(-1 * time.Second)

	for start.Before(cutOff) {
		tbName := fmt.Sprintf("dvc_log_%d%02d%02d", start.Year(), start.Month(), start.Day())
		var tableName string
		switch c.config.DBType {
		case "oracle":
			tableName = strings.ToUpper(tbName)
		case "maria":
			tableName = strings.ToLower(tbName)
		default:
			logger.Println("[Cron] Failed to clean db log_table, invalid config.DBType")
			return
		}
		logger.Println(tableName)
		// if exsits, err := c.dao.DeleteTableLog(tableName); err != nil {
		// 	logger.Printf("[Cron] Failed to clean db log_table, DeleteTableLog: %s", tableName)
		// } else if exsits == 1 {
		// 	logger.Printf("[Cron] Drop the table: %s", tableName)
		// }

		start = start.AddDate(0, 0, 1)
	}
}

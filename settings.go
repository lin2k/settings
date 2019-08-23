/// @Author winjeg,  winjeg@qq.com
/// All rights reserved to winjeg

package settings

import (
	"database/sql"
	"errors"
	"strings"
	"sync"
	"time"

	"github.com/sony/sonyflake"
	"github.com/winjeg/go-commons/log"
)

const (
	getSql                 = "SELECT value FROM settings WHERE name = ?"
	updateSql              = "UPDATE settings SET value = ? WHERE name = ?"
	existSql               = "SELECT COUNT(*) FROM settings WHERE name= ?"
	addSql                 = "INSERT IGNORE INTO settings(name, value) VALUE(?, ?)"
	addSqlWithId           = "INSERT IGNORE INTO settings(id, name, value) VALUE(?, ?, ?)"
	deleteVarSql           = "DELETE FROM settings WHERE name = ?"
	settingsSql            = `SELECT 1 FROM settings`
	settingVar             = "1"
	createSettingsTableSql = "CREATE TABLE `settings` ( `id` int(10) unsigned NOT NULL AUTO_INCREMENT COMMENT 'pk', `name` varchar(200) COLLATE utf8_bin NOT NULL COMMENT 'varname', `value` text COLLATE utf8_bin NOT NULL, PRIMARY KEY (`id`), UNIQUE KEY `name_UNIQUE` (`name`), KEY `idx_name` (`name`)) ENGINE=InnoDB AUTO_INCREMENT=1 DEFAULT CHARSET=utf8 COLLATE=utf8_bin"
	descSettingsSql        = "SELECT * FROM settings LIMIT 1"
	nameCol                = "name"
	valCol                 = "value"
)

var (
	settingsMap         = map[string]string{}
	logger              = log.GetLogger(nil)
	db          *sql.DB = nil
	withId              = false
)

// generate primary key, with this function
// if using databases that won't automatically generated primary key
// this function may suit you, but you must make the primary key at lease 8 byte long

func InitV2(dbConn *sql.DB, autoGenerateId bool) {
	err := Init(dbConn)
	if err != nil {
		return
	}
	withId = autoGenerateId
}

func Init(dbConn *sql.DB) error {
	if dbConn == nil {
		return errors.New("db connection should not be nil")
	}
	if db != nil {
		return errors.New("already initialized")
	}
	var lock sync.Mutex
	lock.Lock()
	defer lock.Unlock()

	// if table not exists create the table for them
	if !tableExists(dbConn) {
		// table not exists
		_, err := dbConn.Exec(createSettingsTableSql)
		if err != nil {
			return err
		}
	}
	rows, descError := dbConn.Query(descSettingsSql)
	if descError != nil {
		return descError
	}
	cols, _ := rows.Columns()
	if !contains(cols, nameCol, valCol) {
		return errors.New("table structure for settings is not supported")
	}
	db = dbConn
	return nil
}

func tableExists(dbConn *sql.DB) bool {
	row := dbConn.QueryRow(settingsSql)
	var re string
	err := row.Scan(&re)
	return err == nil && strings.EqualFold(re, settingVar)
}

func GetVar(name string) string {
	if v, ok := settingsMap[name]; ok {
		return v
	} else {
		var x string
		r := db.QueryRow(getSql, name)
		err := r.Scan(&x)
		if err != nil {
			return ""
		}
		var lock sync.Mutex
		lock.Lock()
		settingsMap[name] = x
		lock.Unlock()
		return x
	}
}

// set variable and update cache
func SetVar(name, value string) {
	var lock sync.Mutex
	lock.Lock()
	defer lock.Unlock()
	settingsMap[name] = value
	row := db.QueryRow(existSql, name)
	var exists int
	if err := row.Scan(&exists); err == nil && exists == 0 {
		var err error
		if withId {
			id := generateId()
			_, err = db.Exec(addSqlWithId, id, name, value)
		} else {
			_, err = db.Exec(addSql, name, value)
		}
		if err != nil {
			logger.Error(err)
		}
	} else {
		_, err = db.Exec(updateSql, value, name)
		if err != nil {
			logger.Error(err)
		}
	}
}

func DelVar(name string) {
	var lock sync.Mutex
	lock.Lock()
	defer lock.Unlock()
	delete(settingsMap, name)
	_, err := db.Exec(deleteVarSql, name)
	if err != nil {
		logger.Error(err)
	}
}

func contains(collection []string, elements ...string) bool {
	if len(elements) == 0 {
		return true
	}
	if len(collection) == len(elements) && len(collection) == 0 {
		return true
	}
	if len(elements) == 0 && len(collection) != 0 {
		return true
	}
	if len(collection) == 0 && len(elements) != 0 {
		return false
	}
	// put elements to map
	elementMap := make(map[string]bool, len(elements))
	for _, v := range elements {
		elementMap[v] = true
	}
	count := 0
	for _, v := range collection {
		if elementMap[v] {
			count++
		}
	}
	return count >= len(elementMap)
}

// below is the way to generate id for

var (
	timeStart, _  = time.Parse("2006-01-02 15:04:05", "2019-07-01 12:00:00")
	snowFlakeInst = sonyflake.NewSonyflake(sonyflake.Settings{
		StartTime:      timeStart, // start time
		MachineID:      nil,       // can be replaced, default is private ip address
		CheckMachineID: nil,       // the method to make sure the machine id is unique
	})
)

func generateId() uint64 {
	id, err := snowFlakeInst.NextID()
	if err != nil {
		logger.Error(err)
		return 0
	}
	return id
}

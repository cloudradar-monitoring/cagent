package mysql

import (
	"database/sql"
	"math"
	"strconv"
	"time"

	// import mysql driver to inject into database/sql
	_ "github.com/go-sql-driver/mysql"
)

type Status struct {
	Selects        int64
	Updates        int64
	Inserts        int64
	Deletes        int64
	Replaces       int64
	CallProcedures int64
	CacheHits      int64
	Commits        int64

	InnoDBReadBytes  int64
	InnoDBWriteBytes int64
	ReadBytes        int64
	WriteBytes       int64
}

func (s *Status) Queries() int64 {
	// according to mysql docs: https://dev.mysql.com/doc/refman/5.7/en/query-cache-status-and-maintenance.html
	//  number of select queries = Com_select + Qcache_hits
	return s.Selects + s.Updates + s.Inserts + s.Deletes + s.Replaces + s.CallProcedures + s.CacheHits
}

func getStatus(db *sql.DB) (*Status, error) {
	rows, err := db.Query(`SHOW GLOBAL STATUS WHERE Variable_name IN (
'Com_select', 'Com_insert', 'Com_update', 'Com_delete', 'Com_replace', 'Com_call_procedure', 
'Qcache_hits', 'Com_commit', 'Innodb_data_read', 'Innodb_data_write', 'Bytes_received', 'Bytes_sent')`)
	if err != nil {
		return nil, err
	}

	var total Status
	for rows.Next() {
		var key, val string
		err = rows.Scan(&key, &val)
		if err != nil {
			log.Error(err.Error())
			continue
		}

		valInt, err := strconv.ParseInt(val, 10, 64)
		if err != nil {
			log.Errorf("failed to convert %s to int: %s", key, err.Error())
			break
		}

		switch key {
		case "Com_select":
			total.Selects = valInt
		case "Com_update":
			total.Updates = valInt
		case "Com_insert":
			total.Inserts = valInt
		case "Com_delete":
			total.Deletes = valInt
		case "Com_replace":
			total.Replaces = valInt
		case "Com_call_procedure":
			total.CallProcedures = valInt
		case "Qcache_hits":
			total.CacheHits = valInt
		case "Com_commit":
			total.Commits = valInt
		case "Innodb_data_read":
			total.InnoDBReadBytes = valInt
		case "Innodb_data_write":
			total.InnoDBWriteBytes = valInt
		case "Bytes_received":
			total.ReadBytes = valInt
		case "Bytes_sent":
			total.WriteBytes = valInt
		}
	}

	return &total, nil
}

func fillResultsPerSecond(new *Status, old *Status, durationBetween time.Duration, result map[string]interface{}) {
	sec := durationBetween.Seconds()
	result["Selects per sec"] = roundUp(onlyPositive(float64(new.Selects-old.Selects) / sec))
	result["Updates per sec"] = roundUp(onlyPositive(float64(new.Updates-old.Updates) / sec))
	result["Inserts per sec"] = roundUp(onlyPositive(float64(new.Inserts-old.Inserts) / sec))
	result["Deletes per sec"] = roundUp(onlyPositive(float64(new.Deletes-old.Deletes) / sec))
	result["Commits per sec"] = roundUp(onlyPositive(float64(new.Commits-old.Commits) / sec))
	result["Innodb data read bps"] = roundUp(onlyPositive(float64(new.InnoDBReadBytes-old.InnoDBReadBytes) / sec))
	result["Innodb data write bps"] = roundUp(onlyPositive(float64(new.InnoDBWriteBytes-old.InnoDBWriteBytes) / sec))
	result["Bytes read bps"] = roundUp(onlyPositive(float64(new.ReadBytes-old.ReadBytes) / sec))
	result["Bytes write bps"] = roundUp(onlyPositive(float64(new.WriteBytes-old.WriteBytes) / sec))
	result["Queries per sec"] = roundUp(onlyPositive(float64(new.Queries()-old.Queries()) / sec))
}

func onlyPositive(f float64) float64 {
	if f < 0 {
		return 0
	}
	return f
}

func roundUp(p float64) float64 {
	k := math.Pow10(2)
	return float64(int64(p*k+0.5)) / k
}

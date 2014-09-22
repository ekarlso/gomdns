package db

import (
	log "code.google.com/p/log4go"
	"database/sql"
	config "github.com/ekarlso/gomdns/config"
)

// Designate Zone
type Zone struct {
	Id      string `db:"id"`
	Version int
	Name    string
	Email   string
	Ttl     uint32
	Serial  int
	Refresh int
	Retry   int
	Expire  int
	Minimum int
}

type RecordSet struct {
	Id       string
	DomainId string `db:"domain_id"`
	Name     string
	Type     string
	Ttl      sql.NullInt64
	Records  []Record
}

type Record struct {
	Id          string
	DomainId    string `db:"domain_id"`
	RecordSetId string `db:"recordset_id"`
	Data        string
	Priority    sql.NullInt64
	Hash        string
}

func GetZoneById(zoneId string) (error, Zone) {
	cfg := config.GetConfig()
	conn := Connect(cfg.StorageDSN)

	var z Zone
	err := conn.Get(&z, "SELECT id, version, name, email, ttl, serial, refresh, retry, expire, minimum FROM domains WHERE id = ?", zoneId)
	return err, z
}

func GetZoneByName(zoneName string) (error, Zone) {
	cfg := config.GetConfig()
	conn := Connect(cfg.StorageDSN)

	log.Info("Fetching domain by name %s", zoneName)

	var z Zone
	err := conn.Get(&z, "SELECT id, version, name, email, ttl, serial, refresh, retry, expire, minimum FROM domains WHERE name = ?", zoneName)

	return err, z
}

func GetRecordSet(zoneId string, recordType string) (rs RecordSet, err error) {
	cfg := config.GetConfig()
	conn := Connect(cfg.StorageDSN)

	err = conn.Get(&rs, "SELECT id, domain_id, name, type, ttl FROM recordsets WHERE domain_id = ? AND type = ?", zoneId, recordType)
	if err != nil {
		return RecordSet{}, err
	}

	records := []Record{}
	err = conn.Select(&records, "SELECT id, domain_id, recordset_id, data, priority, hash FROM records WHERE recordset_id = ?", rs.Id)
	if err != nil {
		log.Error("Error fetching records for RRset %s", err)
	}
	rs.Records = records
	return rs, nil
}

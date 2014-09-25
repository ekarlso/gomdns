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

func GetZoneById(zoneId string) (z Zone, err error) {
	cfg := config.GetConfig()
	conn, _ := Connect(cfg.StorageDSN)

	defer conn.Close()

	err = conn.Get(&z, "SELECT id, version, name, email, ttl, serial, refresh, retry, expire, minimum FROM domains WHERE id = ?", zoneId)
	if err != nil {
		log.Debug("Failed getting zone")
	}

	return z, err
}

func GetZoneByName(zoneName string) (z Zone, err error) {
	cfg := config.GetConfig()
	conn, _ := Connect(cfg.StorageDSN)

	defer conn.Close()

	err = conn.Get(&z, "SELECT id, version, name, email, ttl, serial, refresh, retry, expire, minimum FROM domains WHERE name = ?", zoneName)
	if err != nil {
		log.Debug("Failed getting zone")
	}

	return z, err
}

func GetRecordSet(rrName string, rrType string) (rrSet RecordSet, err error) {
	cfg := config.GetConfig()
	conn, _ := Connect(cfg.StorageDSN)

	defer conn.Close()

	err = conn.Get(&rrSet, "SELECT id, domain_id, name, type, ttl from recordsets WHERE name = ?", rrName)
	if err != nil {
		log.Debug("Failed getting RRSet")
		return rrSet, err
	}

	records := []Record{}

	err = conn.Select(&records, "SELECT id, domain_id, recordset_id, data, priority, hash FROM records WHERE recordset_id = ?", rrSet.Id)
	if err != nil {
		log.Error("Error fetching records for RRset %s", err)
	}
	rrSet.Records = records
	return rrSet, err
}

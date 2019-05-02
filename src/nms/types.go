package nms

import (
	"database/sql"
)

//使用IP版本
type IPVersion int

const (
	//_     IPVersion = 0
	IPV_4 IPVersion = 4
	IPV_6 IPVersion = 6
	//IPV_ALL           = 3
)

//保存模式
type SAVEMODE int

const (
	//每次更新
	UPDATE SAVEMODE = 1
	//每次新增
	NEW SAVEMODE = 2
)

//采集版本
type SNMPVersion int

const (
	SNMP_V1 SNMPVersion = 1
	SNMP_V2 SNMPVersion = 2
	SNMP_V3 SNMPVersion = 3
)

//采集信息
type GInfo struct {
	//TargetKey string
	OID   string
	Val   string
	Type  string
	GTime int64
}

type KeyVal struct {
	Name string
	OID  string
}

type RowHandler func(row []string)

type RowParser func(row *sql.Rows)

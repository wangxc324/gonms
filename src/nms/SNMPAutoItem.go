package nms

import (
	"encoding/json"
)

//自动运行的采集项
type SNMPAutoItem struct {
	//配置项名称（全局唯一）
	Name string
	//索引oid，如果没有索引则填0
	Index string
	//采集的oid列表{name:oid}
	OIDS []KeyVal
	//采集间隔分钟数
	Span int64
	//方法，walk|bulkwalk|getnext
	Method string
	//是否本地存储
	LocalSave bool
	//本地保存方式
	LocalSaveModel SAVEMODE
	//UDP转发配置
	UDPForwards []string
	//描述
	Descr string
	//适用于
	SuitSysoid []string
	//是否开启采集
	Activate bool
}

//设置配置项
func (i SNMPAutoItem) Set() error {
	ldb := NewLDB("ldb/conf")
	err := ldb.Set([]byte(i.Name), i)
	return err
}

//删除指定配置项
func DelAutoConf(name string) error {
	ldb := NewLDB("ldb/conf")
	err := ldb.Del([]byte(name))
	return err
}

//获取所有配置项
func GetAutoConfs() map[string]SNMPAutoItem {
	ldb := NewLDB("ldb/conf")
	dict := ldb.GetAll()
	results := make(map[string]SNMPAutoItem)
	for k, v := range dict {
		var c SNMPAutoItem
		err := json.Unmarshal([]byte(v), &c)
		if err == nil {
			results[k] = c
		}
	}
	delete(results, "")
	return results
}

//初始化配置文件
func InitAutoRunConf() {
	//DelAutoConf("")
	m := GetAutoConfs()
	data := g_conf.Get("autosnmp")
	var dict map[string]SNMPAutoItem
	g_conf.To(data, &dict)
	if len(m) == 0 {
		for _, v := range dict {
			v.Set()
		}
	}
}

package main

import (
	"fmt"
	//"fmt"
	"nms"
	"time"
)

func main() {
	nms.InitConf()
	nms.InitAutoRunConf()

	fmt.Println("NMS STARTING...", time.Now())

	// var i = nms.SNMPAutoItem{
	// 	Name:        "hrStorageEntry",
	// 	Index:       ".1.3.6.1.2.1.25.2.3.1.1",
	// 	OIDS:        []string{".1.3.6.1.2.1.25.2.3.1.2", ".1.3.6.1.2.1.25.2.3.1.3", ".1.3.6.1.2.1.25.2.3.1.4", ".1.3.6.1.2.1.25.2.3.1.5", ".1.3.6.1.2.1.25.2.3.1.6", ".1.3.6.1.2.1.25.2.3.1.7"},
	// 	Headers:     []string{"hrStorageType", "hrStorageDescr", "hrStorageAllocationUnits", "hrStorageSize", "hrStorageUsed", "hrStorageAllocationFailures"},
	// 	Span:        1,
	// 	Method:      "bulkwalk",
	// 	LocalSave:   true,
	// 	UDPForwards: []string{"127.0.0.1:1234", "172.0.0.76:1234"},
	// 	Descr:       "采集主机信息",
	// 	SuitSysoid:  []string{".1.3.6.1.2.1"},
	// }
	// fmt.Println(i)
	// i.Set()
	// nms.DelAutoConf("")
	// l2 := nms.GetAutoConfs()
	// fmt.Println(l2)

	// t := nms.Target{
	// 	IPV:        nms.IPV_4,
	// 	IPV4:       "127.0.0.1",
	// 	SNMPV:      nms.SNMP_V2,
	// 	Community:  "public",
	// 	Activate:   true,
	// 	CREATETIME: time.Now(),
	// }
	// err := t.Reg()

	// fmt.Println(err)

	// l := nms.GetTargets()
	// fmt.Println(l)

	go nms.Start()
	for {

		nms.AutoRun()
		time.Sleep(60 * 1000 * time.Millisecond)
	}

	// nms.Start()

}

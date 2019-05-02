package nms

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/soniah/gosnmp"
)

// const SERVER_PORT = 8080
// const SERVER_DOMAIN = "localhost"
// const RESPONSE_TEMPPATE = "WANGXC"

func HttpApiHander(w http.ResponseWriter, r *http.Request) {
	vars := r.URL.Query()
	for i, v := range vars {
		fmt.Fprintf(w, "%s=%s;", i, v)
	}
}

//配置接口
func httpApiConfSetHander(w http.ResponseWriter, r *http.Request) {
	vars := r.URL.Query()
	c := SNMPAutoItem{}
	c.Activate = (vars["activate"][0] == "true")
	c.Descr = vars["descr"][0]
	c.Index = vars["index"][0]
	c.LocalSave = (vars["local"][0] == "true")
	c.Method = vars["method"][0]
	c.Name = vars["name"][0]
	c.OIDS = make([]KeyVal, 0) //make(map[string]string)
	oids := strings.Split(vars["oids"][0], ",")
	for _, v := range oids {
		kv := strings.Split(v, ":")
		i := KeyVal{
			Name: kv[0],
			OID:  kv[1],
		}
		c.OIDS = append(c.OIDS, i)
	}
	span, err := strconv.ParseInt(vars["span"][0], 10, 64)
	if err == nil {
		c.Span = span
		if len(vars["suit"][0]) > 0 {
			c.SuitSysoid = strings.Split(vars["suit"][0], ",")
		} else {
			c.SuitSysoid = make([]string, 0)
		}
		if len(vars["forwards"][0]) > 0 {
			c.UDPForwards = strings.Split(vars["forwards"][0], ",")
		} else {
			c.UDPForwards = make([]string, 0)
		}
		err := c.Set()
		if err == nil {
			fmt.Fprint(w, "true")
		} else {
			fmt.Fprint(w, "false")
		}
	} else {
		fmt.Fprint(w, "false")
	}
}

//设备注册接口
func httpApiRegHander(w http.ResponseWriter, r *http.Request) {
	t := Target{}
	vars := r.URL.Query()
	ipv, err := strconv.Atoi(vars["ipv"][0])
	if err == nil {
		t.IPV = IPVersion(ipv)
		if t.IPV == IPV_4 {
			t.IPV4 = vars["ip"][0]
		} else {
			t.IPV6 = vars["ip"][0]
		}
	}
	snmpv, err2 := strconv.Atoi(vars["snmpv"][0])
	if err2 == nil {
		t.SNMPV = SNMPVersion(snmpv)
		if t.SNMPV == SNMP_V3 {
			t.USERNAME = vars["user"][0]
			level, _ := strconv.Atoi(vars["level"][0])
			t.LEVEL = gosnmp.SnmpV3MsgFlags(level)
			t.AUTHPASSPHRASE = vars["auth"][0]
			ap, _ := strconv.Atoi(vars["authp"][0])
			t.AUTHPROTOCAL = gosnmp.SnmpV3AuthProtocol(ap)

			pp, _ := strconv.Atoi(vars["privp"][0])
			t.PRIVACYPROTOCAL = gosnmp.SnmpV3PrivProtocol(pp)
			t.PRIVACYPASSPHRASE = vars["priv"][0]
		} else {
			t.Community = vars["c"][0]
		}
	}

	t.CREATETIME = time.Now()
	t.Key = t.GetKey()
	t.Activate = (vars["activate"][0] == "true")

	err3 := t.Reg()
	if err3 == nil {
		fmt.Fprint(w, "true")
	} else {
		fmt.Fprint(w, "false")
	}

}

//取得设备清单的接口
func httpApiTargetsHander(w http.ResponseWriter, r *http.Request) {
	l := GetTargets()
	str, _ := json.Marshal(l)
	fmt.Fprint(w, string(str))
}

//SNMP工具接口
func httpApiSNMPHander(w http.ResponseWriter, r *http.Request) {
	vars := r.URL.Query()
	t := Target{}
	ipv, err := strconv.Atoi(vars["ipv"][0])
	if err == nil {
		t.IPV = IPVersion(ipv)
		if t.IPV == IPV_4 {
			t.IPV4 = vars["ip"][0]
		} else {
			t.IPV6 = vars["ip"][0]
		}
	}
	snmpv, err2 := strconv.Atoi(vars["snmpv"][0])
	if err2 == nil {
		t.SNMPV = SNMPVersion(snmpv)
		if t.SNMPV == SNMP_V3 {
			t.USERNAME = vars["user"][0]
			level, _ := strconv.Atoi(vars["level"][0])
			t.LEVEL = gosnmp.SnmpV3MsgFlags(level)
			t.AUTHPASSPHRASE = vars["auth"][0]
			ap, _ := strconv.Atoi(vars["authp"][0])
			t.AUTHPROTOCAL = gosnmp.SnmpV3AuthProtocol(ap)

			pp, _ := strconv.Atoi(vars["privp"][0])
			t.PRIVACYPROTOCAL = gosnmp.SnmpV3PrivProtocol(pp)
			t.PRIVACYPASSPHRASE = vars["priv"][0]
		} else {
			t.Community = vars["c"][0]
		}
	}

	t.CREATETIME = time.Now()
	t.Key = t.IPV4

	fmt.Println("from:", r.RemoteAddr, time.Now())
	fmt.Println("snmp:", t.ToJson())

	data, err := t.Snmp(vars["op"][0], vars["oid"][0])
	if err != nil {
		fmt.Fprintln(w, err)

	} else {
		js, _ := json.Marshal(data)
		fmt.Fprintln(w, string(js))
	}
}
func Start() {
	http.HandleFunc("/api/index", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintln(w, "当前时间：/api/now")
		fmt.Fprintln(w, "")
		fmt.Fprintln(w, "当前时间戳：/api/timestamp")
		fmt.Fprintln(w, "")
		fmt.Fprintln(w, "配置字典:/api/conf/dict")
		fmt.Fprintln(w, "")
		fmt.Fprintln(w, "配置清单:/api/conf/set?name=&span=1&suit=&index=&oids=&local=true&method=bulkwalk&headers=&activate=&forwards=&descr=")
		fmt.Fprintln(w, "")
		fmt.Fprintln(w, "删除配置:/api/conf/del?key=")
		fmt.Fprintln(w, "")
		fmt.Fprintln(w, "设备清单:/api/targets/list")
		fmt.Fprintln(w, "")
		fmt.Fprintln(w, "读取数据:/api/targets/read?targetkey=IPV4_172.0.0.6&indexes=1,3&s=1555844867&e=1555931267&name=hrStorageEntry&fileds=INDEX,WALKTIME,hrStorageUsed,hrStorageAllocationUnits")
		fmt.Fprintln(w, "")
		fmt.Fprintln(w, "注册设备（v2）:/api/targets/reg?ipv=4&snmpv=2&c=fengyang&ip=172.0.0.1&activate=true")
		fmt.Fprintln(w, "ipv:4,6")
		fmt.Fprintln(w, "")
		fmt.Fprintln(w, "注册设备（v3）:/api/targets/reg?ipv=4&snmpv=3&c=fengyang&ip=172.0.0.1&activate=true&user=test&level=3&auth=2&authp=test&priv=2&privp=test")
		fmt.Fprintln(w, "ipv:4,6")
		fmt.Fprintln(w, "level:0(NoAuthNoPriv),1(AuthNoPriv),3(AuthPriv),4(Reportable)")
		fmt.Fprintln(w, "auth:1(NoAuth),2(MD5),3(SHA)")
		fmt.Fprintln(w, "priv:1(NoPriv),2(DES),3(AES)")

		fmt.Fprintln(w, "")
		fmt.Fprintln(w, "删除设备:/api/targets/unreg?key=IPV4_172.0.0.1")
		fmt.Fprintln(w, "")
		fmt.Fprintln(w, "采集接口：/api/snmp?op=&ipv=4&ip=&snmpv=&c=&user=&level=&authp=&auth=&privp=&priv=&oid=")

	})
	http.HandleFunc("/api/now", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, time.Now().Format("2006-01-02 15:04:05"))
	})
	http.HandleFunc("/api/timestamp", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, time.Now().Unix())
	})
	http.HandleFunc("/api/conf/dict", func(w http.ResponseWriter, r *http.Request) {
		dict := GetAutoConfs()
		js, _ := json.Marshal(dict)
		fmt.Fprintln(w, string(js))
		//fmt.Fprint(w, time.Now().Unix())
	})
	http.HandleFunc("/api/conf/del", func(w http.ResponseWriter, r *http.Request) {
		vars := r.URL.Query()
		key := vars["key"][0]
		err := DelAutoConf(key)
		if err == nil {
			fmt.Fprintf(w, "true")
		} else {
			fmt.Fprint(w, "false")
		}
	})
	http.HandleFunc("/api/conf/set", httpApiConfSetHander)
	http.HandleFunc("/api/targets/read", func(w http.ResponseWriter, r *http.Request) {
		///api/targets/read?targetkey=IPV4_172.0.0.6&name=&indexes=2&s=&e=
		vars := r.URL.Query()
		key := vars["targetkey"][0]
		indexes := vars["indexes"][0]
		s, _ := strconv.ParseInt(vars["s"][0], 10, 64)
		e, _ := strconv.ParseInt(vars["e"][0], 10, 64)
		n := vars["name"][0]
		index := strings.Split(indexes, ",")
		var t Target
		err := GetTarget(key, &t)
		if err == nil {
			result := t.ReadData(n, index, s, e, strings.Split(vars["fileds"][0], ","))
			js, err := json.Marshal(result)
			if err == nil {
				fmt.Fprintf(w, string(js))
			} else {
				fmt.Fprint(w, "[]")
			}
		} else {
			fmt.Fprint(w, "设备不存在")
		}
	})
	http.HandleFunc("/api/targets/reg", httpApiRegHander)
	http.HandleFunc("/api/targets/unreg", func(w http.ResponseWriter, r *http.Request) {
		vars := r.URL.Query()

		key := vars["key"][0]

		err := Unreg(key)
		if err == nil {
			fmt.Fprint(w, "true")
		} else {
			fmt.Fprint(w, "false")
		}

	})
	http.HandleFunc("/api/targets/list", httpApiTargetsHander)
	http.HandleFunc("/api/snmp", httpApiSNMPHander)
	conf := Conf{}
	conf.Load()
	p := conf.Get("hserver", "host")
	host := p.(string)
	fmt.Println("STARTED:", host)
	http.ListenAndServe(host, nil)

}

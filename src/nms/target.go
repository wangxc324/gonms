package nms

import (
	"bytes"
	"encoding/binary"

	//	"encoding/binary"
	"encoding/hex"

	//"unicode/utf8"

	//"reflect"
	// "database/sql"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"net"
	"os"
	"strconv"
	"strings"

	// "fmt"

	// "strconv"

	//	"fmt"
	"log"
	"time"

	g "github.com/soniah/gosnmp"
)

//网络设备
type Target struct {
	//唯一标识
	Key string
	//IP版本
	IPV IPVersion
	//IPV4
	IPV4 string
	//IPV6
	IPV6 string
	//SNMP版本
	SNMPV SNMPVersion
	//Community 仅对V2生效
	Community string
	//V3安全等级
	LEVEL g.SnmpV3MsgFlags
	//V3用户名
	USERNAME string
	//V3认证方式
	AUTHPROTOCAL g.SnmpV3AuthProtocol
	//V3认证密码
	AUTHPASSPHRASE string
	//V3加密方式
	PRIVACYPROTOCAL g.SnmpV3PrivProtocol
	//V3加密密钥
	PRIVACYPASSPHRASE string
	//创建时间
	CREATETIME time.Time
	//是否开启采集
	Activate bool
	//要自动运行的采集项
	AutoRunConfs []string
}

//转JSON字符串
func (target Target) ToJson() string {

	js, _ := json.Marshal(target)
	return string(js)
}

//获取设备的唯一标识
func (target Target) GetKey() string {
	if target.IPV == IPV_4 {
		return "IPV4_" + target.IPV4
	} else {
		return "IPV6_" + target.IPV6
	}
}

//执行Walk的回调
func (target Target) ToGInfo(pdu g.SnmpPDU) GInfo {
	info := GInfo{}
	info.GTime = time.Now().Unix()
	info.OID = pdu.Name
	info.Type = pdu.Type.String()

	hexstr := make(map[string]int)
	//物理地址
	hexstr[".1.3.6.1.2.1.3.1.1.2."] = 1
	hexstr[".1.3.6.1.2.1.4.22.1.2."] = 1
	hexstr[".1.3.6.1.2.1.4.35.1.4."] = 1
	hexstr[".1.3.6.1.2.1.2.2.1.6."] = 1

	//时间
	hexstr[".1.3.6.1.2.1.25.1.2."] = 2
	hexstr[".1.3.6.1.2.1.25.6.3.1.5."] = 2
	hexstr[".1.3.6.1.2.1.25.3.8.1.8."] = 2
	hexstr[".1.3.6.1.2.1.25.3.8.1.9."] = 2

	switch pdu.Type {
	case g.OctetString:
		b := pdu.Value.([]byte)
		ishex := false
		var hextype int = 0
		for k, v := range hexstr {
			if strings.Index(info.OID, k) == 0 {
				ishex = true
				hextype = v
				break
			}
		}
		if ishex {
			if hextype == 1 {
				info.Type += "(Hex-MAC)"
				info.Val = hex.EncodeToString(b)
			} else if hextype == 2 {
				y := b[:2]
				var bf = bytes.NewBuffer(y)
				var year int16
				binary.Read(bf, binary.BigEndian, &year)
				info.Val = strconv.FormatInt(int64(year), 10) + "-" + strconv.Itoa(int(b[2])) + "-" + strconv.Itoa(int(b[3])) + " " + strconv.Itoa(int(b[4])) + ":" + strconv.Itoa(int(b[5])) + ":" + strconv.Itoa(int(b[6])) //+ "." + strconv.Itoa(int(b[7]))
				info.Type += "(Hex-TIME)"

			} else {
				info.Val = hex.EncodeToString(b)
				info.Type += "(Hex-STRING)"
			}
		} else {
			info.Val = string(b)
		}
	case g.BitString:
		info.Val = hex.EncodeToString(pdu.Value.([]byte))
	case g.IPAddress:
		info.Val = pdu.Value.(string)

	default:
		info.Val = g.ToBigInt(pdu.Value).String()
	}
	return info
}

//注册采集
func (target Target) Reg() error {

	ldb := NewLDB("ldb/targets")
	err := ldb.Set([]byte(target.GetKey()), target)
	return err
}

//执行snmp请求
func (target Target) Snmp(op string, oid string) (data []GInfo, err error) {
	var kv []GInfo = make([]GInfo, 0)

	var p *g.GoSNMP //g.Default
	if target.SNMPV == SNMP_V3 {
		p = &g.GoSNMP{
			Version:       g.Version3,
			Timeout:       time.Duration(30) * time.Second,
			SecurityModel: g.UserSecurityModel,
			MsgFlags:      target.LEVEL,
			SecurityParameters: &g.UsmSecurityParameters{
				UserName:                 target.USERNAME,
				AuthenticationProtocol:   target.AUTHPROTOCAL,
				AuthenticationPassphrase: target.AUTHPASSPHRASE,
				PrivacyProtocol:          target.PRIVACYPROTOCAL,
				PrivacyPassphrase:        target.PRIVACYPASSPHRASE,
			},
		}
	} else {
		p = &g.GoSNMP{
			Community: target.Community,
			Version:   g.Version2c,
		}
	}

	if target.IPV == IPV_4 {
		p.Target = target.IPV4
	} else {
		p.Target = target.IPV6
	}
	p.Port = 161
	p.Timeout = time.Duration(10) * time.Second
	e := p.Connect()

	if e != nil {
		log.Fatalf("Connect() err: %v", e)
	}

	defer p.Conn.Close()

	switch op {
	case "walk":
		f := func(pdu g.SnmpPDU) error {
			info := target.ToGInfo(pdu)
			kv = append(kv, info)
			return nil
		}
		p.Walk(oid, f)
	case "bulkwalk":
		f := func(pdu g.SnmpPDU) error {
			info := target.ToGInfo(pdu)
			kv = append(kv, info)
			return nil
		}
		p.BulkWalk(oid, f)
	case "get":
		oids := []string{oid}
		result, err := p.Get(oids)
		if err == nil {
			for _, variable := range result.Variables {
				info := target.ToGInfo(variable)
				kv = append(kv, info)
			}
		}

	case "getnext":
		oids := []string{oid}
		result, err := p.GetNext(oids)
		if err == nil {
			for _, variable := range result.Variables {
				info := target.ToGInfo(variable)
				kv = append(kv, info)
			}
		}
	// case "BULK":
	// 	data,err = p.GetBulk(oid)
	default:

	}

	return kv, nil
}
func (target Target) SnmpDict(op string, oid string) (dict map[string]GInfo, err error) {
	data := make(map[string]GInfo)
	l, err := target.Snmp(op, oid)
	if err == nil {
		for _, v := range l {
			data[v.OID] = v
		}
	}
	return data, err
}

//执行get
func (target Target) Get(oid string) (data []GInfo, err error) {
	return target.Snmp("get", oid)
}

//执行getnext
func (target Target) GetNext(oid string) (data []GInfo, err error) {
	return target.Snmp("getnext", oid)
}

//执行walk
func (target Target) Walk(oid string) (data []GInfo, err error) {
	return target.Snmp("bulkwalk", oid)
}

//根据配置文件自动执行采集
func (target Target) Run(i SNMPAutoItem) (currenttime int64, result map[string][]GInfo, err error) {
	var dict map[string][]GInfo
	dict = make(map[string][]GInfo)
	var minutes int64 = time.Now().Unix() / 60
	// fmt.Println("minutes:", minutes)

	if minutes%i.Span == 0 {
		if i.Index == "0" {
			data_i, err := target.Snmp(i.Method, i.OIDS[0].OID)
			if err == nil {
				dict["0"] = data_i
			}
		} else {

			il, err := target.Snmp(i.Method, i.Index)
			if err == nil {
				for _, oid := range i.OIDS {
					data_i, err := target.SnmpDict(i.Method, oid.OID)
					if err == nil {
						for _, index := range il {
							key := (oid.OID + "." + index.Val)
							value, ok := dict[index.Val]
							var gl []GInfo
							if ok == false {
								gl = make([]GInfo, 0)
								dict[index.Val] = gl
							} else {
								gl = value
							}
							ov, ok := data_i[key]
							if ok {
								gl = append(gl, ov)
								dict[index.Val] = gl
							}
						}
					}
				}
			}
		}
	}

	return minutes * 60, dict, err
}

//获取文件名(并确保目录存在)
func (target Target) GetFileName(typename string, timestamp int64) string {
	dir := g_conf.Get("root").(string) + "src/" + target.GetKey() + "/" + typename + "/"
	fi1, err := os.Stat(dir)
	if err != nil {
		os.MkdirAll(dir, os.ModePerm)
	} else {
		if fi1.IsDir() == false {
			fmt.Println("目录重复！")
		}
	}
	fn := dir + strconv.FormatInt((3600*24)*(timestamp/(3600*24)), 10) + ".csv"
	return fn
}

//运行采集配置
func (target Target) RunConf(node string, dict map[string]SNMPAutoItem) {
	for k, v := range dict {
		if v.Activate {
			v.Name = k
			t, data, err := target.Run(v)
			if err == nil {
				target.runConfDataHandler(node, t, v, data)

			}
			fmt.Println(time.Now(), "RUN:", target.GetKey(), v.Name)
		}
	}
}

//采集的数据转换位字符串数组
func (target Target) runConfDataHandler(node string, timestamp int64, conf SNMPAutoItem, data map[string][]GInfo) {
	var w *csv.Writer
	if conf.LocalSave {
		fn := target.GetFileName(conf.Name, timestamp)
		fi, err := os.Stat(fn)
		if err == nil {
			if fi.IsDir() {
				fmt.Println("文件已经存在！")
			}
		}
		var f *os.File
		var isfirst = false
		if err != nil {
			f, err = os.Create(fn)
			isfirst = true
		} else {
			f, err = os.OpenFile(fn, os.O_WRONLY|os.O_APPEND, os.ModeAppend)
		}
		defer f.Close()

		w = csv.NewWriter(f)
		if isfirst {
			var h []string = []string{"INDEX", "WALKTIME"}
			for _, hn := range conf.OIDS {
				h = append(h, hn.Name, hn.Name+"_TIME")
			}
			w.Write(h)
		}
	}

	var conns []net.Conn

	for _, udp := range conf.UDPForwards {

		conn, err := net.Dial("udp", udp)
		if err == nil {
			conns = append(conns, conn)
		}
		defer conn.Close()
	}

	for k, v := range data {
		row := make([]string, 0)
		row = append(row, k, strconv.FormatInt(timestamp, 10))
		for _, g := range v {
			row = append(row, g.Val, strconv.FormatInt(g.GTime, 10))
		}

		if conf.LocalSave {
			w.Write(row)
		}
		newline := "#ROW," + node + "," + target.GetKey() + "," + conf.Name + "," + strings.Join(row, ",") + ",#ROWEND"
		for _, conn := range conns {
			conn.Write([]byte(newline))
		}
	}
	if conf.LocalSave {
		w.Flush()
	}
}

//读取采集数据
func (target Target) ReadData(typename string, index []string, start int64, end int64, fileds []string) interface{} {
	dict := make(map[string]bool)
	for _, v := range index {
		dict[v] = true
	}

	dict_fileds := make(map[string]bool)
	has_fileds := false
	for _, v := range fileds {
		if len(v) > 0 {
			dict_fileds[v] = true
			has_fileds = true
		}
	}

	mperday := int64(3600 * 24)
	s := mperday * (start / mperday)
	e := mperday * (end / mperday)
	//result := make(map[string]interface)
	var result []interface{} = make([]interface{}, 0)
	for i := s; i <= e; i += mperday {
		fn := target.GetFileName(typename, i)
		fi, err := os.Stat(fn)
		if err == nil {
			if fi.IsDir() == false {
				var f *os.File
				//var isfirst = false
				f, err = os.Open(fn)
				defer f.Close()
				if err == nil {
					r := csv.NewReader(f)
					header, err := r.Read()
					if err == nil {
						for {
							row, err := r.Read()
							if err != nil {
								break
							}
							wtime, err := strconv.ParseInt(row[1], 10, 64) //strconv.Atoi(row[1])
							if err == nil {
								o := make(map[string]string)
								if dict[row[0]] == true && wtime >= start && wtime < end {
									for i, v := range row {

										if has_fileds {
											if dict_fileds[header[i]] == true {
												o[header[i]] = v
											}
										} else {

											o[header[i]] = v
										}
									}
									result = append(result, o)
								}
							}
						}
					}
				}
			}
		}
	}
	return result
}

//获取所有设备
func GetTargets() (targets []Target) {
	ldb := NewLDB("ldb/targets")
	dict := ldb.GetAll()
	var l []Target = make([]Target, 0)
	for _, v := range dict {
		var t Target
		json.Unmarshal([]byte(v), &t)
		l = append(l, t)
	}
	return l
}

//读取指定设备
func GetTarget(key string, target *Target) error {
	ldb := NewLDB("ldb/targets")
	data, err := ldb.Get([]byte(key))
	json.Unmarshal(data, target)
	return err
}
func Unreg(key string) error {
	ldb := NewLDB("ldb/targets")
	err := ldb.Del([]byte(key))
	return err
}
func AutoRun() {
	c := Conf{}
	c.Load()
	var dict map[string]SNMPAutoItem
	dict = GetAutoConfs()
	targets := GetTargets()
	for _, t := range targets {
		if t.Activate {
			go t.RunConf(c.Get("node").(string), dict)
		}
	}
}

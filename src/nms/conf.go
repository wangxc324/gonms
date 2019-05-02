// conf.go
package nms

import (
	"encoding/json"
	// "fmt"
	"io/ioutil"
)

type Conf struct {
	conf_data interface{}
	dict      map[string]interface{}
}

//加载配置文件
func (conf *Conf) Load(filename ...string) error {
	//下面使用的是相对路径，config.json文件和main.go文件处于同一目录下
	//ReadFile函数会读取文件的全部内容，并将结果以[]byte类型返回
	data, err := ioutil.ReadFile("./config.json")
	if err != nil {
		return err
	}
	// jsonstr := string(data)
	// fmt.Println(jsonstr)
	//读取的数据为json格式，需要进行解码
	err = json.Unmarshal(data, &conf.conf_data)
	if err != nil {
		return err
	} else {
		conf.dict, _ = conf.conf_data.(map[string]interface{})
		return nil
	}
}

/*
	读取json值，支持按路径读取
*/
func (conf *Conf) Get(keys ...string) interface{} {
	var d interface{} = conf.dict
	for _, k := range keys {
		d = d.(map[string]interface{})[k]
	}
	return d
}

//读取的项进行类型转换
func (conf *Conf) To(src interface{}, dist interface{}) {
	js, _ := json.Marshal(src)
	json.Unmarshal(js, dist)
}

var g_conf Conf

//初始化配置文件
func InitConf() {
	g_conf = Conf{}
	g_conf.Load()
}

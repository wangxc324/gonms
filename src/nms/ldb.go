package nms

import (
	// "bytes"
	// "encoding/binary"
	"encoding/json"
	"fmt"

	"github.com/syndtr/goleveldb/leveldb"
)

type LDB struct {
	Path string
}

func (ldb LDB) Set(key []byte, val interface{}) error {

	db, err := leveldb.OpenFile(ldb.Path, nil)
	defer db.Close()
	if err != nil {
		fmt.Println("LDB SET ERROR:", err)
	}
	data, _ := Encode(val)
	err = db.Put(key, data, nil)
	return err
}
func (ldb LDB) Del(key []byte) error {
	db, err := leveldb.OpenFile(ldb.Path, nil)
	defer db.Close()
	if err != nil {
		fmt.Println("LDB Del ERROR:", err)
	} else {
		err = db.Delete(key, nil)
	}
	return err
}

func (ldb LDB) Get(key []byte) ([]byte, error) {
	db, err := leveldb.OpenFile(ldb.Path, nil)
	defer db.Close()
	data, err := db.Get(key, nil)
	// if err == nil {
	// 	Decode(data, result)
	// }
	return data, err
}

func (ldb LDB) GetAll() map[string]string {
	dict := make(map[string]string)
	db, err := leveldb.OpenFile(ldb.Path, nil)
	defer db.Close()
	if err == nil {

		iter := db.NewIterator(nil, nil)
		for iter.Next() {
			key := iter.Key()
			value := iter.Value()

			dict[string(key)] = string(value)
		}
		iter.Release()
	}
	return dict
}

func NewLDB(path string) LDB {
	var root = g_conf.Get("root")
	ldb := LDB{
		Path: root.(string) + "conf/" + path,
	}
	return ldb
}

//编码
func Encode(data interface{}) ([]byte, error) {
	r, err := json.Marshal(data)
	return r, err
}

//解码
func Decode(b []byte, result *interface{}) error {
	err := json.Unmarshal(b, result)
	return err
}

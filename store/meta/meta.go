package meta

import (
	"encoding/json"
	"io/ioutil"
	"os"
	"path/filepath"
)

type Meta struct {
	OlderFids []int64 `json:"fids"`
	Dir       string  `json:"dir"`
	ActiveFid int64   `json:"activeFid"`
}

const (
	metaFn     = "meta.json"
	DefaultDir = "/tmp/dkv"
)

func NewMeta(dir string) (*Meta, error) {
	if dir == "" {
		dir = DefaultDir
	}
	fn := filepath.Join(dir, metaFn)
	ok, err := isExist(fn)
	if err != nil {
		return nil, err
	}
	if ok {
		b, err := ioutil.ReadFile(fn)
		if err != nil {
			return nil, err
		}
		m := &Meta{}
		err = json.Unmarshal(b, m)
		if err != nil {
			return nil, err
		}
		return m, nil
	}
	m := &Meta{Dir: dir}
	m.Save()
	return m, nil
}

func (m *Meta) Save() error {
	fn := filepath.Join(m.Dir, metaFn)
	b, err := json.Marshal(m)
	if err != nil {
		return err
	}
	os.Mkdir(m.Dir, 0666)
	return ioutil.WriteFile(fn, b, 0666)
}

func isExist(fn string) (bool, error) {
	_, err := os.Stat(fn)
	if err == nil {
		return true, nil
	}
	if os.IsNotExist(err) {
		return false, nil
	}
	return true, err
}

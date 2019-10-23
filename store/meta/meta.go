package meta

import (
	"encoding/json"
	"io/ioutil"
	"os"
	"path/filepath"
)

var (
	EOF = []byte("CYZEOF")
)

const (
	metaFn     = "meta.json"
	DefaultDir = "/tmp/dkv"
	Master     = 1
	Slave      = 2
)

type SlaveInfo struct {
	Fid int64 `json:"fid"`
	Off int64 `json:"off"`
}

type Meta struct {
	OlderFids  []int64               `json:"fids"`
	Dir        string                `json:"dir"`
	ActiveFid  int64                 `json:"activeFid"`
	SlaveInfos map[string]*SlaveInfo `json:"slaveInfos"`
}

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
	m := &Meta{Dir: dir, SlaveInfos: make(map[string]*SlaveInfo)}
	m.Save()
	return m, nil
}

func (m *Meta) Save() error {
	fn := filepath.Join(m.Dir, metaFn)
	b, err := json.Marshal(m)
	if err != nil {
		return err
	}
	ok, err := isExist(fn)
	if err != nil {
		return err
	}
	if ok == false {
		os.Mkdir(m.Dir, 0755)
	}
	return ioutil.WriteFile(fn, b, 0644)
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

func (m *Meta) SaveSlaveInfo(host string, sI *SlaveInfo) error {
	m.SlaveInfos[host] = sI
	return m.Save()
}

func (m *Meta) GetSalveInfoByHost(host string) (sI *SlaveInfo, state bool) {
	sI, state = m.SlaveInfos[host]
	return
}

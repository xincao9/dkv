package meta

import (
	"encoding/json"
	"io/ioutil"
	"os"
	"path/filepath"
)

type meta struct {
	Fids      []string `json:"fids"`
	Dir       string   `json:"dir"`
	ActiveFid string   `json:"activeFid"`
}

const (
	metaFn     = "meta.json"
	defaultDir = "/tmp"
)

func NewMeta(dir string) (*meta, error) {
	if dir == "" {
		dir = defaultDir
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
		m := &meta{}
		err = json.Unmarshal(b, m)
		if err != nil {
			return nil, err
		}
		return m, nil
	}
	return &meta{Dir:dir}, nil
}

func (m *meta) Save () error {
	fn := filepath.Join(m.Dir, metaFn)
	b, err := json.Marshal(m)
	if err != nil {
		return err
	}
	return ioutil.WriteFile(fn, b, os.ModePerm)
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

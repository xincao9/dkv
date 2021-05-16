package meta

import (
	"dkv/component/constant"
	"dkv/component/logger"
	"encoding/json"
	"io/ioutil"
	"os"
	"path/filepath"
	"sort"
)

var M *meta

func init() {
	var err error
	M, err = New()
	if err != nil {
		logger.L.Fatalf("Fatal error meta: %v\n", err)
	}
}

type (
	i64       []int64
	slaveInfo struct {
		Fid int64 `json:"fid"`
		Off int64 `json:"off"`
	}
	meta struct {
		OlderFids  []int64               `json:"olderFids"`
		ActiveFid  int64                 `json:"activeFid"`
		SlaveInfos map[string]*slaveInfo `json:"slaveInfos"`
	}
)

func New() (*meta, error) {
	fn := filepath.Join(constant.Dir, constant.MetaFn)
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
	m := &meta{SlaveInfos: make(map[string]*slaveInfo)}
	m.Save()
	return m, nil
}

func (m *meta) Save() error {
	fn := filepath.Join(constant.Dir, constant.MetaFn)
	b, err := json.Marshal(m)
	if err != nil {
		return err
	}
	ok, err := isExist(fn)
	if err != nil {
		return err
	}
	if ok == false {
		err = os.MkdirAll(constant.Dir, 0755)
		if err != nil {
			return err
		}
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

func (m *meta) SaveSlaveInfo(host string, fid int64, off int64) error {
	m.SlaveInfos[host] = &slaveInfo{
		Fid: fid,
		Off: off,
	}
	return m.Save()
}

func (m *meta) GetSalveInfoByHost(host string) (sI *slaveInfo, state bool) {
	sI, state = m.SlaveInfos[host]
	return
}

func (m *meta) GetFids() []int64 {
	var fids []int64
	if m.OlderFids != nil {
		sort.Sort(i64(m.OlderFids))
		for _, fid := range m.OlderFids {
			if fid != 0 {
				fids = append(fids, fid)
			}
		}
	}
	if m.ActiveFid != 0 {
		fids = append(fids, m.ActiveFid)
	}
	return fids
}

func (i i64) Len() int {
	return len(i)
}
func (i i64) Swap(x, y int) {
	i[x], i[y] = i[y], i[x]
}

func (i i64) Less(x, y int) bool {
	return i[x] < i[y]
}

//  Copyright (c) 2014 Couchbase, Inc.
//  Licensed under the Apache License, Version 2.0 (the "License");
//  you may not use this file except in compliance with the
//  License. You may obtain a copy of the License at
//    http://www.apache.org/licenses/LICENSE-2.0
//  Unless required by applicable law or agreed to in writing,
//  software distributed under the License is distributed on an "AS
//  IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either
//  express or implied. See the License for the specific language
//  governing permissions and limitations under the License.

package cbgt

import (
	"github.com/couchbase/cbauth/metakv"
	log "github.com/couchbase/clog"
	"math"
	"sync"
)

const (
	BASE_CFG_PATH = "/cbgt/cfg/"
	CAS_FORCE     = math.MaxUint64
)

type CfgMetaKv struct {
	m        sync.Mutex
	path     string
	cancelCh chan struct{}
	rev      interface{}
	cfgMem   *CfgMem
}

// NewCfgMetaKv returns a CfgMetaKv that reads and stores its single
// configuration file in the metakv.
func NewCfgMetaKv() (*CfgMetaKv, error) {
	cfg := &CfgMetaKv{
		path:     BASE_CFG_PATH,
		cancelCh: make(chan struct{}),
		cfgMem:   NewCfgMem(),
	}
	go func() {
		for {
			err := metakv.RunObserveChildren(cfg.path, cfg.metaKVCallback,
				cfg.cancelCh)
			if err == nil {
				return
			} else {
				log.Printf("metakv notifier failed (%v)", err)
			}
		}
	}()
	return cfg, nil
}

func (c *CfgMetaKv) Get(key string, cas uint64) (
	[]byte, uint64, error) {
	c.m.Lock()
	defer c.m.Unlock()
	return c.cfgMem.Get(key, cas)
}

func (c *CfgMetaKv) Set(key string, val []byte, cas uint64) (
	uint64, error) {
	c.m.Lock()
	defer c.m.Unlock()
	rev, err := c.cfgMem.GetRev(key, cas)
	if err != nil {
		return 0, err
	}
	if rev == nil {
		err = metakv.Add(c.makeKey(key), val)
	} else {
		err = metakv.Set(c.makeKey(key), val, rev)
	}
	if err == nil {
		cas, err = c.cfgMem.Set(key, val, CAS_FORCE)
		if err != nil {
			return 0, err
		}
	}
	return cas, err
}

func (c *CfgMetaKv) Del(key string, cas uint64) error {
	c.m.Lock()
	defer c.m.Unlock()
	return c.delUnlocked(key, cas)
}

func (c *CfgMetaKv) delUnlocked(key string, cas uint64) error {
	rev, err := c.cfgMem.GetRev(key, cas)
	if err != nil {
		return err
	}
	err = metakv.Delete(c.makeKey(key), rev)
	if err == nil {
		return c.cfgMem.Del(key, 0)
	}
	return err
}

func (c *CfgMetaKv) Load() error {
	metakv.IterateChildren(c.path, c.metaKVCallback)
	return nil
}

func (c *CfgMetaKv) metaKVCallback(path string, value []byte, rev interface{}) error {
	c.m.Lock()
	defer c.m.Unlock()
	key := c.getMetaKey(path)
	if value == nil {
		// key got deleted
		return c.delUnlocked(key, 0)
	}
	cas, err := c.cfgMem.Set(key, value, CAS_FORCE)
	if err == nil {
		c.cfgMem.SetRev(key, cas, rev)
	}
	return err
}

func (c *CfgMetaKv) Subscribe(key string, ch chan CfgEvent) error {
	c.m.Lock()
	defer c.m.Unlock()

	return c.cfgMem.Subscribe(key, ch)
}

func (c *CfgMetaKv) Refresh() error {
	return nil
}

func (c *CfgMetaKv) OnError(err error) {
	log.Printf("cfg_metakv: OnError, err: %v", err)
}

func (c *CfgMetaKv) DelConf() {
	metakv.RecursiveDelete(c.path)
}

func (c *CfgMetaKv) makeKey(k string) string {
	return c.path + k
}

func (c *CfgMetaKv) getMetaKey(k string) string {
	return k[len(c.path):]
}

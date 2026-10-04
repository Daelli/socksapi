package main

import (
	"log"
	"os"
	"sync"
)

type DataMgr struct {
	canuse *os.File
	fis    map[string]*os.File
	lock   sync.Mutex
}

func NewDataMgr() *DataMgr {
	if err := os.RemoveAll(P("")); err != nil {
		log.Fatalln(err)
	}

	if err := os.MkdirAll(P(""), 0755); err != nil {
		log.Fatalln(err)
	}

	canuse, err := os.Create(P("canuse.txt"))
	if err != nil {
		log.Fatalln(err)
	}

	return &DataMgr{
		canuse: canuse,
		fis:    map[string]*os.File{},
	}
}

func (dm *DataMgr) Close() {
	dm.canuse.Close()
	for _, fi := range dm.fis {
		fi.Close()
	}
}

func (dm *DataMgr) Add(country, proxy string) {
	dm.lock.Lock()
	defer dm.lock.Unlock()

	if _, exist := dm.fis[country]; !exist {
		fi, err := os.Create(P(country + ".txt"))
		if err != nil {
			log.Fatalln(err)
		}

		dm.fis[country] = fi
	}

	dm.canuse.WriteString(proxy + "\n")
	dm.fis[country].WriteString(proxy + "\n")
}

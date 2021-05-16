package load_balance

import (
	"fmt"
	"log"
	"net"
	"reflect"
	"sort"
	"time"
)

const (
	//default check setting
	DefaultCheckMethod    = 0
	DefaultCheckTimeout   = 5
	DefaultCheckMaxErrNum = 2
	DefaultCheckInterval  = 5
)

type LoadBalanceCheckConf struct {
	observers    []Observer
	confIpWeight map[string]string
	activeList   []string
	format       string
	name         string
	closeChan    chan bool
}

func (s *LoadBalanceCheckConf) Attach(o Observer) {
	s.observers = append(s.observers, o)
}

func (s *LoadBalanceCheckConf) NotifyAllObservers() {
	for _, obs := range s.observers {
		obs.Update()
	}
}

func (s *LoadBalanceCheckConf) GetConf() []string {
	confList := []string{}
	for _, ip := range s.activeList {
		weight, ok := s.confIpWeight[ip]
		if !ok {
			weight = "50" //默认weight
		}
		confList = append(confList, fmt.Sprintf(s.format, ip)+","+weight)
	}
	return confList
}

//更新配置时，通知监听者也更新
func (s *LoadBalanceCheckConf) CloseWatch() {
	s.closeChan <- true
	close(s.closeChan)
}

//更新配置时，通知监听者也更新
func (s *LoadBalanceCheckConf) WatchConf() {
	go func() {
		confIpErrNum := map[string]int{}
		log.Printf("%s is checking ip_weight:%v active_list:%v\n", s.name, s.confIpWeight, s.activeList)
	OUTFOR:
		for {
			//log.Printf("begin switch\n")
			select {
			case <-s.closeChan:
				//log.Printf("ip_weight:%v is closed\n", s.confIpWeight)
				break OUTFOR
			default:
				//log.Printf("ip_weight:%v is default\n", s.confIpWeight)
			}
			//log.Printf("confIpWeight:%v is after switch\n", s.confIpWeight)
			//log.Printf("activeList:%v is after switch\n", s.activeList)
			changedList := []string{}
			for rs, _ := range s.confIpWeight {
				conn, err := net.DialTimeout("tcp", rs, time.Duration(DefaultCheckTimeout)*time.Second)
				if err == nil {
					conn.Close()
					if _, ok := confIpErrNum[rs]; ok {
						confIpErrNum[rs] = 0
					}
				}
				if err != nil {
					if _, ok := confIpErrNum[rs]; ok {
						confIpErrNum[rs] += 1
					} else {
						confIpErrNum[rs] = 1
					}
				}
				if confIpErrNum[rs] < DefaultCheckMaxErrNum {
					changedList = append(changedList, rs)
				}
				//log.Printf("rs:%v confIpErrNum:%v changeList:%v\n", rs, confIpErrNum, changedList)
			}
			sort.Strings(changedList)
			sort.Strings(s.activeList)
			if !reflect.DeepEqual(changedList, s.activeList) {
				log.Printf("%s is changed ip_weight:%v active_list:%v changed_list:%v\n", s.name, s.confIpWeight, s.activeList, changedList)
				s.UpdateConf(changedList)
			}
			time.Sleep(time.Duration(DefaultCheckInterval) * time.Second)
		}
	}()
}

//更新配置时，通知监听者也更新
func (s *LoadBalanceCheckConf) UpdateConf(conf []string) {
	s.activeList = conf
	for _, obs := range s.observers {
		obs.Update()
	}
}

func NewLoadBalanceCheckConf(name, format string, conf map[string]string) (*LoadBalanceCheckConf, error) {
	aList := []string{}
	for item, _ := range conf {
		aList = append(aList, item)
	}
	mConf := &LoadBalanceCheckConf{name: name, format: format, activeList: aList, confIpWeight: conf, closeChan: make(chan bool, 1)}
	mConf.WatchConf()
	return mConf, nil
}
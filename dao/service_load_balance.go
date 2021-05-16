package dao

import (
	"fmt"
	"github.com/e421083458/go_gateway/public"
	"github.com/e421083458/go_gateway/reverse_proxy/load_balance"
	"github.com/e421083458/gorm"
	"github.com/gin-gonic/gin"
	"log"
	"net"
	"net/http"
	"strings"
	"sync"
	"time"
)

type LoadBalance struct {
	ID            int64  `json:"id" gorm:"primary_key"`
	ServiceID     int64  `json:"service_id" gorm:"column:service_id" description:"服务id	"`
	CheckMethod   int    `json:"check_method" gorm:"column:check_method" description:"检查方法 tcpchk=检测端口是否握手成功	"`
	CheckTimeout  int    `json:"check_timeout" gorm:"column:check_timeout" description:"check超时时间"`
	CheckInterval int    `json:"check_interval" gorm:"column:check_interval" description:"检查间隔, 单位s"`
	RoundType     int    `json:"round_type" gorm:"column:round_type" description:"轮询方式 round/weight_round/random/ip_hash"`
	IpList        string `json:"ip_list" gorm:"column:ip_list" description:"ip列表"`
	WeightList    string `json:"weight_list" gorm:"column:weight_list" description:"权重列表"`
	ForbidList    string `json:"forbid_list" gorm:"column:forbid_list" description:"禁用ip列表"`

	UpstreamConnectTimeout int `json:"upstream_connect_timeout" gorm:"column:upstream_connect_timeout" description:"下游建立连接超时, 单位s"`
	UpstreamHeaderTimeout  int `json:"upstream_header_timeout" gorm:"column:upstream_header_timeout" description:"下游获取header超时, 单位s	"`
	UpstreamIdleTimeout    int `json:"upstream_idle_timeout" gorm:"column:upstream_idle_timeout" description:"下游链接最大空闲时间, 单位s	"`
	UpstreamMaxIdle        int `json:"upstream_max_idle" gorm:"column:upstream_max_idle" description:"下游最大空闲链接数"`
}

func (t *LoadBalance) TableName() string {
	return "gateway_service_load_balance"
}

func (t *LoadBalance) Find(c *gin.Context, tx *gorm.DB, search *LoadBalance) (*LoadBalance, error) {
	model := &LoadBalance{}
	err := tx.SetCtx(public.GetGinTraceContext(c)).Where(search).Find(model).Error
	return model, err
}

func (t *LoadBalance) Save(c *gin.Context, tx *gorm.DB) error {
	if err := tx.SetCtx(public.GetGinTraceContext(c)).Save(t).Error; err != nil {
		return err
	}
	return nil
}

func (t *LoadBalance) GetIPListByModel() []string {
	//todo disf
	return strings.Split(t.IpList, ",")
}

func (t *LoadBalance) GetWeightListByModel() []string {
	//todo disf
	return strings.Split(t.WeightList, ",")
}

var LoadBalancerHandler *LoadBalancer

type LoadBalancer struct {
	LoadBanlanceMap   map[string]*LoadBalancerItem
	LoadBanlanceSlice []*LoadBalancerItem
	Locker            sync.RWMutex
}

type LoadBalancerItem struct {
	LoadBanlance load_balance.LoadBalance
	ServiceName  string
	UpdatedAt    time.Time
}

func NewLoadBalancer() *LoadBalancer {
	return &LoadBalancer{
		LoadBanlanceMap:   map[string]*LoadBalancerItem{},
		LoadBanlanceSlice: []*LoadBalancerItem{},
		Locker:            sync.RWMutex{},
	}
}

func init() {
	LoadBalancerHandler = NewLoadBalancer()
	ServiceManagerHandler.Regist(LoadBalancerHandler)
}

func (lbr *LoadBalancer) Update(e *ServiceEvent) {
	log.Printf("LoadBalancer.Update\n")
	for _, service := range e.AddService {
		lbr.GetLoadBalancer(service)
	}
	for _, service := range e.UpdateService {
		lbr.GetLoadBalancer(service)
	}
	newLBSlice := []*LoadBalancerItem{}
	for _, lbrItem := range lbr.LoadBanlanceSlice {
		matched := false
		for _, service := range e.DeleteService {
			if lbrItem.ServiceName == service.Info.ServiceName {
				lbrItem.LoadBanlance.Close()
				matched = true
			}
		}
		if matched {
			delete(lbr.LoadBanlanceMap, lbrItem.ServiceName)
		} else {
			newLBSlice = append(newLBSlice, lbrItem)
		}
	}
	lbr.LoadBanlanceSlice = newLBSlice
}

func (lbr *LoadBalancer) GetLoadBalancer(service *ServiceDetail) (load_balance.LoadBalance, error) {
	for _, lbrItem := range lbr.LoadBanlanceSlice {
		if lbrItem.ServiceName == service.Info.ServiceName && lbrItem.UpdatedAt == service.Info.UpdatedAt {
			//log.Println("get store service")
			return lbrItem.LoadBanlance, nil
		}
	}
	schema := "http://"
	if service.HTTPRule.NeedHttps == 1 {
		schema = "https://"
	}
	if service.Info.LoadType == public.LoadTypeTCP || service.Info.LoadType == public.LoadTypeGRPC {
		schema = ""
	}
	ipList := service.LoadBalance.GetIPListByModel()
	weightList := service.LoadBalance.GetWeightListByModel()
	ipConf := map[string]string{}
	for ipIndex, ipItem := range ipList {
		ipConf[ipItem] = weightList[ipIndex]
	}
	//key是ip, value是weight
	var mConf load_balance.LoadBalanceConf

	tmpConf, err := load_balance.NewLoadBalanceCheckConf(service.Info.ServiceName, fmt.Sprintf("%s%s", schema, "%s"), ipConf)
	if err != nil {
		return nil, err
	}
	mConf = tmpConf
	lb := load_balance.LoadBanlanceFactorWithConf(load_balance.LbType(service.LoadBalance.RoundType), mConf)

	//save to map and slice
	matched := false
	for _, lbrItem := range lbr.LoadBanlanceSlice {
		if lbrItem.ServiceName == service.Info.ServiceName {
			matched = true
			lbrItem.LoadBanlance.Close()
			lbrItem.LoadBanlance = lb
			lbrItem.UpdatedAt = service.Info.UpdatedAt
		}
	}
	if !matched {
		lbItem := &LoadBalancerItem{
			LoadBanlance: lb,
			ServiceName:  service.Info.ServiceName,
			UpdatedAt:    service.Info.UpdatedAt,
		}
		lbr.LoadBanlanceSlice = append(lbr.LoadBanlanceSlice, lbItem)
		lbr.Locker.Lock()
		defer lbr.Locker.Unlock()
		lbr.LoadBanlanceMap[service.Info.ServiceName] = lbItem
	}
	return lb, nil
}

var TransportorHandler *Transportor

type Transportor struct {
	TransportMap   map[string]*TransportItem
	TransportSlice []*TransportItem
	Locker         sync.RWMutex
}

type TransportItem struct {
	Trans       *http.Transport
	ServiceName string
	UpdateAt    time.Time
}

func NewTransportor() *Transportor {
	return &Transportor{
		TransportMap:   map[string]*TransportItem{},
		TransportSlice: []*TransportItem{},
		Locker:         sync.RWMutex{},
	}
}

func init() {
	TransportorHandler = NewTransportor()
	ServiceManagerHandler.Regist(TransportorHandler)
}

func (t *Transportor) Update(e *ServiceEvent) {
	log.Printf("Transportor.Update\n")
	for _, service := range e.AddService {
		t.GetTrans(service)
	}
	for _, service := range e.UpdateService {
		t.GetTrans(service)
	}
	newSlice := []*TransportItem{}
	for _, tItem := range t.TransportSlice {
		matched := false
		for _, service := range e.DeleteService {
			if tItem.ServiceName == service.Info.ServiceName {
				matched = true
			}
		}
		if matched {
			delete(t.TransportMap, tItem.ServiceName)
		} else {
			newSlice = append(newSlice, tItem)
		}
	}
	t.TransportSlice = newSlice
}

func (t *Transportor) GetTrans(service *ServiceDetail) (*http.Transport, error) {
	for _, transItem := range t.TransportSlice {
		if transItem.ServiceName == service.Info.ServiceName && transItem.UpdateAt == service.Info.UpdatedAt {
			return transItem.Trans, nil
		}
	}
	if service.LoadBalance.UpstreamConnectTimeout == 0 {
		service.LoadBalance.UpstreamConnectTimeout = 50
	}
	if service.LoadBalance.UpstreamMaxIdle == 0 {
		service.LoadBalance.UpstreamMaxIdle = 4000
	}
	if service.LoadBalance.UpstreamIdleTimeout == 0 {
		service.LoadBalance.UpstreamIdleTimeout = 90
	}
	if service.LoadBalance.UpstreamHeaderTimeout == 0 {
		service.LoadBalance.UpstreamHeaderTimeout = 30
	}
	perhost := 0
	if len(service.LoadBalance.GetIPListByModel()) > 0 {
		perhost = service.LoadBalance.UpstreamMaxIdle / len(service.LoadBalance.GetIPListByModel())
	}
	trans := &http.Transport{
		DialContext: (&net.Dialer{
			Timeout:   time.Duration(service.LoadBalance.UpstreamConnectTimeout) * time.Second,
			KeepAlive: 30 * time.Second,
			DualStack: true,
		}).DialContext,
		ForceAttemptHTTP2:     true,
		MaxIdleConnsPerHost:   perhost,
		MaxIdleConns:          service.LoadBalance.UpstreamMaxIdle,
		WriteBufferSize:       1 << 18, //200m
		ReadBufferSize:        1 << 18, //200m
		IdleConnTimeout:       time.Duration(service.LoadBalance.UpstreamIdleTimeout) * time.Second,
		TLSHandshakeTimeout:   10 * time.Second,
		ResponseHeaderTimeout: time.Duration(service.LoadBalance.UpstreamHeaderTimeout) * time.Second,
	}

	//save to map and slice
	matched := false
	for _, transItem := range t.TransportSlice {
		if transItem.ServiceName == service.Info.ServiceName {
			matched = true
			transItem.Trans = trans
			transItem.UpdateAt = service.Info.UpdatedAt
		}
	}
	if !matched {
		transItem := &TransportItem{
			Trans:       trans,
			ServiceName: service.Info.ServiceName,
			UpdateAt:    service.Info.UpdatedAt,
		}
		t.TransportSlice = append(t.TransportSlice, transItem)
		t.Locker.Lock()
		defer t.Locker.Unlock()
		t.TransportMap[service.Info.ServiceName] = transItem
	}
	return trans, nil
}

package load_balance

// 配置主题
type LoadBalanceConf interface {
	Attach(o Observer)
	GetConf() []string
	WatchConf()
	UpdateConf(conf []string)
	CloseWatch()
}

type Observer interface {
	Update()
}

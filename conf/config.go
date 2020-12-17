package conf

// AppConf app配置结构体
type AppConf struct {
	KafkaConf `ini:"kafka"`
	EtcdConf  `ini:"etcd"`
}

// KafkaConf kafka配置结构体
type KafkaConf struct {
	Address string `ini:"address"`
	MaxSize int    `ini:"chan_max_size"`
}

// EtcdConf etcd集群配置结构体
type EtcdConf struct {
	Address string `ini:"address"`
	TimeOut int    `ini:"timeout"`
	Key     string `ini:"collect_log_key"`
}

// ---- unused ⬇️ ----

// TaillogConf taillog配置结构体
type TaillogConf struct {
	FileName string `ini:"filename"`
}

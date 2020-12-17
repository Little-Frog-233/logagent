package main

import (
	"fmt"
	"logagent_v2/conf"
	"logagent_v2/etcd"
	"logagent_v2/kafka"
	"logagent_v2/taillog"
	"sync"
	"time"

	"gopkg.in/ini.v1"
)

var (
	cfg = new(conf.AppConf)
)

func main() {
	var wg sync.WaitGroup
	wg.Add(1)
	// 0.加载配置文件
	err := ini.MapTo(cfg, "./conf/config.ini")
	if err != nil {
		fmt.Println("load ini failed, err:", err)
		return
	}

	// 1.初始化kafka连接
	err = kafka.Init([]string{cfg.KafkaConf.Address}, cfg.KafkaConf.MaxSize)
	if err != nil {
		fmt.Println("Init Kafka failed, err:", err)
		return
	}
	fmt.Println("init kafka success")

	// 2.初始化etcd
	err = etcd.Init(cfg.EtcdConf.Address, time.Duration(cfg.EtcdConf.TimeOut)*time.Second)
	if err != nil {
		fmt.Println("Init Etcd failed, err:", err)
		return
	}
	fmt.Println("init etcd success")

	// 2.1 从etcd中获取日志收集项的配置信息
	// 注: 为了实现每个logagent都拉取自己独有的配置，所以要以自己的ip地址作为区分
	// ipStr, err := utils.GetOutboundIP()
	// if err != nil {
	// 	panic(err)
	// }
	// etcdCondKey := fmt.Sprintf(cfg.EtcdConf.Key, ipStr)

	etcdCondKey := cfg.EtcdConf.Key
	logEntryConf, err := etcd.GetConf(etcdCondKey)
	if err != nil {
		fmt.Println("get conf from etcd failed, err:", err)
		return
	}
	fmt.Println("get conf from etcd success, value:", logEntryConf)

	// 3. 收集日志发往kafka
	taillog.Init(logEntryConf)

	// 2.2 派一个哨兵去监视日志收集项的变化(有变化即使通知logAgent实现热加载)
	newConfChan := taillog.PushNewConfChan()    //从taillog包中获取对外暴露的通道
	go etcd.WatchConf(etcdCondKey, newConfChan) //哨兵发现最新的配置信息会通知上面的那个通道
	wg.Wait()                                   //阻塞
}

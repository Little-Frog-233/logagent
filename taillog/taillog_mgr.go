package taillog

import (
	"fmt"
	"logagent_v2/etcd"
	"time"
)

var tskMgr *taillogMgr

type taillogMgr struct {
	logEntry    []*etcd.LogEntry
	tskMap      map[string]*TailTask
	newConfChan chan []*etcd.LogEntry
}

func Init(logEntryConf []*etcd.LogEntry) {
	tskMgr = &taillogMgr{
		logEntry:    logEntryConf, //把当前的日志收集项的配置信息保存起来
		tskMap:      make(map[string]*TailTask, 16),
		newConfChan: make(chan []*etcd.LogEntry),
	}

	for _, logEntry := range logEntryConf {
		// 初始化的时候起了多少个tailtask都要记下来，为了后续判断方便
		tailObj := NewTailTask(logEntry.Path, logEntry.Topic)
		mk := fmt.Sprintf("%s:%s", logEntry.Path, logEntry.Topic)
		tskMgr.tskMap[mk] = tailObj
	}
	go tskMgr.run()
}

// 监听自己的newConfChan，有了新的配置过来之后就做对应的处理
func (t *taillogMgr) run() {
	for {
		select {
		case newConf := <-t.newConfChan:
			for _, conf := range newConf {
				mk := fmt.Sprintf("%s:%s", conf.Path, conf.Topic)
				_, ok := t.tskMap[conf.Path]
				if ok {
					continue
				} else {
					tailObj := NewTailTask(conf.Path, conf.Topic)
					t.tskMap[mk] = tailObj
				}
			}
			// 找出原来t.tskMap有，但是newConf中没有的要删掉
			for _, c1 := range t.logEntry { //从原来的配置中依次拿出配置项
				isDelete := true
				for _, c2 := range newConf { //去新的配置中逐一进行比较
					if c2.Path == c1.Path && c2.Topic == c1.Topic {
						isDelete = true
						continue
					}
				}
				if isDelete {
					// 把c1对应的tailObj给停掉
					mk := fmt.Sprintf("%s:%s", c1.Path, c1.Topic)
					t.tskMap[mk].cancelFunc()
				}
			}
			// 1.配置新增
			// 2.配置删除
			// 3.配置变更
			fmt.Println("新配置来了, 配置:", newConf)
		default:
			time.Sleep(time.Second)
		}
	}
}

// PushNewConfChan 向外暴露tskMgr的newConfChan
func PushNewConfChan() chan<- []*etcd.LogEntry {
	return tskMgr.newConfChan
}

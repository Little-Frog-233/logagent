package taillog

import (
	"context"
	"fmt"
	"logagent_v2/kafka"
	"time"

	"github.com/hpcloud/tail"
)

// 专门从日志文件收集日志的模块

// TailTask 一个日志收集的任务
type TailTask struct {
	Path     string
	Topic    string
	Instance *tail.Tail
	// 为了能够实现退出t.run
	ctx        context.Context
	cancelFunc context.CancelFunc
}

// NewTailTask 构造函数
func NewTailTask(path, topic string) (tailObj *TailTask) {
	ctx, cancel := context.WithCancel(context.Background())
	tailObj = &TailTask{
		Path:       path,
		Topic:      topic,
		ctx:        ctx,
		cancelFunc: cancel,
	}
	tailObj.init()
	return
}

// Init 初始化Tail.Task.Instance
func (t *TailTask) init() {
	config := tail.Config{
		ReOpen:    true,
		Follow:    true,
		Location:  &tail.SeekInfo{Offset: 0, Whence: 2},
		MustExist: false,
		Poll:      true,
	}
	var err error

	// 打开文件开始读取数据
	t.Instance, err = tail.TailFile(t.Path, config)
	if err != nil {
		fmt.Println("tail file failed, err:", err)
	}
	go t.Run()
}

// ReadChan 返回一个chan
func (t *TailTask) ReadChan() <-chan *tail.Line {
	return t.Instance.Lines
}

// Run 向kafka发送消息
func (t *TailTask) Run() {
	for {
		select {
		// 利用context关闭任务
		case <-t.ctx.Done():
			fmt.Printf("tail task:%v over...\n", t.Path+"_"+t.Topic)
			return
		case line := <-t.Instance.Lines:
			// kafka.SendToKafka(t.Topic, line.Text) //函数调用函数
			// 先把日志数据发到一个通道中
			// kafka包中有单独的goroutine去取日志数据发往kafka
			kafka.SendToChan(t.Topic, line.Text)
		default:
			time.Sleep(time.Second)
		}
	}
}

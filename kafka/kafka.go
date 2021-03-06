package kafka

/*
	专门往kafka里写日志的模块
*/

import (
	"fmt"
	"time"

	"github.com/Shopify/sarama"
)

type logData struct {
	topic string
	data  string
}

var (
	client      sarama.SyncProducer //声明一个全局的连接kafka的生产者
	logDataChan chan *logData
)

// Init 初始化client
func Init(addrs []string, maxSize int) (err error) {
	config := sarama.NewConfig()
	config.Producer.RequiredAcks = sarama.WaitForAll          //发送完数据需要leader和follow都确认
	config.Producer.Partitioner = sarama.NewRandomPartitioner //轮询选分区
	config.Producer.Return.Successes = true                   //成功交付的消息将在success channel返回

	// 3.连接kafka
	client, err = sarama.NewSyncProducer(addrs, config)
	if err != nil {
		fmt.Println("producer closed, err:", err)
		return
	}

	//初始化logDataChan
	logDataChan = make(chan *logData, maxSize)
	//开启后台的goroutine从通道中取数据发往kafka
	go sendToKafka()
	return
}

// SendToChan 对外部暴露，把日志数据发送到内部chan中
func SendToChan(topic, data string) {
	msg := &logData{
		topic: topic,
		data:  data,
	}
	logDataChan <- msg
}

// SendToKafka 往kafka发送消息
func sendToKafka() {
	for {
		select {
		case ld := <-logDataChan:
			msg := &sarama.ProducerMessage{}
			msg.Topic = ld.topic
			msg.Value = sarama.StringEncoder(ld.data)
			// 发送到kafka
			pid, offset, err := client.SendMessage(msg)
			if err != nil {
				fmt.Println("send msg failed, err:", err)
				return
			}
			fmt.Printf("pid: %v offset: %v\n", pid, offset)
		default:
			time.Sleep(time.Microsecond * 50)
		}
	}

}

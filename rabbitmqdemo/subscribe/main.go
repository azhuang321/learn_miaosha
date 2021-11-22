package main

import (
	"demo-rabbitmq/service/mq"
	"fmt"
	"strconv"
	"time"
)

/**
订阅模式
多个队列可以订阅同一个交换机,则队列可同时被多个消费者消费
**/
// 生产队列中的消息
func PublishMq() {
	go func() {
		count := 0
		for {
			mq.PublishEx("fyouku.demo.fanout", "fanout", "", "hello"+strconv.Itoa(count))
			count++
			time.Sleep(time.Second)
		}
	}()
}

func ConsumerMq(wokerName string) {
	mq.ConsumerEx("fyouku.demo.fanout", "fanout", "", func(msg string) {
		fmt.Printf("worker-%s-msg is : %s \n", wokerName, msg)
	})
}

//消费队列中的消息

func main() {
	PublishMq()

	go func() {
		ConsumerMq("1")
	}()

	ConsumerMq("2")
}

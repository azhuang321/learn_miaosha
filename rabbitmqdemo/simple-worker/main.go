package main

import (
	"demo-rabbitmq/service/mq"
	"fmt"
	"strconv"
	"time"
)

//简单工作模式和worker工作模式 push方法
/**

简单模式和工作模式
简单模式:直接生产,直接消费队列中的值
worker模式:直接生产,竞争的从队列中消费值


**/
// 生产队列中的消息
func PublishMq() {
	go func() {
		count := 0
		for {
			mq.Publish("", "fyouku_demo", "hello"+strconv.Itoa(count))
			count++
			time.Sleep(time.Second)
		}
	}()
}

func ConsumerMq(wokerName string) {
	mq.Consumer("", "fyouku_demo", func(msg string) {
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

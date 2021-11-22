package main

import (
	"demo-rabbitmq/service/mq"
	"fmt"
	"strconv"
	"time"
)

/**
路由模式
可根据设置的路由进行消费
**/
// 生产队列中的消息
func PublishMq() {
	go func() {
		count := 0
		for {
			if count%2 != 0 {
				mq.PublishEx("fyouku.demo.direct", "direct", "one", "hello"+strconv.Itoa(count))
			} else {
				mq.PublishEx("fyouku.demo.direct", "direct", "two", "hello"+strconv.Itoa(count))
			}
			count++
			time.Sleep(time.Second)
		}
	}()
}

func ConsumerMq(wokerName string) {
	mq.ConsumerEx("fyouku.demo.direct", "direct", wokerName, func(msg string) {
		fmt.Printf("worker-%s-msg is : %s \n", wokerName, msg)
	})
}

//消费队列中的消息

func main() {
	PublishMq()

	go func() {
		ConsumerMq("one")
	}()

	ConsumerMq("two")
}

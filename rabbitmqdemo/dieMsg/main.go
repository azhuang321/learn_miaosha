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
				mq.PublishDlx("fyouku.dxl.a", "dxl"+strconv.Itoa(count))
			} else {
				mq.PublishEx("fyouku.dxl.b", "fanout", "", "dxltwo"+strconv.Itoa(count))
			}
			count++
			time.Sleep(time.Second)
		}
	}()
}

func ConsumerMq(wokerName string) {
	mq.ConsumerDlx("fyouku.dxl.a", "fyouku_dxl_a", "fyouku.dxl.b", "fyouku_dxl_b", 10000, func(msg string) {
		fmt.Printf("worker-%s-msg is : %s \n", wokerName, msg)
	})
}

//消费队列中的消息

func main() {
	PublishMq()

	ConsumerMq("two")
}

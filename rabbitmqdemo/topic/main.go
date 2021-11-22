package main

import (
	"demo-rabbitmq/service/mq"
	"fmt"
	"strconv"
	"time"
)

/**
主题模式

*/
func PublishMq() {
	go func() {
		count := 0
		for {
			if count%2 != 0 {
				mq.PublishEx("fyouku.demo.topic", "topic", "fyouku.video", "fyouku.video"+strconv.Itoa(count))
			} else {
				mq.PublishEx("fyouku.demo.topic", "topic", "user.fyouku", "user.fyouku"+strconv.Itoa(count))
			}
			count++
			time.Sleep(time.Second)
		}
	}()
}

func PublishMqTwo() {
	go func() {
		count := 0
		for {
			if count%2 != 0 {
				mq.PublishEx("fyouku.demo.topic", "topic", "a.frog.name", "a.frog.name"+strconv.Itoa(count))
			} else {
				mq.PublishEx("fyouku.demo.topic", "topic", "b.frog.uid", "b.frog.uid"+strconv.Itoa(count))
			}
			count++
			time.Sleep(time.Second)
		}
	}()
}

func ConsumerMq(wokerName string) {
	mq.ConsumerEx("fyouku.demo.topic", "topic", wokerName, func(msg string) {
		fmt.Printf("worker-%s-msg is : %s \n", wokerName, msg)
	})
}

//消费队列中的消息

func main() {
	PublishMq()
	PublishMqTwo()

	go func() {
		ConsumerMq("#")
	}()

	go func() {
		ConsumerMq("*.frog.*")
	}()

	go func() {
		ConsumerMq("fyouku.*")
	}()

	ConsumerMq("two")
}

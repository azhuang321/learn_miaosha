package main

import (
	"imooc-product/demo"
	"strconv"
	"time"
)

func main() {
	rabbitmq := demo.NewRabbitMQSimple("imoocSimple")
	for i := 0; i <= 100; i++ {
		rabbitmq.PublishSimple("hello world" + strconv.Itoa(i))
		time.Sleep(time.Second)
	}
}

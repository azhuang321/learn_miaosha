package main

import (
	"fmt"
	"imooc-product/demo"
)

func main() {
	rabbitmq := demo.NewRabbitMQSimple("imoocSimple")
	rabbitmq.PublishSimple("hello world")
	fmt.Println("send success")
}

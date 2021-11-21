package main

import (
	"fmt"
	"miaosha/demo"
)

func main() {
	rabbitmq := demo.NewRabbitMQSimple("imoocSimple")
	rabbitmq.PublishSimple("hello world")
	fmt.Println("send success")
}

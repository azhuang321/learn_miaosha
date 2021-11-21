package main

import (
	"imooc-product/demo"
)

func main() {
	rabbitmq := demo.NewRabbitMQSimple("imoocSimple")
	rabbitmq.ConsumeSimple()
}

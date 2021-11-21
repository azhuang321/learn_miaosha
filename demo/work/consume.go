package main

import (
	"miaosha/demo"
)

func main() {
	rabbitmq := demo.NewRabbitMQSimple("imoocSimple")
	rabbitmq.ConsumeSimple()
}

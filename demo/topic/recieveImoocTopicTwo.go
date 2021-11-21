package main

import (
	"miaosha/demo"
)

func main() {
	imoocOne := demo.NewRabbitMQTopic("exImoocTopic", "imooc.*.two")
	imoocOne.RecieveTopic()
}

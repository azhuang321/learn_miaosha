package main

import (
	"miaosha/demo"
)

func main() {
	imoocOne := demo.NewRabbitMQTopic("exImoocTopic", "#")
	imoocOne.RecieveTopic()
}

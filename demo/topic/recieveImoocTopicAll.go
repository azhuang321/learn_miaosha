package main

import (
	"imooc-product/demo"
)

func main() {
	imoocOne := demo.NewRabbitMQTopic("exImoocTopic", "#")
	imoocOne.RecieveTopic()
}

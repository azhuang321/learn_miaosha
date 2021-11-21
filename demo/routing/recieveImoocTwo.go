package main

import (
	"miaosha/demo"
)

func main() {
	imoocOne := demo.NewRabbitMQRouting("exImooc", "imooc_two")
	imoocOne.RecieveRouting()
}

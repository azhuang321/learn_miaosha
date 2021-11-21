package main

import (
	"imooc-product/demo"
)

func main() {
	imoocOne := demo.NewRabbitMQRouting("exImooc", "imooc_one")
	imoocOne.RecieveRouting()
}

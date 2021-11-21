package main

import "imooc-product/demo"

func main() {
	rabbitmq := demo.NewRabbitMQPubSub("" +
		"newProduct")
	rabbitmq.RecieveSub()
}

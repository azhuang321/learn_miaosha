package main

import "miaosha/demo"

func main() {
	rabbitmq := demo.NewRabbitMQPubSub("" +
		"newProduct")
	rabbitmq.RecieveSub()
}

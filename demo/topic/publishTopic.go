package main

import (
	"fmt"
	"miaosha/demo"
	"strconv"
	"time"
)

func main() {
	imoocOne := demo.NewRabbitMQTopic("exImoocTopic", "imooc.topic.one")
	imoocTwo := demo.NewRabbitMQTopic("exImoocTopic", "imooc.topic.two")
	for i := 0; i <= 10; i++ {
		imoocOne.PublishTopic("Hello imooc topic one!" + strconv.Itoa(i))
		imoocTwo.PublishTopic("Hello imooc topic Two!" + strconv.Itoa(i))
		time.Sleep(1 * time.Second)
		fmt.Println(i)
	}

}

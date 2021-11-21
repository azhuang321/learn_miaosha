package main

import (
	"fmt"
	"miaosha/demo"
	"strconv"
	"time"
)

func main() {
	imoocOne := demo.NewRabbitMQRouting("exImooc", "imooc_one")
	imoocTwo := demo.NewRabbitMQRouting("exImooc", "imooc_two")
	for i := 0; i <= 10; i++ {
		imoocOne.PublishRouting("Hello imooc one!" + strconv.Itoa(i))
		imoocTwo.PublishRouting("Hello imooc Two!" + strconv.Itoa(i))
		time.Sleep(1 * time.Second)
		fmt.Println(i)
	}

}

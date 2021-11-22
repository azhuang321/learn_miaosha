package mq

import (
	"bytes"
	"fmt"

	"github.com/streadway/amqp"
)

type Callback func(msg string)

// 创建连接
func Connect() (*amqp.Connection, error) {
	conn, err := amqp.Dial("amqp://guest:guest@127.0.0.1:5672")
	return conn, err
}

// 简单的生产者(用于简单模式和worker模式)
func Publish(exchange string, queueName string, body string) error {
	// 建立连接
	conn, err := Connect()
	defer conn.Close()
	if err != nil {
		return err
	}

	// 创建通道channel
	channel, err := conn.Channel()
	defer channel.Close()
	if err != nil {
		return err
	}
	// 创建队列
	q, err := channel.QueueDeclare(
		queueName, // 队列名称
		true,      // 是否持久化: 当rabbitmq重启, 队列中的数据仍然保留  必须发送端和接受端都设置为持久化才生效
		false,     // 是否自动删除(应答)
		false,     // 是否具有排他性
		false,     // 是否阻塞处理
		nil,       // 额外的属性
	)
	if err != nil {
		return err
	}
	// 发送消息
	err = channel.Publish(exchange, q.Name, false, false, amqp.Publishing{
		DeliveryMode: amqp.Persistent, // 持久化是需要该参数
		ContentType:  "text/plain",
		Body:         []byte(body),
	})
	return err
}

// 简单的消费者(用于简单模式和worker模式)
func Consumer(exchange string, queueName string, callback Callback) {
	// 建立连接
	conn, err := Connect()
	defer conn.Close()
	if err != nil {
		fmt.Println(err)
		return
	}

	// 创建通道channel
	channel, err := conn.Channel()
	defer channel.Close()
	if err != nil {
		fmt.Println(err)
		return
	}
	// 创建队列
	q, err := channel.QueueDeclare(
		queueName,
		true, // 是否持久化: 当rabbitmq重启, 队列中的数据仍然保留
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		fmt.Println(err)
		return
	}

	/*
		第三个参数:是否自动应答(当消息被取出后,不管消费端是否操作成功,队列直接删除消息即为自动应答,)
		当第三个参数设置为false的话, 如果消息没有收到手动应答该消息会在Unacked状态中, 即:未应答
		手动应答参考代码中的d.Ack(false)
	*/
	msg, err := channel.Consume(q.Name, "", false, false, false, false, nil)
	if err != nil {
		fmt.Println(err)
		return
	}
	forever := make(chan bool)
	go func() {
		for d := range msg {
			s := BytesToString(&(d.Body))
			callback(*s)

			d.Ack(false) // 手动应答
		}
	}()
	fmt.Printf("Waiting for messages")
	<-forever
}

// 制定模式和路由的 - 生产者(用于发布订阅模式/路由模式/主题模式)
func PublishEx(exchange string, types string, routingKey string, body string) error {
	// 建立连接
	conn, err := Connect()
	defer conn.Close()
	if err != nil {
		return err
	}
	// 创建channel
	channel, err := conn.Channel()
	defer channel.Close()
	if err != nil {
		return err
	}
	// 创建交换机
	err = channel.ExchangeDeclare(
		exchange,
		types,
		true, // 持久化
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		return err
	}
	err = channel.Publish(exchange, routingKey, false, false, amqp.Publishing{
		DeliveryMode: amqp.Persistent, // 持久化是需要该参数
		ContentType:  "text/plain",
		Body:         []byte(body),
	})
	return err
}

// 制定模式和路由的 - 消费者(用于发布订阅模式/路由模式/主题模式)
func ConsumerEx(exchange string, types string, routingKey string, callback Callback) {
	// 建立连接
	conn, err := Connect()
	defer conn.Close()
	if err != nil {
		fmt.Println(err)
		return
	}
	// 创建通道channel
	channel, err := conn.Channel()
	defer channel.Close()
	if err != nil {
		fmt.Println(err)
		return
	}
	// 创建交换机
	err = channel.ExchangeDeclare(
		exchange,
		types,
		true,
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		fmt.Println(err)
		return
	}
	// 创建队列
	q, err := channel.QueueDeclare(
		"",
		false,
		false,
		true,
		false,
		nil,
	)
	if err != nil {
		fmt.Println(err)
		return
	}
	// 绑定队列到交换机
	err = channel.QueueBind(
		q.Name,
		routingKey,
		exchange,
		false,
		nil,
	)
	if err != nil {
		fmt.Println(err)
		return
	}

	msgs, err := channel.Consume(q.Name, "", false, false, false, false, nil)
	if err != nil {
		fmt.Println(err)
		return
	}

	forever := make(chan bool)
	go func() {
		for d := range msgs {
			s := BytesToString(&(d.Body))
			callback(*s)
			d.Ack(false)
		}
	}()
	fmt.Printf("Waiting for messages\n")
	<-forever
}

func BytesToString(b *[]byte) *string {
	s := bytes.NewBuffer(*b)
	r := s.String()
	return &r
}

// 死信队列 - 消费者
func ConsumerDlx(exchangeA string, queueAName string, exchangeB string, queueBName string, ttl int, callback Callback) {
	// 建立连接
	conn, err := Connect()
	defer conn.Close()
	if err != nil {
		fmt.Println(err)
		return
	}
	// 创建channel
	channel, err := conn.Channel()
	defer channel.Close()
	if err != nil {
		fmt.Println(err)
		return
	}
	// 1)创建A交换机 & 2)A队列 & 3)A交换机和A队列绑定
	// 1)创建A交换机
	err = channel.ExchangeDeclare(
		exchangeA,
		"fanout",
		true,
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		fmt.Println(err)
		return
	}
	// 2)A队列
	queueA, err := channel.QueueDeclare(
		queueAName,
		true,
		false,
		false,
		false,
		amqp.Table{
			"x-message-ttl":          ttl,       // 过期时间
			"x-dead-letter-exchange": exchangeB, // 交换机
			// "x-dead-letter-queue": "", // 绑定的队列
			// "x-dead-letter-routing-key": "", // 路由关键字
		},
	)
	if err != nil {
		fmt.Println(err)
		return
	}
	// 3)A交换机和A队列绑定
	err = channel.QueueBind(
		queueA.Name,
		"",
		exchangeA,
		false,
		nil,
	)
	if err != nil {
		fmt.Println(err)
		return
	}

	// 1)创建B交换机 & 2)B队列 & 3)B交换机和B队列绑定
	// 1)创建B交换机
	err = channel.ExchangeDeclare(
		exchangeB,
		"fanout",
		true,
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		fmt.Println(err)
		return
	}
	// 2)B队列
	queueB, err := channel.QueueDeclare(
		queueBName,
		true,
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		fmt.Println(err)
		return
	}
	// 3)B交换机和B队列绑定
	err = channel.QueueBind(
		queueB.Name,
		"",
		exchangeB,
		false,
		nil,
	)
	if err != nil {
		fmt.Println(err)
		return
	}

	// 接收消息
	msgs, err := channel.Consume(queueB.Name, "", false, false, false, false, nil)
	if err != nil {
		fmt.Println(err)
		return
	}

	forever := make(chan bool)
	go func() {
		for d := range msgs {
			s := BytesToString(&(d.Body))
			callback(*s) // 回调函数
			d.Ack(false) // 应答
		}
	}()
	fmt.Printf("Waiting for messages")
	<-forever
}

// 死信队列 - 生产者
func PublishDlx(exchangeA string, body string) error {
	// 建立连接
	conn, err := Connect()
	defer conn.Close()
	if err != nil {
		return err
	}
	// 创建channel
	channel, err := conn.Channel()
	defer channel.Close()
	if err != nil {
		return err
	}
	err = channel.Publish(exchangeA, "", false, false, amqp.Publishing{
		DeliveryMode: amqp.Persistent,
		ContentType:  "text/plain",
		Body:         []byte(body),
	})
	return err
}

package demo

import (
	"fmt"
	"github.com/streadway/amqp"
	"log"
)

const MQURL = "amqp:imoocuser:imoocuser@127.0.0.1:5672/imooc"

type RabbitMQ struct {
	conn      *amqp.Connection
	channel   *amqp.Channel
	QueueName string //队列名称
	Exchange  string //交换机
	Key       string
	Mqurl     string //链接信息
}

func NewRabbitMQ(queueName string, exchange string, key string) *RabbitMQ {
	rabbitmq := &RabbitMQ{
		QueueName: queueName,
		Exchange:  exchange,
		Key:       key,
		Mqurl:     MQURL,
	}
	var err error
	rabbitmq.conn, err = amqp.Dial(rabbitmq.Mqurl)
	rabbitmq.failOnErr(err, "创建链接错误!")
	rabbitmq.channel, err = rabbitmq.conn.Channel()
	rabbitmq.failOnErr(err, "获取Channel失败")
	return rabbitmq
}

func (r *RabbitMQ) Destroy() {
	r.channel.Close()
	r.conn.Close()
}

func (r *RabbitMQ) failOnErr(err error, message string) {
	if err != nil {
		log.Fatalf("%s:%s", message, err)
		panic(fmt.Sprintf("%s:%s", message, err))
	}
}

// NewRabbitMQSimple 创建简单模式下rabbitMq的实例
// @param queueName
// @return *RabbitMQ
func NewRabbitMQSimple(queueName string) *RabbitMQ {
	return NewRabbitMQ(queueName, "", "")
}

func (r *RabbitMQ) PublishSimple(message string) {
	// 1.申请队列,如果队列不存在会自动创建,如果存在则跳过创建(保证队列存在,消息能够发送到队列中
	_, err := r.channel.QueueDeclare(
		r.QueueName,
		false, //是否持久化
		false, //是否自动删除
		false, //是否具有排他性
		false, //是否阻塞
		nil,   //额外属性
	)
	if err != nil {
		fmt.Println(err)
	}
	// 2.发送消息到队列中
	r.channel.Publish(
		r.Exchange,
		r.QueueName,
		false, //如果为true,根据exchange类型和route key规则,如果无法找到符合条件的队列,那么会把发送的消息返回给发送者
		false, //如果为true,当exchange发送消息到队列后发现队列上没有绑定消费者,则会把消息返回给发送者
		amqp.Publishing{
			ContentType: "text/plain",
			Body:        []byte(message),
		},
	)

}

func (r *RabbitMQ) ConsumeSimple() {
	// 1.申请队列,如果队列不存在会自动创建,如果存在则跳过创建(保证队列存在,消息能够发送到队列中
	_, err := r.channel.QueueDeclare(
		r.QueueName,
		false, //是否持久化
		false, //是否自动删除
		false, //是否具有排他性
		false, //是否阻塞
		nil,   //额外属性
	)
	if err != nil {
		fmt.Println(err)
	}
	// 2.接收消息
	msgs, err := r.channel.Consume(
		r.QueueName,
		"",    //区分多个消费者
		true,  //是否自动应答
		false, //是否具有排他性
		false, //如果为true,表示不能将同一个 connection 中发送的消息传递给这个connection中的消费者
		false, //是否阻塞消费
		nil,   //额外参数
	)
	if err != nil {
		fmt.Println(err)
	}

	forever := make(chan bool)
	// 启用协程处理消息
	go func() {
		for d := range msgs {
			log.Printf("Received a message: %s", d.Body)
		}
	}()
	log.Printf("[*] Waiting for messages,To exit press CTRL + C")
	<-forever
}

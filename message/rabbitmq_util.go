package message

import (
	"github.com/streadway/amqp"
	"log"
	"mini-seckill/service"
	"strconv"
)

func failOnError(err error, msg string) {
	if err != nil {
		log.Printf("%s: %s", msg, err)
	}
}

func PublishCacheDeleteMessage(cacheKey string) error {
	conn, err := amqp.Dial("amqp://guest:guest@localhost:5672/")
	if err != nil {
		failOnError(err, "Failed to connect to RabbitMQ")
		return err
	}
	defer conn.Close()

	ch, err := conn.Channel()
	if err != nil {
		failOnError(err, "Failed to open a channel")
		return err
	}
	defer ch.Close()

	q, err := ch.QueueDeclare(
		"cacheDeleteQueue", false, false, false, false, nil)
	if err != nil {
		failOnError(err, "Failed to declare a queue")
		return err
	}

	err = ch.Publish(
		"",
		q.Name,
		false,
		false,
		amqp.Publishing{
			ContentType: "text/plain",
			Body:        []byte(cacheKey),
		})
	if err != nil {
		failOnError(err, "Failed to publish a message")
		return err
	}
	return nil
}

func ConsumerForCacheDeleteMessage() {
	conn, err := amqp.Dial("amqp://guest:guest@localhost:5672/")
	failOnError(err, "Failed to connect to RabbitMQ")
	defer conn.Close()

	ch, err := conn.Channel()
	failOnError(err, "Failed to open a channel")
	defer ch.Close()

	q, err := ch.QueueDeclare(
		"cacheDeleteQueue", false, false, false, false, nil)
	failOnError(err, "Failed to declare a queue")

	msgs, err := ch.Consume(
		q.Name, "", true, false, false, false, nil)
	failOnError(err, "Failed to register a consumer")

	forever := make(chan bool)

	go func() {
		for d := range msgs {
			stockId, _ := strconv.Atoi(string(d.Body))
			res := service.DeleteStockCountCache(stockId)
			if !res {
				// retry
				err = ch.Publish(
					"",
					q.Name,
					false,
					false,
					amqp.Publishing{
						ContentType: "text/plain",
						Body:        []byte(strconv.Itoa(stockId)),
					})
				failOnError(err, "Failed to publish a message")
			} else {
				log.Printf("删除缓存重试成功，key: %d", stockId)
			}
		}
	}()

	log.Printf(" [*] Waiting for messages. To exit press CTRL+C")
	<-forever
}

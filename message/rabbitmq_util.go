package message

import (
	"context"
	"encoding/json"
	"github.com/streadway/amqp"
	"log"
	"mini-seckill/config"
	"mini-seckill/domain"
	"mini-seckill/service"
	"mini-seckill/util"
	"os"
	"os/signal"
	"time"
)

func failOnError(err error, msg string) {
	if err != nil {
		log.Printf("%s: %s", msg, err)
	}
}

func connect() (*amqp.Connection, error) {
	return amqp.Dial("amqp://guest:guest@host.docker.internal:5672/")
}

func PublishMessage(message []byte, queueName string) error {
	conn, err := connect()
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
		queueName, false, false, false, false, nil)
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
			ContentType: "text/json",
			Body:        message,
		})
	if err != nil {
		failOnError(err, "Failed to publish a message")
		return err
	}
	return nil
}

func PublishCacheDeleteMessage(cacheKey string) error {
	conn, err := connect()
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
		config.StockCacheDeleteQueueName, false, false, false, false, nil)
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

func ConsumerForOrderCreate() {
	conn, err := connect()
	failOnError(err, "Failed to connect to RabbitMQ")
	defer conn.Close()

	ch, err := conn.Channel()
	failOnError(err, "Failed to open a channel")
	defer ch.Close()

	q, err := ch.QueueDeclare(
		config.OrderCreateQueueName, false, false, false, false, nil)
	failOnError(err, "Failed to declare a queue")

	msgs, err := ch.Consume(
		q.Name, "", true, false, false, false, nil)
	failOnError(err, "Failed to register a consumer")

	stop := make(chan os.Signal)
	signal.Notify(stop, os.Interrupt)

	go func() {
		for d := range msgs {
			go func(delivery amqp.Delivery) {
				userOrderInfo := &domain.UserOrderInfo{}
				err = json.Unmarshal(delivery.Body, userOrderInfo)
				if err != nil {
					log.Println("invalid message content")
					return
				}
				// create order
				_, err := service.CreateOrder(userOrderInfo.Sid, userOrderInfo.UserId)
				if err != nil {
					log.Println("fail to create order")
					return
				}

				for {
					err := PublishCacheDeleteMessage(config.GenerateStockKey(userOrderInfo.Sid))
					if err == nil {
						break
					}
					select {
					case sig := <-stop:
						log.Printf("got %s signal, clean all resources", sig)
						ch.Close()
						conn.Close()
					}
				}
			}(d)
		}
	}()

	log.Printf(" [*] Waiting for messages. To exit press CTRL+C")
	select {
	case sig := <-stop:
		log.Printf("got %s signal, clean all resources", sig)
		ch.Close()
		conn.Close()
	}

}

func ConsumerForStockCacheDelete() {
	conn, err := connect()
	failOnError(err, "Failed to connect to RabbitMQ")
	defer conn.Close()

	ch, err := conn.Channel()
	failOnError(err, "Failed to open a channel")
	defer ch.Close()

	q, err := ch.QueueDeclare(
		config.StockCacheDeleteQueueName, false, false, false, false, nil)
	failOnError(err, "Failed to declare a queue")

	msgs, err := ch.Consume(
		q.Name, "", true, false, false, false, nil)
	failOnError(err, "Failed to register a consumer")

	stop := make(chan os.Signal)
	signal.Notify(stop, os.Interrupt)

	go func() {
		for d := range msgs {
			go func(delivery amqp.Delivery) {
				key := string(delivery.Body)
				time.Sleep(time.Second)
				err := util.RedisCli.Del(context.Background(), key).Err()
				if err != nil {
					for {
						err := PublishCacheDeleteMessage(key)
						if err == nil {
							break
						}
						select {
						case sig := <-stop:
							log.Printf("got %s signal, clean all resources", sig)
							ch.Close()
							conn.Close()
						}
					}
				}
			}(d)
		}
	}()

	log.Printf(" [*] Waiting for messages. To exit press CTRL+C")
	select {
	case sig := <-stop:
		log.Printf("got %s signal, clean all resources", sig)
		ch.Close()
		conn.Close()
	}

}

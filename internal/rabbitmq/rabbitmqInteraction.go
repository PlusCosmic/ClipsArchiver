package rabbitmq

import (
	"context"
	"log"
	"strconv"
	"time"

	amqp "github.com/rabbitmq/amqp091-go"
)

type RequestEntry struct {
	Id          int `json:"id"`
	RequestType int `json:"requestType"`
}

const connStr = "amqp://clipsArchiver:clipsArchiver@10.0.0.10:5672/"

var channel *amqp.Channel
var queue *amqp.Queue

func init() {
	conn, err := amqp.Dial(connStr)
	if err != nil {
		log.Fatal(err.Error())
	}

	ch, err := conn.Channel()
	if err != nil {
		log.Fatal(err.Error())
	}

	q, err := ch.QueueDeclare(
		"clips_transcode_queue",
		true,
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		log.Fatal(err.Error())
	}

	channel = ch
	queue = &q
}

func PublishToTranscodeQueue(requestEntry RequestEntry) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	err := channel.PublishWithContext(ctx,
		"",         // exchange
		queue.Name, // routing key
		false,      // mandatory
		false,
		amqp.Publishing{
			DeliveryMode: amqp.Persistent,
			ContentType:  "text/plain",
			Body:         []byte(strconv.Itoa(requestEntry.Id) + "," + strconv.Itoa(requestEntry.RequestType)),
		})
	return err
}

func GetConsumeChannel() (<-chan amqp.Delivery, error) {
	transcodes, err := channel.Consume(queue.Name, "", false, false, false, false, nil)
	return transcodes, err
}

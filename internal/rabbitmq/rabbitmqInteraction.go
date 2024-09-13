package rabbitmq

import (
	"context"
	"encoding/json"
	"log"
	"strconv"
	"time"

	amqp "github.com/rabbitmq/amqp091-go"
)

type RequestEntry struct {
	Id          int `json:"id"`
	RequestType int `json:"requestType"`
}

type MapUpdateNotification struct {
	MapName         string `json:"mapName"`
	DurationMinutes int    `json:"durationMinutes"`
}

const connStr = "amqp://clipsArchiver:clipsArchiver@10.0.0.10:5672/"

var channel *amqp.Channel
var transcodeQueue *amqp.Queue
var mapUpdateNotificationQueue *amqp.Queue

func init() {
	conn, err := amqp.Dial(connStr)
	if err != nil {
		log.Fatal(err.Error())
	}

	ch, err := conn.Channel()
	if err != nil {
		log.Fatal(err.Error())
	}

	err = ch.ExchangeDeclare(
		"video_requests", // name
		"direct",         // type
		true,             // durable
		false,            // auto-deleted
		false,            // internal
		false,            // no-wait
		nil,              // arguments
	)
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

	err = ch.QueueBind(q.Name, "", "video_requests", false, nil)

	if err != nil {
		log.Fatal(err.Error())
	}
	transcodeQueue = &q

	muq, err := ch.QueueDeclare(
		"map_update_notification_queue",
		true,
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		log.Fatal(err.Error())
	}

	mapUpdateNotificationQueue = &muq

	channel = ch

}

func PublishToTranscodeQueue(requestEntry RequestEntry) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	err := channel.PublishWithContext(ctx,
		"video_requests",    // exchange
		transcodeQueue.Name, // routing key
		false,               // mandatory
		false,
		amqp.Publishing{
			DeliveryMode: amqp.Persistent,
			ContentType:  "text/plain",
			Body:         []byte(strconv.Itoa(requestEntry.Id) + "," + strconv.Itoa(requestEntry.RequestType)),
		})
	return err
}

func PublishMapUpdateNotification(mapUpdateNotification MapUpdateNotification) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	jsonBytes, err := json.Marshal(mapUpdateNotification)
	if err != nil {
		return err
	}

	err = channel.PublishWithContext(ctx,
		"",                              // exchange
		mapUpdateNotificationQueue.Name, // routing key
		false,                           // mandatory
		false,
		amqp.Publishing{
			DeliveryMode: amqp.Persistent,
			ContentType:  "text/plain",
			Body:         jsonBytes,
		})
	return err
}

func GetConsumeChannel() (<-chan amqp.Delivery, error) {
	transcodes, err := channel.Consume(transcodeQueue.Name, "", false, false, false, false, nil)
	return transcodes, err
}

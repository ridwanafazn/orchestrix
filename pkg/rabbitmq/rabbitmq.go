package rabbitmq

import (
	"log"

	amqp "github.com/rabbitmq/amqp091-go"
)

type RabbitMQClient struct {
	Conn    *amqp.Connection
	Channel *amqp.Channel
}

func InitRabbitMQ(url string) *RabbitMQClient {
	conn, err := amqp.Dial(url)
	if err != nil {
		log.Fatalf("❌ Gagal terhubung ke RabbitMQ: %v", err)
	}

	ch, err := conn.Channel()
	if err != nil {
		log.Fatalf("❌ Gagal membuka channel RabbitMQ: %v", err)
	}

	log.Println("✅ RabbitMQ Connected")

	return &RabbitMQClient{
		Conn:    conn,
		Channel: ch,
	}
}

// Publish mengirim pesan ke antrean yang spesifik
func (r *RabbitMQClient) Publish(queueName string, payload []byte) error {
	_, err := r.Channel.QueueDeclare(
		queueName,
		true,  // durable
		false, // auto-delete
		false, // exclusive
		false, // no-wait
		nil,   // arguments
	)
	if err != nil {
		return err
	}

	err = r.Channel.Publish(
		"",        // exchange default
		queueName, // routing key
		false,     // mandatory
		false,     // immediate
		amqp.Publishing{
			ContentType:  "application/json",
			DeliveryMode: amqp.Persistent,
			Body:         payload,
		},
	)
	return err
}

// Consume menarik pesan dari antrean secara terus-menerus
func (r *RabbitMQClient) Consume(queueName string, prefetchCount int) (<-chan amqp.Delivery, error) {
	_, err := r.Channel.QueueDeclare(
		queueName,
		true,
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		return nil, err
	}

	// QoS membatasi pesan yang ditarik Worker agar CPU tidak hang.
	// Jika diset 10, worker maksimal hanya memegang 10 pekerjaan bersamaan.
	err = r.Channel.Qos(
		prefetchCount,
		0,     // prefetch size
		false, // global
	)
	if err != nil {
		return nil, err
	}

	return r.Channel.Consume(
		queueName,
		"",    // consumer tag
		false, // auto-ack: FALSE! Kita pakai manual Ack di Usecase
		false, // exclusive
		false, // no-local
		false, // no-wait
		nil,   // args
	)
}

func (r *RabbitMQClient) Close() {
	r.Channel.Close()
	r.Conn.Close()
}

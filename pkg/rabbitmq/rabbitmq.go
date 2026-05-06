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

	// Setup Dead Letter Exchange (DLX) secara global saat inisiasi
	err = ch.ExchangeDeclare("dlx_exchange", "direct", true, false, false, false, nil)
	if err != nil {
		log.Fatalf("❌ Gagal membuat DLX: %v", err)
	}

	log.Println("✅ RabbitMQ Connected (with DLX ready)")

	return &RabbitMQClient{
		Conn:    conn,
		Channel: ch,
	}
}

func (r *RabbitMQClient) Publish(queueName string, payload []byte) error {
	// Menambahkan argumen DLX agar jika terjadi reject/failure,
	// pesan dilempar ke dlx_exchange
	args := amqp.Table{
		"x-dead-letter-exchange":    "dlx_exchange",
		"x-dead-letter-routing-key": "dead_" + queueName,
	}

	_, err := r.Channel.QueueDeclare(
		queueName,
		true,  // durable
		false, // auto-delete
		false, // exclusive
		false, // no-wait
		args,  // IMPLEMENTASI DLX
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

func (r *RabbitMQClient) Consume(queueName string, prefetchCount int) (<-chan amqp.Delivery, error) {
	args := amqp.Table{
		"x-dead-letter-exchange":    "dlx_exchange",
		"x-dead-letter-routing-key": "dead_" + queueName,
	}

	_, err := r.Channel.QueueDeclare(
		queueName,
		true,
		false,
		false,
		false,
		args, // BINDING DLX DI KONSUMEN JUGA
	)
	if err != nil {
		return nil, err
	}

	err = r.Channel.Qos(prefetchCount, 0, false)
	if err != nil {
		return nil, err
	}

	return r.Channel.Consume(
		queueName,
		"",    // consumer tag
		false, // auto-ack: FALSE!
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

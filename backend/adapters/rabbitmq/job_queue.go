package rabbitmq

import (
	"context"
	"encoding/json"
	"fmt"

	"bloodconnect/application"
	"bloodconnect/application/domain"

	amqp "github.com/rabbitmq/amqp091-go"
)

type rabbitMQJobQueue struct {
	conn    *amqp.Connection
	channel *amqp.Channel
	queues  map[domain.JobType]string
}

func NewJobQueue(url string) (application.JobQueue, error) {
	conn, err := amqp.Dial(url)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to RabbitMQ: %w", err)
	}

	ch, err := conn.Channel()
	if err != nil {
		conn.Close()
		return nil, fmt.Errorf("failed to open a channel: %w", err)
	}

	queues := map[domain.JobType]string{
		domain.JobTypeWaveSearch:   "wave_search_jobs",
		"notification":             "notification_jobs", // Added for refactor
	}

	for _, qName := range queues {
		_, err := ch.QueueDeclare(
			qName, // name
			true,  // durable
			false, // delete when unused
			false, // exclusive
			false, // no-wait
			nil,   // arguments
		)
		if err != nil {
			ch.Close()
			conn.Close()
			return nil, fmt.Errorf("failed to declare queue %s: %w", qName, err)
		}
	}

	return &rabbitMQJobQueue{
		conn:    conn,
		channel: ch,
		queues:  queues,
	}, nil
}

func (q *rabbitMQJobQueue) Enqueue(ctx context.Context, job *domain.Job) error {
	qName, ok := q.queues[job.Type]
	if !ok {
		return fmt.Errorf("unsupported job type: %s", job.Type)
	}

	body, err := json.Marshal(job)
	if err != nil {
		return fmt.Errorf("failed to marshal job: %w", err)
	}

	return q.channel.PublishWithContext(ctx,
		"",    // exchange
		qName, // routing key
		false, // mandatory
		false, // immediate
		amqp.Publishing{
			ContentType: "application/json",
			Body:        body,
		})
}

// These methods are no longer used with RabbitMQ as it handles queuing natively
func (q *rabbitMQJobQueue) FetchNextAvailable(ctx context.Context) (*domain.Job, error) {
	return nil, nil
}

func (q *rabbitMQJobQueue) MarkStatus(ctx context.Context, id domain.JobID, status domain.JobStatus) error {
	return nil
}

func (q *rabbitMQJobQueue) Delete(ctx context.Context, id domain.JobID) error {
	return nil
}

func (q *rabbitMQJobQueue) Consume(ctx context.Context, jobType domain.JobType) (<-chan *domain.Job, error) {
	qName, ok := q.queues[jobType]
	if !ok {
		return nil, fmt.Errorf("unsupported job type: %s", jobType)
	}

	msgs, err := q.channel.Consume(
		qName, // queue
		"",    // consumer
		true,  // auto-ack
		false, // exclusive
		false, // no-local
		false, // no-wait
		nil,   // args
	)
	if err != nil {
		return nil, fmt.Errorf("failed to register consumer for %s: %w", qName, err)
	}

	jobChan := make(chan *domain.Job)
	go func() {
		defer close(jobChan)
		for d := range msgs {
			var job domain.Job
			if err := json.Unmarshal(d.Body, &job); err == nil {
				jobChan <- &job
			}
		}
	}()

	return jobChan, nil
}

func (q *rabbitMQJobQueue) Close() error {
	if q.channel != nil {
		q.channel.Close()
	}
	if q.conn != nil {
		q.conn.Close()
	}
	return nil
}

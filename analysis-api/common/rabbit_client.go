package common

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"os"
	"runtime"
	"sync"
	"time"

	"github.com/rs/zerolog"
	"github.com/streadway/amqp"
)

var (
	ErrDisconnected = errors.New("disconnected from rabbitmq, trying to reconnect")
)

const (
	// When reconnecting to the server after connection failure
	reconnectDelay = 5 * time.Second

	// When resending messages the server didn't confirm
	resendDelay = 5 * time.Second
)

type RabbitClient struct {
	pushQueue     string
	streamQueue   string
	logger        zerolog.Logger
	connection    *amqp.Connection
	channel       *amqp.Channel
	done          chan os.Signal
	notifyClose   chan *amqp.Error
	notifyConfirm chan amqp.Confirmation
	isConnected   bool
	alive         bool
	threads       int
	wg            *sync.WaitGroup
}

// NewRabbitClient is a constructor that takes address, push and listen queue names, logger,
// and a channel that will notify rabbitmq client on server shutdown.
// We calculate the number of threads, create the client, and start the connection process.
// Connect method connects to the rabbitmq server and creates push/listen channels if they don't exist.
func NewRabbitClient(streamQueue, pushQueue, addr string, done chan os.Signal, l zerolog.Logger) *RabbitClient {
	threads := runtime.GOMAXPROCS(0)
	if numCPU := runtime.NumCPU(); numCPU > threads {
		threads = numCPU
	}

	client := RabbitClient{
		logger:      l,
		threads:     threads,
		pushQueue:   pushQueue,
		streamQueue: streamQueue,
		done:        done,
		alive:       true,
		wg:          &sync.WaitGroup{},
	}
	client.wg.Add(threads)

	go client.handleReconnect(addr)
	return &client
}

// handleReconnect will wait for a connection error on
// notifyClose, and then continuously attempt to reconnect.
func (c *RabbitClient) handleReconnect(addr string) {
	for c.alive {
		c.isConnected = false
		t := time.Now()
		fmt.Printf("Attempting to connect to rabbitMQ: %s\n", addr)
		var retryCount int
		for !c.connect(addr) {
			if !c.alive {
				return
			}
			select {
			case <-c.done:
				return
			case <-time.After(reconnectDelay + time.Duration(retryCount)*time.Second):
				c.logger.Printf("disconnected from rabbitMQ and failed to connect")
				retryCount++
			}
		}
		c.logger.Printf("Connected to rabbitMQ in: %vms", time.Since(t).Milliseconds())
		select {
		case <-c.done:
			return
		case <-c.notifyClose:
		}
	}
}

// connect will make a single attempt to connect to
// RabbitMq. It returns the success of the attempt.
func (c *RabbitClient) connect(addr string) bool {
	conn, err := amqp.Dial(addr)
	if err != nil {
		c.logger.Printf("failed to dial rabbitMQ server: %v", err)
		return false
	}
	ch, err := conn.Channel()
	if err != nil {
		c.logger.Printf("failed connecting to channel: %v", err)
		return false
	}
	ch.Confirm(false)
	_, err = ch.QueueDeclare(
		c.streamQueue,
		true,  // Durable
		false, // Delete when unused
		false, // Exclusive
		false, // No-wait
		nil,   // Arguments
	)
	if err != nil {
		c.logger.Printf("failed to declare stream queue: %v", err)
		return false
	}

	_, err = ch.QueueDeclare(
		c.pushQueue,
		true,  // Durable
		false, // Delete when unused
		false, // Exclusive
		false, // No-wait
		nil,   // Arguments
	)
	if err != nil {
		c.logger.Printf("failed to declare push queue: %v", err)
		return false
	}
	c.changeConnection(conn, ch)
	c.isConnected = true
	return true
}

// changeConnection takes a new connection to the queue,
// and updates the channel listeners to reflect this.
func (c *RabbitClient) changeConnection(connection *amqp.Connection, channel *amqp.Channel) {
	c.connection = connection
	c.channel = channel
	c.notifyClose = make(chan *amqp.Error)
	c.notifyConfirm = make(chan amqp.Confirmation)
	c.channel.NotifyClose(c.notifyClose)
	c.channel.NotifyPublish(c.notifyConfirm)
}

// Push will push data onto the queue, and wait for a confirmation.
// If no confirms are received until within the resendTimeout,
// it continuously resends messages until a confirmation is received.
// This will block until the server sends a confirm.
func (c *RabbitClient) Push(data []byte) error {
	if !c.isConnected {
		return errors.New("failed to push push: not connected")
	}
	for {
		err := c.UnsafePush(data)
		if err != nil {
			if err == ErrDisconnected {
				continue
			}
			return err
		}
		select {
		case confirm := <-c.notifyConfirm:
			if confirm.Ack {
				return nil
			}
		case <-time.After(resendDelay):
		}
	}
}

// UnsafePush will push to the queue without checking for
// confirmation. It returns an error if it fails to connect.
// No guarantees are provided for whether the server will
// receive the message.
func (c *RabbitClient) UnsafePush(data []byte) error {
	if !c.isConnected {
		return ErrDisconnected
	}
	return c.channel.Publish(
		"",          // Exchange
		c.pushQueue, // Routing key
		false,       // Mandatory
		false,       // Immediate
		amqp.Publishing{
			ContentType: "text/plain",
			Body:        data,
		},
	)
}

func (c *RabbitClient) Stream(cancelCtx context.Context) error {
	for {
		if c.isConnected {
			break
		}
		time.Sleep(1 * time.Second)
	}

	err := c.channel.Qos(1, 0, false)
	if err != nil {
		return err
	}

	var connectionDropped bool

	for i := 1; i <= c.threads; i++ {
		msgs, err := c.channel.Consume(
			c.streamQueue,
			consumerName(i), // Consumer
			false,           // Auto-Ack
			false,           // Exclusive
			false,           // No-local
			false,           // No-Wait
			nil,             // Args
		)
		if err != nil {
			return err
		}

		go func() {
			defer c.wg.Done()
			for {
				select {
				case <-cancelCtx.Done():
					return
				case msg, ok := <-msgs:
					if !ok {
						connectionDropped = true
						return
					}
					c.parseEvent(msg)
				}
			}
		}()

	}

	c.wg.Wait()

	if connectionDropped {
		return ErrDisconnected
	}

	return nil
}

type event struct {
	Job  string `json:"job"`
	Data string `json:"data"`
}

func (c *RabbitClient) parseEvent(msg amqp.Delivery) {
	l := c.logger.Log().Timestamp()
	startTime := time.Now()

	var evt event
	err := json.Unmarshal(msg.Body, &evt)
	if err != nil {
		logAndNack(msg, l, startTime, "unmarshalling body: %s - %s", string(msg.Body), err.Error())
		return
	}

	if evt.Data == "" {
		logAndNack(msg, l, startTime, "received event without data")
		return
	}

	defer func(e event, m amqp.Delivery, logger *zerolog.Event) {
		if err := recover(); err != nil {
			stack := make([]byte, 8096)
			stack = stack[:runtime.Stack(stack, false)]
			logger.Bytes("stack", stack).Str("level", "fatal").Interface("error", err).Msg("panic recovery for rabbitMQ message")
			msg.Nack(false, false)
		}
	}(evt, msg, l)

	switch evt.Job {
	case "job1":
		// Call an actual function
		log.Print("received job1")
	default:
		msg.Reject(false)
		return
	}

	//
	//if err != nil {
	//	logAndNack(msg, l, startTime, err.Error())
	//	return
	//}

	l.Str("level", "info").Int64("took-ms", time.Since(startTime).Milliseconds()).Msgf("%s succeeded", evt.Job)
	msg.Ack(false)
}

func logAndNack(msg amqp.Delivery, l *zerolog.Event, t time.Time, err string, args ...interface{}) {
	msg.Nack(false, false)
	l.Int64("took-ms", time.Since(t).Milliseconds()).Str("level", "error").Msg(fmt.Sprintf(err, args...))
}

func (c *RabbitClient) Close() error {
	if !c.isConnected {
		return nil
	}
	c.alive = false
	fmt.Println("Waiting for current messages to be processed...")
	c.wg.Wait()
	for i := 1; i <= c.threads; i++ {
		fmt.Println("Closing consumer: ", i)
		err := c.channel.Cancel(consumerName(i), false)
		if err != nil {
			return fmt.Errorf("error canceling consumer %s: %v", consumerName(i), err)
		}
	}
	err := c.channel.Close()
	if err != nil {
		return err
	}
	err = c.connection.Close()
	if err != nil {
		return err
	}
	c.isConnected = false
	fmt.Println("gracefully stopped rabbitMQ connection")
	return nil
}

func consumerName(i int) string {
	return fmt.Sprintf("go-consumer-%v", i)
}

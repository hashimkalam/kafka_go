package main

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/IBM/sarama"
)

func main() {
	topic := "comments"
	worker, err := connectConsumer([]string{"localhost:29092"})
	if err != nil {
		panic(err)
	}
	defer func() {
		if err := worker.Close(); err != nil {
			fmt.Println("Failed to close consumer:", err)
		}
	}()

	consumer, err := worker.ConsumePartition(topic, 0, sarama.OffsetOldest)
	if err != nil {
		panic(err)
	}
	defer func() {
		if err := consumer.Close(); err != nil {
			fmt.Println("Failed to close partition consumer:", err)
		}
	}()

	fmt.Println("Consumer started")
	sigchan := make(chan os.Signal, 1)
	signal.Notify(sigchan, syscall.SIGINT, syscall.SIGTERM)

	msgCount := 0
	doneCh := make(chan struct{})

	go func() {
		for {
			select {
			case err := <-consumer.Errors():
				fmt.Println("Consumer error:", err)
			case msg := <-consumer.Messages():
				msgCount++
				fmt.Printf("Received message count: %d: | Topic (%s) | Message (%s)\n", msgCount, string(msg.Topic), string(msg.Value))
			case <-sigchan:
				fmt.Println("Interruption detected")
				close(doneCh)
				return
			}
		}
	}()

	<-doneCh
	fmt.Println("Processed", msgCount, "messages")
}

func connectConsumer(brokersUrl []string) (sarama.Consumer, error) {
	config := sarama.NewConfig()
	config.Consumer.Return.Errors = true
	conn, err := sarama.NewConsumer(brokersUrl, config)
	if err != nil {
		return nil, err
	}
	return conn, nil
}

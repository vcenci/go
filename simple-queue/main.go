package main

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"example.com/simple-queue/handlers"
	"example.com/simple-queue/queue"

	"github.com/prometheus/client_golang/prometheus/promhttp"
)

func main() {
	fmt.Println("Iniciando servidor na porta 8080")

	ctx := context.Background()
	q := queue.NewQueue(queue.ShardsQuantity, 30*time.Second, ctx)

	http.HandleFunc("/queue", handlers.EnqueueHandler(q))
	http.HandleFunc("/consume", handlers.DequeueHandler(q))
	http.HandleFunc("/ack", handlers.AckHandler(q))
	http.HandleFunc("/queue/size", handlers.SizeHandler(q))
	http.HandleFunc("/dlq", handlers.DLQHandler(q))
	http.HandleFunc("/dlq/size", handlers.DLQSizeHandler(q))
	http.Handle("/metrics", promhttp.Handler())

	err := http.ListenAndServe(":8080", nil)
	if err != nil {
		fmt.Println("Erro ao criar servidor na porta 8080: ", err)
	}
}

package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"log"

	"github.com/hibiken/asynq"
	"api-students/app/model"
	"api-students/config"
	"api-students/database"
)

func main() {
	outboxID := flag.Int64("outbox", 0, "ID baris outbox yang ingin disalin")
	copies := flag.Int("copies", 5, "Jumlah salinan task yang dikirim ke Redis")
	flag.Parse()

	if *outboxID == 0 {
		log.Fatal("Parameter -outbox wajib diisi, contoh: go run ./tools/redeliver -outbox 8 -copies 5")
	}

	config.LoadEnv()
	ctx := context.Background()

	pool, err := database.NewPool(ctx)
	if err != nil {
		log.Fatalf("Gagal koneksi ke database: %v", err)
	}
	defer pool.Close()

	var eventType string
	var payload []byte
	err = pool.QueryRow(ctx, "SELECT event_type, payload FROM outbox WHERE id = $1", *outboxID).Scan(&eventType, &payload)
	if err != nil {
		log.Fatalf("Gagal mengambil event outbox id %d: %v", *outboxID, err)
	}

	job := model.NotificationJob{
		OutboxID:  *outboxID,
		EventType: eventType,
		Payload:   payload,
	}

	payloadBytes, err := json.Marshal(job)
	if err != nil {
		log.Fatalf("Gagal marshal job: %v", err)
	}

	client := asynq.NewClient(config.RedisOpt())
	defer client.Close()

	for i := 1; i <= *copies; i++ {
		task := asynq.NewTask(
			model.TaskSendNotification,
			payloadBytes,
			asynq.Queue(model.QueueNotifications),
			asynq.TaskID(fmt.Sprintf("redeliver-%d-%d", *outboxID, i)),
		)
		_, err := client.Enqueue(task)
		if err != nil {
			log.Fatalf("Gagal enqueue salinan ke-%d: %v", i, err)
		}
		fmt.Printf("Berhasil memasukkan salinan task ke-%d ke Redis\n", i)
	}

	fmt.Printf("Selesai! Sebanyak %d salinan task untuk outbox ID %d telah dimasukkan ke Redis.\n", *copies, *outboxID)
}
package database // Pastikan nama package-nya sama dengan postgres.go

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// MongoClient adalah variabel global untuk diakses dari repositori nanti
var MongoClient *mongo.Client
var MongoDB *mongo.Database

func ConnectMongoDB() {
	// Ambil URI dan DB Name dari file .env (yang sudah diload sebelumnya)
	mongoURI := os.Getenv("MONGO_URI")
	dbName := os.Getenv("MONGO_DB_NAME")

	if mongoURI == "" {
		log.Fatal("MONGO_URI is not set in .env file")
	}

	// Set opsi koneksi klien
	clientOptions := options.Client().ApplyURI(mongoURI)

	// Buat konteks dengan timeout (maksimal 10 detik untuk mencoba connect)
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Mulai koneksi
	client, err := mongo.Connect(ctx, clientOptions)
	if err != nil {
		log.Fatal("Failed to connect to MongoDB:", err)
	}

	// Cek apakah database benar-benar hidup/merespons
	err = client.Ping(ctx, nil)
	if err != nil {
		log.Fatal("Failed to ping MongoDB:", err)
	}

	fmt.Println("🚀 Connected to MongoDB successfully!")
	
	// Simpan koneksi ke variabel global agar bisa dipakai fungsi lain
	MongoClient = client
	MongoDB = client.Database(dbName)
}
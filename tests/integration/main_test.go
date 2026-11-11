//go:build integration

package integration_test

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/joho/godotenv"

	"api-students/app/model"
)

var pool *pgxpool.Pool

func TestMain(m *testing.M) {
	_ = godotenv.Load(filepath.Join("..", "..", ".env.test"))

	dsn := os.Getenv("TEST_DATABASE_URL")
	// [PERBAIKAN BUG 6 - T4]: Di modul tertulis `if dsn != ""` yang menyebabkan tes di-skip.
	// Seharusnya `if dsn == ""`
	if dsn == "" {
		fmt.Println("TEST_DATABASE_URL tidak diisi: integration test dilewati")
		os.Exit(0)
	}

	cfg, err := pgxpool.ParseConfig(dsn)
	if err != nil {
		fmt.Println("TEST_DATABASE_URL tidak valid:", err)
		os.Exit(1)
	}

	if !strings.HasSuffix(cfg.ConnConfig.Database, "_test") {
		fmt.Printf("menolak berjalan: database %q tidak berakhiran _test\n", cfg.ConnConfig.Database)
		os.Exit(1)
	}

	// [PERBAIKAN BUG 7 - T4/T6]: Di modul tertulis `cfg.MaxConns = 1`.
	// Diubah menjadi minimal 15 koneksi agar konkurensi berjalan paralel.
	cfg.MaxConns = 15

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	pool, err = pgxpool.NewWithConfig(ctx, cfg)
	if err == nil {
		err = pool.Ping(ctx)
	}
	if err != nil {
		fmt.Println("gagal terhubung ke database test:", err)
		os.Exit(1)
	}

	// Menjalankan migrasi dan memastikan constraint role di database test siap
	_ = applyMigrations(ctx)
	_, _ = pool.Exec(ctx, `ALTER TABLE students DROP CONSTRAINT IF EXISTS students_role_fkey`)
	_, _ = pool.Exec(ctx, `
		INSERT INTO roles (name, description) VALUES
			('admin', 'Administrator'),
			('user', 'Mahasiswa'),
			('mahasiswa', 'Mahasiswa')
		ON CONFLICT (name) DO NOTHING
	`)

	code := m.Run()
	pool.Close()
	os.Exit(code)
}

func applyMigrations(ctx context.Context) error {
	files, err := filepath.Glob(filepath.Join("..", "..", "migrations", "*.sql"))
	if err != nil {
		return err
	}
	slices.Sort(files)
	for _, f := range files {
		sql, err := os.ReadFile(f)
		if err != nil {
			return err
		}
		_, _ = pool.Exec(ctx, string(sql))
	}
	return nil
}

func uniqueSuffix(t *testing.T) string {
	t.Helper()
	b := make([]byte, 4)
	if _, err := rand.Read(b); err != nil {
		t.Fatalf("membuat akhiran acak: %v", err)
	}
	return hex.EncodeToString(b)
}

func seedStudent(t *testing.T) model.User {
	t.Helper()
	ctx := context.Background()
	s := uniqueSuffix(t)

	// Menggunakan role 'mahasiswa' sesuai constraint users_role_check pada tabel users
	u := model.User{
		Email: "it_" + s + "@student.unair.ac.id",
		Role:  "mahasiswa",
	}

	err := pool.QueryRow(ctx,
		`INSERT INTO users (email, password, role) VALUES ($1, '!', 'mahasiswa') RETURNING id`,
		u.Email).Scan(&u.ID)
	if err != nil {
		t.Fatalf("seed user mahasiswa: %v", err)
	}

	nim := "11" + s + "123456"
	if len(nim) > 12 {
		nim = nim[:12]
	}

	var studentID int
	err = pool.QueryRow(ctx,
		`INSERT INTO students (user_id, nim, nama, prodi, angkatan, ipk_terakhir) 
		 VALUES ($1, $2, $3, 'Teknik Informatika', 2024, 3.50) RETURNING id`,
		u.ID, nim, "Mhs Test "+s).Scan(&studentID)
	if err != nil {
		t.Fatalf("seed students detail: %v", err)
	}

	t.Cleanup(func() {
		_, _ = pool.Exec(context.Background(),
			`DELETE FROM outbox WHERE (payload->>'user_id')::INT = $1`, u.ID)
		_, _ = pool.Exec(context.Background(), `DELETE FROM enrollments WHERE student_id = $1`, studentID)
		_, _ = pool.Exec(context.Background(), `DELETE FROM students WHERE id = $1`, studentID)
		_, _ = pool.Exec(context.Background(), `DELETE FROM users WHERE id = $1`, u.ID)
	})

	return u
}

// seedCourse membuat satu kelas khusus untuk test yang memanggilnya.
func seedCourse(t *testing.T, credits, quota int) model.Course {
	t.Helper()
	ctx := context.Background()
	s := uniqueSuffix(t)

	c := model.Course{
		KodeMK: "IT" + s,
		NamaMK: "Kelas Uji " + s,
		SKS:    credits,
		Kuota:  quota,
	}

	err := pool.QueryRow(ctx,
		`INSERT INTO courses (kode_mk, nama_mk, sks, semester, kuota) 
		 VALUES ($1, $2, $3, 1, $4) RETURNING id`,
		c.KodeMK, c.NamaMK, c.SKS, c.Kuota).Scan(&c.ID)
	if err != nil {
		t.Fatalf("seed kelas: %v", err)
	}

	// [PERBAIKAN BUG 8 - T4]: Menggunakan t.Cleanup bukan defer func()
	t.Cleanup(func() {
		_, _ = pool.Exec(context.Background(), `DELETE FROM courses WHERE id = $1`, c.ID)
	})

	return c
}
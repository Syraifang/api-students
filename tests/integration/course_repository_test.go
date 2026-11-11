//go:build integration

package integration_test

import (
	"context"
	"errors"
	"testing"

	"api-students/app/model"
	"api-students/app/repository"
)

func TestIncrementEnrolled_StopsAtQuota(t *testing.T) {
	ctx := context.Background()
	course := seedCourse(t, 3, 2)
	enrollRepo := repository.NewEnrollmentRepository(pool)

	for i := 1; i <= 2; i++ {
		student := seedStudent(t)
		tx, err := pool.Begin(ctx)
		if err != nil {
			t.Fatalf("begin tx: %v", err)
		}
		c, terisi, err := enrollRepo.LockAndGetCourse(ctx, tx, course.ID)
		if err != nil {
			_ = tx.Rollback(ctx)
			t.Fatalf("kursi ke-%d: lock and get course: %v", i, err)
		}
		if terisi >= c.Kuota {
			_ = tx.Rollback(ctx)
			t.Fatalf("kursi ke-%d: seharusnya belum penuh", i)
		}
		var studentID int
		_ = pool.QueryRow(ctx, "SELECT id FROM students WHERE user_id = $1", student.ID).Scan(&studentID)
		if err := enrollRepo.Create(ctx, tx, studentID, course.ID, "2026/2027"); err != nil {
			_ = tx.Rollback(ctx)
			t.Fatalf("kursi ke-%d: create enrollment: %v", i, err)
		}
		_ = tx.Commit(ctx)
	}

	tx, err := pool.Begin(ctx)
	if err != nil {
		t.Fatalf("begin tx: %v", err)
	}
	defer tx.Rollback(ctx)

	c, terisi, err := enrollRepo.LockAndGetCourse(ctx, tx, course.ID)
	if err != nil {
		t.Fatalf("kursi ke-3: lock and get course: %v", err)
	}
	if terisi < c.Kuota {
		t.Fatalf("kursi ke-3: terisi = %d, kuota = %d, seharusnya sudah penuh", terisi, c.Kuota)
	}

	var totalTerisi int
	_ = pool.QueryRow(ctx, `SELECT COUNT(*) FROM enrollments WHERE course_id = $1`, course.ID).Scan(&totalTerisi)
	if totalTerisi != 2 {
		t.Errorf("terisi = %d, seharusnya 2", totalTerisi)
	}
}

func TestInsertEnrollment_Duplicate(t *testing.T) {
	ctx := context.Background()
	course := seedCourse(t, 3, 40)
	student := seedStudent(t)
	enrollRepo := repository.NewEnrollmentRepository(pool)

	var studentID int
	_ = pool.QueryRow(ctx, "SELECT id FROM students WHERE user_id = $1", student.ID).Scan(&studentID)

	isDup, err := enrollRepo.CheckDuplicate(ctx, studentID, course.ID, "2026/2027")
	if err != nil {
		t.Fatalf("check duplicate 1: %v", err)
	}
	if isDup {
		t.Fatalf("seharusnya belum duplikat")
	}

	tx, err := pool.Begin(ctx)
	if err != nil {
		t.Fatalf("begin tx: %v", err)
	}
	if err := enrollRepo.Create(ctx, tx, studentID, course.ID, "2026/2027"); err != nil {
		_ = tx.Rollback(ctx)
		t.Fatalf("create enrollment 1: %v", err)
	}
	_ = tx.Commit(ctx)

	isDup2, err := enrollRepo.CheckDuplicate(ctx, studentID, course.ID, "2026/2027")
	if err != nil {
		t.Fatalf("check duplicate 2: %v", err)
	}
	if !isDup2 {
		t.Errorf("pendaftaran kedua: seharusnya terdeteksi duplikat")
	}
}

func TestWithTx_RollsBackEveryStep(t *testing.T) {
	ctx := context.Background()
	course := seedCourse(t, 3, 40)
	student := seedStudent(t)

	var studentID int
	_ = pool.QueryRow(ctx, "SELECT id FROM students WHERE user_id = $1", student.ID).Scan(&studentID)

	outboxRepo := repository.NewOutboxRepository(pool)
	txm := repository.NewTxManager(pool)

	errStop := errors.New("gagal yang disengaja")
	err := txm.WithTx(ctx, func(tx repository.DBTX) error {
		if err := outboxRepo.WithTx(tx).Insert(ctx, model.EventEnrollmentCreated,
			model.EnrollmentEvent{UserID: student.ID, CourseID: course.ID}); err != nil {
			return err
		}

		// Simulasi error di tengah transaksi
		return errStop
	})

	if !errors.Is(err, errStop) {
		t.Fatalf("WithTx: error = %v, seharusnya errStop", err)
	}

	var enrollCount int
	_ = pool.QueryRow(ctx, `SELECT COUNT(*) FROM enrollments WHERE course_id = $1 AND student_id = $2`, course.ID, studentID).Scan(&enrollCount)
	if enrollCount != 0 {
		t.Errorf("enrollments = %d, seharusnya 0 setelah rollback", enrollCount)
	}

	var events int
	_ = pool.QueryRow(ctx, `SELECT COUNT(*) FROM outbox WHERE (payload->>'user_id')::INT = $1`, student.ID).Scan(&events)
	if events != 0 {
		t.Errorf("baris outbox = %d, seharusnya 0 setelah rollback", events)
	}
}

func TestClaim_OnlyOnce(t *testing.T) {
	ctx := context.Background()
	student := seedStudent(t)
	repo := repository.NewNotificationRepository(pool)

	var outboxID int64
	err := pool.QueryRow(ctx,
		`INSERT INTO outbox (event_type, payload) VALUES ($1, jsonb_build_object('user_id', $2::INT)) RETURNING id`,
		model.EventEnrollmentCreated, student.ID).Scan(&outboxID)
	if err != nil {
		t.Fatalf("seed outbox: %v", err)
	}

	n := model.Notification{OutboxID: outboxID, UserID: student.ID, Channel: "email", Subject: "s", Body: "b"}
	first, err := repo.Claim(ctx, n)
	if err != nil || !first {
		t.Fatalf("claim pertama = %v, %v; seharusnya true, nil", first, err)
	}
	second, err := repo.Claim(ctx, n)
	if err != nil || second {
		t.Errorf("claim kedua = %v, %v; seharusnya false, nil", second, err)
	}
}
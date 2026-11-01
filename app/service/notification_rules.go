package service

import (
	"errors"
	"fmt"

	"api-students/app/model"
)

var errUnknownEvent = errors.New("jenis event tidak dikenal")

func BuildMessage(eventType string, e model.EnrollmentEvent, user model.User) (model.Message, error) {
	course := e.CourseCode + " " + e.CourseName
	var subject, body string

	switch eventType {
	case model.EventEnrollmentCreated:
		subject = "KRS: berhasil terdaftar di " + course
		body = fmt.Sprintf("Halo %s, Anda tercatat sebagai peserta kelas %s.", user.Email, course)
	case model.EventEnrollmentDropped:
		subject = "KRS: keluar dari " + course
		body = fmt.Sprintf("Halo %s, Anda sudah tidak terdaftar di kelas %s.", user.Email, course)
	case model.EventEnrollmentSwapped:
		from := e.FromCourseCode + " " + e.FromCourseName
		subject = "KRS: pindah dari " + from + " ke " + course
		body = fmt.Sprintf("Halo %s, Anda telah pindah dari kelas %s ke kelas %s.", user.Email, from, course)
	default:
		return model.Message{}, fmt.Errorf("%w: %q", errUnknownEvent, eventType)
	}

	return model.Message{To: user.Email, Subject: subject, Body: body}, nil
}
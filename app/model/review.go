package model

import "time"

type Reply struct {
	ID        string    `json:"id"`
	UserID    int       `json:"user_id"`
	Comment   string    `json:"comment"`
	CreatedAt time.Time `json:"created_at"`
}

type Review struct {
	ID        string    `json:"id"`
	CourseID  int       `json:"course_id"`
	UserID    int       `json:"user_id"`
	Rating    int       `json:"rating"`
	Comment   string    `json:"comment"`
	Tags      []string  `json:"tags"`
	Replies   []Reply   `json:"replies"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type CreateReviewRequest struct {
	Rating  int      `json:"rating"`
	Comment string   `json:"comment"`
	Tags    []string `json:"tags"`
}

// ReviewPage adalah response GET /courses/{id}/reviews (Langkah 10).
type ReviewPage struct {
	Success bool           `json:"success"`
	Message string         `json:"message"`
	Data    []Review       `json:"data"`
	Meta    ReviewPageMeta `json:"meta"`
}

type ReviewPageMeta struct {
	Limit      int    `json:"limit"`
	HasMore    bool   `json:"has_more"`
	NextBefore string `json:"next_before"`
}
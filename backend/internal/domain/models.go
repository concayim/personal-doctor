package domain

import "time"

type Patient struct {
	ID           string     `json:"id"`
	Name         string     `json:"name"`
	Gender       string     `json:"gender"`
	Birthday     string     `json:"birthday"`
	Phone        string     `json:"phone"`
	Allergies    string     `json:"allergies"`
	Notes        string     `json:"notes"`
	CreatedAt    time.Time  `json:"createdAt"`
	UpdatedAt    time.Time  `json:"updatedAt"`
	LastRecordAt *time.Time `json:"lastRecordAt,omitempty"`
}

type MedicalRecord struct {
	ID         string    `json:"id"`
	PatientID  string    `json:"patientId"`
	Kind       string    `json:"kind"`
	Title      string    `json:"title"`
	Content    string    `json:"content"`
	RecordedAt time.Time `json:"recordedAt"`
	CreatedAt  time.Time `json:"createdAt"`
}

type ChatMessage struct {
	ID        string    `json:"id"`
	PatientID string    `json:"patientId"`
	Role      string    `json:"role"`
	Content   string    `json:"content"`
	CreatedAt time.Time `json:"createdAt"`
}

type ChatInput struct {
	Patient Patient
	Records []MedicalRecord
	History []ChatMessage
	Message string
}

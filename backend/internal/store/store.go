package store

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/google/uuid"
	_ "modernc.org/sqlite"

	"personal-doctor/backend/internal/domain"
)

type Store struct {
	db *sql.DB
}

func Open(path string) (*Store, error) {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return nil, err
	}
	db, err := sql.Open("sqlite", path)
	if err != nil {
		return nil, err
	}
	db.SetMaxOpenConns(1)
	s := &Store{db: db}
	if err := s.migrate(context.Background()); err != nil {
		_ = db.Close()
		return nil, err
	}
	return s, nil
}

func (s *Store) Close() error {
	return s.db.Close()
}

func (s *Store) migrate(ctx context.Context) error {
	_, err := s.db.ExecContext(ctx, `
CREATE TABLE IF NOT EXISTS patients (
	id TEXT PRIMARY KEY,
	name TEXT NOT NULL,
	gender TEXT NOT NULL DEFAULT '',
	birthday TEXT NOT NULL DEFAULT '',
	phone TEXT NOT NULL DEFAULT '',
	allergies TEXT NOT NULL DEFAULT '',
	notes TEXT NOT NULL DEFAULT '',
	created_at DATETIME NOT NULL,
	updated_at DATETIME NOT NULL
);

CREATE TABLE IF NOT EXISTS medical_records (
	id TEXT PRIMARY KEY,
	patient_id TEXT NOT NULL,
	kind TEXT NOT NULL,
	title TEXT NOT NULL,
	content TEXT NOT NULL,
	recorded_at DATETIME NOT NULL,
	created_at DATETIME NOT NULL,
	FOREIGN KEY(patient_id) REFERENCES patients(id) ON DELETE CASCADE
);

CREATE TABLE IF NOT EXISTS chat_messages (
	id TEXT PRIMARY KEY,
	patient_id TEXT NOT NULL,
	role TEXT NOT NULL,
	content TEXT NOT NULL,
	created_at DATETIME NOT NULL,
	FOREIGN KEY(patient_id) REFERENCES patients(id) ON DELETE CASCADE
);
`)
	return err
}

func (s *Store) ListPatients(ctx context.Context) ([]domain.Patient, error) {
	rows, err := s.db.QueryContext(ctx, `
SELECT p.id, p.name, p.gender, p.birthday, p.phone, p.allergies, p.notes, p.created_at, p.updated_at, MAX(r.recorded_at)
FROM patients p
LEFT JOIN medical_records r ON r.patient_id = p.id
GROUP BY p.id
ORDER BY p.updated_at DESC`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var patients []domain.Patient
	for rows.Next() {
		patient, err := scanPatient(rows)
		if err != nil {
			return nil, err
		}
		patients = append(patients, patient)
	}
	return patients, rows.Err()
}

func (s *Store) CreatePatient(ctx context.Context, patient domain.Patient) (domain.Patient, error) {
	now := time.Now().UTC()
	patient.ID = uuid.NewString()
	patient.CreatedAt = now
	patient.UpdatedAt = now
	if patient.Name == "" {
		return patient, errors.New("name is required")
	}
	_, err := s.db.ExecContext(ctx, `
INSERT INTO patients (id, name, gender, birthday, phone, allergies, notes, created_at, updated_at)
VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		patient.ID, patient.Name, patient.Gender, patient.Birthday, patient.Phone, patient.Allergies, patient.Notes, patient.CreatedAt, patient.UpdatedAt)
	return patient, err
}

func (s *Store) GetPatient(ctx context.Context, id string) (domain.Patient, error) {
	row := s.db.QueryRowContext(ctx, `
SELECT p.id, p.name, p.gender, p.birthday, p.phone, p.allergies, p.notes, p.created_at, p.updated_at, MAX(r.recorded_at)
FROM patients p
LEFT JOIN medical_records r ON r.patient_id = p.id
WHERE p.id = ?
GROUP BY p.id`, id)
	return scanPatient(row)
}

func (s *Store) ListRecords(ctx context.Context, patientID string) ([]domain.MedicalRecord, error) {
	rows, err := s.db.QueryContext(ctx, `
SELECT id, patient_id, kind, title, content, recorded_at, created_at
FROM medical_records
WHERE patient_id = ?
ORDER BY recorded_at DESC, created_at DESC`, patientID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var records []domain.MedicalRecord
	for rows.Next() {
		var record domain.MedicalRecord
		if err := rows.Scan(&record.ID, &record.PatientID, &record.Kind, &record.Title, &record.Content, &record.RecordedAt, &record.CreatedAt); err != nil {
			return nil, err
		}
		records = append(records, record)
	}
	return records, rows.Err()
}

func (s *Store) CreateRecord(ctx context.Context, record domain.MedicalRecord) (domain.MedicalRecord, error) {
	if record.PatientID == "" {
		return record, errors.New("patientId is required")
	}
	if record.Title == "" || record.Content == "" {
		return record, errors.New("title and content are required")
	}
	if record.Kind == "" {
		record.Kind = "condition"
	}
	now := time.Now().UTC()
	record.ID = uuid.NewString()
	record.CreatedAt = now
	if record.RecordedAt.IsZero() {
		record.RecordedAt = now
	}
	_, err := s.db.ExecContext(ctx, `
INSERT INTO medical_records (id, patient_id, kind, title, content, recorded_at, created_at)
VALUES (?, ?, ?, ?, ?, ?, ?)`,
		record.ID, record.PatientID, record.Kind, record.Title, record.Content, record.RecordedAt, record.CreatedAt)
	if err != nil {
		return record, err
	}
	_, err = s.db.ExecContext(ctx, `UPDATE patients SET updated_at = ? WHERE id = ?`, now, record.PatientID)
	return record, err
}

func (s *Store) ListMessages(ctx context.Context, patientID string) ([]domain.ChatMessage, error) {
	rows, err := s.db.QueryContext(ctx, `
SELECT id, patient_id, role, content, created_at
FROM chat_messages
WHERE patient_id = ?
ORDER BY created_at ASC`, patientID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var messages []domain.ChatMessage
	for rows.Next() {
		var message domain.ChatMessage
		if err := rows.Scan(&message.ID, &message.PatientID, &message.Role, &message.Content, &message.CreatedAt); err != nil {
			return nil, err
		}
		messages = append(messages, message)
	}
	return messages, rows.Err()
}

func (s *Store) CreateMessage(ctx context.Context, message domain.ChatMessage) (domain.ChatMessage, error) {
	if message.PatientID == "" || message.Role == "" || message.Content == "" {
		return message, errors.New("patientId, role and content are required")
	}
	now := time.Now().UTC()
	message.ID = uuid.NewString()
	message.CreatedAt = now
	_, err := s.db.ExecContext(ctx, `
INSERT INTO chat_messages (id, patient_id, role, content, created_at)
VALUES (?, ?, ?, ?, ?)`, message.ID, message.PatientID, message.Role, message.Content, message.CreatedAt)
	if err != nil {
		return message, err
	}
	_, err = s.db.ExecContext(ctx, `UPDATE patients SET updated_at = ? WHERE id = ?`, now, message.PatientID)
	return message, err
}

type patientScanner interface {
	Scan(dest ...any) error
}

func scanPatient(scanner patientScanner) (domain.Patient, error) {
	var patient domain.Patient
	var lastRecordAt sql.NullString
	err := scanner.Scan(
		&patient.ID,
		&patient.Name,
		&patient.Gender,
		&patient.Birthday,
		&patient.Phone,
		&patient.Allergies,
		&patient.Notes,
		&patient.CreatedAt,
		&patient.UpdatedAt,
		&lastRecordAt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return patient, fmt.Errorf("patient not found")
		}
		return patient, err
	}
	if lastRecordAt.Valid {
		parsed, err := parseDBTime(lastRecordAt.String)
		if err == nil {
			patient.LastRecordAt = &parsed
		}
	}
	return patient, nil
}

func parseDBTime(value string) (time.Time, error) {
	layouts := []string{
		time.RFC3339Nano,
		"2006-01-02 15:04:05.999999999Z07:00",
		"2006-01-02 15:04:05.999999999-07:00",
		"2006-01-02 15:04:05.999999999",
		"2006-01-02 15:04:05",
	}
	for _, layout := range layouts {
		parsed, err := time.Parse(layout, value)
		if err == nil {
			return parsed, nil
		}
	}
	return time.Time{}, fmt.Errorf("parse database time %q", value)
}

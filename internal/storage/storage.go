package storage

import (
	"context"
	"university/internal/model"

	"github.com/jackc/pgx/v5"
)

type Repository struct {
	db *pgx.Conn
}

func NewRepository(db *pgx.Conn) *Repository {
	return &Repository{db: db}
}

func (r *Repository) getStudentByID(id string) (*model.StudentResponse, error) {

	query := `
	SELECT s.id, s.first_name, s.last_name, s.gender, s.birth_date, g.name
	FROM students s
	JOIN groups g ON s.group_id = g.id
	WHERE s.id = $1
	`

	var student model.StudentResponse

	err := r.db.QueryRow(
		context.Background(),
		query,
		id,
	).Scan(
		&student.ID,
		&student.FirstName,
		&student.LastName,
		&student.Gender,
		&student.BirthDate,
		&student.GroupName,
	)

	if err != nil {
		return nil, err
	}

	return &student, nil
}

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

func (r *Repository) GetStudentByID(id string) (*model.StudentResponse, error) {
	var err error

	query := `
	SELECT s.id, s.first_name, s.last_name, s.gender, s.birth_date, g.name
	FROM students s
	JOIN groups g ON s.group_id = g.id
	WHERE s.id = $1
	`

	var student model.StudentResponse

	err = r.db.QueryRow(
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

func (r *Repository) GetAllSchedules() ([]model.ScheduleResponse, error) {
	var err error

	query := `
	SELECT sc.id, f.name, g.name, s.name, sc.class_time
	FROM schedule sc
	JOIN faculties f ON sc.faculty_id = f.id
	JOIN groups g ON sc.group_id = g.id
	JOIN subjects s ON sc.subject_id = s.id
	`

	rows, err := r.db.Query(context.Background(), query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var schedules []model.ScheduleResponse

	for rows.Next() {
		var schedule model.ScheduleResponse
		err := rows.Scan(
			&schedule.ID,
			&schedule.Faculty,
			&schedule.Group,
			&schedule.Subject,
			&schedule.ClassTime,
		)
		if err != nil {
			return nil, err
		}
		schedules = append(schedules, schedule)
	}

	return schedules, nil
}

func (r *Repository) GetGroupSchedule(groupID string) ([]model.ScheduleResponse, error) {
	query := `
	SELECT sc.id, f.name, g.name, s.name, sc.class_time
	FROM schedule sc
	JOIN faculties f ON sc.faculty_id = f.id
	JOIN groups g ON sc.group_id = g.id
	JOIN subjects s ON sc.subject_id = s.id
	WHERE sc.group_id = $1
	`

	rows, err := r.db.Query(context.Background(), query, groupID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var schedules []model.ScheduleResponse
	for rows.Next() {
		var schedule model.ScheduleResponse
		err := rows.Scan(
			&schedule.ID,
			&schedule.Faculty,
			&schedule.Group,
			&schedule.Subject,
			&schedule.ClassTime,
		)
		if err != nil {
			return nil, err
		}
		schedules = append(schedules, schedule)
	}

	return schedules, nil
}

func (r *Repository) CreateAttendanceRecord(record *model.AttendanceRecord) error {
	query := `
	INSERT INTO attendance (student_id, subject_id, visit_day, visited)
	VALUES ($1, $2, $3, $4)
	`

	_, err := r.db.Exec(
		context.Background(),
		query,
		record.StudentID,
		record.SubjectID,
		record.VisitDay,
		record.Visited,
	)

	return err
}

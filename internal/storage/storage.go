package storage

import (
	"context"
	"fmt"
	"university/internal/model"

	"github.com/jackc/pgx/v5"
)

type Repository struct {
	db *pgx.Conn
}

func NewRepository(db *pgx.Conn) *Repository {
	return &Repository{db: db}
}

func (r *Repository) InitDB() error {
	query := `
    CREATE TABLE IF NOT EXISTS faculties (
        id SERIAL PRIMARY KEY,
        name VARCHAR(50) NOT NULL
    );

    CREATE TABLE IF NOT EXISTS groups (
        id SERIAL PRIMARY KEY,
        name VARCHAR(20) NOT NULL,
        faculty_id INT REFERENCES faculties(id)
    );

    CREATE TABLE IF NOT EXISTS users (
        id SERIAL PRIMARY KEY,
        email VARCHAR(100) UNIQUE NOT NULL,
        password_hash TEXT NOT NULL,
        is_active BOOLEAN DEFAULT TRUE,
        created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
    );

    CREATE TABLE IF NOT EXISTS roles (
        id SERIAL PRIMARY KEY,
        name VARCHAR(30) UNIQUE NOT NULL
    );

    CREATE TABLE IF NOT EXISTS user_roles (
        user_id INT REFERENCES users(id) ON DELETE CASCADE,
        role_id INT REFERENCES roles(id) ON DELETE CASCADE,
        PRIMARY KEY (user_id, role_id)
    );

    CREATE TABLE IF NOT EXISTS students (
        id SERIAL PRIMARY KEY,
        first_name VARCHAR(50),
        last_name VARCHAR(50),
        gender VARCHAR(10),
        birth_date DATE,
        group_id INT REFERENCES groups(id),
        user_id INT UNIQUE REFERENCES users(id) ON DELETE SET NULL
    );

    CREATE TABLE IF NOT EXISTS staff (
        id SERIAL PRIMARY KEY,
        user_id INT UNIQUE REFERENCES users(id) ON DELETE CASCADE,
        first_name VARCHAR(50),
        last_name VARCHAR(50),
        faculty_id INT REFERENCES faculties(id),
        position VARCHAR(50)
    );

    CREATE TABLE IF NOT EXISTS subjects (
        id SERIAL PRIMARY KEY,
        name VARCHAR(100) NOT NULL
    );

    CREATE TABLE IF NOT EXISTS schedule (
        id SERIAL PRIMARY KEY,
        faculty_id INT REFERENCES faculties(id),
        group_id INT REFERENCES groups(id),
        subject_id INT REFERENCES subjects(id),
        class_time VARCHAR(50)
    );

    CREATE TABLE IF NOT EXISTS attendance (
        id SERIAL PRIMARY KEY,
        student_id INT NOT NULL REFERENCES students(id),
        subject_id INT NOT NULL REFERENCES subjects(id),
        visit_day DATE NOT NULL,
        visited BOOLEAN NOT NULL
    );

    CREATE TABLE IF NOT EXISTS grades (
        id SERIAL PRIMARY KEY,
        student_id INT REFERENCES students(id) ON DELETE CASCADE,
        subject_id INT REFERENCES subjects(id) ON DELETE CASCADE,
        grade NUMERIC(4,2),
        graded_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
    );

    `

	_, err := r.db.Exec(context.Background(), query)
	if err != nil {
		return fmt.Errorf("failed to initialize database: %w", err)
	}

	return nil
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

func (r *Repository) GetAllStudents() ([]model.StudentListResponse, error) {
	query := `
	SELECT s.id, s.first_name, s.last_name,
	       g.name AS group,
	       u.email
	FROM students s
	LEFT JOIN groups g ON s.group_id = g.id
	LEFT JOIN users u ON s.user_id = u.id
	`

	rows, err := r.db.Query(context.Background(), query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var students []model.StudentListResponse
	for rows.Next() {
		var student model.StudentListResponse
		if err := rows.Scan(
			&student.ID,
			&student.FirstName,
			&student.LastName,
			&student.Group,
			&student.Email,
		); err != nil {
			return nil, err
		}
		students = append(students, student)
	}

	return students, rows.Err()
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

func (r *Repository) GetAttendanceRecordsByStudentID(studentID string) ([]model.AttendanceRecord, error) {
	query := `
	SELECT id, student_id, subject_id, visit_day, visited
	FROM attendance
	WHERE student_id = $1
	LIMIT 5
	`

	rows, err := r.db.Query(context.Background(), query, studentID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var records []model.AttendanceRecord
	for rows.Next() {
		var record model.AttendanceRecord
		err := rows.Scan(
			&record.ID,
			&record.StudentID,
			&record.SubjectID,
			&record.VisitDay,
			&record.Visited,
		)
		if err != nil {
			return nil, err
		}
		records = append(records, record)
	}

	return records, nil
}

func (r *Repository) GetAttendanceRecordsBySubjectID(subjectID string) ([]model.AttendanceRecord, error) {
	query := `
	SELECT id, student_id, subject_id, visit_day, visited
	FROM attendance
	WHERE subject_id = $1
	LIMIT 5
	`

	rows, err := r.db.Query(context.Background(), query, subjectID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var records []model.AttendanceRecord
	for rows.Next() {
		var record model.AttendanceRecord
		err := rows.Scan(
			&record.ID,
			&record.StudentID,
			&record.SubjectID,
			&record.VisitDay,
			&record.Visited,
		)
		if err != nil {
			return nil, err
		}
		records = append(records, record)
	}

	return records, nil
}

// GetUserByEmail retrieves a user by email
func (r *Repository) GetUserByEmail(email string) (*model.User, error) {
	query := `
	SELECT id, email, password_hash, is_active, created_at
	FROM users
	WHERE email = $1
	`

	var user model.User
	err := r.db.QueryRow(context.Background(), query, email).Scan(
		&user.ID,
		&user.Email,
		&user.PasswordHash,
		&user.IsActive,
		&user.CreatedAt,
	)

	if err != nil {
		return nil, err
	}

	return &user, nil
}

// CreateUser creates a new user account
func (r *Repository) CreateUser(email, passwordHash string) (*model.User, error) {
	query := `
	INSERT INTO users (email, password_hash)
	VALUES ($1, $2)
	RETURNING id, email, password_hash, is_active, created_at
	`

	var user model.User
	err := r.db.QueryRow(
		context.Background(),
		query,
		email,
		passwordHash,
	).Scan(
		&user.ID,
		&user.Email,
		&user.PasswordHash,
		&user.IsActive,
		&user.CreatedAt,
	)

	if err != nil {
		return nil, err
	}

	return &user, nil
}

// GetUserByID retrieves a user by ID with roles
func (r *Repository) GetUserByID(userID string) (*model.UserResponse, error) {
	query := `
	SELECT u.id, u.email, u.is_active, u.created_at
	FROM users u
	WHERE u.id = $1
	`

	var user model.UserResponse
	err := r.db.QueryRow(context.Background(), query, userID).Scan(
		&user.ID,
		&user.Email,
		&user.IsActive,
		&user.CreatedAt,
	)

	if err != nil {
		return nil, err
	}

	// Get user roles
	rolesQuery := `
	SELECT r.name
	FROM roles r
	JOIN user_roles ur ON r.id = ur.role_id
	WHERE ur.user_id = $1
	`

	rows, err := r.db.Query(context.Background(), rolesQuery, userID)
	if err != nil {
		user.Roles = []string{}
		return &user, nil
	}
	defer rows.Close()

	user.Roles = []string{}
	for rows.Next() {
		var roleName string
		if err := rows.Scan(&roleName); err != nil {
			continue
		}
		user.Roles = append(user.Roles, roleName)
	}

	return &user, nil
}

func (r *Repository) GetStudentsGPA() ([]model.StudentGPAResponse, error) {
	query := `
	SELECT s.id,
	       ROUND(AVG(g.grade)::NUMERIC, 2) AS gpa
	FROM students s
	INNER JOIN grades g ON g.student_id = s.id
	GROUP BY s.id
	`

	rows, err := r.db.Query(context.Background(), query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var results []model.StudentGPAResponse
	for rows.Next() {
		var result model.StudentGPAResponse
		if err := rows.Scan(&result.ID, &result.GPA); err != nil {
			return nil, err
		}
		results = append(results, result)
	}

	return results, rows.Err()
}

func (r *Repository) GetSubjectStats() ([]model.SubjectStatsResponse, error) {
	query := `
	SELECT sub.name,
	       COUNT(g.grade) AS graded_students,
	       ROUND(AVG(g.grade)::NUMERIC, 2) AS avg_grade
	FROM subjects sub
	INNER JOIN grades g ON g.subject_id = sub.id
	GROUP BY sub.name
	`

	rows, err := r.db.Query(context.Background(), query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var results []model.SubjectStatsResponse
	for rows.Next() {
		var result model.SubjectStatsResponse
		if err := rows.Scan(&result.Name, &result.GradedStudents, &result.AverageGrade); err != nil {
			return nil, err
		}
		results = append(results, result)
	}

	return results, rows.Err()
}

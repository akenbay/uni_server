package storage

import (
	"context"
	"fmt"
	"university/internal/model"

	"github.com/jackc/pgx/v5/pgxpool"
)

type Repository struct {
	pool *pgxpool.Pool
}

func NewRepository(pool *pgxpool.Pool) *Repository {
	return &Repository{pool: pool}
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

	_, err := r.pool.Exec(context.Background(), query)
	if err != nil {
		return fmt.Errorf("failed to initialize database: %w", err)
	}

	return nil
}

func (r *Repository) GetGroupIDByName(name string) (int, error) {
	var id int
	err := r.pool.QueryRow(context.Background(), `SELECT id FROM groups WHERE name = $1`, name).Scan(&id)
	return id, err
}

func (r *Repository) CreateStudent(req *model.CreateStudentRequest) (*model.StudentResponse, error) {
	groupID := req.GroupID
	if groupID == 0 && req.GroupName != "" {
		id, err := r.GetGroupIDByName(req.GroupName)
		if err != nil {
			return nil, fmt.Errorf("group not found: %s", req.GroupName)
		}
		groupID = id
	}

	query := `
	INSERT INTO students (first_name, last_name, gender, birth_date, group_id)
	VALUES ($1, $2, $3, $4, $5)
	RETURNING id, first_name, last_name, gender, COALESCE(birth_date::text, ''),
	          (SELECT name FROM groups WHERE id = $5)
	`

	var student model.StudentResponse
	err := r.pool.QueryRow(
		context.Background(),
		query,
		req.FirstName,
		req.LastName,
		req.Gender,
		req.BirthDate,
		groupID,
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

func (r *Repository) UpdateStudent(id string, req *model.UpdateStudentRequest) (*model.StudentResponse, error) {
	// Build dynamic update query
	query := `UPDATE students SET `
	args := []interface{}{}
	argNum := 1

	if req.FirstName != nil {
		query += fmt.Sprintf("first_name = $%d, ", argNum)
		args = append(args, *req.FirstName)
		argNum++
	}
	if req.LastName != nil {
		query += fmt.Sprintf("last_name = $%d, ", argNum)
		args = append(args, *req.LastName)
		argNum++
	}
	if req.Gender != nil {
		query += fmt.Sprintf("gender = $%d, ", argNum)
		args = append(args, *req.Gender)
		argNum++
	}
	if req.BirthDate != nil {
		query += fmt.Sprintf("birth_date = $%d, ", argNum)
		args = append(args, *req.BirthDate)
		argNum++
	}
	if req.GroupID != nil {
		query += fmt.Sprintf("group_id = $%d, ", argNum)
		args = append(args, *req.GroupID)
		argNum++
	}
	if req.GroupName != nil && *req.GroupName != "" {
		groupID, err := r.GetGroupIDByName(*req.GroupName)
		if err != nil {
			return nil, fmt.Errorf("group not found: %s", *req.GroupName)
		}
		query += fmt.Sprintf("group_id = $%d, ", argNum)
		args = append(args, groupID)
		argNum++
	}

	if len(args) == 0 {
		return r.GetStudentByID(id)
	}

	query = query[:len(query)-2]
	query += fmt.Sprintf(" WHERE id = $%d RETURNING id, first_name, last_name, gender, COALESCE(birth_date::text, ''), (SELECT name FROM groups g WHERE g.id = students.group_id)", argNum)
	args = append(args, id)

	var student model.StudentResponse
	err := r.pool.QueryRow(context.Background(), query, args...).Scan(
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

func (r *Repository) DeleteStudent(id string) error {
	query := `DELETE FROM students WHERE id = $1`
	_, err := r.pool.Exec(context.Background(), query, id)
	return err
}

func (r *Repository) GetStudentByID(id string) (*model.StudentResponse, error) {
	query := `
	SELECT s.id, s.first_name, s.last_name, s.gender, COALESCE(s.birth_date::text, ''), COALESCE(g.name, '')
	FROM students s
	LEFT JOIN groups g ON s.group_id = g.id
	WHERE s.id = $1
	`

	var student model.StudentResponse

	err := r.pool.QueryRow(
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
	       COALESCE(g.name, '') AS group_name,
	       COALESCE(u.email, '') AS email
	FROM students s
	LEFT JOIN groups g ON s.group_id = g.id
	LEFT JOIN users u ON s.user_id = u.id
	`

	rows, err := r.pool.Query(context.Background(), query)
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
			&student.GroupName,
			&student.Email,
		); err != nil {
			return nil, err
		}
		students = append(students, student)
	}

	return students, rows.Err()
}

func (r *Repository) CreateSchedule(req *model.CreateScheduleRequest) (*model.ScheduleResponse, error) {
	query := `
	INSERT INTO schedule (faculty_id, group_id, subject_id, class_time)
	VALUES ($1, $2, $3, $4)
	RETURNING id,
	          (SELECT name FROM faculties WHERE id = $1),
	          (SELECT name FROM groups WHERE id = $2),
	          (SELECT name FROM subjects WHERE id = $3),
	          class_time
	`

	var schedule model.ScheduleResponse
	err := r.pool.QueryRow(
		context.Background(),
		query,
		req.FacultyID,
		req.GroupID,
		req.SubjectID,
		req.ClassTime,
	).Scan(
		&schedule.ID,
		&schedule.Faculty,
		&schedule.Group,
		&schedule.Subject,
		&schedule.ClassTime,
	)
	if err != nil {
		return nil, err
	}
	return &schedule, nil
}

func (r *Repository) UpdateSchedule(id string, req *model.UpdateScheduleRequest) (*model.ScheduleResponse, error) {
	query := `SELECT faculty_id, group_id, subject_id, class_time FROM schedule WHERE id = $1`
	var facultyID, groupID, subjectID int
	var classTime string
	err := r.pool.QueryRow(context.Background(), query, id).Scan(&facultyID, &groupID, &subjectID, &classTime)
	if err != nil {
		return nil, err
	}

	if req.FacultyID != nil {
		facultyID = *req.FacultyID
	}
	if req.GroupID != nil {
		groupID = *req.GroupID
	}
	if req.SubjectID != nil {
		subjectID = *req.SubjectID
	}
	if req.ClassTime != nil {
		classTime = *req.ClassTime
	}

	updateQuery := `
	UPDATE schedule SET faculty_id = $1, group_id = $2, subject_id = $3, class_time = $4
	WHERE id = $5
	RETURNING id,
	          (SELECT name FROM faculties WHERE id = $1),
	          (SELECT name FROM groups WHERE id = $2),
	          (SELECT name FROM subjects WHERE id = $3),
	          class_time
	`
	var schedule model.ScheduleResponse
	err = r.pool.QueryRow(
		context.Background(),
		updateQuery,
		facultyID,
		groupID,
		subjectID,
		classTime,
		id,
	).Scan(
		&schedule.ID,
		&schedule.Faculty,
		&schedule.Group,
		&schedule.Subject,
		&schedule.ClassTime,
	)
	if err != nil {
		return nil, err
	}
	return &schedule, nil
}

func (r *Repository) DeleteSchedule(id string) error {
	query := `DELETE FROM schedule WHERE id = $1`
	_, err := r.pool.Exec(context.Background(), query, id)
	return err
}

func (r *Repository) GetScheduleByID(id string) (*model.ScheduleResponse, error) {
	query := `
	SELECT sc.id, f.name, g.name, s.name, sc.class_time
	FROM schedule sc
	JOIN faculties f ON sc.faculty_id = f.id
	JOIN groups g ON sc.group_id = g.id
	JOIN subjects s ON sc.subject_id = s.id
	WHERE sc.id = $1
	`
	var schedule model.ScheduleResponse
	err := r.pool.QueryRow(context.Background(), query, id).Scan(
		&schedule.ID,
		&schedule.Faculty,
		&schedule.Group,
		&schedule.Subject,
		&schedule.ClassTime,
	)
	if err != nil {
		return nil, err
	}
	return &schedule, nil
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

	rows, err := r.pool.Query(context.Background(), query)
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

	rows, err := r.pool.Query(context.Background(), query, groupID)
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

func (r *Repository) CreateFaculty(req *model.CreateFacultyRequest) (*model.FacultyResponse, error) {
	query := `INSERT INTO faculties (name) VALUES ($1) RETURNING id, name`
	var faculty model.FacultyResponse
	err := r.pool.QueryRow(context.Background(), query, req.Name).Scan(&faculty.ID, &faculty.Name)
	if err != nil {
		return nil, err
	}
	return &faculty, nil
}

func (r *Repository) GetFacultyByID(id string) (*model.FacultyResponse, error) {
	query := `SELECT id, name FROM faculties WHERE id = $1`
	var faculty model.FacultyResponse
	err := r.pool.QueryRow(context.Background(), query, id).Scan(&faculty.ID, &faculty.Name)
	if err != nil {
		return nil, err
	}
	return &faculty, nil
}

func (r *Repository) GetAllFaculties() ([]model.FacultyResponse, error) {
	query := `SELECT id, name FROM faculties ORDER BY id`
	rows, err := r.pool.Query(context.Background(), query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var faculties []model.FacultyResponse
	for rows.Next() {
		var f model.FacultyResponse
		if err := rows.Scan(&f.ID, &f.Name); err != nil {
			return nil, err
		}
		faculties = append(faculties, f)
	}
	return faculties, rows.Err()
}

func (r *Repository) CreateGroup(req *model.CreateGroupRequest) (*model.GroupResponse, error) {
	query := `
	INSERT INTO groups (name, faculty_id) VALUES ($1, $2)
	RETURNING id, name, faculty_id, COALESCE((SELECT name FROM faculties WHERE id = $2), '')
	`
	var group model.GroupResponse
	err := r.pool.QueryRow(context.Background(), query, req.Name, req.FacultyID).Scan(
		&group.ID, &group.Name, &group.FacultyID, &group.FacultyName,
	)
	if err != nil {
		return nil, err
	}
	return &group, nil
}

func (r *Repository) GetGroupByID(id string) (*model.GroupResponse, error) {
	query := `
	SELECT g.id, g.name, g.faculty_id, COALESCE(f.name, '') FROM groups g
	LEFT JOIN faculties f ON g.faculty_id = f.id
	WHERE g.id = $1
	`
	var group model.GroupResponse
	err := r.pool.QueryRow(context.Background(), query, id).Scan(
		&group.ID, &group.Name, &group.FacultyID, &group.FacultyName,
	)
	if err != nil {
		return nil, err
	}
	return &group, nil
}

func (r *Repository) GetAllGroups() ([]model.GroupResponse, error) {
	query := `
	SELECT g.id, g.name, g.faculty_id, COALESCE(f.name, '') FROM groups g
	LEFT JOIN faculties f ON g.faculty_id = f.id
	ORDER BY g.id
	`
	rows, err := r.pool.Query(context.Background(), query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var groups []model.GroupResponse
	for rows.Next() {
		var g model.GroupResponse
		if err := rows.Scan(&g.ID, &g.Name, &g.FacultyID, &g.FacultyName); err != nil {
			return nil, err
		}
		groups = append(groups, g)
	}
	return groups, rows.Err()
}

func (r *Repository) CreateSubject(req *model.CreateSubjectRequest) (*model.SubjectResponse, error) {
	query := `INSERT INTO subjects (name) VALUES ($1) RETURNING id, name`
	var subject model.SubjectResponse
	err := r.pool.QueryRow(context.Background(), query, req.Name).Scan(&subject.ID, &subject.Name)
	if err != nil {
		return nil, err
	}
	return &subject, nil
}

func (r *Repository) GetSubjectByID(id string) (*model.SubjectResponse, error) {
	query := `SELECT id, name FROM subjects WHERE id = $1`
	var subject model.SubjectResponse
	err := r.pool.QueryRow(context.Background(), query, id).Scan(&subject.ID, &subject.Name)
	if err != nil {
		return nil, err
	}
	return &subject, nil
}

func (r *Repository) GetAllSubjects() ([]model.SubjectResponse, error) {
	query := `SELECT id, name FROM subjects ORDER BY id`
	rows, err := r.pool.Query(context.Background(), query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var subjects []model.SubjectResponse
	for rows.Next() {
		var s model.SubjectResponse
		if err := rows.Scan(&s.ID, &s.Name); err != nil {
			return nil, err
		}
		subjects = append(subjects, s)
	}
	return subjects, rows.Err()
}

func (r *Repository) CreateAttendanceRecord(req *model.CreateAttendanceRequest) (*model.AttendanceRecord, error) {
	query := `
	INSERT INTO attendance (student_id, subject_id, visit_day, visited)
	VALUES ($1, $2, $3, $4)
	RETURNING id, student_id, subject_id, visit_day, visited
	`

	var created model.AttendanceRecord
	err := r.pool.QueryRow(
		context.Background(),
		query,
		req.StudentID,
		req.SubjectID,
		req.VisitDay,
		req.Visited,
	).Scan(
		&created.ID,
		&created.StudentID,
		&created.SubjectID,
		&created.VisitDay,
		&created.Visited,
	)
	if err != nil {
		return nil, err
	}
	return &created, nil
}

func (r *Repository) UpdateAttendanceRecord(id string, req *model.UpdateAttendanceRequest) (*model.AttendanceRecord, error) {
	query := `SELECT student_id, subject_id, visit_day, visited FROM attendance WHERE id = $1`
	var studentID, subjectID int
	var visitDay string
	var visited bool
	err := r.pool.QueryRow(context.Background(), query, id).Scan(&studentID, &subjectID, &visitDay, &visited)
	if err != nil {
		return nil, err
	}

	if req.StudentID != nil {
		studentID = *req.StudentID
	}
	if req.SubjectID != nil {
		subjectID = *req.SubjectID
	}
	if req.VisitDay != nil {
		visitDay = *req.VisitDay
	}
	if req.Visited != nil {
		visited = *req.Visited
	}

	updateQuery := `
	UPDATE attendance SET student_id = $1, subject_id = $2, visit_day = $3, visited = $4
	WHERE id = $5
	RETURNING id, student_id, subject_id, visit_day, visited
	`
	var record model.AttendanceRecord
	err = r.pool.QueryRow(
		context.Background(),
		updateQuery,
		studentID,
		subjectID,
		visitDay,
		visited,
		id,
	).Scan(
		&record.ID,
		&record.StudentID,
		&record.SubjectID,
		&record.VisitDay,
		&record.Visited,
	)
	if err != nil {
		return nil, err
	}
	return &record, nil
}

func (r *Repository) DeleteAttendanceRecord(id string) error {
	query := `DELETE FROM attendance WHERE id = $1`
	_, err := r.pool.Exec(context.Background(), query, id)
	return err
}

func (r *Repository) GetAttendanceByID(id string) (*model.AttendanceRecord, error) {
	query := `SELECT id, student_id, subject_id, visit_day, visited FROM attendance WHERE id = $1`
	var record model.AttendanceRecord
	err := r.pool.QueryRow(context.Background(), query, id).Scan(
		&record.ID,
		&record.StudentID,
		&record.SubjectID,
		&record.VisitDay,
		&record.Visited,
	)
	if err != nil {
		return nil, err
	}
	return &record, nil
}

func (r *Repository) GetAllAttendanceRecords() ([]model.AttendanceRecord, error) {
	query := `SELECT id, student_id, subject_id, visit_day, visited FROM attendance`
	rows, err := r.pool.Query(context.Background(), query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var records []model.AttendanceRecord
	for rows.Next() {
		var record model.AttendanceRecord
		if err := rows.Scan(
			&record.ID,
			&record.StudentID,
			&record.SubjectID,
			&record.VisitDay,
			&record.Visited,
		); err != nil {
			return nil, err
		}
		records = append(records, record)
	}
	return records, rows.Err()
}

func (r *Repository) GetAttendanceRecordsByStudentID(studentID string) ([]model.AttendanceRecord, error) {
	query := `
	SELECT id, student_id, subject_id, visit_day, visited
	FROM attendance
	WHERE student_id = $1
	LIMIT 5
	`

	rows, err := r.pool.Query(context.Background(), query, studentID)
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

	rows, err := r.pool.Query(context.Background(), query, subjectID)
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
	err := r.pool.QueryRow(context.Background(), query, email).Scan(
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
	err := r.pool.QueryRow(
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
	err := r.pool.QueryRow(context.Background(), query, userID).Scan(
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

	rows, err := r.pool.Query(context.Background(), rolesQuery, userID)
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

	rows, err := r.pool.Query(context.Background(), query)
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

	rows, err := r.pool.Query(context.Background(), query)
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

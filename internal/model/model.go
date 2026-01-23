package model

import "time"

type StudentResponse struct {
	ID        int    `json:"id"`
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
	Gender    string `json:"gender"`
	BirthDate string `json:"birth_date"`
	GroupName string `json:"group_name"`
}

type StudentListResponse struct {
	ID        int    `json:"id"`
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
	Group     string `json:"group"`
	Email     string `json:"email"`
}

type StudentGPAResponse struct {
	ID  int     `json:"id"`
	GPA float64 `json:"gpa"`
}

type SubjectStatsResponse struct {
	Name           string  `json:"name"`
	GradedStudents int     `json:"graded_students"`
	AverageGrade   float64 `json:"avg_grade"`
}

type Subject struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
}

type ScheduleResponse struct {
	ID        int    `json:"id"`
	Faculty   string `json:"faculty"`
	Group     string `json:"group"`
	Subject   string `json:"subject"`
	ClassTime string `json:"class_time"`
}

type AttendanceRecord struct {
	ID        int    `json:"id"`
	StudentID int    `json:"student_id"`
	SubjectID int    `json:"subject_id"`
	VisitDay  string `json:"visit_day"`
	Visited   bool   `json:"visited"`
}

// User represents a user account
type User struct {
	ID           int       `json:"id"`
	Email        string    `json:"email"`
	PasswordHash string    `json:"-"`
	IsActive     bool      `json:"is_active"`
	CreatedAt    time.Time `json:"created_at"`
}

// AuthRequest is the payload for both registration and login
type AuthRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

// LoginResponse is returned after successful login
type LoginResponse struct {
	Token string `json:"token"`
	User  *User  `json:"user"`
}

// UserResponse is the user info returned in /api/users/me
type UserResponse struct {
	ID        int       `json:"id"`
	Email     string    `json:"email"`
	IsActive  bool      `json:"is_active"`
	CreatedAt time.Time `json:"created_at"`
	Roles     []string  `json:"roles"`
}

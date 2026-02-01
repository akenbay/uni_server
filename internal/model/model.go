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
	GroupName string `json:"group_name"`
	Email     string `json:"email"`
}

type CreateStudentRequest struct {
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
	Gender    string `json:"gender"`
	BirthDate string `json:"birth_date"`
	GroupID   int    `json:"group_id"`
	GroupName string `json:"group_name,omitempty"` // optional: used when group_id is 0
}

type UpdateStudentRequest struct {
	FirstName *string `json:"first_name,omitempty"`
	LastName  *string `json:"last_name,omitempty"`
	Gender    *string `json:"gender,omitempty"`
	BirthDate *string `json:"birth_date,omitempty"`
	GroupID   *int    `json:"group_id,omitempty"`
	GroupName *string `json:"group_name,omitempty"`
}

type CreateScheduleRequest struct {
	FacultyID int    `json:"faculty_id"`
	GroupID   int    `json:"group_id"`
	SubjectID int    `json:"subject_id"`
	ClassTime string `json:"class_time"`
}

type UpdateScheduleRequest struct {
	FacultyID *int    `json:"faculty_id,omitempty"`
	GroupID   *int    `json:"group_id,omitempty"`
	SubjectID *int    `json:"subject_id,omitempty"`
	ClassTime *string `json:"class_time,omitempty"`
}

type CreateAttendanceRequest struct {
	StudentID int    `json:"student_id"`
	SubjectID int    `json:"subject_id"`
	VisitDay  string `json:"visit_day"`
	Visited   bool   `json:"visited"`
}

type UpdateAttendanceRequest struct {
	StudentID *int    `json:"student_id,omitempty"`
	SubjectID *int    `json:"subject_id,omitempty"`
	VisitDay  *string `json:"visit_day,omitempty"`
	Visited   *bool   `json:"visited,omitempty"`
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

// Faculty response and create request
type FacultyResponse struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
}

type CreateFacultyRequest struct {
	Name string `json:"name"`
}

// Group response and create request
type GroupResponse struct {
	ID          int    `json:"id"`
	Name        string `json:"name"`
	FacultyID   int    `json:"faculty_id"`
	FacultyName string `json:"faculty_name,omitempty"`
}

type CreateGroupRequest struct {
	Name      string `json:"name"`
	FacultyID int    `json:"faculty_id"`
}

// Subject response and create request
type SubjectResponse struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
}

type CreateSubjectRequest struct {
	Name string `json:"name"`
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

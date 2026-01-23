package service

import (
	"errors"
	"fmt"
	"os"
	"regexp"
	"time"
	"university/internal/model"
	"university/internal/storage"

	"github.com/golang-jwt/jwt/v5"
	"github.com/jackc/pgx/v5"
	"golang.org/x/crypto/bcrypt"
)

type Service struct {
	repo      *storage.Repository
	jwtSecret string
}

func NewService(repo *storage.Repository) *Service {
	jwtSecret := os.Getenv("JWT_SECRET")
	if jwtSecret == "" {
		jwtSecret = "secretkey0909"
	}

	return &Service{
		repo:      repo,
		jwtSecret: jwtSecret,
	}
}

func (s *Service) GetStudentByID(id string) (*model.StudentResponse, error) {
	return s.repo.GetStudentByID(id)
}

func (s *Service) GetAllStudents() ([]model.StudentListResponse, error) {
	return s.repo.GetAllStudents()
}

func (s *Service) GetStudentsGPA() ([]model.StudentGPAResponse, error) {
	return s.repo.GetStudentsGPA()
}

func (s *Service) GetSubjectStats() ([]model.SubjectStatsResponse, error) {
	return s.repo.GetSubjectStats()
}

func (s *Service) GetAllSchedules() ([]model.ScheduleResponse, error) {
	return s.repo.GetAllSchedules()
}

func (s *Service) GetGroupSchedule(groupID string) ([]model.ScheduleResponse, error) {
	return s.repo.GetGroupSchedule(groupID)
}

func (s *Service) CreateAttendanceRecord(record *model.AttendanceRecord) error {
	return s.repo.CreateAttendanceRecord(record)
}

func (s *Service) GetAttendanceRecordsByStudentID(studentID string) ([]model.AttendanceRecord, error) {
	return s.repo.GetAttendanceRecordsByStudentID(studentID)
}

func (s *Service) GetAttendanceRecordsBySubjectID(subjectID string) ([]model.AttendanceRecord, error) {
	return s.repo.GetAttendanceRecordsBySubjectID(subjectID)
}

// Register creates a new user account
func (s *Service) Register(req *model.AuthRequest) (*model.User, error) {
	// Validate email format
	if !isValidEmail(req.Email) {
		return nil, errors.New("invalid email format")
	}

	// Validate password length
	if len(req.Password) < 6 {
		return nil, errors.New("password must be at least 6 characters")
	}

	// Check if user already exists
	_, err := s.repo.GetUserByEmail(req.Email)
	if err == nil {
		return nil, errors.New("user already exists")
	}
	if err != pgx.ErrNoRows {
		return nil, err
	}

	// Hash password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, fmt.Errorf("failed to hash password: %w", err)
	}

	// Create user
	user, err := s.repo.CreateUser(req.Email, string(hashedPassword))
	if err != nil {
		return nil, fmt.Errorf("failed to create user: %w", err)
	}

	return user, nil
}

// Login authenticates user and returns JWT token
func (s *Service) Login(req *model.AuthRequest) (*model.LoginResponse, error) {
	// Get user by email
	user, err := s.repo.GetUserByEmail(req.Email)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, errors.New("invalid email or password")
		}
		return nil, fmt.Errorf("failed to get user: %w", err)
	}

	// Check if user is active
	if !user.IsActive {
		return nil, errors.New("user account is inactive")
	}

	// Compare passwords
	err = bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(req.Password))
	if err != nil {
		return nil, errors.New("invalid email or password")
	}

	// Generate JWT token
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"user_id": user.ID,
		"email":   user.Email,
		"exp":     time.Now().Add(24 * time.Hour).Unix(),
		"iat":     time.Now().Unix(),
	})

	tokenString, err := token.SignedString([]byte(s.jwtSecret))
	if err != nil {
		return nil, fmt.Errorf("failed to sign token: %w", err)
	}

	return &model.LoginResponse{
		Token: tokenString,
		User:  user,
	}, nil
}

// GetCurrentUser retrieves user info by ID
func (s *Service) GetCurrentUser(userID string) (*model.UserResponse, error) {
	return s.repo.GetUserByID(userID)
}

// ValidateToken validates JWT token and returns user ID as string
func (s *Service) ValidateToken(tokenString string) (string, error) {
	token, err := jwt.ParseWithClaims(tokenString, &jwt.MapClaims{}, func(token *jwt.Token) (interface{}, error) {
		// Verify signing method
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(s.jwtSecret), nil
	})

	if err != nil {
		return "", fmt.Errorf("failed to parse token: %w", err)
	}

	claims, ok := token.Claims.(*jwt.MapClaims)
	if !ok || !token.Valid {
		return "", errors.New("invalid token claims")
	}

	userID, ok := (*claims)["user_id"].(float64)
	if !ok {
		return "", errors.New("invalid user_id in token")
	}

	return fmt.Sprintf("%d", int(userID)), nil
}

// isValidEmail validates email format using regex
func isValidEmail(email string) bool {
	pattern := `^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`
	re := regexp.MustCompile(pattern)
	return re.MatchString(email)
}

package service

import (
	"university/internal/model"
	"university/internal/storage"
)

type Service struct {
	repo *storage.Repository
}

func NewService(repo *storage.Repository) *Service {
	return &Service{repo: repo}
}

func (s *Service) GetStudentByID(id string) (*model.StudentResponse, error) {
	return s.repo.GetStudentByID(id)
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

func (s *Service) GetAttendanceRecordsByStudentID(studentID int) ([]model.AttendanceRecord, error) {
	return s.repo.GetAttendanceRecordsByStudentID(studentID)
}

func (s *Service) GetAttendanceRecordsBySubjectID(subjectID int) ([]model.AttendanceRecord, error) {
	return s.repo.GetAttendanceRecordsBySubjectID(subjectID)
}

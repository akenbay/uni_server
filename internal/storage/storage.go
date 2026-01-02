package storage

import "university/internal/model"

type Repository interface {
	GetStudentByID(id string) (*model.StudentResponse, error)
	GetAllSchedules() ([]model.ScheduleResponse, error)
	GetGroupSchedule(groupID string) ([]model.ScheduleResponse, error)
}

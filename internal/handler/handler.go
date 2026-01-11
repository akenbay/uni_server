package handler

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/jackc/pgx/v5"
	"github.com/labstack/echo/v4"

	"university/internal/model"
	"university/internal/service"
)

type Handler struct {
	service *service.Service
}

func NewHandler(service *service.Service) *Handler {
	return &Handler{service: service}
}

// Register registers HTTP routes on the provided Echo instance.
func (h *Handler) Register(e *echo.Echo) {
	e.GET("/student/:id", h.GetStudentByID)
	e.GET("/all_class_schedule", h.GetAllSchedules)
	e.GET("/schedule/group/:id", h.GetGroupSchedule)
}

func (h *Handler) GetStudentByID(c echo.Context) error {
	id := c.Param("id")
	if id == "" {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "id is required"})
	}

	student, err := h.service.GetStudentByID(id)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return c.JSON(http.StatusNotFound, map[string]string{"error": "student not found"})
		}
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}

	return c.JSON(http.StatusOK, student)
}

func (h *Handler) GetAllSchedules(c echo.Context) error {
	schedules, err := h.service.GetAllSchedules()
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}
	return c.JSON(http.StatusOK, schedules)
}

func (h *Handler) GetGroupSchedule(c echo.Context) error {
	groupID := c.Param("id")
	if groupID == "" {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "group id is required"})
	}

	schedules, err := h.service.GetGroupSchedule(groupID)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}
	return c.JSON(http.StatusOK, schedules)
}

func (h *Handler) CreateAttendanceRecord(c echo.Context) error {
	var record model.AttendanceRecord
	if err := c.Bind(&record); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid request body"})
	}

	err := h.service.CreateAttendanceRecord(&record)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}

	return c.JSON(http.StatusOK, map[string]string{"message": "attendance record created successfully"})
}

func (h *Handler) GetAttendanceRecordsByStudentID(c echo.Context) error {
	studentIDParam := c.Param("id")
	if studentIDParam == "" {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "id is required"})
	}
	studentID, err := strconv.Atoi(studentIDParam)
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid student id"})
	}

	records, err := h.service.GetAttendanceRecordsByStudentID(studentID)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}

	return c.JSON(http.StatusOK, records)
}

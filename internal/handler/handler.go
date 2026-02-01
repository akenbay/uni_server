package handler

import (
	"errors"
	"net/http"

	"github.com/jackc/pgx/v5"
	"github.com/labstack/echo/v4"

	"university/internal/middleware"
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
	// Public auth routes
	e.POST("/api/auth/register", h.Register_User)
	e.POST("/api/auth/login", h.Login)

	// Protected routes
	e.GET("/api/users/me", h.GetCurrentUser, middleware.AuthMiddleware(h.service))

	// Public student/schedule routes
	e.GET("/student/:id", h.GetStudentByID)
	e.GET("/students", h.GetAllStudents)
	e.POST("/students", h.CreateStudent)
	e.PATCH("/students/:id", h.UpdateStudent)
	e.DELETE("/students/:id", h.DeleteStudent)
	e.GET("/students/gpa", h.GetStudentsGPA)
	e.GET("/subjects/stats", h.GetSubjectStats)
	e.POST("/faculties", h.CreateFaculty)
	e.GET("/faculties", h.GetAllFaculties)
	e.GET("/faculties/:id", h.GetFacultyByID)
	e.POST("/groups", h.CreateGroup)
	e.GET("/groups", h.GetAllGroups)
	e.GET("/groups/:id", h.GetGroupByID)
	e.POST("/subjects", h.CreateSubject)
	e.GET("/subjects", h.GetAllSubjects)
	e.GET("/subjects/:id", h.GetSubjectByID)
	e.GET("/all_class_schedule", h.GetAllSchedules)
	e.GET("/schedule/group/:id", h.GetGroupSchedule)
	e.GET("/schedule/:id", h.GetScheduleByID)
	e.POST("/schedule", h.CreateSchedule)
	e.PATCH("/schedule/:id", h.UpdateSchedule)
	e.DELETE("/schedule/:id", h.DeleteSchedule)
	e.GET("/attendance", h.GetAllAttendanceRecords)
	e.GET("/attendance/student/:id", h.GetAttendanceRecordsByStudentID)
	e.GET("/attendance/subject/:id", h.GetAttendanceRecordsBySubjectID)
	e.GET("/attendance/:id", h.GetAttendanceByID)
	e.POST("/attendance", h.CreateAttendanceRecord)
	e.PATCH("/attendance/:id", h.UpdateAttendanceRecord)
	e.DELETE("/attendance/:id", h.DeleteAttendanceRecord)
}

// GetStudentByID godoc
// @Summary      Get student by ID
// @Tags         students
// @Param        id   path      string  true  "Student ID"
// @Success      200  {object}  model.StudentResponse
// @Failure      404  {object}  map[string]string
// @Router       /student/{id} [get]
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

// GetAllStudents godoc
// @Summary      Get all students
// @Tags         students
// @Success      200  {array}   model.StudentListResponse
// @Router       /students [get]
func (h *Handler) GetAllStudents(c echo.Context) error {
	students, err := h.service.GetAllStudents()
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}
	return c.JSON(http.StatusOK, students)
}

// CreateStudent godoc
// @Summary      Create a student
// @Tags         students
// @Accept       json
// @Param        body  body  model.CreateStudentRequest  true  "Student data"
// @Success      201   {object}  model.StudentResponse
// @Router       /students [post]
func (h *Handler) CreateStudent(c echo.Context) error {
	var req model.CreateStudentRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid request body"})
	}

	student, err := h.service.CreateStudent(&req)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}
	return c.JSON(http.StatusCreated, student)
}

// UpdateStudent godoc
// @Summary      Update a student
// @Tags         students
// @Param        id    path      string  true  "Student ID"
// @Param        body  body      model.UpdateStudentRequest  true  "Update data"
// @Success      200   {object}  model.StudentResponse
// @Router       /students/{id} [patch]
func (h *Handler) UpdateStudent(c echo.Context) error {
	id := c.Param("id")
	if id == "" {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "id is required"})
	}

	var req model.UpdateStudentRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid request body"})
	}

	student, err := h.service.UpdateStudent(id, &req)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return c.JSON(http.StatusNotFound, map[string]string{"error": "student not found"})
		}
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}
	return c.JSON(http.StatusOK, student)
}

// DeleteStudent godoc
// @Summary      Delete a student
// @Tags         students
// @Param        id   path  string  true  "Student ID"
// @Success      204
// @Router       /students/{id} [delete]
func (h *Handler) DeleteStudent(c echo.Context) error {
	id := c.Param("id")
	if id == "" {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "id is required"})
	}

	err := h.service.DeleteStudent(id)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}
	return c.NoContent(http.StatusNoContent)
}

func (h *Handler) GetStudentsGPA(c echo.Context) error {
	gpaList, err := h.service.GetStudentsGPA()
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}
	return c.JSON(http.StatusOK, gpaList)
}

func (h *Handler) GetSubjectStats(c echo.Context) error {
	stats, err := h.service.GetSubjectStats()
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}
	return c.JSON(http.StatusOK, stats)
}

func (h *Handler) CreateFaculty(c echo.Context) error {
	var req model.CreateFacultyRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid request body"})
	}
	faculty, err := h.service.CreateFaculty(&req)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}
	return c.JSON(http.StatusCreated, faculty)
}

func (h *Handler) GetAllFaculties(c echo.Context) error {
	faculties, err := h.service.GetAllFaculties()
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}
	return c.JSON(http.StatusOK, faculties)
}

func (h *Handler) GetFacultyByID(c echo.Context) error {
	id := c.Param("id")
	if id == "" {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "id is required"})
	}
	faculty, err := h.service.GetFacultyByID(id)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return c.JSON(http.StatusNotFound, map[string]string{"error": "faculty not found"})
		}
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}
	return c.JSON(http.StatusOK, faculty)
}

func (h *Handler) CreateGroup(c echo.Context) error {
	var req model.CreateGroupRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid request body"})
	}
	group, err := h.service.CreateGroup(&req)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}
	return c.JSON(http.StatusCreated, group)
}

func (h *Handler) GetAllGroups(c echo.Context) error {
	groups, err := h.service.GetAllGroups()
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}
	return c.JSON(http.StatusOK, groups)
}

func (h *Handler) GetGroupByID(c echo.Context) error {
	id := c.Param("id")
	if id == "" {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "id is required"})
	}
	group, err := h.service.GetGroupByID(id)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return c.JSON(http.StatusNotFound, map[string]string{"error": "group not found"})
		}
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}
	return c.JSON(http.StatusOK, group)
}

func (h *Handler) CreateSubject(c echo.Context) error {
	var req model.CreateSubjectRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid request body"})
	}
	subject, err := h.service.CreateSubject(&req)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}
	return c.JSON(http.StatusCreated, subject)
}

func (h *Handler) GetAllSubjects(c echo.Context) error {
	subjects, err := h.service.GetAllSubjects()
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}
	return c.JSON(http.StatusOK, subjects)
}

func (h *Handler) GetSubjectByID(c echo.Context) error {
	id := c.Param("id")
	if id == "" {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "id is required"})
	}
	subject, err := h.service.GetSubjectByID(id)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return c.JSON(http.StatusNotFound, map[string]string{"error": "subject not found"})
		}
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}
	return c.JSON(http.StatusOK, subject)
}

// GetAllSchedules godoc
// @Summary      Get all schedules
// @Tags         schedules
// @Success      200  {array}  model.ScheduleResponse
// @Router       /all_class_schedule [get]
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

func (h *Handler) GetScheduleByID(c echo.Context) error {
	id := c.Param("id")
	if id == "" {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "id is required"})
	}

	schedule, err := h.service.GetScheduleByID(id)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return c.JSON(http.StatusNotFound, map[string]string{"error": "schedule not found"})
		}
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}
	return c.JSON(http.StatusOK, schedule)
}

func (h *Handler) CreateSchedule(c echo.Context) error {
	var req model.CreateScheduleRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid request body"})
	}

	schedule, err := h.service.CreateSchedule(&req)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}
	return c.JSON(http.StatusCreated, schedule)
}

func (h *Handler) UpdateSchedule(c echo.Context) error {
	id := c.Param("id")
	if id == "" {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "id is required"})
	}

	var req model.UpdateScheduleRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid request body"})
	}

	schedule, err := h.service.UpdateSchedule(id, &req)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return c.JSON(http.StatusNotFound, map[string]string{"error": "schedule not found"})
		}
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}
	return c.JSON(http.StatusOK, schedule)
}

func (h *Handler) DeleteSchedule(c echo.Context) error {
	id := c.Param("id")
	if id == "" {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "id is required"})
	}

	err := h.service.DeleteSchedule(id)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}
	return c.NoContent(http.StatusNoContent)
}

// GetAllAttendanceRecords godoc
// @Summary      Get all attendance records
// @Tags         attendance
// @Success      200  {array}  model.AttendanceRecord
// @Router       /attendance [get]
func (h *Handler) GetAllAttendanceRecords(c echo.Context) error {
	records, err := h.service.GetAllAttendanceRecords()
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}
	return c.JSON(http.StatusOK, records)
}

func (h *Handler) GetAttendanceByID(c echo.Context) error {
	id := c.Param("id")
	if id == "" {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "id is required"})
	}

	record, err := h.service.GetAttendanceByID(id)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return c.JSON(http.StatusNotFound, map[string]string{"error": "attendance record not found"})
		}
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}
	return c.JSON(http.StatusOK, record)
}

// CreateAttendanceRecord godoc
// @Summary      Create attendance record
// @Tags         attendance
// @Accept       json
// @Param        body  body  model.CreateAttendanceRequest  true  "Attendance data"
// @Success      201   {object}  model.AttendanceRecord
// @Router       /attendance [post]
func (h *Handler) CreateAttendanceRecord(c echo.Context) error {
	var req model.CreateAttendanceRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid request body"})
	}

	created, err := h.service.CreateAttendanceRecord(&req)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}
	return c.JSON(http.StatusCreated, created)
}

func (h *Handler) UpdateAttendanceRecord(c echo.Context) error {
	id := c.Param("id")
	if id == "" {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "id is required"})
	}

	var req model.UpdateAttendanceRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid request body"})
	}

	record, err := h.service.UpdateAttendanceRecord(id, &req)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return c.JSON(http.StatusNotFound, map[string]string{"error": "attendance record not found"})
		}
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}
	return c.JSON(http.StatusOK, record)
}

func (h *Handler) DeleteAttendanceRecord(c echo.Context) error {
	id := c.Param("id")
	if id == "" {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "id is required"})
	}

	err := h.service.DeleteAttendanceRecord(id)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}
	return c.NoContent(http.StatusNoContent)
}

// Register_User godoc
// @Summary      Register a new user
// @Tags         auth
// @Accept       json
// @Param        body  body  model.AuthRequest  true  "Email and password"
// @Success      201   {object}  map[string]interface{}
// @Router       /api/auth/register [post]
func (h *Handler) Register_User(c echo.Context) error {
	var req model.AuthRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid request body"})
	}

	user, err := h.service.Register(&req)
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": err.Error()})
	}

	return c.JSON(http.StatusCreated, map[string]interface{}{
		"id":    user.ID,
		"email": user.Email,
	})
}

// Login godoc
// @Summary      Login
// @Tags         auth
// @Accept       json
// @Param        body  body  model.AuthRequest  true  "Email and password"
// @Success      200   {object}  model.LoginResponse
// @Router       /api/auth/login [post]
func (h *Handler) Login(c echo.Context) error {
	var req model.AuthRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid request body"})
	}

	response, err := h.service.Login(&req)
	if err != nil {
		return c.JSON(http.StatusUnauthorized, map[string]string{"error": err.Error()})
	}

	return c.JSON(http.StatusOK, response)
}

// GetCurrentUser returns current user info (protected endpoint)
func (h *Handler) GetCurrentUser(c echo.Context) error {
	userID, ok := c.Get("user_id").(string)
	if !ok {
		return c.JSON(http.StatusUnauthorized, map[string]string{"error": "invalid user context"})
	}

	user, err := h.service.GetCurrentUser(userID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return c.JSON(http.StatusNotFound, map[string]string{"error": "user not found"})
		}
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}

	return c.JSON(http.StatusOK, user)
}

func (h *Handler) GetAttendanceRecordsByStudentID(c echo.Context) error {
	studentID := c.Param("id")
	if studentID == "" {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "id is required"})
	}

	records, err := h.service.GetAttendanceRecordsByStudentID(studentID)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}

	return c.JSON(http.StatusOK, records)
}

func (h *Handler) GetAttendanceRecordsBySubjectID(c echo.Context) error {
	subjectID := c.Param("id")
	if subjectID == "" {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "id is required"})
	}

	records, err := h.service.GetAttendanceRecordsBySubjectID(subjectID)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}

	return c.JSON(http.StatusOK, records)
}

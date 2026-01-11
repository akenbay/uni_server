package model

type StudentResponse struct {
	ID        int    `json:"id"`
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
	Gender    string `json:"gender"`
	BirthDate string `json:"birth_date"`
	GroupName string `json:"group_name"`
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

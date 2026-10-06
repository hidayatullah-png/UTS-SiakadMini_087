package service

import "siakad-mini/app/model"

func CanAccessOwnStudent(current model.AuthUser, targetUserID int) bool {
	return current.Role == "admin" || current.UserID == targetUserID
}

func CanAccessOwnEnrollment(current model.AuthUser, enrollmentStudentUserID int) bool {
	return current.UserID == enrollmentStudentUserID
}
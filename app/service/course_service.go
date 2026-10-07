package service

import (
	"github.com/gofiber/fiber/v2"

	"siakad-mini/app/repository"
	"siakad-mini/helper"
)

type CourseService struct {
	repo *repository.CourseRepository
}

func NewCourseService(repo *repository.CourseRepository) *CourseService {
	return &CourseService{repo: repo}
}

func (s *CourseService) List(c *fiber.Ctx) error {
	ctx, cancel := helper.RequestContext(c)
	defer cancel()

	q := repository.CourseListQuery{
		Semester:  c.QueryInt("semester", 0),
		Search:    c.Query("search"),
		Available: c.Query("available") == "true",
	}

	courses, err := s.repo.FindAll(ctx, q)
	if err != nil {
		return helper.Fail(c, fiber.StatusInternalServerError, "gagal mengambil data mata kuliah")
	}

	return helper.Success(c, fiber.StatusOK, "daftar mata kuliah berhasil diambil", courses)
}
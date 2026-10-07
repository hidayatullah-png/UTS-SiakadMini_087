package service

import (
	"errors"
	"fmt"

	"github.com/gofiber/fiber/v2"

	"siakad-mini/app/model"
	"siakad-mini/app/repository"
	"siakad-mini/helper"
)

type EnrollmentService struct {
	repo     *repository.EnrollmentRepository
	students *repository.StudentRepository
}

func NewEnrollmentService(
	repo *repository.EnrollmentRepository, students *repository.StudentRepository,
) *EnrollmentService {
	return &EnrollmentService{repo: repo, students: students}
}

func (s *EnrollmentService) Create(c *fiber.Ctx) error {
	ctx, cancel := helper.RequestContext(c)
	defer cancel()

	current, ok := helper.CurrentUser(c)
	if !ok {
		return helper.Fail(c, fiber.StatusUnauthorized, "belum terautentikasi")
	}

	student, err := s.students.FindByUserID(ctx, current.UserID)
	if err != nil {
		return helper.Fail(c, fiber.StatusNotFound, "data mahasiswa tidak ditemukan")
	}

	var req model.CreateEnrollmentRequest
	if err := c.BodyParser(&req); err != nil {
		return helper.Fail(c, fiber.StatusBadRequest, "body harus berupa JSON yang valid")
	}
	if errs := helper.ValidateStruct(req); errs != nil {
		return helper.FailValidation(c, errs)
	}

	batasSKS := BatasSKS(student.IPKTerakhir)
	enrollment, err := s.repo.Create(ctx, student.ID, req.CourseID, req.TahunAkademik, batasSKS)
	if err != nil {
		switch {
		case errors.Is(err, repository.ErrNotFound):
			return helper.Fail(c, fiber.StatusNotFound, "mata kuliah tidak ditemukan")
		case errors.Is(err, repository.ErrKuotaPenuh):
			return helper.Fail(c, fiber.StatusUnprocessableEntity, "kuota mata kuliah penuh")
		case errors.Is(err, repository.ErrMelebihiSKS):
			totalSKS, _ := s.repo.TotalSKS(ctx, student.ID, req.TahunAkademik)
			sisa := SisaSKS(totalSKS, batasSKS)
			return helper.Fail(c, fiber.StatusUnprocessableEntity,
				fmt.Sprintf("melebihi batas SKS, sisa %d SKS", sisa))
		case errors.Is(err, repository.ErrDuplicate):
			return helper.Fail(c, fiber.StatusConflict, "mata kuliah sudah pernah diambil tahun ini")
		default:
			return helper.Fail(c, fiber.StatusInternalServerError, "gagal memproses pengambilan mata kuliah")
		}
	}

	return helper.Success(c, fiber.StatusCreated, "mata kuliah berhasil diambil", enrollment)
}

func (s *EnrollmentService) Delete(c *fiber.Ctx) error {
	ctx, cancel := helper.RequestContext(c)
	defer cancel()

	current, ok := helper.CurrentUser(c)
	if !ok {
		return helper.Fail(c, fiber.StatusUnauthorized, "belum terautentikasi")
	}
	id, valid := helper.ParamID(c)
	if !valid {
		return helper.Fail(c, fiber.StatusBadRequest, "id harus berupa angka positif")
	}

	enrollment, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return helper.Fail(c, fiber.StatusNotFound, "enrollment tidak ditemukan")
	}

	owner, err := s.students.FindByID(ctx, enrollment.StudentID)
	if err != nil {
		return helper.Fail(c, fiber.StatusNotFound, "enrollment tidak ditemukan")
	}
	if !CanAccessOwnEnrollment(current, owner.UserID) {
		return helper.Fail(c, fiber.StatusForbidden, "tidak berhak membatalkan enrollment milik orang lain")
	}

	if err := s.repo.Delete(ctx, id); err != nil {
		return helper.Fail(c, fiber.StatusInternalServerError, "gagal membatalkan mata kuliah")
	}

	return helper.NoContent(c)
}

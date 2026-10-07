package service

import (
	"errors"
	"math"

	"github.com/gofiber/fiber/v2"

	"siakad-mini/app/model"
	"siakad-mini/app/repository"
	"siakad-mini/helper"
)

type StudentService struct {
	repo *repository.StudentRepository
}

func NewStudentService(repo *repository.StudentRepository) *StudentService {
	return &StudentService{repo: repo}
}

func (s *StudentService) List(c *fiber.Ctx) error {
	ctx, cancel := helper.RequestContext(c)
	defer cancel()

	q := helper.ParseListQuery(c)
	students, total, err := s.repo.FindAll(ctx, q)
	if err != nil {
		return helper.Fail(c, fiber.StatusInternalServerError, "gagal mengambil data mahasiswa")
	}

	lastPage := int(math.Ceil(float64(total) / float64(q.PerPage)))
	if lastPage < 1 {
		lastPage = 1
	}

	return helper.SuccessList(c, "Data mahasiswa berhasil diambil", students, &model.Meta{
		CurrentPage: q.Page, PerPage: q.PerPage, Total: total, LastPage: lastPage,
	})
}

func (s *StudentService) Create(c *fiber.Ctx) error {
	ctx, cancel := helper.RequestContext(c)
	defer cancel()

	var req model.CreateStudentRequest
	if err := c.BodyParser(&req); err != nil {
		return helper.Fail(c, fiber.StatusBadRequest, "body harus berupa JSON yang valid")
	}
	if errs := helper.ValidateStruct(req); errs != nil {
		return helper.FailValidation(c, errs)
	}

	student, err := s.repo.CreateWithUser(ctx, req)
	if err != nil {
		switch {
		case errors.Is(err, repository.ErrDuplicateNIM):
			return helper.FailValidation(c, map[string][]string{"nim": {"NIM sudah terdaftar"}})
		case errors.Is(err, repository.ErrDuplicateEmail):
			return helper.FailValidation(c, map[string][]string{"email": {"Email sudah terdaftar"}})
		default:
			return helper.Fail(c, fiber.StatusInternalServerError, "gagal membuat mahasiswa")
		}
	}

	return helper.Success(c, fiber.StatusCreated, "mahasiswa berhasil ditambahkan", student)
}

func (s *StudentService) Get(c *fiber.Ctx) error {
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

	student, err := s.repo.FindByIDWithSKS(ctx, id)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return helper.Fail(c, fiber.StatusNotFound, "mahasiswa tidak ditemukan")
		}
		return helper.Fail(c, fiber.StatusInternalServerError, "gagal mengambil data mahasiswa")
	}

	if !CanAccessOwnStudent(current, student.UserID) {
		return helper.Fail(c, fiber.StatusForbidden, "tidak berhak mengakses data mahasiswa lain")
	}

	student.BatasSKS = BatasSKS(student.IPKTerakhir)
	return helper.Success(c, fiber.StatusOK, "mahasiswa ditemukan", student)
}

func (s *StudentService) Replace(c *fiber.Ctx) error {
	ctx, cancel := helper.RequestContext(c)
	defer cancel()

	id, valid := helper.ParamID(c)
	if !valid {
		return helper.Fail(c, fiber.StatusBadRequest, "id harus berupa angka positif")
	}

	var req model.ReplaceStudentRequest
	if err := c.BodyParser(&req); err != nil {
		return helper.Fail(c, fiber.StatusBadRequest, "body harus berupa JSON yang valid")
	}
	if errs := helper.ValidateStruct(req); errs != nil {
		return helper.FailValidation(c, errs)
	}

	student, err := s.repo.Replace(ctx, id, req)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return helper.Fail(c, fiber.StatusNotFound, "mahasiswa tidak ditemukan")
		}
		return helper.Fail(c, fiber.StatusInternalServerError, "gagal memperbarui mahasiswa")
	}

	return helper.Success(c, fiber.StatusOK, "mahasiswa berhasil diperbarui", student)
}

func (s *StudentService) SoftDelete(c *fiber.Ctx) error {
	ctx, cancel := helper.RequestContext(c)
	defer cancel()

	id, valid := helper.ParamID(c)
	if !valid {
		return helper.Fail(c, fiber.StatusBadRequest, "id harus berupa angka positif")
	}

	if err := s.repo.SoftDelete(ctx, id); err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return helper.Fail(c, fiber.StatusNotFound, "mahasiswa tidak ditemukan")
		}
		return helper.Fail(c, fiber.StatusInternalServerError, "gagal menghapus mahasiswa")
	}

	return helper.NoContent(c)
}
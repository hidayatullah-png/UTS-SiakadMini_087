package service

import (
	"github.com/gofiber/fiber/v2"

	"siakad-mini/app/model"
	"siakad-mini/app/repository"
	"siakad-mini/helper"
)

type AuthService struct {
	users    *repository.UserRepository
	students *repository.StudentRepository
	jwt      *helper.JWTManager
}

func NewAuthService(
	users *repository.UserRepository, students *repository.StudentRepository, jwt *helper.JWTManager,
) *AuthService {
	return &AuthService{users: users, students: students, jwt: jwt}
}

func (s *AuthService) Login(c *fiber.Ctx) error {
	ctx, cancel := helper.RequestContext(c)
	defer cancel()

	var req model.LoginRequest
	if err := c.BodyParser(&req); err != nil {
		return helper.Fail(c, fiber.StatusBadRequest, "body harus berupa JSON yang valid")
	}
	if errs := helper.ValidateStruct(req); errs != nil {
		return helper.FailValidation(c, errs)
	}

	user, err := s.users.FindByEmailForLogin(ctx, req.Email)
	if err != nil {
		// user tidak ada ATAU password salah membalas pesan yang SAMA,
		// mencegah user enumeration (pola dari Modul 5).
		return helper.Fail(c, fiber.StatusUnauthorized, "email atau password salah")
	}
	if user.StudentDeleted {
		// mahasiswa yang sudah di-soft-delete diperlakukan sama seperti
		// kredensial salah - bukan pesan khusus "akun dinonaktifkan",
		// supaya tidak membocorkan status akun ke pemanggil anonim.
		return helper.Fail(c, fiber.StatusUnauthorized, "email atau password salah")
	}
	if !helper.VerifyPassword(user.Password, req.Password) {
		return helper.Fail(c, fiber.StatusUnauthorized, "email atau password salah")
	}

	token, err := s.jwt.GenerateAccess(user.User)
	if err != nil {
		return helper.Fail(c, fiber.StatusInternalServerError, "gagal membuat token")
	}

	return helper.Success(c, fiber.StatusOK, "login berhasil", fiber.Map{
		"access_token": token,
		"token_type":   "Bearer",
		"expires_in":   s.jwt.AccessTTLSeconds(),
		"user": fiber.Map{
			"id":    user.ID,
			"email": user.Email,
			"role":  user.Role,
		},
	})
}

func (s *AuthService) Me(c *fiber.Ctx) error {
	ctx, cancel := helper.RequestContext(c)
	defer cancel()

	current, ok := helper.CurrentUser(c)
	if !ok {
		return helper.Fail(c, fiber.StatusUnauthorized, "belum terautentikasi")
	}

	user, err := s.users.FindByID(ctx, current.UserID)
	if err != nil {
		return helper.Fail(c, fiber.StatusInternalServerError, "gagal mengambil profil")
	}

	response := fiber.Map{
		"id":    user.ID,
		"email": user.Email,
		"role":  user.Role,
	}

	if user.Role == "mahasiswa" {
		student, err := s.students.FindByUserID(ctx, user.ID)
		if err == nil {
			response["students"] = fiber.Map{
				"nim":      student.NIM,
				"nama":     student.Nama,
				"prodi":    student.Prodi,
				"angkatan": student.Angkatan,
			}
		}
	}

	return helper.Success(c, fiber.StatusOK, "profil berhasil diambil", response)
}
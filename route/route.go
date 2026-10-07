package route

import (
	"github.com/gofiber/fiber/v2"
	"github.com/jackc/pgx/v5/pgxpool"

	"siakad-mini/app/service"
	"siakad-mini/helper"
	"siakad-mini/middleware"
)

type Dependencies struct {
	Pool              *pgxpool.Pool
	JWT               *helper.JWTManager
	AuthService       *service.AuthService
	StudentService    *service.StudentService
	CourseService     *service.CourseService
	EnrollmentService *service.EnrollmentService
}

func Register(app *fiber.App, deps Dependencies) {
	api := app.Group("/api/v1")

	auth := api.Group("/auth", middleware.RequireJSON)
	auth.Post("/login", middleware.LoginRateLimiter(), deps.AuthService.Login) // 1
	auth.Get("/me", middleware.RequireAuth(deps.JWT), deps.AuthService.Me)     // 2

	students := api.Group("/students", middleware.RequireJSON, middleware.RequireAuth(deps.JWT))
	students.Get("/", middleware.RequireRole("admin"), deps.StudentService.List)             // 3
	students.Post("/", middleware.RequireRole("admin"), deps.StudentService.Create)          // 4
	students.Get("/:id", deps.StudentService.Get)                                            // 5 - ownership di service
	students.Put("/:id", middleware.RequireRole("admin"), deps.StudentService.Replace)       // 6
	students.Delete("/:id", middleware.RequireRole("admin"), deps.StudentService.SoftDelete) // 7

	courses := api.Group("/courses", middleware.RequireAuth(deps.JWT))
	courses.Get("/", deps.CourseService.List) // 8 - semua role

	enrollments := api.Group("/enrollments", middleware.RequireJSON, middleware.RequireAuth(deps.JWT))
	enrollments.Post("/", middleware.RequireRole("mahasiswa"), deps.EnrollmentService.Create) // 9
	enrollments.Delete("/:id", deps.EnrollmentService.Delete)                                 // 10 - ownership di service
}

package helper

import (
	"reflect"
	"strings"
	"time"

	"github.com/go-playground/validator/v10"
)

var validate = validator.New()

func init() {
	// Pesan pakai nama field JSON (mis. "nim"), bukan nama field Go ("NIM").
	validate.RegisterTagNameFunc(func(fld reflect.StructField) string {
		name := strings.SplitN(fld.Tag.Get("json"), ",", 2)[0]
		if name == "-" {
			return ""
		}
		return name
	})
	validate.RegisterValidation("angkatan_valid", func(fl validator.FieldLevel) bool {
		year := fl.Field().Int()
		return year >= 1900 && int(year) <= time.Now().Year()
	})
}

// ValidateStruct mengembalikan map[string][]string - kosong berarti lolos.
func ValidateStruct(s any) map[string][]string {
	err := validate.Struct(s)
	if err == nil {
		return nil
	}

	errs := map[string][]string{}
	for _, fe := range err.(validator.ValidationErrors) {
		field := fe.Field()
		errs[field] = append(errs[field], messageFor(fe))
	}
	return errs
}

func messageFor(fe validator.FieldError) string {
	switch fe.Tag() {
	case "required":
		return "wajib diisi"
	case "email":
		return "format email tidak valid"
	case "min":
		return fe.Field() + " minimal " + fe.Param()
	case "max":
		return fe.Field() + " maksimal " + fe.Param()
	case "len":
		return "harus " + fe.Param() + " karakter"
	case "numeric":
		return "harus berupa angka"
	case "angkatan_valid":
		return "angkatan harus 4 digit dan tidak lebih dari tahun berjalan"
	default:
		return "tidak valid"
	}
}

package main

import (
	"context"
	"fmt"
	"log"

	"siakad-mini/config"
	"siakad-mini/database"
	"siakad-mini/helper"
)

func main() {
	config.LoadEnv()

	pool, err := database.NewPool(context.Background())
	if err != nil {
		log.Fatalf("gagal konek database: %v", err)
	}
	defer pool.Close()
	ctx := context.Background()
	//admin
	adminHash, err := helper.HashPassword("Admin123!")
	if err != nil {
		log.Fatalf("gagal hash password: %v", err)
	}
	_, err = pool.Exec(ctx,
		`INSERT INTO users (email, password, role) VALUES ($1,$2,'admin')
	ON CONFLICT (LOWER(email)) DO NOTHING`,
		"sukma@siakad.com", adminHash)
	if err != nil {
		log.Fatalf("gagal seed admin: %v", err)
	}
	log.Println("admin siap: sukma@siakad.com / Admin123!")

	// 20 mahasiswa
	for i := 1; i <= 20; i++ {
		nim := fmt.Sprintf("1872210%05d", i)
		email := fmt.Sprintf("mahasiswa%d@siakad.com", i)

		passwordHash, err := helper.HashPassword(nim)
		if err != nil {
			log.Fatalf("gagal hash password %d: %v", i, err)
		}

		var userID int
		err = pool.QueryRow(ctx,
			`INSERT INTO users (email, password, role) VALUES ($1, $2, 'mahasiswa')
			ON CONFLICT (LOWER(email)) DO NOTHING RETURNING id`,
			email, passwordHash).Scan(&userID)
		if err != nil {
			//user sudah ada, ambil id user
			err = pool.QueryRow(ctx, `SELECT id FROM users WHERE LOWER(email) = LOWER($1)`, email).Scan(&userID)
			if err != nil {
				log.Fatalf("gagal ambil user mahasiswa %d: %v", i, err)
			}
		}
		ipk := 2.30 + float64(i%18)*0.1 //variasi IPK antara 2.30 - 3.00
		_, err = pool.Exec(ctx,
			`INSERT INTO students (user_id, nim, nama, prodi, angkatan, ipk_terakhir)
			VALUES ($1, $2, $3, $4, $5, $6)
			ON CONFLICT (nim) DO NOTHING`,
			userID, nim, fmt.Sprintf("Mahasiswa %02d", i), "Sistem Informasi", 2021+i%4, ipk)
		if err != nil {
			log.Fatalf("gagal seed student %d: %v", i, err)
		}
	}
	log.Println("20 mahasiswa siap (password awal = NIM masing-masing)")

	// 10 matkul
	courses := []struct {
		Kode, Nama           string
		SKS, Semester, Kuota int
	}{
		{"IF123", "Pemrograman Dasar", 3, 1, 30},
		{"IF124", "Matematika Diskret", 3, 1, 30},
		{"IF125", "Algoritma dan Struktur Data", 3, 1, 30},
		{"IF126", "Basis Data", 3, 1, 30},
		{"IF127", "Jaringan Komputer", 3, 1, 30},
		{"IF128", "Sistem Operasi", 3, 1, 30},
		{"IF129", "Kecerdasan Buatan", 3, 1, 30},
		{"IF130", "Pengantar Teknologi Informasi", 3, 1, 30},
		{"IF131", "Pemrograman Web", 3, 1, 2},
		{"IF132", "Mobile Programming", 3, 1, 30},
	}
	for _, mk := range courses {
		_, err = pool.Exec(ctx,
			`INSERT INTO courses (kode_mk, nama_mk, sks, semester, kuota)
				VALUES ($1, $2, $3, $4, $5)
				ON CONFLICT (kode_mk) DO NOTHING`,
			mk.Kode, mk.Nama, mk.SKS, mk.Semester, mk.Kuota)
		if err != nil {
			log.Fatalf("gagal seed course %s: %v", mk.Kode, err)
		}
	}
	log.Println("10 mata kuliah siap")

	log.Println("seeding selesai")
}

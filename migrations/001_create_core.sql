CREATE TABLE users (
    id SERIAL PRIMARY KEY,
    email VARCHAR(255) NOT NULL,
    password VARCHAR(255) NOT NULL,
    role VARCHAR (20) NOT NULL CHECK (role IN ('admin','mahasiswa')),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
CREATE UNIQUE INDEX users_email_lower_key ON users (LOWER(email));

CREATE TABLE students (
    id SERIAL PRIMARY KEY,
    user_id INT NOT NULL UNIQUE REFERENCES users(id) ON DELETE CASCADE,
    nim VARCHAR(12) NOT NULL,
    nama VARCHAR(150) NOT NULL,
    prodi VARCHAR(100) NOT NULL,
    angkatan INT NOT NULL,
    ipk_terakhir NUMERIC(3,2) CHECK (ipk_terakhir >= 0 AND ipk_terakhir <=4.00),
    deleted_at TIMESTAMPTZ
);
CREATE UNIQUE INDEX students_nim_key ON students (nim);

CREATE TABLE courses (
    id SERIAL PRIMARY KEY,
    kode_mk VARCHAR(20) NOT NULL,
    nama_mk VARCHAR(150) NOT NULL,
    sks INT NOT NULL CHECK (sks > 0),
    semester INT NOT NULL CHECK (semester > 0 AND semester <= 8),
    kuota INT NOT NULL CHECK (kuota >= 0)
);
CREATE UNIQUE INDEX courses_kode_mk_key ON courses (kode_mk);

CREATE TABLE enrollments (
    id SERIAL PRIMARY KEY,
    student_id INT NOT NULL REFERENCES students(id),
    course_id INTEGER NOT NULL REFERENCES courses(id),
    tahun_akademik VARCHAR(20) NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE (student_id, course_id, tahun_akademik)
);
CREATE INDEX enrollments_student_idx ON enrollments (student_id);
CREATE INDEX enrollments_course_idx ON enrollments (course_id);
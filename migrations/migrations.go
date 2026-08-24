package migrations

import "embed"

// FS ฝังไฟล์ .sql ทั้งหมดในโฟลเดอร์นี้เข้าไปใน Binary
//
//go:embed *.sql
var FS embed.FS

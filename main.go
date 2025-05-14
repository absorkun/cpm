package main

import (
	"database/sql"
	"fmt"
	"log"
	"os"
	"time"

	_ "github.com/go-sql-driver/mysql"
	"github.com/gofiber/fiber/v3"
	"github.com/joho/godotenv"
)

var db *sql.DB

func initDB() *sql.DB {
	dsn := fmt.Sprintf("%s:%s@tcp(%s)/%s?parseTime=true&loc=Asia%%2FJakarta",
		os.Getenv("USERNAME"),
		os.Getenv("PASSWORD"),
		os.Getenv("HOSTNAME"),
		os.Getenv("DATABASE"),
	)
	db, err := sql.Open("mysql", dsn)
	if err != nil {
		log.Fatalf("Database connection error: %v", err)
	}
	db.SetConnMaxLifetime(time.Minute * 3)
	db.SetMaxOpenConns(10)
	db.SetMaxIdleConns(10)

	_, err = db.Exec("CREATE TABLE IF NOT EXISTS clipboard (id INT PRIMARY KEY, content VARCHAR(255))")
	if err != nil {
		log.Fatalf("Failed to create table: %v", err)
	}

	_, _ = db.Exec("INSERT IGNORE INTO clipboard (id, content) VALUES (1, '')")

	return db
}

func main() {
	godotenv.Load()
	db = initDB()

	app := fiber.New()

	// app.Use(basicauth.New(basicauth.Config{
	// 	Users: map[string]string{
	// 		os.Getenv("AUTHUSERNAME"): os.Getenv("AUTHPASSWORD"),
	// 	},
	// }))

	// GET untuk menampilkan konten dan form
	app.Get("/", func(c fiber.Ctx) error {
		var content string
		err := db.QueryRow("SELECT content FROM clipboard WHERE id=1").Scan(&content)
		if err != nil {
			return c.Status(500).SendString("Database error: " + err.Error())
		}

		// HTML sederhana untuk menampilkan form
		html := fmt.Sprintf(`
			<!DOCTYPE html>
			<html>
			<head><title>Clipboard</title></head>
			<body>
				<center>
					<h2>Silakan Kopi:</h2>
					<p><strong>%s</strong></p>
					<form method="POST" action="/reset">
						<button type="submit">Reset</button>
					</form>
					<h2>Paste Di Sini</h2>
					<form method="POST" action="/">
						<textarea name="content" rows="4" cols="50"></textarea><br><br>
						<button type="submit">Update</button>
					</form>
				</center>
			</body>
			</html>
		`, content)

		return c.Type("html").SendString(html)
	})

	// POST untuk menerima data form
	app.Post("/", func(c fiber.Ctx) error {
		content := c.FormValue("content")
		_, err := db.Exec("UPDATE clipboard SET content=? WHERE id=1", content)
		if err != nil {
			return c.Status(500).SendString("Update error: " + err.Error())
		}
		return c.Redirect().To("/")
	})

	app.Post("/reset", func(c fiber.Ctx) error {
		_, err := db.Exec("UPDATE clipboard SET content='' WHERE id=1")
		if err != nil {
			return c.Status(500).SendString("Update error: " + err.Error())
		}
		return c.Redirect().To("/")
	})

	port := os.Getenv("PORT")
	if port == "" {
		port = "80"
	}

	app.Listen(fmt.Sprintf("0.0.0.0:%s", port))
}

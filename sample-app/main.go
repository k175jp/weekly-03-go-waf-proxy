package main

import (
	"database/sql"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"os"
	"time"

	_ "github.com/go-sql-driver/mysql"
)

type User struct {
	ID    int
	Name  string
	Email string
}

var db *sql.DB

func main() {
	dsn := os.Getenv("DATABASE_URL")
	if dsn == "" {
		dsn = "root:rootpass@tcp(mariadb:3306)/security_lab"
	}

	var err error
	for i := 0; i < 10; i++ {
		db, err = sql.Open("mysql", dsn)
		if err == nil {
			err = db.Ping()
			if err == nil {
				break
			}
		}
		log.Printf("⏳ Waiting for MariaDB to be ready... (%d/10)", i+1)
		time.Sleep(3 * time.Second)
	}

	if err != nil {
		log.Fatalf("❌ Failed to connect to MariaDB: %v", err)
	}
	log.Println("✅ Successfully connected to MariaDB")

	http.HandleFunc("/", indexHandler)
	http.HandleFunc("/search", searchHandler)

	log.Println("🚀 Backend vulnerable app sharing on :8080...")
	if err := http.ListenAndServe(":8080", nil); err != nil {
		log.Fatal(err)
	}
}

func indexHandler(w http.ResponseWriter, r *http.Request) {
	http.Redirect(w, r, "/search", http.StatusSeeOther)
}

func searchHandler(w http.ResponseWriter, r *http.Request) {
	var searchInput string
	if r.Method == http.MethodPost {
		searchInput = r.FormValue("q")
	} else {
		searchInput = r.URL.Query().Get("q")
	}

	var results []User
	var generatedQuery string
	hasSearched := false

	if searchInput != "" {
		hasSearched = true

		generatedQuery = fmt.Sprintf("SELECT id, name, email FROM users WHERE name = '%s';", searchInput)
		log.Printf("[BACKEND DB] Executing Query: %s", generatedQuery)

		rows, err := db.Query(generatedQuery)
		if err == nil {
			defer rows.Close()
			for rows.Next() {
				var u User
				if err := rows.Scan(&u.ID, &u.Name, &u.Email); err == nil {
					results = append(results, u)
				}
			}
		} else {
			log.Printf("[DB ERROR] %v", err)
		}
	}

	tmpl, err := template.New("webpage").Parse(htmlTemplate)
	if err != nil {
		http.Error(w, "Template Parse Error", http.StatusInternalServerError)
		return
	}

	data := struct {
		RawQuery    string
		Results     []User
		HasSearched bool
		UserQuery   template.HTML
	}{
		RawQuery:    generatedQuery,
		Results:     results,
		HasSearched: hasSearched,
		UserQuery:   template.HTML(searchInput),
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	tmpl.Execute(w, data)
}

const htmlTemplate = `
<!DOCTYPE html>
<html lang="ja">
<head>
    <meta charset="UTF-8">
    <title>Vulnerable Search App</title>
    <style>
        body { font-family: 'Segoe UI', Tahoma, Geneva, Verdana, sans-serif; margin: 40px; background-color: #f5f7fb; color: #333; }
        .container { max-width: 800px; background: white; padding: 30px; border-radius: 8px; box-shadow: 0 4px 6px rgba(0,0,0,0.05); }
        h1 { color: #2c3e50; border-bottom: 2px solid #ecf0f1; padding-bottom: 10px; }
        .search-box { margin: 20px 0; display: flex; gap: 10px; }
        input[type="text"] { flex: 1; padding: 12px; border: 1px solid #ccc; border-radius: 4px; font-size: 16px; }
        button { padding: 12px 24px; background-color: #3498db; color: white; border: none; border-radius: 4px; cursor: pointer; font-size: 16px; }
        button:hover { background-color: #2980b9; }
        .meta-info { background-color: #f8f9fa; border-left: 4px solid #e74c3c; padding: 15px; margin: 20px 0; border-radius: 4px; font-family: monospace; word-break: break-all; }
        table { width: 100%; border-collapse: collapse; margin-top: 20px; }
        th, td { border: 1px solid #ddd; padding: 12px; text-align: left; }
        th { background-color: #34495e; color: white; }
        tr:nth-child(even) { background-color: #f9f9f9; }
        .no-results { color: #7f8c8d; font-style: italic; margin-top: 20px; }
    </style>
</head>
<body>
    <div class="container">
        <h1>🔍 ユーザー検索システム (検証用環境)</h1>

        <form action="/search" method="POST" class="search-box">
            <input type="text" name="q" placeholder="ユーザー名を入力してください..." required>
            <button type="submit">検索</button>
        </form>

        {{if .HasSearched}}
            <h2>Search Results for: {{.UserQuery}}</h2>

            <div class="meta-info">
                <strong>[DEBUG] 実際に発行されたSQLクエリ:</strong><br>
                {{if .RawQuery}}{{.RawQuery}}{{else}}（クエリ生成失敗または空）{{end}}
            </div>

            {{if .Results}}
                <table>
                    <thead>
                        <tr>
                            <th>ID</th>
                            <th>Name</th>
                            <th>Email</th>
                        </tr>
                    </thead>
                    <tbody>
                        {{range .Results}}
                        <tr>
                            <td>{{.ID}}</td>
                            <td>{{.Name}}</td>
                            <td>{{.Email}}</td>
                        </tr>
                        {{end}}
                    </tbody>
                </table>
            {{else}}
                <p class="no-results">❌ 該当するユーザーは見つかりませんでした。</p>
            {{end}}
        {{end}}
    </div>
</body>
</html>
`
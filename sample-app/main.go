package main

import (
	"fmt"
	"net/http"
)

func main() {
	// 検索画面 兼 結果表示のエンドポイント
	http.HandleFunc("/search", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")

		// クエリパラメータ "q" を取得
		keyword := r.URL.Query().Get("q")

		// 検索結果のメッセージを動的に生成
		var resultMessage string
		if keyword != "" {
			// ⚠️ エスケープせずにそのままHTMLに埋め込む（反射型XSSの脆弱性）
			resultMessage = fmt.Sprintf("<div style='color: red; padding: 10px; background: #eee;'>検索結果: %s</div>", keyword)
		}

		// フォームと結果をあわせたHTMLを出力
		fmt.Fprintf(w, `
			<!DOCTYPE html>
			<html>
			<head>
				<title>Vulnerable Search App</title>
			</head>
			<body style="font-family: sans-serif; margin: 40px;">
				<h1>🔍 脆弱な検索システム</h1>
				<p>WAFのテスト用に、自由なペイロードを入力してください。</p>

				<form action="/search" method="GET">
					<input type="text" name="q" value="%s" placeholder="キーワードを入力..." style="padding: 8px; width: 300px;">
					<button type="submit" style="padding: 8px 15px;">検索</button>
				</form>

				<br>
				%s </body>
			</html>
		`, keyword, resultMessage)
	})

	fmt.Println("🎯 Vulnerable Sample App running on port 8080...")
	if err := http.ListenAndServe(":8080", nil); err != nil {
		panic(err)
	}
}
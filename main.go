package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	_ "github.com/go-sql-driver/mysql"
)

// ==================== HTML テンプレート ====================

const indexHTML = `<!DOCTYPE html>
<html lang="ja">
<head>
  <meta charset="UTF-8">
  <meta name="viewport" content="width=device-width, initial-scale=1.0">
  <title>Deploy Test - Go App</title>
  <style>
    * { margin: 0; padding: 0; box-sizing: border-box; }
    body { font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', sans-serif; background: #f0f4f8; min-height: 100vh; }
    .container { max-width: 800px; margin: 0 auto; padding: 32px 16px; }
    h1 { font-size: 2rem; color: #1a202c; margin-bottom: 8px; }
    .subtitle { color: #718096; margin-bottom: 32px; }
    .card { background: #fff; border-radius: 16px; box-shadow: 0 4px 20px rgba(0,0,0,.06); padding: 24px; margin-bottom: 20px; }
    .card h2 { font-size: 1.1rem; color: #2d3748; margin-bottom: 12px; display: flex; align-items: center; gap: 8px; }
    .endpoint { display: flex; align-items: center; gap: 12px; padding: 10px 16px; background: #f7fafc; border-radius: 10px; margin-bottom: 8px; cursor: pointer; transition: background .2s; text-decoration: none; color: inherit; }
    .endpoint:hover { background: #edf2f7; }
    .method { font-size: 11px; font-weight: 700; padding: 3px 8px; border-radius: 6px; color: #fff; }
    .get { background: #38a169; }
    .path { font-family: 'SF Mono', monospace; font-size: 14px; color: #2d3748; }
    .desc { font-size: 12px; color: #a0aec0; margin-left: auto; }
    .status { display: inline-flex; align-items: center; gap: 6px; padding: 6px 14px; border-radius: 20px; font-size: 13px; font-weight: 600; }
    .status.ok { background: #c6f6d5; color: #22543d; }
    .status.err { background: #fed7d7; color: #9b2c2c; }
    .result { margin-top: 16px; background: #1a202c; border-radius: 12px; padding: 16px; overflow-x: auto; }
    .result pre { color: #68d391; font-family: 'SF Mono', monospace; font-size: 13px; white-space: pre-wrap; }
    .info { font-size: 13px; color: #718096; line-height: 1.8; }
    .info code { background: #edf2f7; padding: 2px 6px; border-radius: 4px; font-size: 12px; }
    .btn { display: inline-flex; align-items: center; gap: 6px; padding: 8px 18px; background: #4299e1; color: #fff; border: none; border-radius: 10px; font-size: 13px; font-weight: 600; cursor: pointer; transition: background .2s; }
    .btn:hover { background: #3182ce; }
    .btn.green { background: #38a169; } .btn.green:hover { background: #2f855a; }
    .btn.orange { background: #dd6b20; } .btn.orange:hover { background: #c05621; }
    .time { font-size: 12px; color: #a0aec0; text-align: right; margin-top: 8px; }
    .grid { display: grid; grid-template-columns: 1fr 1fr; gap: 16px; }
    @media (max-width: 640px) { .grid { grid-template-columns: 1fr; } }
  </style>
</head>
<body>
  <div class="container">
    <h1>🚀 Deploy Test</h1>
    <p class="subtitle">Go (net/http) Web アプリケーション</p>

    <!-- ステータス -->
    <div class="card">
      <h2>📊 サーバー情報</h2>
      <div class="grid">
        <div>
          <p class="info">
            <strong>言語:</strong> Go (net/http)<br>
            <strong>ポート:</strong> {{.Port}}<br>
            <strong>起動時刻:</strong> {{.StartTime}}
          </p>
        </div>
        <div>
          <p class="info">
            <strong>RDS:</strong> <span id="rds-status">確認中...</span><br>
            <strong>S3:</strong> <span id="s3-status">確認中...</span>
          </p>
        </div>
      </div>
    </div>

    <!-- API エンドポイント -->
    <div class="card">
      <h2>🔌 API エンドポイント</h2>
      <a class="endpoint" href="/api/info" target="_blank">
        <span class="method get">GET</span>
        <span class="path">/api/info</span>
        <span class="desc">サーバー情報</span>
      </a>
      <a class="endpoint" href="/api/hello?name=太郎" target="_blank">
        <span class="method get">GET</span>
        <span class="path">/api/hello?name=太郎</span>
        <span class="desc">挨拶API</span>
      </a>
      <a class="endpoint" href="/api/health" target="_blank">
        <span class="method get">GET</span>
        <span class="path">/api/health</span>
        <span class="desc">ヘルスチェック</span>
      </a>
      <a class="endpoint" href="/api/db" target="_blank">
        <span class="method get">GET</span>
        <span class="path">/api/db</span>
        <span class="desc">RDS (MySQL) 接続テスト</span>
      </a>
      <a class="endpoint" href="/api/s3" target="_blank">
        <span class="method get">GET</span>
        <span class="path">/api/s3</span>
        <span class="desc">S3 バケット一覧</span>
      </a>
    </div>

    <!-- ライブテスト -->
    <div class="card">
      <h2>🧪 ライブテスト</h2>
      <div style="display:flex;gap:8px;flex-wrap:wrap">
        <button class="btn" onclick="testAPI('/api/info')">📊 Info</button>
        <button class="btn green" onclick="testAPI('/api/hello?name=太郎')">👋 Hello</button>
        <button class="btn orange" onclick="testAPI('/api/db')">🗄️ DB接続</button>
        <button class="btn" onclick="testAPI('/api/s3')" style="background:#805ad5">📦 S3一覧</button>
      </div>
      <div class="result" id="result" style="display:none">
        <pre id="result-text"></pre>
      </div>
    </div>

    <p class="time">現在時刻: <span id="clock"></span></p>
  </div>

  <script>
    // 時計
    setInterval(function() {
      document.getElementById('clock').textContent = new Date().toLocaleString('ja-JP');
    }, 1000);
    document.getElementById('clock').textContent = new Date().toLocaleString('ja-JP');

    // API テスト
    async function testAPI(url) {
      var el = document.getElementById('result');
      var text = document.getElementById('result-text');
      el.style.display = 'block';
      text.textContent = '⏳ リクエスト中...';
      try {
        var res = await fetch(url);
        var data = await res.json();
        text.textContent = JSON.stringify(data, null, 2);
      } catch (e) {
        text.textContent = '❌ エラー: ' + e.message;
      }
    }

    // 自動チェック
    (async function() {
      try {
        var r = await fetch('/api/db');
        var d = await r.json();
        document.getElementById('rds-status').innerHTML = d.connected
          ? '<span class="status ok">✅ 接続OK</span>'
          : '<span class="status err">❌ ' + (d.error || '未設定') + '</span>';
      } catch(e) { document.getElementById('rds-status').innerHTML = '<span class="status err">❌ 未設定</span>'; }
      try {
        var r2 = await fetch('/api/s3');
        var d2 = await r2.json();
        document.getElementById('s3-status').innerHTML = d2.error
          ? '<span class="status err">❌ ' + d2.error + '</span>'
          : '<span class="status ok">✅ ' + (d2.buckets ? d2.buckets.length : 0) + ' バケット</span>';
      } catch(e) { document.getElementById('s3-status').innerHTML = '<span class="status err">❌ エラー</span>'; }
    })();
  </script>
</body>
</html>`

// ==================== Main ====================

var startTime = time.Now()

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = "3000"
	}

	mux := http.NewServeMux()

	// HTML トップページ
	mux.HandleFunc("/", handleIndex)

	// API エンドポイント
	mux.HandleFunc("/api/info", handleInfo)
	mux.HandleFunc("/api/hello", handleHello)
	mux.HandleFunc("/api/health", handleHealth)
	mux.HandleFunc("/api/db", handleDB)
	mux.HandleFunc("/api/s3", handleS3)

	log.Printf("🌐 Go サーバー起動: http://0.0.0.0:%s\n", port)
	if err := http.ListenAndServe(":"+port, mux); err != nil {
		log.Fatal(err)
	}
}

// ==================== Handlers ====================

// トップページ (HTML)
func handleIndex(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		http.NotFound(w, r)
		return
	}
	tmpl := template.Must(template.New("index").Parse(indexHTML))
	port := os.Getenv("PORT")
	if port == "" {
		port = "3000"
	}
	data := map[string]string{
		"Port":      port,
		"StartTime": startTime.Format("2006-01-02 15:04:05"),
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	tmpl.Execute(w, data)
}

// サーバー情報
func handleInfo(w http.ResponseWriter, r *http.Request) {
	jsonResp(w, map[string]interface{}{
		"app":     "deploy-test",
		"lang":    "Go (net/http)",
		"port":    getPort(),
		"time":    time.Now().Format("2006-01-02 15:04:05"),
		"uptime":  time.Since(startTime).String(),
		"message": "🚀 Go デプロイ成功！",
	})
}

// 挨拶
func handleHello(w http.ResponseWriter, r *http.Request) {
	name := r.URL.Query().Get("name")
	if name == "" {
		name = "World"
	}
	jsonResp(w, map[string]string{
		"message": fmt.Sprintf("こんにちは、%sさん！", name),
	})
}

// ヘルスチェック
func handleHealth(w http.ResponseWriter, r *http.Request) {
	jsonResp(w, map[string]string{"status": "ok"})
}

// ==================== DB 接続テスト ====================

func handleDB(w http.ResponseWriter, r *http.Request) {
	host := os.Getenv("RDS_HOST")
	port := os.Getenv("RDS_PORT")
	user := os.Getenv("RDS_USER")
	pass := os.Getenv("RDS_PASSWORD")
	dbName := os.Getenv("RDS_DATABASE")

	if host == "" {
		jsonResp(w, map[string]interface{}{
			"connected": false,
			"error":     "RDS_HOST が設定されていません（環境変数を確認してください）",
		})
		return
	}
	if port == "" {
		port = "3306"
	}
	if user == "" {
		user = "admin"
	}
	if dbName == "" {
		dbName = "trainingdb"
	}

	dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?timeout=5s&parseTime=true", user, pass, host, port, dbName)
	db, err := sql.Open("mysql", dsn)
	if err != nil {
		jsonResp(w, map[string]interface{}{
			"connected": false,
			"error":     fmt.Sprintf("接続オープン失敗: %v", err),
		})
		return
	}
	defer db.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := db.PingContext(ctx); err != nil {
		jsonResp(w, map[string]interface{}{
			"connected": false,
			"error":     fmt.Sprintf("Ping失敗: %v", err),
			"host":      host,
			"port":      port,
			"database":  dbName,
		})
		return
	}

	// バージョン取得
	var version string
	db.QueryRowContext(ctx, "SELECT VERSION()").Scan(&version)

	// テーブル一覧
	rows, err := db.QueryContext(ctx, "SHOW TABLES")
	var tables []string
	if err == nil {
		defer rows.Close()
		for rows.Next() {
			var t string
			rows.Scan(&t)
			tables = append(tables, t)
		}
	}

	jsonResp(w, map[string]interface{}{
		"connected": true,
		"host":      host,
		"port":      port,
		"database":  dbName,
		"version":   version,
		"tables":    tables,
		"message":   "✅ RDS 接続成功！",
	})
}

// ==================== S3 接続テスト ====================

func handleS3(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	cfg, err := config.LoadDefaultConfig(ctx, config.WithRegion("ap-northeast-1"))
	if err != nil {
		jsonResp(w, map[string]interface{}{
			"error": fmt.Sprintf("AWS設定読み込み失敗: %v", err),
		})
		return
	}

	client := s3.NewFromConfig(cfg)
	result, err := client.ListBuckets(ctx, &s3.ListBucketsInput{})
	if err != nil {
		jsonResp(w, map[string]interface{}{
			"error": fmt.Sprintf("S3一覧取得失敗: %v", err),
		})
		return
	}

	buckets := []map[string]string{}
	for _, b := range result.Buckets {
		buckets = append(buckets, map[string]string{
			"name":    *b.Name,
			"created": b.CreationDate.Format("2006-01-02"),
		})
	}

	// 特定バケットの操作（S3_BUCKET が設定されている場合）
	var bucketInfo map[string]interface{}
	if bucket := os.Getenv("S3_BUCKET"); bucket != "" {
		listResult, err := client.ListObjectsV2(ctx, &s3.ListObjectsV2Input{
			Bucket:  &bucket,
			MaxKeys: toInt32Ptr(10),
		})
		if err != nil {
			bucketInfo = map[string]interface{}{
				"bucket": bucket,
				"error":  err.Error(),
			}
		} else {
			objects := []map[string]interface{}{}
			for _, obj := range listResult.Contents {
				objects = append(objects, map[string]interface{}{
					"key":  *obj.Key,
					"size": obj.Size,
				})
			}
			bucketInfo = map[string]interface{}{
				"bucket":  bucket,
				"count":   listResult.KeyCount,
				"objects": objects,
			}
		}
	}

	resp := map[string]interface{}{
		"buckets": buckets,
		"count":   len(buckets),
		"message": fmt.Sprintf("✅ S3 接続成功！(%d バケット)", len(buckets)),
	}
	if bucketInfo != nil {
		resp["bucket_detail"] = bucketInfo
	}
	jsonResp(w, resp)
}

// ==================== Helpers ====================

func jsonResp(w http.ResponseWriter, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	json.NewEncoder(w).Encode(data)
}

func getPort() string {
	port := os.Getenv("PORT")
	if port == "" {
		return "3000"
	}
	return port
}

func toInt32Ptr(v int32) *int32 {
	return &v
}

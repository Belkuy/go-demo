// go-demo：最小可运行 HTTP 服务，用于演示 CI/CD 全链路
// 编译时可通过 -ldflags "-X main.version=v1.0.0" 注入版本号
// 2026-11-05 新增 CORS 支持：允许 GitHub Pages 页面跨域调用本服务
package main

import (
	"fmt"
	"log"
	"net/http"
)

// version 是构建时注入的版本号，默认 dev 表示未经 CI 构建的本地开发版本
var version = "dev"

// allowedOrigins 是允许跨域访问的来源白名单
// 原则：只放行确需调用的前端站点，不使用 "*" 通配
var allowedOrigins = map[string]bool{
	"https://belkuy.github.io": true, // GitHub Pages 线上页面
	"http://localhost:8080":    true, // 本地调试页面
	"http://127.0.0.1:8080":    true, // 本地调试页面
}

// withCORS 是 CORS 中间件：给白名单来源的请求回写跨域头，并处理预检请求
// 浏览器跨域请求前会先发一个 OPTIONS 预检，必须正确响应，否则后续请求直接被拦截
func withCORS(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		origin := r.Header.Get("Origin")
		if allowedOrigins[origin] {
			// ACAO 回显请求方的具体 origin；白名单模式下回显比 "*" 更安全
			w.Header().Set("Access-Control-Allow-Origin", origin)
			w.Header().Set("Vary", "Origin") // 响应随 Origin 变化，告知 CDN 不要缓存单一版本
		}

		// 预检请求：直接返回 204，不带业务内容
		if r.Method == http.MethodOptions {
			w.Header().Set("Access-Control-Allow-Methods", "GET, OPTIONS")
			w.Header().Set("Access-Control-Max-Age", "86400") // 预检结果缓存 24h，减少预检次数
			w.WriteHeader(http.StatusNoContent)
			return
		}

		next(w, r)
	}
}

func main() {
	// 根路径：返回纯文本问候与版本号
	http.HandleFunc("/", withCORS(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, "hello from go-demo, version: %s\n", version)
	}))

	// JSON 接口：给前端 fetch 演示用
	http.HandleFunc("/api/hello", withCORS(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json; charset=utf-8")
		fmt.Fprintf(w, `{"message":"hello from go-demo","version":%q}`, version)
	}))

	// 监听 8080 端口；ListenAndServe 会阻塞，因此放在 log.Fatal 里
	// 一旦返回错误（如端口被占用）直接退出并打印原因
	log.Println("listening on :8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}

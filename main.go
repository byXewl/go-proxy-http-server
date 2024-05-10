package main

import (
	"io"
	"log"
	"net/http"
	"net/url"
)

func main() {
	// 监听的代理端口
	proxyPort := ":33129"

	// 启动代理服务器
	log.Printf("Proxy server listening on %s", proxyPort)
	if err := http.ListenAndServe(proxyPort, http.HandlerFunc(handleRequest)); err != nil {
		log.Fatalf("Error starting proxy server: %v", err)
	}
}

func handleRequest(w http.ResponseWriter, r *http.Request) {
	// 打印收到的请求信息
	log.Printf("Received request: %s %s", r.Method, r.URL)

	// 创建一个新的请求
	targetURL := r.URL.String()

	// 解析请求 URL
	target, err := url.Parse(targetURL)
	if err != nil {
		log.Printf("Error parsing request URL: %v", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	// 构建目标地址
	target.Scheme = "http"
	target.Host = r.Host

	// 创建目标请求
	req, err := http.NewRequest(r.Method, target.String(), r.Body)
	if err != nil {
		log.Printf("Error creating target request: %v", err)
		http.Error(w, "Bad Gateway", http.StatusBadGateway)
		return
	}

	// 复制请求头部
	req.Header = make(http.Header)
	for key, values := range r.Header {
		for _, value := range values {
			req.Header.Add(key, value)
		}
	}

	// 发送目标请求
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		log.Printf("Error forwarding the request: %v", err)
		http.Error(w, "Bad Gateway", http.StatusBadGateway)
		return
	}
	defer resp.Body.Close()

	// 复制响应头部
	for key, values := range resp.Header {
		for _, value := range values {
			w.Header().Add(key, value)
		}
	}

	// 设置响应状态码
	w.WriteHeader(resp.StatusCode)

	// 将响应体写入代理服务器的连接
	if _, err := io.Copy(w, resp.Body); err != nil {
		log.Printf("Error copying response body: %v", err)
	}
}

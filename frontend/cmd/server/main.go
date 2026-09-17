package main

import (
	"log"
	"net/http"
)

/**
 * Máy chủ phát triển tĩnh phục vụ nạp ứng dụng Go WebAssembly tại cổng 3000.
 */
func main() {
	thuMucTinh := http.FileServer(http.Dir("."))
	http.Handle("/", thuMucTinh)

	log.Println("🌐 [DevOps Sentinel Frontend] Đang phục vụ tại: http://localhost:3000")
	if err := http.ListenAndServe(":3000", nil); err != nil {
		log.Fatalf("Lỗi máy chủ frontend: %v", err)
	}
}

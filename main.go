package main

import (
	"fmt"
	"io"
	"net"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"syscall"
	"time"
)

const PORT = ":9090"
const UPLOAD_DIR = "./uploads"
const TASK_NAME = "LanMediaServer"

func getLocalIP() string {
	conn, err := net.Dial("udp", "8.8.8.8:80")
	if err == nil {
		defer conn.Close()
		return conn.LocalAddr().(*net.UDPAddr).IP.String()
	}
	ifaces, err := net.Interfaces()
	if err != nil {
		return "127.0.0.1"
	}
	for _, iface := range ifaces {
		if iface.Flags&net.FlagUp == 0 || iface.Flags&net.FlagLoopback != 0 {
			continue
		}
		name := strings.ToLower(iface.Name)
		if strings.Contains(name, "docker") || strings.Contains(name, "vbox") ||
			strings.Contains(name, "veth") || strings.Contains(name, "vmnet") ||
			strings.Contains(name, "wsl") {
			continue
		}
		addrs, err := iface.Addrs()
		if err != nil {
			continue
		}
		for _, addr := range addrs {
			var ip net.IP
			switch v := addr.(type) {
			case *net.IPNet:
				ip = v.IP
			case *net.IPAddr:
				ip = v.IP
			}
			if ip == nil || ip.IsLoopback() {
				continue
			}
			if ip = ip.To4(); ip == nil {
				continue
			}
			return ip.String()
		}
	}
	return "127.0.0.1"
}

func enableCORS(w *http.ResponseWriter) {
	(*w).Header().Set("Access-Control-Allow-Origin", "*")
	(*w).Header().Set("Access-Control-Allow-Methods", "POST, GET, OPTIONS")
	(*w).Header().Set("Access-Control-Allow-Headers", "Content-Type")
}

func uploadHandler(w http.ResponseWriter, r *http.Request) {
	enableCORS(&w)
	if r.Method == "OPTIONS" {
		w.WriteHeader(http.StatusOK)
		return
	}
	if r.Method != "POST" {
		http.Error(w, "仅支持 POST", http.StatusMethodNotAllowed)
		return
	}
	err := r.ParseMultipartForm(1024 << 20)
	if err != nil {
		http.Error(w, "文件过大", http.StatusBadRequest)
		return
	}
	file, handler, err := r.FormFile("file")
	if err != nil {
		http.Error(w, "参数错误", http.StatusBadRequest)
		return
	}
	defer file.Close()
	os.MkdirAll(UPLOAD_DIR, os.ModePerm)
	filename := fmt.Sprintf("%d_%s", time.Now().Unix(), handler.Filename)
	savePath := filepath.Join(UPLOAD_DIR, filename)
	dst, err := os.Create(savePath)
	if err != nil {
		http.Error(w, "写入失败", http.StatusInternalServerError)
		return
	}
	defer dst.Close()
	io.Copy(dst, file)
	localIP := getLocalIP()
	fileURL := fmt.Sprintf("http://%s%s/files/%s", localIP, PORT, filename)
	w.Header().Set("Content-Type", "application/json")
	fmt.Fprintf(w, `{"code":0,"msg":"success","url":"%s"}`, fileURL)
}

func install() {
	if runtime.GOOS != "windows" {
		return
	}
	exePath, err := os.Executable()
	if err != nil {
		return
	}
	exePath, _ = filepath.Abs(exePath)

	// 注册开机自启 (当前用户)
	key := `HKCU\Software\Microsoft\Windows\CurrentVersion\Run`
	exec.Command("reg", "add", key, "/v", TASK_NAME, "/t", "REG_SZ", "/d", `"`+exePath+`"`, "/f").Run()

	// 静默启动服务（无窗口）
	cmd := exec.Command(exePath)
	cmd.SysProcAttr = &syscall.SysProcAttr{HideWindow: true, CreationFlags: 0x08000000}
	cmd.Start()
}

func uninstall() {
	if runtime.GOOS != "windows" {
		return
	}
	key := `HKCU\Software\Microsoft\Windows\CurrentVersion\Run`
	exec.Command("reg", "delete", key, "/v", TASK_NAME, "/f").Run()
	exec.Command("taskkill", "/f", "/im", "server.exe").Run()
}

func main() {
	if len(os.Args) > 1 {
		switch os.Args[1] {
		case "-install":
			install()
			return
		case "-uninstall":
			uninstall()
			return
		}
	}

	fs := http.FileServer(http.Dir(UPLOAD_DIR))
	http.Handle("/files/", http.StripPrefix("/files/", fs))
	http.HandleFunc("/upload", uploadHandler)

	fmt.Printf("服务已启动 http://%s%s\n", getLocalIP(), PORT)
	http.ListenAndServe(PORT, nil)
}

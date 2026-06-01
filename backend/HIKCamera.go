package backend

import (
	"changeme/backend/clib"
	"fmt"
	"net/http"
	"strconv"
	"sync"
	"sync/atomic"
)

type HIKCameraService struct {
	camera     *clib.Camera
	imgBuffer  []byte
	imgMu      sync.Mutex
	frameId    uint64
	httpServer *http.Server
}

func NewHIKCamera() *HIKCameraService {
	service := &HIKCameraService{
		camera: clib.NewCamera(),
	}

	// 设置图像回调 - camera.go 采集到图像后会调用此回调
	service.camera.SetImageCallback(service.onImageReceived)

	// 初始化相机 SDK
	service.camera.Init()

	// 启动 HTTP 图像流服务器
	service.startLiveStreamServer(9099)

	return service
}

// onImageReceived 图像数据回调（由 camera.go 调用）
func (s *HIKCameraService) onImageReceived(data []byte, frameId uint64) {
	s.imgMu.Lock()
	defer s.imgMu.Unlock()

	// 更新帧缓冲区（用于 HTTP 图像流）
	if cap(s.imgBuffer) < len(data) {
		s.imgBuffer = make([]byte, len(data))
	}
	s.imgBuffer = s.imgBuffer[:len(data)]
	copy(s.imgBuffer, data)
	atomic.StoreUint64(&s.frameId, frameId)
}

// startLiveStreamServer 启动本地 HTTP 服务器提供图像流
func (s *HIKCameraService) startLiveStreamServer(port int) {
	mux := http.NewServeMux()
	mux.HandleFunc("/live", func(w http.ResponseWriter, r *http.Request) {
		s.imgMu.Lock()
		defer s.imgMu.Unlock()

		if len(s.imgBuffer) == 0 {
			w.WriteHeader(http.StatusNoContent)
			return
		}

		w.Header().Set("Content-Type", "image/jpeg")
		w.Header().Set("Content-Length", strconv.Itoa(len(s.imgBuffer)))
		w.Header().Set("Cache-Control", "no-cache, no-store, must-revalidate")
		w.Header().Set("X-Frame-Id", strconv.FormatUint(atomic.LoadUint64(&s.frameId), 10))
		w.Write(s.imgBuffer)
	})

	addr := fmt.Sprintf("127.0.0.1:%d", port)
	s.httpServer = &http.Server{Addr: addr, Handler: mux}
	fmt.Printf("图像流 HTTP 服务器启动: http://%s/live\n", addr)
	go s.httpServer.ListenAndServe()
}

func (s *HIKCameraService) GetHIKCameraDevices() []string {
	devices, err := s.camera.GetCameraDevices()
	if err != nil {
		fmt.Println(err)
		return nil
	}
	return devices
}

func (s *HIKCameraService) OpenHIKCamera(name string) error {
	return s.camera.OpenCamera(name)
}

func (s *HIKCameraService) CloseHIKCamera() error {
	// 停止 HTTP 服务器
	if s.httpServer != nil {
		s.httpServer.Close()
		s.httpServer = nil
	}

	// 关闭相机
	return s.camera.CloseCamera()
}

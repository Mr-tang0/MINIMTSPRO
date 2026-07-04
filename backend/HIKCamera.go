package backend

import (
	"changeme/backend/extensometer"
	"changeme/backend/hik"
	"encoding/base64"
	"fmt"
	"image"
	"image/color"
	"math"
	"sync"
	"time"

	"github.com/wailsapp/wails/v3/pkg/application"
	"gocv.io/x/gocv"
)

type HIKCameraParams struct {
	ExposureAuto      uint32  `json:"exposureAuto"`
	ExposureTime      float64 `json:"exposureTime"`
	GainAuto          uint32  `json:"gainAuto"`
	Gain              float64 `json:"gain"`
	DigitalGain       float64 `json:"digitalGain"`
	DigitalGainEnable bool    `json:"digitalGainEnable"`
	FrameRate         float64 `json:"frameRate"`
	GammaEnable       bool    `json:"gammaEnable"`
	Gamma             float64 `json:"gamma"`
	BalanceWhiteAuto  uint32  `json:"balanceWhiteAuto"`
	BalanceWhite      int64   `json:"balanceWhite"`
}

type ROIRect struct {
	X      float64 `json:"x"`
	Y      float64 `json:"y"`
	Width  float64 `json:"width"`
	Height float64 `json:"height"`
}

type HIKCameraService struct {
	camera *hik.Camera

	frameMu     sync.RWMutex
	latestFrame []byte
	frameWidth  int
	frameHeight int

	roiMu sync.Mutex
	roiA  *extensometer.ExtensometerService
	roiB  *extensometer.ExtensometerService
	// roiDistanceLogCounter int

	previewMu       sync.Mutex
	lastPreviewEmit time.Time
	previewInterval time.Duration
}

func NewHIKCamera() *HIKCameraService {
	service := &HIKCameraService{
		camera:          hik.NewCamera(),
		previewInterval: 100 * time.Millisecond,
	}

	// 设置图像回调 - camera.go 采集到图像后会调用此回调
	service.camera.SetImageCallback(service.onImageReceived)

	// 初始化相机 SDK
	service.camera.Init()

	return service
}

// onImageReceived 图像数据回调（由 camera.go 调用）
func (s *HIKCameraService) onImageReceived(data []byte, frameId uint64) {
	app := application.Get()
	if len(data) == 0 {
		return
	}

	emitPreview := s.shouldEmitPreview()
	frameData := s.processFrame(data, emitPreview)
	if app == nil {
		return
	}
	if !emitPreview {
		return
	}

	app.Event.Emit("hik_camera_frame", map[string]any{
		"frameId": frameId,
		"image":   "data:image/jpeg;base64," + base64.StdEncoding.EncodeToString(frameData),
	})
}

func (s *HIKCameraService) processFrame(data []byte, drawOverlay bool) []byte {
	mat, err := gocv.IMDecode(data, gocv.IMReadColor)
	if err != nil || mat.Empty() {
		s.storeLatestFrame(data, 0, 0)
		return data
	}
	defer mat.Close()

	s.storeLatestFrame(data, mat.Cols(), mat.Rows())
	s.updateTrackedROIs(mat)
	if !drawOverlay || !s.hasTrackedROIs() {
		return data
	}
	s.drawTrackedROIs(&mat)

	buf, err := gocv.IMEncode(gocv.JPEGFileExt, mat)
	if err != nil {
		return data
	}
	defer buf.Close()

	out := make([]byte, len(buf.GetBytes()))
	copy(out, buf.GetBytes())
	return out
}

func (s *HIKCameraService) shouldEmitPreview() bool {
	s.previewMu.Lock()
	defer s.previewMu.Unlock()

	now := time.Now()
	if s.previewInterval <= 0 || s.lastPreviewEmit.IsZero() || now.Sub(s.lastPreviewEmit) >= s.previewInterval {
		s.lastPreviewEmit = now
		return true
	}
	return false
}

func (s *HIKCameraService) hasTrackedROIs() bool {
	s.roiMu.Lock()
	defer s.roiMu.Unlock()
	return s.roiA != nil || s.roiB != nil
}

func (s *HIKCameraService) storeLatestFrame(data []byte, width, height int) {
	s.frameMu.Lock()
	defer s.frameMu.Unlock()

	s.latestFrame = make([]byte, len(data))
	copy(s.latestFrame, data)
	s.frameWidth = width
	s.frameHeight = height
}

func (s *HIKCameraService) updateTrackedROIs(frame gocv.Mat) {
	s.roiMu.Lock()
	roiA := s.roiA
	roiB := s.roiB
	s.roiMu.Unlock()
	if roiA == nil || roiB == nil {
		return
	}

	grayFrame := gocv.NewMat()
	defer grayFrame.Close()
	if frame.Channels() > 1 {
		gocv.CvtColor(frame, &grayFrame, gocv.ColorBGRToGray)
	} else {
		frame.CopyTo(&grayFrame)
	}

	preparedFrame := roiA.PreprocessGrayForDIC(grayFrame)
	defer preparedFrame.Close()

	var wg sync.WaitGroup
	runTracker := func(tracker *extensometer.ExtensometerService) {
		defer wg.Done()
		_, _ = tracker.RunDICPreparedGrayMat(preparedFrame)
	}

	wg.Add(1)
	go runTracker(roiA)
	wg.Add(1)
	go runTracker(roiB)
	wg.Wait()
}

func (s *HIKCameraService) drawTrackedROIs(frame *gocv.Mat) {
	s.roiMu.Lock()
	defer s.roiMu.Unlock()

	drawROI := func(label string, tracker *extensometer.ExtensometerService, c color.RGBA) {
		if tracker == nil {
			return
		}
		roi := tracker.CurrentROI()

		rect := image.Rect(int(roi.X), int(roi.Y), int(roi.X+roi.Width), int(roi.Y+roi.Height))
		_ = gocv.Rectangle(frame, rect, c, 2)
		_ = gocv.PutText(frame, label, image.Pt(rect.Max.X+6, rect.Min.Y+18), gocv.FontHersheySimplex, 0.7, c, 2)
	}

	drawROI("A", s.roiA, color.RGBA{R: 34, G: 197, B: 94, A: 255})
	drawROI("B", s.roiB, color.RGBA{R: 59, G: 130, B: 246, A: 255})

	if s.roiA != nil && s.roiB != nil {
		roiA := s.roiA.CurrentROI()
		roiB := s.roiB.CurrentROI()
		centerAX := roiA.X + roiA.Width/2
		centerAY := roiA.Y + roiA.Height/2
		centerBX := roiB.X + roiB.Width/2
		centerBY := roiB.Y + roiB.Height/2
		dx := centerBX - centerAX
		dy := centerBY - centerAY
		distance := math.Sqrt(dx*dx + dy*dy)
		fmt.Printf("ROI A-B center distance: %.4f px, time: %s\n", distance, time.Now())
	}
}

func (s *HIKCameraService) GetLatestFrameForROI() (map[string]any, error) {
	s.frameMu.RLock()
	defer s.frameMu.RUnlock()

	if len(s.latestFrame) == 0 {
		return nil, fmt.Errorf("暂无相机图像，请先打开相机并等待图像流")
	}

	return map[string]any{
		"image":  "data:image/jpeg;base64," + base64.StdEncoding.EncodeToString(s.latestFrame),
		"width":  s.frameWidth,
		"height": s.frameHeight,
	}, nil
}

func (s *HIKCameraService) OpenROISelector(label string) error {
	app := application.Get()
	if app == nil {
		return fmt.Errorf("application not initialized")
	}

	_, err := s.GetLatestFrameForROI()
	if err != nil {
		return err
	}

	app.Window.NewWithOptions(application.WebviewWindowOptions{
		Title:            "选择标记 " + label,
		Width:            1200,
		Height:           850,
		BackgroundColour: application.NewRGB(15, 23, 42),
		URL:              "/roi-selector?label=" + label,
	})
	return nil
}

func (s *HIKCameraService) SetROI(label string, rect ROIRect) error {
	if rect.Width <= 0 || rect.Height <= 0 {
		return fmt.Errorf("ROI 区域无效")
	}
	s.frameMu.RLock()
	latestFrame := make([]byte, len(s.latestFrame))
	copy(latestFrame, s.latestFrame)
	s.frameMu.RUnlock()
	if len(latestFrame) == 0 {
		return fmt.Errorf("暂无相机图像")
	}

	mat, err := gocv.IMDecode(latestFrame, gocv.IMReadColor)
	if err != nil || mat.Empty() {
		return fmt.Errorf("解析相机图像失败")
	}
	defer mat.Close()

	tracker := extensometer.NewExtensometerService()
	tracker.SetOrignalImg(mat)
	tracker.SetROI(extensometer.Rect2f{
		X:      rect.X,
		Y:      rect.Y,
		Width:  rect.Width,
		Height: rect.Height,
	})

	s.roiMu.Lock()
	defer s.roiMu.Unlock()
	switch label {
	case "A":
		s.roiA = tracker
	case "B":
		s.roiB = tracker
	default:
		return fmt.Errorf("未知 ROI 标记: %s", label)
	}

	s.updateTrackerAxesLocked()

	if app := application.Get(); app != nil {
		app.Event.Emit("hik_roi_selected", map[string]any{
			"label": label,
			"roi":   rect,
		})
	}
	return nil
}

func (s *HIKCameraService) updateTrackerAxesLocked() {
	if s.roiA == nil || s.roiB == nil {
		return
	}
	axis := roiPairTrackingAxis(s.roiA.CurrentROI(), s.roiB.CurrentROI())
	s.roiA.SetTrackingAxis(axis)
	s.roiB.SetTrackingAxis(axis)
}

func roiPairTrackingAxis(roiA, roiB extensometer.Rect2f) extensometer.TrackingAxis {
	centerAX := roiA.X + roiA.Width/2
	centerAY := roiA.Y + roiA.Height/2
	centerBX := roiB.X + roiB.Width/2
	centerBY := roiB.Y + roiB.Height/2
	if math.Abs(centerBX-centerAX) >= math.Abs(centerBY-centerAY) {
		return extensometer.TrackingAxisVertical
	}
	return extensometer.TrackingAxisHorizontal
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
	return s.camera.CloseCamera()
}

func cameraModeToString(value uint32) string {
	switch value {
	case 1:
		return "once_auto"
	case 2:
		return "continuous_auto"
	default:
		return "manual"
	}
}

func cameraModeToValue(mode string) uint32 {
	switch mode {
	case "once_auto":
		return 1
	case "continuous_auto":
		return 2
	default:
		return 0
	}
}

func (s *HIKCameraService) GetCameraParams() HIKCameraParams {
	var p HIKCameraParams

	if value, _, err := s.camera.GetEnumValue("ExposureAuto"); err == nil {
		fmt.Println("GetExposureAuto:", cameraModeToString(value))
		p.ExposureAuto = value
	}
	if value, _, _, err := s.camera.GetFloatValue("ExposureTime"); err == nil {
		fmt.Println("GetExposureTime:", value)
		p.ExposureTime = value
	}
	if value, _, err := s.camera.GetEnumValue("GainAuto"); err == nil {
		fmt.Println("GetGainAuto:", cameraModeToString(value))
		p.GainAuto = value
	}
	if value, _, _, err := s.camera.GetFloatValue("Gain"); err == nil {
		fmt.Println("GetGain:", value)
		p.Gain = value
	}
	if value, _, _, err := s.camera.GetFloatValue("DigitalShift"); err == nil {
		fmt.Println("GetDigitalGain:", value)
		p.DigitalGain = value
		p.DigitalGainEnable = value > 0
	}
	if value, err := s.camera.GetBoolValue("DigitalShiftEnable"); err == nil {
		fmt.Println("GetDigitalGainEnable:", value)
		p.DigitalGainEnable = value
	}
	if value, _, _, err := s.camera.GetFloatValue("ResultingFrameRate"); err == nil {
		fmt.Println("GetFrameRate:", value)
		p.FrameRate = value
	}
	if value, err := s.camera.GetBoolValue("GammaEnable"); err == nil {
		fmt.Println("GetGammaEnable:", value)
		p.GammaEnable = value
	}
	if value, _, _, err := s.camera.GetFloatValue("Gamma"); err == nil {
		fmt.Println("GetGamma:", value)
		p.Gamma = value
	}
	if value, _, err := s.camera.GetEnumValue("BalanceWhiteAuto"); err == nil {
		fmt.Println("GetBalanceWhiteAuto:", cameraModeToString(value))
		p.BalanceWhiteAuto = value
	}
	if err := s.camera.SetEnumValue("BalanceRatioSelector", 0); err == nil {
		fmt.Println("SetBalanceRatioSelector:", 0)
		if value, _, _, _, err := s.camera.GetIntValue("BalanceRatio"); err == nil {
			fmt.Println("GetBalanceWhite:", value)
			p.BalanceWhite = value
		}
	}

	return p
}

func (s *HIKCameraService) SetExposureAuto(mode string) error {

	return s.camera.SetEnumValue("ExposureAuto", cameraModeToValue(mode))
}

func (s *HIKCameraService) SetExposureTime(value float64) error {
	fmt.Println("SetExposureTime:", value)
	return s.camera.SetFloatValue("ExposureTime", value)
}

func (s *HIKCameraService) SetGainAuto(mode string) error {
	fmt.Println("SetGainAuto:", cameraModeToString(cameraModeToValue(mode)))
	return s.camera.SetEnumValue("GainAuto", cameraModeToValue(mode))
}

func (s *HIKCameraService) SetGain(value float64) error {
	fmt.Println("SetGain:", value)
	return s.camera.SetFloatValue("Gain", value)
}

func (s *HIKCameraService) SetDigitalGainEnabled(enabled bool) error {
	return s.camera.SetBoolValue("DigitalShiftEnable", enabled)
}

func (s *HIKCameraService) SetDigitalGain(value float64) error {
	fmt.Println("SetDigitalGain:", value)
	return s.camera.SetFloatValue("DigitalShift", value)
}

func (s *HIKCameraService) SetGammaEnabled(enabled bool) error {
	return s.camera.SetBoolValue("GammaEnable", enabled)
}

func (s *HIKCameraService) SetGamma(value float64) error {
	return s.camera.SetFloatValue("Gamma", value)
}

func (s *HIKCameraService) SetBalanceWhiteAuto(mode string) error {
	return s.camera.SetEnumValue("BalanceWhiteAuto", cameraModeToValue(mode))
}

func (s *HIKCameraService) SetBalanceWhiteAutoByString(value string) error {
	return s.camera.SetEnumValueByString("BalanceWhiteAuto", value)
}

func (s *HIKCameraService) SetBalanceRatio(selector string, value int64) error {
	if err := s.camera.SetEnumValueByString("BalanceRatioSelector", selector); err != nil {
		return err
	}
	return s.camera.SetIntValue("BalanceRatio", value)
}

func (s *HIKCameraService) SetWhiteBalanceRed(value int64) error {
	return s.SetBalanceRatio("Red", value)
}

func (s *HIKCameraService) SetWhiteBalanceGreen(value int64) error {
	return s.SetBalanceRatio("Green", value)
}

func (s *HIKCameraService) SetWhiteBalanceBlue(value int64) error {
	return s.SetBalanceRatio("Blue", value)
}

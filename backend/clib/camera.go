package clib

/*
#cgo CFLAGS: -I${SRCDIR}/includes
#cgo LDFLAGS: -L${SRCDIR} -lMvCameraControl
#include "hik_wrapper.h"

#include <stdio.h>
#include <stdlib.h>
#include <string.h>

// 声明外部 Go 函数
extern void goOnImageReceived(unsigned char * pData, MV_FRAME_OUT_INFO_EX* pFrameInfo, void* pUser);

// C 端回调桥接函数
static void __stdcall OnImageReceivedCBridge(unsigned char * pData, MV_FRAME_OUT_INFO_EX* pFrameInfo, void* pUser) {
    goOnImageReceived(pData, pFrameInfo, pUser);
}

// 注册回调的辅助函数
static int RegisterCallbackBridge(void* handle, void* pUser) {
    return MV_CC_RegisterImageCallBackEx(handle, OnImageReceivedCBridge, pUser);
}

// JPEG 转换压缩函数
static int ConvertAndSaveToJpeg(void* handle, MV_FRAME_OUT_INFO_EX* pFrameInfo, unsigned char* pSrcData, unsigned char* pDstBuffer, unsigned int nDstBufferSize, unsigned int* nActualLen) {
    MV_SAVE_IMAGE_PARAM_EX stSaveParam = {0};
    stSaveParam.enImageType = MV_Image_Jpeg;
    stSaveParam.nWidth = pFrameInfo->nWidth;
    stSaveParam.nHeight = pFrameInfo->nHeight;
    stSaveParam.pData = pSrcData;
    stSaveParam.nDataLen = pFrameInfo->nFrameLen;
    stSaveParam.enPixelType = pFrameInfo->enPixelType;
    stSaveParam.pImageBuffer = pDstBuffer;
    stSaveParam.nBufferSize = nDstBufferSize;
    stSaveParam.nJpgQuality = 75;

    int ret = MV_CC_SaveImageEx2(handle, &stSaveParam);
    if (ret == 0) {
        *nActualLen = stSaveParam.nImageLen;
    }
    return ret;
}
*/
import "C"
import (
	"fmt"
	"sync"
	"sync/atomic"
	"unsafe"
)

// ImageCallback 图像数据回调函数类型
type ImageCallback func(data []byte, frameId uint64)

// Camera 相机结构体 - 只负责采集，不与前端直接交互
type Camera struct {
	OpenFlag     bool
	GrabbingFlag bool
	handle       unsafe.Pointer
	devicesCache map[string]*C.MV_CC_DEVICE_INFO

	// 高性能图像缓冲机制
	ImgMu     sync.Mutex
	ImgBuffer []byte
	FrameId   uint64
	ImageCb   ImageCallback // 图像数据回调
}

// 全局变量
var (
	cameraMap = make(map[unsafe.Pointer]*Camera)
	mapMu     sync.RWMutex
)

// NewCamera 创建相机实例
func NewCamera() *Camera {
	return &Camera{
		handle: nil,
	}
}

// SetImageCallback 设置图像数据回调
func (c *Camera) SetImageCallback(cb ImageCallback) {
	c.ImageCb = cb
}

// Init 初始化 SDK
func (c *Camera) Init() error {
	ret := C.MV_CC_Initialize()
	if uint32(ret) != 0 {
		return fmt.Errorf("SDK 初始化失败: 0x%08x", uint32(ret))
	}
	fmt.Println("SDK 初始化成功")
	return nil
}

// Finalize 释放 SDK 资源
func (c *Camera) Finalize() error {
	C.MV_CC_Finalize()
	fmt.Println("SDK 资源已释放")
	return nil
}

// GetCameraDevices 枚举设备
func (c *Camera) GetCameraDevices() ([]string, error) {
	var stDeviceList C.MV_CC_DEVICE_INFO_LIST
	C.memset(unsafe.Pointer(&stDeviceList), 0, C.sizeof_MV_CC_DEVICE_INFO_LIST)

	var nTLayerType uint32 = C.MV_GIGE_DEVICE | C.MV_USB_DEVICE |
		C.MV_GENTL_GIGE_DEVICE | C.MV_GENTL_CAMERALINK_DEVICE |
		C.MV_GENTL_CXP_DEVICE | C.MV_GENTL_XOF_DEVICE

	ret := C.MV_CC_EnumDevices(C.uint(nTLayerType), &stDeviceList)
	if uint32(ret) != 0 {
		return nil, fmt.Errorf("枚举失败: 0x%08x", uint32(ret))
	}

	c.devicesCache = make(map[string]*C.MV_CC_DEVICE_INFO)
	var names []string

	count := int(stDeviceList.nDeviceNum)
	for i := 0; i < count; i++ {
		pDeviceInfo := stDeviceList.pDeviceInfo[i]
		if pDeviceInfo == nil {
			continue
		}

		var modelName, serialNum string

		switch uint32(pDeviceInfo.nTLayerType) {
		case uint32(C.MV_GIGE_DEVICE):
			gigeInfo := (*C.MV_GIGE_DEVICE_INFO)(unsafe.Pointer(&pDeviceInfo.SpecialInfo[0]))
			modelName = C.GoString((*C.char)(unsafe.Pointer(&gigeInfo.chModelName[0])))
			serialNum = C.GoString((*C.char)(unsafe.Pointer(&gigeInfo.chSerialNumber[0])))
		case uint32(C.MV_USB_DEVICE):
			usbInfo := (*C.MV_USB3_DEVICE_INFO)(unsafe.Pointer(&pDeviceInfo.SpecialInfo[0]))
			modelName = C.GoString((*C.char)(unsafe.Pointer(&usbInfo.chModelName[0])))
			serialNum = C.GoString((*C.char)(unsafe.Pointer(&usbInfo.chSerialNumber[0])))
		default:
			modelName = "Unknown"
			serialNum = fmt.Sprintf("Idx_%d", i)
		}

		displayName := fmt.Sprintf("[%s] %s", modelName, serialNum)
		c.devicesCache[displayName] = pDeviceInfo
		names = append(names, displayName)
	}

	return names, nil
}

// OpenCamera 打开设备
func (c *Camera) OpenCamera(name string) error {
	pstDeviceInfo, ok := c.devicesCache[name]
	if !ok {
		fmt.Printf("设备名 '%s' 不在缓存中，无法打开\n", name)
		c.GetCameraDevices()
		pstDeviceInfo, ok = c.devicesCache[name]
		if !ok {
			return fmt.Errorf("设备 '%s' 不存在", name)
		}
	}

	if pstDeviceInfo == nil {
		return fmt.Errorf("设备信息为空")
	}

	ret := C.MV_CC_CreateHandle(&c.handle, pstDeviceInfo)
	if uint32(ret) != 0 {
		return fmt.Errorf("创建句柄失败: 0x%08x", uint32(ret))
	}

	ret = C.MV_CC_OpenDevice(c.handle, 1, 0)
	if uint32(ret) != 0 {
		C.MV_CC_DestroyHandle(c.handle)
		c.handle = nil
		return fmt.Errorf("打开设备失败: 0x%08x", uint32(ret))
	}

	mapMu.Lock()
	cameraMap[c.handle] = c
	mapMu.Unlock()

	C.RegisterCallbackBridge(c.handle, c.handle)

	c.CameraStartGrabbing()
	c.OpenFlag = true

	fmt.Printf("相机 [%s] 打开成功\n", name)
	return nil
}

// CloseCamera 关闭设备
func (c *Camera) CloseCamera() error {
	if c.handle == nil {
		return nil
	}

	if c.GrabbingFlag {
		c.CameraStopGrabbing()
	}

	C.MV_CC_CloseDevice(c.handle)

	mapMu.Lock()
	delete(cameraMap, c.handle)
	mapMu.Unlock()

	C.MV_CC_DestroyHandle(c.handle)
	c.handle = nil
	c.OpenFlag = false
	fmt.Println("相机已关闭并销毁句柄")
	return nil
}

// CameraStartGrabbing 开始取流
func (c *Camera) CameraStartGrabbing() error {
	ret := C.MV_CC_StartGrabbing(c.handle)
	if uint32(ret) != 0 {
		return fmt.Errorf("开启取流失败: 0x%08x", uint32(ret))
	}
	c.GrabbingFlag = true
	return nil
}

// CameraStopGrabbing 停止取流
func (c *Camera) CameraStopGrabbing() error {
	ret := C.MV_CC_StopGrabbing(c.handle)
	if uint32(ret) != 0 {
		return fmt.Errorf("停止取流失败: 0x%08x", uint32(ret))
	}
	c.GrabbingFlag = false
	return nil
}

// IsOpened 返回相机打开状态
func (c *Camera) IsOpened() bool {
	return c.OpenFlag
}

//export goOnImageReceived
func goOnImageReceived(pData *C.uchar, pFrameInfo *C.MV_FRAME_OUT_INFO_EX, pUser unsafe.Pointer) {
	mapMu.RLock()
	cam, ok := cameraMap[pUser]
	mapMu.RUnlock()

	if !ok || cam == nil {
		return
	}

	// 1. 估算输出缓冲区大小并实施复用机制
	destSize := uint32(pFrameInfo.nWidth)*uint32(pFrameInfo.nHeight)*3 + 2048

	// 加锁安全写入该相机的私有缓冲区
	cam.ImgMu.Lock()
	if uint32(cap(cam.ImgBuffer)) < destSize {
		cam.ImgBuffer = make([]byte, destSize)
	}
	cam.ImgBuffer = cam.ImgBuffer[:destSize]

	var actualLen C.uint

	// 2. 将 C 数据直接压缩并灌入预留的 Go 内存切片中
	ret := C.ConvertAndSaveToJpeg(
		cam.handle,
		pFrameInfo,
		pData,
		(*C.uchar)(unsafe.Pointer(&cam.ImgBuffer[0])),
		C.uint(destSize),
		&actualLen,
	)

	if uint32(ret) != 0 {
		cam.ImgMu.Unlock()
		fmt.Printf("图像转换/压缩失败: 0x%08x\n", uint32(ret))
		return
	}

	// 3. 截取实际有效长度的 JPEG 数据
	cam.ImgBuffer = cam.ImgBuffer[:int(actualLen)]
	atomic.AddUint64(&cam.FrameId, 1)
	frameId := atomic.LoadUint64(&cam.FrameId)

	// 复制数据用于回调（避免持有锁）
	dataCopy := make([]byte, len(cam.ImgBuffer))
	copy(dataCopy, cam.ImgBuffer)
	cam.ImgMu.Unlock()

	// 4. 通过回调将数据传递给上层（HIKCamera.go）
	if cam.ImageCb != nil {
		cam.ImageCb(dataCopy, frameId)
	}
}

package clib

/*
#cgo CFLAGS: -I${SRCDIR}/includes
#cgo LDFLAGS: -L${SRCDIR} -lMvCameraControl
#include "hik_wrapper.h"

#include <stdio.h>
#include <stdlib.h> // 用于 C.free
#include <string.h>

// --- C 桥接逻辑：用于处理 C 到 Go 的回调转发 ---
// 声明一个外部 Go 函数
extern void goOnImageReceived(unsigned char * pData, MV_FRAME_OUT_INFO_EX* pFrameInfo, void* pUser);

// 定义 C 端的回调函数，它会调用上面的 Go 函数
static void __stdcall OnImageReceivedCBridge(unsigned char * pData, MV_FRAME_OUT_INFO_EX* pFrameInfo, void* pUser) {
    goOnImageReceived(pData, pFrameInfo, pUser);
}

// 辅助函数：封装注册回调的逻辑
static int RegisterCallbackBridge(void* handle, void* pUser) {
    return MV_CC_RegisterImageCallBackEx(handle, OnImageReceivedCBridge, pUser);
}

// 在你的 C 块中添加
static int ConvertAndSaveToJpeg(void* handle, MV_FRAME_OUT_INFO_EX* pFrameInfo, unsigned char* pSrcData, unsigned char* pDstBuffer, unsigned int nDstBufferSize, unsigned int* nActualLen) {
    // 1. 准备 JPEG 转换参数
    MV_SAVE_IMAGE_PARAM_EX stSaveParam = {0};
    stSaveParam.enImageType = MV_Image_Jpeg;
    stSaveParam.nWidth = pFrameInfo->nWidth;
    stSaveParam.nHeight = pFrameInfo->nHeight;
    stSaveParam.pData = pSrcData;
    stSaveParam.nDataLen = pFrameInfo->nFrameLen;
    stSaveParam.enPixelType = pFrameInfo->enPixelType;

    stSaveParam.pImageBuffer = pDstBuffer;
    stSaveParam.nBufferSize = nDstBufferSize;
    stSaveParam.nJpgQuality = 75; // 质量设为 75，平衡清晰度与传输速度

    int ret = MV_CC_SaveImageEx2(handle, &stSaveParam);
    if (ret == 0) {
        *nActualLen = stSaveParam.nImageLen;
    }
    return ret;
}
*/
import "C"
import (
	"context"
	"fmt"
	"sync"
	"unsafe"
)

// IndustryCamara 模拟 C++ 的 IndustryCamara 类
type Camera struct {
	OpenFlag     bool
	GrabbingFlag bool

	handle       unsafe.Pointer
	devicesCache map[string]*C.MV_CC_DEVICE_INFO

	// 图片接收回调函数，参数改为：数据, 宽, 高, 像素格式
	OnImageReceived func(data []byte, width, height int, pixelType uint32)

	ctx context.Context
}

// 全局变量用于管理相机实例，供回调时查找对应的 Go 对象
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

func (c *Camera) SetContext(ctx context.Context) {
	c.ctx = ctx
}

// InitSDK 初始化 SDK 环境
func (c *Camera) Init() error {
	ret := C.MV_CC_Initialize()
	if uint32(ret) != 0 {
		return fmt.Errorf("SDK 初始化失败: 0x%08x", uint32(ret))
	}
	fmt.Println("SDK 初始化成功")
	return nil
}

// FinalizeSDK 释放 SDK 资源
func (c *Camera) Finalize() error {
	C.MV_CC_Finalize()
	fmt.Println("SDK 资源已释放")
	return nil
}

// EnumDevices 枚举设备
func (c *Camera) GetCameraDevices() ([]string, error) {
	var stDeviceList C.MV_CC_DEVICE_INFO_LIST
	C.memset(unsafe.Pointer(&stDeviceList), 0, C.sizeof_MV_CC_DEVICE_INFO_LIST)

	// 涵盖所有可能的传输层
	var nTLayerType uint32 = C.MV_GIGE_DEVICE | C.MV_USB_DEVICE |
		C.MV_GENTL_GIGE_DEVICE | C.MV_GENTL_CAMERALINK_DEVICE |
		C.MV_GENTL_CXP_DEVICE | C.MV_GENTL_XOF_DEVICE

	ret := C.MV_CC_EnumDevices(C.uint(nTLayerType), &stDeviceList)
	if uint32(ret) != 0 {
		return nil, fmt.Errorf("枚举失败: 0x%08x", uint32(ret))
	}

	// 清空旧缓存
	c.devicesCache = make(map[string]*C.MV_CC_DEVICE_INFO)
	var names []string

	count := int(stDeviceList.nDeviceNum)
	for i := 0; i < count; i++ {
		pDeviceInfo := stDeviceList.pDeviceInfo[i]
		if pDeviceInfo == nil {
			continue
		}

		var modelName, serialNum string

		// 根据传输层类型提取信息 (对应 C++ 的 SpecialInfo 处理)
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

		// 生成展示名称：[型号] 序列号
		displayName := fmt.Sprintf("[%s] %s", modelName, serialNum)

		// 保存到缓存映射
		c.devicesCache[displayName] = pDeviceInfo
		names = append(names, displayName)
	}

	return names, nil
}

// Open 打开设备
func (c *Camera) OpenCamera(name string) error {
	// 1. 从缓存中获取真正的 C 指针
	pstDeviceInfo, ok := c.devicesCache[name]
	if !ok {
		fmt.Printf("设备名 '%s' 不在缓存中，无法打开\n", name)
		c.GetCameraDevices() // 刷新设备列表并缓存
		pstDeviceInfo, ok = c.devicesCache[name]
		if !ok {
			return fmt.Errorf("设备 '%s' 不存在", name)
		}
	}

	if pstDeviceInfo == nil {
		return fmt.Errorf("设备信息为空")
	}

	// 2. 创建句柄
	ret := C.MV_CC_CreateHandle(&c.handle, pstDeviceInfo)
	if uint32(ret) != 0 {
		return fmt.Errorf("创建句柄失败: 0x%08x", uint32(ret))
	}

	// 3. 打开设备 (独占访问)
	ret = C.MV_CC_OpenDevice(c.handle, 1, 0)
	if uint32(ret) != 0 {
		C.MV_CC_DestroyHandle(c.handle)
		c.handle = nil
		return fmt.Errorf("打开设备失败: 0x%08x", uint32(ret))
	}

	// 4. 注册全局映射用于回调
	mapMu.Lock()
	cameraMap[c.handle] = c
	mapMu.Unlock()

	// 5. 注册桥接回调
	C.RegisterCallbackBridge(c.handle, c.handle)

	c.OpenFlag = true
	fmt.Printf("相机 [%s] 打开成功\n", name)
	return nil
}

// Close 关闭设备
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

// StartGrabbing 开始取流
func (c *Camera) CameraStartGrabbing() error {
	ret := C.MV_CC_StartGrabbing(c.handle)
	if uint32(ret) != 0 {
		return fmt.Errorf("开启取流失败: 0x%08x", uint32(ret))
	}
	c.GrabbingFlag = true
	return nil
}

// StopGrabbing 停止取流
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

func (c *Camera) GetDICStrain() (float64, error) {
	// 这里可以添加获取 DIC 应变数据的逻辑
	// 目前返回一个模拟值
	return 0.0, nil
}

//export goOnImageReceived
func goOnImageReceived(pData *C.uchar, pFrameInfo *C.MV_FRAME_OUT_INFO_EX, pUser unsafe.Pointer) {
	mapMu.RLock()
	cam, ok := cameraMap[pUser]
	mapMu.RUnlock()

	if !ok || cam.ctx == nil {
		return
	}

	// 1. 估算输出缓冲区大小 (Width * Height * 3 对于 JPG 来说绰绰有余)
	// 如果你的相机分辨率很大，建议把这个 buffer 变成 cam 结构体里的复用成员
	destSize := uint32(pFrameInfo.nWidth)*uint32(pFrameInfo.nHeight)*3 + 2048
	dstBuffer := make([]byte, destSize)
	var actualLen C.uint

	// 2. 调用封装好的 C 函数进行格式转换并压缩为 JPEG
	// 海康 SDK 的 SaveImageEx2 会自动处理从各种 PixelType (Bayer/Mono/YUV) 到 JPEG 的转换
	ret := C.ConvertAndSaveToJpeg(
		cam.handle,
		pFrameInfo,
		pData,
		(*C.uchar)(unsafe.Pointer(&dstBuffer[0])),
		C.uint(destSize),
		&actualLen,
	)

	if uint32(ret) != 0 {
		fmt.Printf("图像转换/压缩失败: 0x%08x\n", uint32(ret))
		return
	}

	// 3. 截取有效数据并转 Base64
	// 注意：在 Wails 中，如果帧率太高，Base64 会导致 CPU 占用极高
	// imgBase64 := base64.StdEncoding.EncodeToString(dstBuffer[:int(actualLen)])

	// 4. 推送到前端
	// runtime.EventsEmit(cam.ctx, "live_frame", "data:image/jpeg;base64,"+imgBase64)
}

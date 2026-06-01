#ifndef DICEXTENSOMETER_H
#define DICEXTENSOMETER_H

#ifdef __cplusplus
extern "C" {
#endif

#include "opencv2/opencv.hpp"
#include <vector>

typedef struct DICExtensometer DICExtensometer;

// 回调函数类型
typedef void (*ResultCallback)(float x, float y, void* userData);
typedef void (*ROINotFoundCallback)(void* userData);

struct DICExtensometer {
    // 从 Extensometer 继承的字段
    cv::Rect ROI;        // 感兴趣区域，粗精度
    cv::Rect2f ROIF;     // 感兴趣区域，细精度
    bool debugFlag;
    bool smoothFlag;
    
    // 回调函数
    ResultCallback resultCallback;
    ROINotFoundCallback roiNotFoundCallback;
    void* userData;
    
    // DICExtensometer 自身的字段
    cv::Mat originalImg;       // 永远的第一张图片 (绝对参考帧)
    cv::Mat lastFrameMat;      // 保持接口兼容保留
    
    // 多点阵列相关配置
    int pointsCount;           // 一列上的子区数量
    int spacing;               // 子区中心点之间的像素间距
    cv::Size subsetSize;       // 每个 DIC 子区的大小
    
    // 保存初始帧（参考帧）中各个子区的模板图像和初始坐标
    std::vector<cv::Mat> refSubsets;
    std::vector<cv::Point2f> initialPoints;
    
    bool isInitialized;        // 标记是否已经初始化了多点阵列
};

// 创建和销毁函数
DICExtensometer* DICExtensometer_Create();
void DICExtensometer_Destroy(DICExtensometer* dic);

// 公共方法
void DICExtensometer_SetOrignalImg(DICExtensometer* dic, cv::Mat ori);
void DICExtensometer_SetROI(DICExtensometer* dic, cv::Rect roi);
void DICExtensometer_Uninit(DICExtensometer* dic);
void DICExtensometer_Init(DICExtensometer* dic);
void DICExtensometer_CaculateImg(DICExtensometer* dic, cv::Mat img);

// 设置回调函数
void DICExtensometer_SetResultCallback(DICExtensometer* dic, ResultCallback cb, void* userData);
void DICExtensometer_SetROINotFoundCallback(DICExtensometer* dic, ROINotFoundCallback cb, void* userData);

#ifdef __cplusplus
}
#endif

#endif // DICEXTENSOMETER_H

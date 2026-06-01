#include "dicextensometer.h"
#include <iostream>
#include <algorithm>

using namespace std;

DICExtensometer* DICExtensometer_Create() {
    DICExtensometer* dic = new DICExtensometer();
    
    // 配置 DIC 参数
    dic->pointsCount = 9;
    dic->spacing = 15;
    dic->subsetSize = cv::Size(31, 31);
    dic->isInitialized = false;
    
    // 初始化基础字段
    dic->debugFlag = false;
    dic->smoothFlag = true;
    dic->resultCallback = NULL;
    dic->roiNotFoundCallback = NULL;
    dic->userData = NULL;
    
    return dic;
}

void DICExtensometer_Destroy(DICExtensometer* dic) {
    if (dic) {
        dic->refSubsets.clear();
        dic->initialPoints.clear();
        delete dic;
    }
}

void DICExtensometer_SetResultCallback(DICExtensometer* dic, ResultCallback cb, void* userData) {
    dic->resultCallback = cb;
    dic->userData = userData;
}

void DICExtensometer_SetROINotFoundCallback(DICExtensometer* dic, ROINotFoundCallback cb, void* userData) {
    dic->roiNotFoundCallback = cb;
    dic->userData = userData;
}

void DICExtensometer_SetOrignalImg(DICExtensometer* dic, cv::Mat ori) {
    dic->originalImg = ori.clone();
    if (dic->originalImg.channels() > 1) {
        cv::cvtColor(dic->originalImg, dic->originalImg, cv::COLOR_BGR2GRAY);
    }
    dic->lastFrameMat = dic->originalImg.clone();
    dic->isInitialized = false;
}

void DICExtensometer_SetROI(DICExtensometer* dic, cv::Rect roi) {
    dic->ROI = roi;
    dic->ROIF = roi;
    dic->isInitialized = false;
}

void DICExtensometer_Uninit(DICExtensometer* dic) {
    dic->isInitialized = false;
    dic->refSubsets.clear();
    dic->initialPoints.clear();
}

void DICExtensometer_Init(DICExtensometer* dic) {
    dic->lastFrameMat = dic->originalImg.clone();
    dic->isInitialized = false;
    std::cout << "DICExtensometer init" << std::endl;
}

// 核心多点 DIC 追踪逻辑
static bool calculateDisplacement(
    DICExtensometer* dic,
    const cv::Mat& img_prev,
    const cv::Mat& img_next,
    const cv::Point2f& roi_center,
    const cv::Size& roi_size,
    cv::Point2f& new_position,
    cv::Point2f& displacement
) {
    CV_Assert(img_prev.type() == CV_8UC1 && img_next.type() == CV_8UC1);

    // 1. 初始化多点阵列 (仅在第一帧执行一次)
    if (!dic->isInitialized) {
        dic->refSubsets.clear();
        dic->initialPoints.clear();

        int start_offset = -(dic->pointsCount / 2) * dic->spacing;

        for (int i = 0; i < dic->pointsCount; i++) {
            float px = roi_center.x;
            float py = roi_center.y + start_offset + i * dic->spacing;

            cv::Rect subsetRect(
                static_cast<int>(px - dic->subsetSize.width / 2.0f),
                static_cast<int>(py - dic->subsetSize.height / 2.0f),
                dic->subsetSize.width,
                dic->subsetSize.height
            );

            if (subsetRect.x >= 0 && subsetRect.y >= 0 &&
                subsetRect.x + subsetRect.width < img_prev.cols &&
                subsetRect.y + subsetRect.height < img_prev.rows)
            {
                dic->initialPoints.push_back(cv::Point2f(px, py));
                dic->refSubsets.push_back(img_prev(subsetRect).clone());
            }
        }

        if (dic->initialPoints.empty()) {
            std::cerr << "[DIC] Failed to initialize any subsets within image bounds." << std::endl;
            return false;
        }
        dic->isInitialized = true;
    }

    // 2. 对每个点执行追踪
    std::vector<float> dx_list;
    std::vector<float> dy_list;

    for (size_t i = 0; i < dic->initialPoints.size(); ++i) {
        cv::Mat refPatch = dic->refSubsets[i];
        cv::Point2f initPt = dic->initialPoints[i];

        cv::Point2f predictedPt = initPt + displacement;

        int searchRadius = 15;
        cv::Rect searchRect(
            static_cast<int>(predictedPt.x - searchRadius - dic->subsetSize.width / 2.0f),
            static_cast<int>(predictedPt.y - searchRadius - dic->subsetSize.height / 2.0f),
            dic->subsetSize.width + 2 * searchRadius,
            dic->subsetSize.height + 2 * searchRadius
        );

        searchRect &= cv::Rect(0, 0, img_next.cols, img_next.rows);

        if (searchRect.width < refPatch.cols || searchRect.height < refPatch.rows) {
            continue;
        }

        cv::Mat searchArea = img_next(searchRect);
        cv::Mat matchResult;
        cv::matchTemplate(searchArea, refPatch, matchResult, cv::TM_CCOEFF_NORMED);

        double minVal, maxVal;
        cv::Point minLoc, maxLoc;
        cv::minMaxLoc(matchResult, &minVal, &maxVal, &minLoc, &maxLoc);

        if (maxVal < 0.6) {
            continue;
        }

        cv::Point2f coarsePt(searchRect.x + maxLoc.x + dic->subsetSize.width / 2.0f,
                             searchRect.y + maxLoc.y + dic->subsetSize.height / 2.0f);

        cv::Rect currentPatchRect(
            static_cast<int>(coarsePt.x - dic->subsetSize.width / 2.0f),
            static_cast<int>(coarsePt.y - dic->subsetSize.height / 2.0f),
            dic->subsetSize.width,
            dic->subsetSize.height
        );

        currentPatchRect &= cv::Rect(0, 0, img_next.cols, img_next.rows);
        if (currentPatchRect.size() != dic->subsetSize) continue;

        cv::Mat currentPatch = img_next(currentPatchRect);

        cv::Mat warpMatrix = cv::Mat::eye(2, 3, CV_32F);

        try {
            cv::findTransformECC(
                refPatch,
                currentPatch,
                warpMatrix,
                cv::MOTION_TRANSLATION,
                cv::TermCriteria(cv::TermCriteria::COUNT + cv::TermCriteria::EPS, 15, 0.001)
            );

            float final_dx = (coarsePt.x - initPt.x) + warpMatrix.at<float>(0, 2);
            float final_dy = (coarsePt.y - initPt.y) + warpMatrix.at<float>(1, 2);

            dx_list.push_back(final_dx);
            dy_list.push_back(final_dy);

        } catch (cv::Exception& e) {
            continue;
        }
    }

    // 3. 统计剔除与融合
    if (dx_list.empty()) {
        return false;
    }

    auto getMedian = [](std::vector<float>& v) -> float {
        if (v.empty()) return 0.0f;
        size_t n = v.size() / 2;
        std::nth_element(v.begin(), v.begin() + n, v.end());
        return v[n];
    };

    displacement.x = getMedian(dx_list);
    displacement.y = getMedian(dy_list);
    new_position = roi_center + displacement;

    // debug 绘制逻辑
    if (dic->debugFlag) {
        cv::Mat debugImg;
        cv::cvtColor(img_next, debugImg, cv::COLOR_GRAY2BGR);

        for (const auto& pt : dic->initialPoints) {
            cv::circle(debugImg, pt, 2, cv::Scalar(0, 0, 255), -1);
        }
        cv::circle(debugImg, new_position, 4, cv::Scalar(0, 255, 0), -1);
        cv::rectangle(debugImg,
                      cv::Rect(new_position.x - roi_size.width/2, new_position.y - roi_size.height/2, roi_size.width, roi_size.height),
                      cv::Scalar(255, 0, 0), 1);

        cv::imshow("DIC Tracking", debugImg);
        cv::waitKey(1);
    }

    return true;
}

void DICExtensometer_CaculateImg(DICExtensometer* dic, cv::Mat img) {
    if (img.empty()) return;

    cv::Mat grayImg;
    if (img.channels() == 3 || img.channels() == 4) {
        cv::cvtColor(img, grayImg, cv::COLOR_BGR2GRAY);
    } else {
        grayImg = img.clone();
    }

    if (dic->originalImg.empty()) {
        dic->originalImg = grayImg.clone();
        return;
    }

    double coarse_center_x = dic->ROI.x + dic->ROI.width / 2.0;
    double coarse_center_y = dic->ROI.y + dic->ROI.height / 2.0;
    cv::Point2f original_roi_center(coarse_center_x, coarse_center_y);
    cv::Size roi_size(dic->ROI.width, dic->ROI.height);

    cv::Point2f new_pos;
    cv::Point2f disp(dic->ROIF.x - dic->ROI.x, dic->ROIF.y - dic->ROI.y);

    bool success = calculateDisplacement(dic, dic->originalImg, grayImg, original_roi_center, roi_size, new_pos, disp);

    if (success) {
        dic->ROIF.x = dic->ROI.x + disp.x;
        dic->ROIF.y = dic->ROI.y + disp.y;

        double final_x = dic->ROIF.x + dic->ROIF.width / 2.0;
        double final_y = dic->ROIF.y + dic->ROIF.height / 2.0;

        // 调用回调函数
        if (dic->resultCallback) {
            dic->resultCallback((float)final_x, (float)final_y, dic->userData);
        }

        if (dic->debugFlag) {
            std::cout << final_x << " " << final_y << " ROIF" << std::endl;
        }
    } else {
        std::cerr << "[DIC] Tracking failed!" << std::endl;
        // 调用回调函数
        if (dic->roiNotFoundCallback) {
            dic->roiNotFoundCallback(dic->userData);
        }
    }
}

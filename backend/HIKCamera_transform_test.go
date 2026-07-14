package backend

import (
	"changeme/backend/extensometer"
	"testing"

	"gocv.io/x/gocv"
)

func TestApplyCalibrationTransformsUsesStoredPoseMatrix(t *testing.T) {
	service := &HIKCameraService{calibration: extensometer.NewCalibration()}

	src := gocv.NewMatWithSize(5, 5, gocv.MatTypeCV8U)
	defer src.Close()
	src.SetUCharAt(2, 2, 200)

	pose := gocv.Eye(3, 3, gocv.MatTypeCV64F)
	defer pose.Close()
	pose.SetDoubleAt(0, 2, 1)
	service.setPoseTransform(pose)
	defer service.clearCalibrationTransforms()

	got, err := service.applyCalibrationTransforms(src)
	if err != nil {
		t.Fatal(err)
	}
	defer got.Close()

	if got.GetUCharAt(2, 3) != 200 {
		t.Fatalf("translated pixel at (3,2) = %d, want 200", got.GetUCharAt(2, 3))
	}
}

func TestClearCalibrationTransformsEmptiesStoredMatrices(t *testing.T) {
	service := &HIKCameraService{calibration: extensometer.NewCalibration()}

	camera := gocv.Eye(3, 3, gocv.MatTypeCV64F)
	defer camera.Close()
	pose := gocv.Eye(3, 3, gocv.MatTypeCV64F)
	defer pose.Close()

	service.setCameraTransform(camera)
	service.setPoseTransform(pose)
	service.clearCalibrationTransforms()

	if !service.cameraTransformMat.Empty() {
		t.Fatal("camera transform was not cleared")
	}
	if !service.poseTransformMat.Empty() {
		t.Fatal("pose transform was not cleared")
	}
}

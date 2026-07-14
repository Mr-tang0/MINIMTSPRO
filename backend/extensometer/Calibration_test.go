package extensometer

import (
	"image"
	"reflect"
	"testing"

	"gocv.io/x/gocv"
)

func TestCalibrationIsStateless(t *testing.T) {
	if got := reflect.TypeOf(Calibration{}).NumField(); got != 0 {
		t.Fatalf("Calibration has %d cached fields, want 0", got)
	}
}

func TestCorrectImageAppliesTransform(t *testing.T) {
	src := gocv.NewMatWithSize(12, 16, gocv.MatTypeCV8U)
	defer src.Close()
	src.SetUCharAt(6, 8, 200)

	transform := gocv.Eye(3, 3, gocv.MatTypeCV64F)
	defer transform.Close()

	calibration := NewCalibration()
	got, err := calibration.CorrectImage(src, transform)
	if err != nil {
		t.Fatal(err)
	}
	defer got.Close()

	if got.Empty() {
		t.Fatal("corrected image is empty")
	}
	if got.Rows() != src.Rows() || got.Cols() != src.Cols() {
		t.Fatalf("corrected size = %dx%d, want %dx%d", got.Cols(), got.Rows(), src.Cols(), src.Rows())
	}
	if got.GetUCharAt(6, 8) != 200 {
		t.Fatalf("identity transform changed pixel = %d, want 200", got.GetUCharAt(6, 8))
	}
}

func TestComputePoseNormalizesVisualCornerOrder(t *testing.T) {
	corners := matFromPoints(t, []gocv.Point2f{
		{X: 10, Y: 90}, {X: 10, Y: 50}, {X: 10, Y: 10},
		{X: 50, Y: 90}, {X: 50, Y: 50}, {X: 50, Y: 10},
		{X: 90, Y: 90}, {X: 90, Y: 50}, {X: 90, Y: 10},
	})
	defer corners.Close()

	pose, err := NewCalibration().ComputePose(corners, image.Pt(3, 3), 1)
	if err != nil {
		t.Fatal(err)
	}
	defer pose.Close()

	topLeft := transformPoint(pose, gocv.Point2f{X: 10, Y: 10})
	topRight := transformPoint(pose, gocv.Point2f{X: 90, Y: 10})
	bottomRight := transformPoint(pose, gocv.Point2f{X: 90, Y: 90})
	bottomLeft := transformPoint(pose, gocv.Point2f{X: 10, Y: 90})

	if !(topLeft.X < topRight.X && topLeft.Y < bottomLeft.Y) {
		t.Fatalf("visual top-left was rotated: tl=%+v tr=%+v bl=%+v", topLeft, topRight, bottomLeft)
	}
	if !(topRight.Y < bottomRight.Y && bottomLeft.X < bottomRight.X) {
		t.Fatalf("visual corners do not preserve upright orientation: tr=%+v br=%+v bl=%+v", topRight, bottomRight, bottomLeft)
	}
}

func matFromPoints(t *testing.T, points []gocv.Point2f) gocv.Mat {
	t.Helper()

	vec := gocv.NewPoint2fVectorFromPoints(points)
	defer vec.Close()
	return gocv.NewMatFromPoint2fVector(vec, true)
}

func transformPoint(m gocv.Mat, p gocv.Point2f) gocv.Point2f {
	x := float64(p.X)
	y := float64(p.Y)
	w := m.GetDoubleAt(2, 0)*x + m.GetDoubleAt(2, 1)*y + m.GetDoubleAt(2, 2)
	return gocv.Point2f{
		X: float32((m.GetDoubleAt(0, 0)*x + m.GetDoubleAt(0, 1)*y + m.GetDoubleAt(0, 2)) / w),
		Y: float32((m.GetDoubleAt(1, 0)*x + m.GetDoubleAt(1, 1)*y + m.GetDoubleAt(1, 2)) / w),
	}
}

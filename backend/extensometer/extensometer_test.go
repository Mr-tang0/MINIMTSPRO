package extensometer

import (
	"image"
	"testing"

	"gocv.io/x/gocv"
)

func TestEffectiveSearchRadius(t *testing.T) {
	param := DefaultDICParam()

	if got := effectiveSearchRadius(param, 0, 0); got != param.SearchRadius {
		t.Fatalf("first search radius = %d, want %d", got, param.SearchRadius)
	}

	if got := effectiveSearchRadius(param, 2.2, -1.1); got != param.TrackingSearchRadius {
		t.Fatalf("tracking search radius = %d, want %d", got, param.TrackingSearchRadius)
	}

	param.TrackingSearchRadius = 0
	if got := effectiveSearchRadius(param, 2.2, -1.1); got != param.SearchRadius {
		t.Fatalf("disabled tracking radius = %d, want %d", got, param.SearchRadius)
	}
}

func TestOddKernel(t *testing.T) {
	tests := []struct {
		in   int
		want int
	}{
		{0, 3},
		{2, 3},
		{3, 3},
		{4, 5},
		{5, 5},
	}

	for _, tt := range tests {
		if got := oddKernel(tt.in); got != tt.want {
			t.Fatalf("oddKernel(%d) = %d, want %d", tt.in, got, tt.want)
		}
	}
}

func TestPreprocessGrayForDICReturnsSameSizeType(t *testing.T) {
	src := gocv.NewMatWithSize(32, 40, gocv.MatTypeCV8U)
	defer src.Close()
	for y := 0; y < src.Rows(); y++ {
		for x := 0; x < src.Cols(); x++ {
			src.SetUCharAt(y, x, uint8((x*9+y*7)%251))
		}
	}

	tracker := NewExtensometerService()
	out := tracker.PreprocessGrayForDIC(src)
	defer out.Close()

	if out.Empty() {
		t.Fatal("preprocessed image is empty")
	}
	if out.Rows() != src.Rows() || out.Cols() != src.Cols() {
		t.Fatalf("preprocessed size = %dx%d, want %dx%d", out.Cols(), out.Rows(), src.Cols(), src.Rows())
	}
	if out.Type() != src.Type() {
		t.Fatalf("preprocessed type = %d, want %d", out.Type(), src.Type())
	}
}

func TestRunDICGrayMatTracksShift(t *testing.T) {
	ref := gocv.NewMatWithSize(120, 120, gocv.MatTypeCV8U)
	defer ref.Close()
	def := gocv.NewMatWithSize(120, 120, gocv.MatTypeCV8U)
	defer def.Close()

	for y := 40; y < 80; y++ {
		for x := 45; x < 75; x++ {
			v := uint8((x*7 + y*11) % 251)
			ref.SetUCharAt(y, x, v)
			def.SetUCharAt(y+4, x+6, v)
		}
	}

	tracker := NewExtensometerService()
	tracker.SetOrignalImg(ref)
	tracker.SetROI(Rect2f{X: 50, Y: 45, Width: 20, Height: 30})

	result, err := tracker.RunDICGrayMat(def)
	if err != nil {
		t.Fatal(err)
	}
	got := result.FrameResults[len(result.FrameResults)-1]
	if got.SubsetCount == 0 {
		t.Fatal("no tracked subsets")
	}
	if len(tracker.mEngines) != len(tracker.mSubsets) || len(tracker.mEngines) == 0 {
		t.Fatalf("IC-GN engines = %d, subsets = %d; want one engine per subset", len(tracker.mEngines), len(tracker.mSubsets))
	}
	if got.U < 5.5 || got.U > 6.5 || got.V < 3.5 || got.V > 4.5 {
		t.Fatalf("shift = (%.3f, %.3f), want about (6, 4)", got.U, got.V)
	}

	_ = image.Point{}
}

func TestRunDICGrayBytesTracksShift(t *testing.T) {
	ref := gocv.NewMatWithSize(120, 120, gocv.MatTypeCV8U)
	defer ref.Close()
	def := gocv.NewMatWithSize(120, 120, gocv.MatTypeCV8U)
	defer def.Close()

	for y := 35; y < 85; y++ {
		for x := 40; x < 80; x++ {
			v := uint8((x*5 + y*17 + x*y) % 251)
			ref.SetUCharAt(y, x, v)
			def.SetUCharAt(y+3, x+5, v)
		}
	}

	tracker := NewExtensometerService()
	tracker.SetOrignalImg(ref)
	tracker.SetROI(Rect2f{X: 50, Y: 45, Width: 20, Height: 30})

	result, err := tracker.RunDICGrayBytes(def.ToBytes(), def.Cols(), def.Rows())
	if err != nil {
		t.Fatal(err)
	}
	got := result.FrameResults[len(result.FrameResults)-1]
	if got.U < 4.5 || got.U > 5.5 || got.V < 2.5 || got.V > 3.5 {
		t.Fatalf("shift = (%.3f, %.3f), want about (5, 3)", got.U, got.V)
	}
}

func TestRunDICPreparedGrayMatTracksShift(t *testing.T) {
	ref := gocv.NewMatWithSize(120, 120, gocv.MatTypeCV8U)
	defer ref.Close()
	def := gocv.NewMatWithSize(120, 120, gocv.MatTypeCV8U)
	defer def.Close()

	for y := 35; y < 85; y++ {
		for x := 40; x < 80; x++ {
			v := uint8((x*13 + y*19 + x*y) % 251)
			ref.SetUCharAt(y, x, v)
			def.SetUCharAt(y+2, x+4, v)
		}
	}

	tracker := NewExtensometerService()
	tracker.SetOrignalImg(ref)
	tracker.SetROI(Rect2f{X: 50, Y: 45, Width: 20, Height: 30})

	prepared := tracker.PreprocessGrayForDIC(def)
	defer prepared.Close()
	result, err := tracker.RunDICPreparedGrayMat(prepared)
	if err != nil {
		t.Fatal(err)
	}
	got := result.FrameResults[len(result.FrameResults)-1]
	if got.U < 3.5 || got.U > 4.5 || got.V < 1.5 || got.V > 2.5 {
		t.Fatalf("shift = (%.3f, %.3f), want about (4, 2)", got.U, got.V)
	}
}

func TestInitMultiPointsUsesContinuousVerticalColumnInsideROI(t *testing.T) {
	ref := gocv.NewMatWithSize(120, 120, gocv.MatTypeCV8U)
	defer ref.Close()

	for y := 0; y < ref.Rows(); y++ {
		for x := 0; x < ref.Cols(); x++ {
			ref.SetUCharAt(y, x, uint8((x*x+y*13+x*y)%251))
		}
	}

	tracker := NewExtensometerService()
	tracker.SetOrignalImg(ref)
	tracker.SetROI(Rect2f{X: 40, Y: 30, Width: 20, Height: 40})

	if !tracker.initMultiPoints() {
		t.Fatal("initMultiPoints failed")
	}

	half := tracker.dicParam.SubsetSize / 2
	want := int(tracker.ROIF.Height) - 2*half + 1
	if len(tracker.mInitialPoints) != want {
		t.Fatalf("points = %d, want %d", len(tracker.mInitialPoints), want)
	}
	for i, pt := range tracker.mInitialPoints {
		if pt[0] != tracker.ROIF.X+tracker.ROIF.Width/2 {
			t.Fatalf("point %d x = %.1f, want center x %.1f", i, pt[0], tracker.ROIF.X+tracker.ROIF.Width/2)
		}
		if pt[1]-float64(half) < tracker.ROIF.Y || pt[1]+float64(half) > tracker.ROIF.Y+tracker.ROIF.Height {
			t.Fatalf("point %d y %.1f allows subset outside ROI %+v", i, pt[1], tracker.ROIF)
		}
		if i > 0 && pt[1]-tracker.mInitialPoints[i-1][1] != 1 {
			t.Fatalf("point %d spacing = %.1f, want 1", i, pt[1]-tracker.mInitialPoints[i-1][1])
		}
	}
}

func TestInitMultiPointsUsesContinuousHorizontalRowInsideROI(t *testing.T) {
	ref := gocv.NewMatWithSize(120, 120, gocv.MatTypeCV8U)
	defer ref.Close()

	for y := 0; y < ref.Rows(); y++ {
		for x := 0; x < ref.Cols(); x++ {
			ref.SetUCharAt(y, x, uint8((x*17+y*y+x*y)%251))
		}
	}

	tracker := NewExtensometerService()
	tracker.SetOrignalImg(ref)
	tracker.SetROI(Rect2f{X: 30, Y: 40, Width: 50, Height: 24})
	tracker.SetTrackingAxis(TrackingAxisHorizontal)

	if !tracker.initMultiPoints() {
		t.Fatal("initMultiPoints failed")
	}

	half := tracker.dicParam.SubsetSize / 2
	want := int(tracker.ROIF.Width) - 2*half + 1
	if len(tracker.mInitialPoints) != want {
		t.Fatalf("points = %d, want %d", len(tracker.mInitialPoints), want)
	}
	for i, pt := range tracker.mInitialPoints {
		if pt[1] != tracker.ROIF.Y+tracker.ROIF.Height/2 {
			t.Fatalf("point %d y = %.1f, want center y %.1f", i, pt[1], tracker.ROIF.Y+tracker.ROIF.Height/2)
		}
		if pt[0]-float64(half) < tracker.ROIF.X || pt[0]+float64(half) > tracker.ROIF.X+tracker.ROIF.Width {
			t.Fatalf("point %d x %.1f allows subset outside ROI %+v", i, pt[0], tracker.ROIF)
		}
		if i > 0 && pt[0]-tracker.mInitialPoints[i-1][0] != 1 {
			t.Fatalf("point %d spacing = %.1f, want 1", i, pt[0]-tracker.mInitialPoints[i-1][0])
		}
	}
}

package backend

import (
	"changeme/backend/clib"
	"fmt"
)

type HIKCamera struct {
	camera *clib.Camera
}

func NewHIKCamera() *HIKCamera {
	camera := clib.NewCamera()
	camera.Init()
	devices, _ := camera.GetCameraDevices()
	fmt.Println(devices)
	return &HIKCamera{
		camera: camera,
	}
}

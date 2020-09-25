package image

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"path/filepath"
	"runtime"
	"sort"
	"strings"
)

type computeImage struct {
	CreationTimestamp string
	Image             string
	Family            string
	DiskSizeGib       int64
}

// ImageInfo holds information about compute images.
type ImageInfo struct {
	imagesByFamily map[string][]computeImage
	imagesDiskSize map[string]int64
}

// ReadComputeImagesInfo reads the JSON file with information about compute images.
func ReadComputeImagesInfo() (*ImageInfo, error) {
	imgInfo := &ImageInfo{}
	// Get path of the json file relative to this directory.
	_, callerFile, _, _ := runtime.Caller(0)
	inputPath := filepath.Dir(callerFile) + "/compute_images.json"

	data, err := ioutil.ReadFile(inputPath)
	if err != nil {
		return nil, err
	}

	var jsonMap []computeImage
	json.Unmarshal(data, &jsonMap)

	imgInfo.imagesByFamily = map[string][]computeImage{}
	imgInfo.imagesDiskSize = map[string]int64{}
	for _, img := range jsonMap {
		if imgInfo.imagesByFamily[img.Family] == nil {
			imgInfo.imagesByFamily[img.Family] = []computeImage{}
		}
		imgInfo.imagesByFamily[img.Family] = append(imgInfo.imagesByFamily[img.Family], img)
		imgInfo.imagesDiskSize[img.Image] = img.DiskSizeGib
	}

	for k := range imgInfo.imagesByFamily {
		sort.SliceStable(imgInfo.imagesByFamily[k], func(i, j int) bool {
			return imgInfo.imagesByFamily[k][i].CreationTimestamp > imgInfo.imagesByFamily[k][j].CreationTimestamp
		})
	}
	return imgInfo, nil
}

func concreteImageVal(s string) string {
	i := strings.LastIndex(s, "/")
	if i < 0 {
		return s
	}
	return s[i+1:]
}

// GetImageDiskSize return the disk size for an image specified in any format allowed in google_compute_image resource.
func GetImageDiskSize(imgInfo *ImageInfo, img string) (int64, error) {
	if imgInfo == nil {
		return 0, fmt.Errorf("image information was not initialized")
	}

	img = concreteImageVal(img)

	// Check if it is family and return size if so.
	if l, ok := imgInfo.imagesByFamily[img]; ok {
		latest := l[0].Image
		return imgInfo.imagesDiskSize[latest], nil
	}

	// Check it is image type.
	size, ok := imgInfo.imagesDiskSize[img]
	if !ok {
		return 0, fmt.Errorf("invalid image specification '" + img + "'")
	}
	return size, nil
}

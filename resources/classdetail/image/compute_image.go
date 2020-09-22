package classdetail

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

type imageByDate []computeImage

var imagesByFamily map[string][]computeImage
var imagesDiskSize map[string]int64

func setComputeImagesInfo() error {
	_, callerFile, _, _ := runtime.Caller(0)
	inputPath := filepath.Dir(callerFile) + "/compute_images.json"

	data, err := ioutil.ReadFile(inputPath)
	if err != nil {
		return err
	}

	var jsonMap []computeImage
	json.Unmarshal(data, &jsonMap)

	imagesByFamily = map[string][]computeImage{}
	imagesDiskSize = map[string]int64{}
	for _, img := range jsonMap {
		if imagesByFamily[img.Family] == nil {
			imagesByFamily[img.Family] = []computeImage{}
		}
		imagesByFamily[img.Family] = append(imagesByFamily[img.Family], img)

		imagesDiskSize[img.Image] = img.DiskSizeGib
	}

	for k := range imagesByFamily {
		sort.SliceStable(imagesByFamily[k], func(i, j int) bool {
			return imagesByFamily[k][i].CreationTimestamp > imagesByFamily[k][j].CreationTimestamp
		})
	}
	return nil
}

func concreteImageVal(s string) string {
	i := strings.LastIndex(s, "/")
	if i < 0 {
		return s
	}
	return s[i+1:]
}

// GetImageDiskSize return the disk size for an image specified in any format allowed in google_compute_image resource.
func GetImageDiskSize(img string) (int64, error) {
	if imagesByFamily == nil || imagesDiskSize == nil {
		err := setComputeImagesInfo()
		if err != nil {
			return 0, err
		}
	}

	img = concreteImageVal(img)

	if l, ok := imagesByFamily[img]; ok {
		latest := l[0].Image
		return imagesDiskSize[latest], nil
	}

	size, ok := imagesDiskSize[img]
	if !ok {
		return 0, fmt.Errorf("invalid image specification '" + img + "'")
	}
	return size, nil
}

package classdetail

import (
	"fmt"
	"testing"
)

func TestGetImageDiskSize(t *testing.T) {
	tests := []struct {
		name      string
		imgFormat string
		size      int64
		err       error
	}{
		{"empty", "", 0, fmt.Errorf("invalid image specification ''")},
		{"invalid_0", "abcd", 0, fmt.Errorf("invalid image specification 'abcd'")},
		{"invalid_1", "rhel", 0, fmt.Errorf("invalid image specification 'rhel'")},
		{"image_0", "projects/{project}/global/images/rhel-8-v20200902", 20, nil},
		{"image_1", "global/images/windows-server-1909-dc-core-v20200813", 32, nil},
		{"image_2", "project/sql-2017-express-windows-2019-dc-v20200813", 50, nil},
		{"image_3", "cos-stable-81-12871-1190-0", 10, nil},
		{"family_0", "projects/{project}/global/images/family/centos-7", 20, nil},
		{"family_1", "global/images/family/sles-15-sp2-sap", 10, nil},
		{"family_2", "family/fedora-coreos-next", 8, nil},
		{"family_3", "project/ubuntu-minimal-2004-lts", 10, nil},
		{"family_4", "centos-7", 20, nil},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			size, err := GetImageDiskSize(test.imgFormat)

			// Test fails if errors have different values, messages or the return value differs from the expected one.
			f1 := (test.err == nil && err != nil) || (test.err != nil && err == nil)
			f2 := test.err != nil && err != nil && test.err.Error() != err.Error()
			f3 := size != test.size
			if f1 || f2 || f3 {
				t.Errorf("GetImageDiskSize(%s)= %+v, %+v ; want %+v %+v",
					test.imgFormat, size, err, test.size, test.err)
			}
		})
	}
}

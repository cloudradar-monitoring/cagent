package cagent

import (
	"reflect"
	"testing"

	"github.com/cloudradar-monitoring/cagent/pkg/common"
)

func Test_parseMdstatIntoMeasurements(t *testing.T) {
	tests := []struct {
		name    string
		data    string
		want    common.MeasurementsMap
		wantErr bool
	}{
		{"test1", `Personalities : [raid1] [raid6] [raid5] [raid4]
md_d0 : active raid5 sde1[0] sdf1[4] sdb1[5] sdd1[2] sdc1[1]
      1250241792 blocks super 1.2 level 5, 64k chunk, algorithm 2 [5/5] [UUUUU]
      bitmap: 0/10 pages [0KB], 16384KB chunk

unused devices: <none>`, common.MeasurementsMap{
			"md_d0.degraded":                  0,
			"md_d0.physicaldevice.state.sdb1": "active",
			"md_d0.physicaldevice.state.sdc1": "active",
			"md_d0.physicaldevice.state.sdd1": "active",
			"md_d0.physicaldevice.state.sde1": "active",
			"md_d0.physicaldevice.state.sdf1": "active",
			"md_d0.state":                     "active",
			"md_d0.type":                      "raid5"}, false},

		{"test2", `Personalities : [raid6] [raid5] [raid4]
md0 : active raid5 sda1[0] sdd1[2] sdb1[1]
     1465151808 blocks level 5, 64k chunk, algorithm 2 [4/3] [UUU_]
    unused devices: <none>`, common.MeasurementsMap{
			"md0.degraded":                  1,
			"md0.physicaldevice.missing":    1,
			"md0.physicaldevice.state.sda1": "active",
			"md0.physicaldevice.state.sdb1": "active",
			"md0.physicaldevice.state.sdd1": "active",
			"md0.state":                     "active",
			"md0.type":                      "raid5"}, false},

		{"test3", `Personalities : [raid1] [raid6] [raid5] [raid4]
md1 : active raid1 sdb2[1] sda2[0]
      136448 blocks [2/2] [UU]

md2 : active raid1 sdb3[1] sda3[0]
      129596288 blocks [2/2] [UU]

md3 : active raid5 sdl1[9] sdk1[8] sdj1[7] sdi1[6] sdh1[5] sdg1[4] sdf1[3] sde1[2] sdd1[1] sdc1[0]
      1318680576 blocks level 5, 1024k chunk, algorithm 2 [10/10] [UUUUUUUUUU]

md0 : active raid1 sdb1[1] sda1[0]
      16787776 blocks [2/2] [UU]

unused devices: <none> `, common.MeasurementsMap{
			"md0.degraded":                  0,
			"md0.physicaldevice.state.sda1": "active",
			"md0.physicaldevice.state.sdb1": "active",
			"md0.state":                     "active", "md0.type": "raid1",
			"md1.degraded":                  0,
			"md1.physicaldevice.state.sda2": "active",
			"md1.physicaldevice.state.sdb2": "active",
			"md1.state":                     "active",
			"md1.type":                      "raid1",
			"md2.degraded":                  0,
			"md2.physicaldevice.state.sda3": "active",
			"md2.physicaldevice.state.sdb3": "active",
			"md2.state":                     "active",
			"md2.type":                      "raid1",
			"md3.degraded":                  0,
			"md3.physicaldevice.state.sdc1": "active",
			"md3.physicaldevice.state.sdd1": "active",
			"md3.physicaldevice.state.sde1": "active",
			"md3.physicaldevice.state.sdf1": "active",
			"md3.physicaldevice.state.sdg1": "active",
			"md3.physicaldevice.state.sdh1": "active",
			"md3.physicaldevice.state.sdi1": "active",
			"md3.physicaldevice.state.sdj1": "active",
			"md3.physicaldevice.state.sdk1": "active",
			"md3.physicaldevice.state.sdl1": "active",
			"md3.state":                     "active",
			"md3.type":                      "raid5"}, false},

		{"test4", `Personalities : [raid1] [raid6] [raid5] [raid4]
md127 : active raid5 sdh1[6] sdg1[4] sdf1[3] sde1[2] sdd1[1] sdc1[0]
      1464725760 blocks level 5, 64k chunk, algorithm 2 [6/5] [UUUUU_]
      [==>..................]  recovery = 12.6% (37043392/292945152) finish=127.5min speed=33440K/sec

unused devices: <none>`, common.MeasurementsMap{
			"md127.degraded": 1, "md127.physicaldevice.missing": 0, "md127.physicaldevice.state.sdc1": "failed", "md127.physicaldevice.state.sdd1": "active", "md127.physicaldevice.state.sde1": "active", "md127.physicaldevice.state.sdf1": "active", "md127.physicaldevice.state.sdg1": "active", "md127.physicaldevice.state.sdh1": "active", "md127.state": "active", "md127.type": "raid5"}, false},

		{"test5", `Personalities : [linear] [raid0] [raid1] [raid5] [raid4] [raid6]
md0 : active raid6 sdf1[0] sde1[1] sdd1[2] sdc1[3] sdb1[4] sda1[5] hdb1[6]
      1225557760 blocks level 6, 256k chunk, algorithm 2 [7/7] [UUUUUUU]
      bitmap: 0/234 pages [0KB], 512KB chunk

unused devices: <none>`, common.MeasurementsMap{
			"md0.degraded":                  0,
			"md0.physicaldevice.state.hdb1": "active",
			"md0.physicaldevice.state.sda1": "active",
			"md0.physicaldevice.state.sdb1": "active",
			"md0.physicaldevice.state.sdc1": "active",
			"md0.physicaldevice.state.sdd1": "active",
			"md0.physicaldevice.state.sde1": "active",
			"md0.physicaldevice.state.sdf1": "active",
			"md0.state":                     "active",
			"md0.type":                      "raid6"}, false},

		{"test6", `Personalities : [raid1]
md1 : active raid1 sde1[6](F) sdg1[1] sdb1[4] sdd1[3] sdc1[2]
      488383936 blocks [6/4] [_UUUU_]

unused devices: <none>`, common.MeasurementsMap{
			"md1.degraded":                  1,
			"md1.physicaldevice.missing":    1,
			"md1.physicaldevice.state.sdb1": "active",
			"md1.physicaldevice.state.sdc1": "active",
			"md1.physicaldevice.state.sdd1": "active",
			"md1.physicaldevice.state.sde1": "failed",
			"md1.physicaldevice.state.sdg1": "active",
			"md1.state":                     "active",
			"md1.type":                      "raid1"}, false}}

	for _, tt := range tests {
		got := parseMdstat(tt.data)
		measurements := got.Measurements()
		if !reflect.DeepEqual(measurements, tt.want) {
			t.Errorf("%q. parseMdstat() = %v, want %v", tt.name, measurements, tt.want)
		}
	}
}

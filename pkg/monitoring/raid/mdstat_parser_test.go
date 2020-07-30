package raid

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// examples taken from https://raid.wiki.kernel.org/index.php/Mdstat

func TestParseMdstatSingleRaid(t *testing.T) {

	s := `Personalities : [raid1] [raid6] [raid5] [raid4]
md_d0 : active raid5 sde1[0] sdf1[4] sdb1[5] sdd1[2] sdc1[1]
		1250241792 blocks super 1.2 level 5, 64k chunk, algorithm 2 [5/5] [UUUUU]
		bitmap: 0/10 pages [0KB], 16384KB chunk

unused devices: <none>`

	ra := parseMdstat(s)

	assert.Equal(t, 1, len(ra))
	assert.Equal(t, "md_d0", ra[0].Name)
	assert.Equal(t, "raid5", ra[0].Type)
	assert.Equal(t, "active", ra[0].State)
	assert.Equal(t, 5, ra[0].RaidLevel)
	assert.Equal(t, []string{"sde1", "sdf1", "sdb1", "sdd1", "sdc1"}, ra[0].Devices)
	assert.Equal(t, []int(nil), ra[0].Inactive)
	assert.Equal(t, []int{0, 1, 2, 3, 4}, ra[0].Active)
	assert.Equal(t, []int(nil), ra[0].Failed)
	assert.Equal(t, false, ra[0].IsRebuilding)
}

func TestParseMdstatMultipleRaids(t *testing.T) {
	s := `Personalities : [raid1] [raid6] [raid5] [raid4]
md1 : active raid1 sdb2[1] sda2[0]
		136448 blocks [2/2] [UU]

md2 : active raid1 sdb3[1] sda3[0]
		129596288 blocks [2/2] [UU]

md3 : active raid5 sdl1[9] sdk1[8] sdj1[7] sdi1[6] sdh1[5] sdg1[4] sdf1[3] sde1[2] sdd1[1] sdc1[0]
		1318680576 blocks level 5, 1024k chunk, algorithm 2 [10/10] [UUUUUUUUUU]

md0 : active raid1 sdb1[1] sda1[0]
		16787776 blocks [2/2] [UU]

unused devices: <none>`

	ra := parseMdstat(s)

	assert.Equal(t, 4, len(ra))
	assert.Equal(t, "md1", ra[0].Name)
	assert.Equal(t, "raid1", ra[0].Type)

	assert.Equal(t, "md2", ra[1].Name)
	assert.Equal(t, "raid1", ra[1].Type)

	assert.Equal(t, "md3", ra[2].Name)
	assert.Equal(t, "raid5", ra[2].Type)

	assert.Equal(t, "md0", ra[3].Name)
	assert.Equal(t, "raid1", ra[3].Type)
}

func TestParseMdstatRaidLevelIssue(t *testing.T) {
	// this test asserts that issue in DEV-1593 is handled correctly

	s := `Personalities : [raid1] [raid0] [linear] [multipath] [raid6] [raid5] [raid4] [raid10]
md2 : inactive sdc5[1](S) sda5[3](S) sdb5[0](S)
		2942976 blocks super 1.2

md1 : active raid0 sdc2[1] sdb2[0]
		1019904 blocks super 1.2 512k chunks

md0 : active raid1 sdb1[0] sdc1[1]
		510976 blocks super 1.2 [2/2] [UU]

unused devices: <none>`

	ra := parseMdstat(s)

	assert.Equal(t, 3, len(ra))
	assert.Equal(t, "md2", ra[0].Name)
	assert.Equal(t, "", ra[0].Type)
	assert.Equal(t, "inactive", ra[0].State)
}

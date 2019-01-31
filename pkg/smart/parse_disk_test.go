// +build darwin

package smart

import (
	"bufio"
	"bytes"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSmartDisksParseEmpty(t *testing.T) {
	disk := ""
	buf := bytes.Buffer{}
	wr := bufio.NewWriter(&buf)
	_, err := wr.Write([]byte(disk))
	assert.NoError(t, err)
	err = wr.Flush()
	assert.NoError(t, err)

	var disks []string
	disks, err = parseDisks(&buf)
	assert.Error(t, err)
	assert.Equal(t, 0, len(disks))
}

func TestSmartDisksParseValidOneEntryMacOS(t *testing.T) {
	disk := `/dev/disk0 (internal):`
	buf := bytes.Buffer{}
	wr := bufio.NewWriter(&buf)
	_, err := wr.Write([]byte(disk))
	assert.NoError(t, err)
	err = wr.Flush()
	assert.NoError(t, err)

	var disks []string
	disks, err = parseDisks(&buf)
	assert.NoError(t, err)
	assert.Equal(t, 1, len(disks))
	assert.Equal(t, "/dev/disk0", disks[0])
}

func TestSmartDisksParseValidOneEntry(t *testing.T) {
	disk := `/dev/sda -d ata # /dev/sda, ATA device`
	buf := bytes.Buffer{}
	wr := bufio.NewWriter(&buf)
	_, err := wr.Write([]byte(disk))
	assert.NoError(t, err)
	err = wr.Flush()
	assert.NoError(t, err)

	var disks []string
	disks, err = parseDisks(&buf)
	assert.NoError(t, err)
	assert.Equal(t, 1, len(disks))
	assert.Equal(t, "/dev/sda", disks[0])
}

func TestSmartDisksParseValidTwoEntriesMacOS(t *testing.T) {
	disk := `/dev/sda -d ata # /dev/sda, ATA device
/dev/sdb -d sat # /dev/sdb [SAT], ATA device`
	buf := bytes.Buffer{}
	wr := bufio.NewWriter(&buf)
	_, err := wr.Write([]byte(disk))
	assert.NoError(t, err)
	err = wr.Flush()
	assert.NoError(t, err)

	var disks []string
	disks, err = parseDisks(&buf)
	assert.NoError(t, err)
	assert.Equal(t, 2, len(disks))
	assert.Equal(t, "/dev/sda", disks[0])
	assert.Equal(t, "/dev/sdb", disks[1])
}

func TestSmartDisksParsePartialValidOneEntry(t *testing.T) {
	disk := `/dev/disk0 (internal):/dev/disk2 (disk image):`
	buf := bytes.Buffer{}
	wr := bufio.NewWriter(&buf)
	_, err := wr.Write([]byte(disk))
	assert.NoError(t, err)
	err = wr.Flush()
	assert.NoError(t, err)

	var disks []string
	disks, err = parseDisks(&buf)
	assert.NoError(t, err)
	assert.Equal(t, 1, len(disks))
	assert.Equal(t, "/dev/disk0", disks[0])
}

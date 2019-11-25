package jobmon

import (
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/nightlyone/lockfile"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"

	"github.com/cloudradar-monitoring/cagent/pkg/common"
)

const (
	markerRunning         = "0"
	markerFinished        = "1"
	spoolEntryPermissions = 0666
	jsonExtension         = "json"
)

var ErrJobAlreadyRunning = errors.New("A job with same ID is already running")

type SpoolManager struct {
	dirPath string
	lock    *lockfile.Lockfile
	logger  *logrus.Logger
}

// NewSpoolManager creates a new object to manage jobmon spool
// dirPath must be absolute path
func NewSpoolManager(dirPath string, logger *logrus.Logger) *SpoolManager {
	lockFile, _ := lockfile.New(getLockFilePath(dirPath))
	return &SpoolManager{dirPath, &lockFile, logger}
}

func (s *SpoolManager) NewJob(r *JobRun) (string, error) {
	err := s.getLock()
	if err != nil {
		return "", err
	}
	defer s.releaseLock()

	alreadyRunning, err := s.isJobAlreadyRunning(r.ID)
	if err != nil {
		return "", err
	}
	if alreadyRunning {
		r.AddError(ErrJobAlreadyRunning.Error())
	}

	jsonBytes, err := json.Marshal(r)
	if err != nil {
		return "", err
	}

	uniqID := getUniqJobRunID(r.ID, alreadyRunning, r.StartedAt)
	filePath := s.getFilePath(uniqID)
	err = saveJSONFile(filePath, jsonBytes)
	if err != nil {
		return "", err
	}

	if alreadyRunning {
		err = ErrJobAlreadyRunning
	}

	return uniqID, err
}

func saveJSONFile(filePath string, jsonBytes []byte) error {
	err := ioutil.WriteFile(filePath, jsonBytes, spoolEntryPermissions)
	if err != nil {
		return errors.Wrapf(err, "while writing new file %s", filePath)
	}
	return nil
}

func (s *SpoolManager) FinishJob(uniqID string, r *JobRun) error {
	jsonBytes, err := json.Marshal(r)
	if err != nil {
		return err
	}

	err = s.getLock()
	if err != nil {
		return err
	}
	defer s.releaseLock()

	filePath := s.getFilePath(uniqID)
	newFilePath := s.getFilePath(getUniqJobRunID(r.ID, true, r.StartedAt))
	err = os.Rename(filePath, newFilePath)
	if err != nil {
		return errors.Wrapf(err, "could not mark job %s as finished", uniqID)
	}

	return saveJSONFile(newFilePath, jsonBytes)
}

func (s *SpoolManager) isJobAlreadyRunning(jobID string) (bool, error) {
	encodedJobID := encodeJobID(jobID)
	pattern := fmt.Sprintf("%s/%s_*_%s.%s", s.dirPath, markerRunning, encodedJobID, jsonExtension)
	matches, err := filepath.Glob(pattern)
	if err != nil {
		return false, errors.Wrapf(err, "while searching %s", pattern)
	}
	return len(matches) > 0, nil
}

func (s *SpoolManager) getLock() error {
	err := s.lock.TryLock()
	if err != nil {
		err = errors.Wrap(err, "could not get lock")
	}
	return err
}

func (s *SpoolManager) releaseLock() {
	err := s.lock.Unlock()
	if err != nil {
		s.logger.WithError(err).Error("could not release lock")
	}
}

func (s *SpoolManager) getFilePath(id string) string {
	return fmt.Sprintf("%s/%s.%s", s.dirPath, id, jsonExtension)
}

func getLockFilePath(dirPath string) string {
	return fmt.Sprintf("%s/spool.lock", dirPath)
}

func getUniqJobRunID(jobID string, isJobFinished bool, jobStartedAt common.Timestamp) string {
	marker := markerRunning
	if isJobFinished {
		marker = markerFinished
	}
	parts := []string{
		marker,
		strconv.FormatInt(time.Time(jobStartedAt).Unix(), 10),
		encodeJobID(jobID),
	}
	return strings.Join(parts, "_")
}

func encodeJobID(id string) string {
	return hex.EncodeToString([]byte(id))
}

package process

import (
	"fmt"
	"moduleab_agent/client"
	"moduleab_agent/common"
	"moduleab_agent/logger"
	"moduleab_server/models"
	"path"
	"path/filepath"
	"strings"
	"time"

	"github.com/aliyun/aliyun-oss-go-sdk/oss"
	"golang.org/x/exp/inotify"
)

type BackupManager struct {
	JobList []string
	client.AliConfig
	Watcher *inotify.Watcher
	host    *models.Hosts
}

func NewBackupManager(config client.AliConfig) (*BackupManager, error) {
	var err error
	b := new(BackupManager)
	b.JobList = make([]string, 0)
	b.AliConfig = config
	b.Watcher, err = inotify.NewWatcher()
	if err != nil {
		return nil, err
	}
	return b, nil
}

func (b *BackupManager) Update(ps []*models.Paths) error {
	for _, v := range ps {
		found := false
		for _, v0 := range b.JobList {
			if v.Path == v0 {
				found = true
				break
			}
		}
		if found {
			continue
		}
		err := b.Watcher.AddWatch(
			v.Path, inotify.IN_CLOSE_WRITE|inotify.IN_MOVED_TO)
		if err != nil {
			logger.AppLog.Warning("Monitor start failed:", err)
			continue
		}
		logger.AppLog.Info("Monitor for", v.Path, "started.")
		b.JobList = append(b.JobList, v.Path)
	}
	for k, v := range b.JobList {
		found := false
		for _, v0 := range ps {
			if v0.Path == v {
				found = true
				break
			}
		}
		if !found {
			err := b.Watcher.RemoveWatch(v)
			if err != nil {
				logger.AppLog.Warning("Monitor stop failed:", err)
				continue
			}
			logger.AppLog.Info("Monitor for", k, "stopped.")
			b.JobList = append(b.JobList[:k], b.JobList[k+1:]...)
		}
	}
	return nil
}

func (b *BackupManager) Run(h *models.Hosts) {
	logger.AppLog.Info("Backup process started.")
	for {
		select {
		case event := <-b.Watcher.Event:
			logger.AppLog.Debug("Get event:", event)
			for _, v := range h.Paths {
				filename, err := filepath.Abs(event.Name)
				if err != nil {
					logger.AppLog.Warn("Failed to fetch path:", err)
					continue
				}
				if strings.HasPrefix(filename, v.Path) {
					if v.BackupSet == nil {
						logger.AppLog.Warn("No BackupSet or AppSet got, skip")
						continue
					}
					_, file := path.Split(filename)
					record := &models.Records{
						Filename:   file,
						Host:       h,
						BackupSet:  v.BackupSet,
						AppSet:     h.AppSet,
						Path:       v,
						Type:       models.RecordTypeBackup,
						BackupTime: time.Now(),
					}
					logger.AppLog.Debug("Now is:", time.Now().Format(time.RFC3339))
					logger.AppLog.Debug("Record:", record)
					ossclient, err := oss.New(
						v.BackupSet.Oss.Endpoint, b.ApiKey, b.ApiSecret)
					if err != nil {
						logger.AppLog.Warn("Error while connect to oss:", err)
						fmt.Println("Error while connect to oss:", err)
						continue
					}
					bucket, err := ossclient.Bucket(v.BackupSet.Oss.BucketName)
					if err != nil {
						logger.AppLog.Warn(
							"Error while retrievaling bucket:", err)
						fmt.Println(
							"Error while retrievaling bucket:", err)
						continue
					}

					ps := strings.Split(path.Dir(record.GetFullPath()), "/")
					var (
						dir        string
						dirCreated = true
					)
					for _, p := range ps {
						dir = fmt.Sprintf("%s%s/", dir, p)
						err = bucket.PutObject(dir, strings.NewReader(""))
						if err != nil {
							logger.AppLog.Warn(
								"Error while making dir on bucket:", err)
							fmt.Println(
								"Error while making dir on bucket:", err)
							dirCreated = false
							break
						}
					}
					if !dirCreated {
						continue
					}
					err = bucket.PutObjectFromFile(
						record.GetFullPath(),
						event.Name,
						oss.Routines(common.UploadThreads),
					)
					if err != nil {
						logger.AppLog.Warn(
							"Error while uploading:", err)
						fmt.Println(
							"Error while uploading:", err)
						continue
					}
					err = client.UploadRecord(record)
					if err != nil {
						logger.AppLog.Warn(
							"Error while recording:", err)
						fmt.Println(
							"Error while recording:", err)
						continue
					}
				}
			}
		}
	}
}

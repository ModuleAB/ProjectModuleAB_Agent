package process

import (
	"fmt"
	"moduleab_agent/client"
	"moduleab_agent/logger"
	"moduleab_server/models"
	"path"
	"strings"
	"time"

	"github.com/aliyun/aliyun-oss-go-sdk/oss"
	"golang.org/x/exp/inotify"
)

type BackupManager struct {
	JobList []string
	Oss     *oss.Client
	Watcher *inotify.Watcher
}

func NewBackupManager(
	endpoint, apikey, apisecret string) (*BackupManager, error) {
	var err error
	b := new(BackupManager)
	b.JobList = make([]string, 0)
	b.Oss, err = oss.New(endpoint, apikey, apisecret)
	if err != nil {
		return nil, err
	}
	b.bucket = bucket
	b.Watcher, err = inotify.NewWatcher()
	if err != nil {
		return nil, err
	}
	return b
}

func (b *BackupManager) Update(ps []*models.Paths) error {
	for _, v := range ps {
		err := b.Watcher.AddWatch(
			v.Path, inotify.IN_CLOSE_WRITE|inotify.IN_MOVED_TO)
		if err != nil {
			logger.AppLog.Warning("Monitor start failed:", err)
			continue
		}
		b.JobList = append(b.JobList, v.Path)
	}
	for k, v := range b.JobList {
		found := false
		for _, v := range ps {
			if v.Path == v {
				found = true
				break
			}
		}
		if !found {
			b.Watcher.RemoveWatch(v.Path)
			if err != nil {
				logger.AppLog.Warning("Monitor stop failed:", err)
				continue
			}
			logger.AppLog.Info("Monitor for", k, "stopped.")
		}
		b.JobList[k] = ""
	}
	nJobList := make([]string, 0)
	for _, v := range b.JobList {
		if v != "" {
			nJobList = append(nJobList, v)
		}
	}
	b.JobList = nJobList
}

func (b *BackupManager) Run(h *models.Hosts) error {
	for {
		select {
		case event := <-b.Watcher.Event:
			for _, v := range h.Paths {
				if strings.HasPrefix(event.Name, v.Path) {
					record := &models.Records{
						Filename:   strings.Replace(event.Name, v.Path, "", -1),
						Host:       h,
						BackupSet:  v.BackupSet,
						AppSet:     h.AppSet,
						Path:       v.Path,
						Type:       models.RecordTypeBackup,
						BackupTime: time.Now(),
					}
					bucket, err := b.Oss.Bucket(v.BackupSet.Oss.BucketName)
					if err != nil {
						logger.AppLog.Warn("Error while retrievaling bucket:", err)
						continue
					}

					ps := strings.Split(path.Dir(record.GetFullPath()), "/")
					var (
						dir        string
						dirCreated = true
					)
					for _, p := range ps {
						dir = fmt.Sprintf("%s%s/", dir, p)
						err := bucket.PutObject(dir, strings.NewReader(""))
						if err != nil {
							logger.AppLog.Warn(
								"Error while making dir on bucket:", err)
							dirCreated = false
							break
						}
					}
					if !dirCreated {
						continue
					}
					err = bucket.UploadFile(
						record.GetFullPath(),
						event.Name,
						512*1024,
						oss.Routines(3),
						oss.Checkpoint(true, ""),
					)
					if err != nil {
						logger.AppLog.Warn(
							"Error while uploading:", err)
						continue
					}
					// TODO 注册相关信息
					err = client.UploadRecord(record)
					if err != nil {
						logger.AppLog.Warn(
							"Error while recording:", err)
						continue
					}
				}
			}
		}
	}
}

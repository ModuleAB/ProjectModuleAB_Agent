package process

import (
	"fmt"
	"moduleab_agent/client"
	"moduleab_agent/common"
	"moduleab_agent/logger"
	"moduleab_server/models"
	"path"
	"strings"

	"github.com/aliyun/aliyun-oss-go-sdk/oss"
	"github.com/gorilla/websocket"
)

func RunWebsocket(h *models.Hosts, apikey, apisecret string) {
	var err error
	defer logger.AppLog.Warn("Websocket got error:", err)
	defer func() {
		x := recover()
		if x != nil {
			err = fmt.Errorf("%v", x)
		}
	}()
	url := fmt.Sprintf(
		"/api/v1/client/signal/%s/ws",
		h.Name,
	)
	req, err := client.MakeRequest("GET", url, nil)
	if err != nil {
		return
	}
	wsurl := strings.Replace(req.URL.String(), "http://", "ws://", 1)
	conn, _, err := websocket.DefaultDialer.Dial(wsurl, req.Header)
	if err != nil {
		return
	}
	defer conn.Close()
	for {
		msg := make(models.Signal)
		err := conn.ReadJSON(&msg)
		if websocket.IsUnexpectedCloseError(err) {
			return
		} else {
			continue
		}
		go DoDownload(msg, apikey, apisecret)
		err = conn.WriteMessage(websocket.TextMessage, []byte("DONE"))
		if websocket.IsUnexpectedCloseError(err) {
			return
		} else {
			continue
		}
	}

}

func DoDownload(s models.Signal, apikey, apisecret string) {
	var downType int
	if Type, found := s["type"]; found {
		if _, ok := Type.(int); ok {
			downType = Type.(int)
		}
	}
	if downType == models.SignalTypeDownload {
		ossclient, err := oss.New(s["endpoint"].(string), apikey, apisecret)
		if err != nil {
			logger.AppLog.Warn("Got error:", err)
			return
		}
		bucket, err := ossclient.Bucket(s["bucket"].(string))
		if err != nil {
			logger.AppLog.Warn("Got error:", err)
			return
		}
		localPath := "/" + path.Join(
			strings.Split(s["path"].(string), "/")[2:]...,
		)
		err = bucket.DownloadFile(
			s["path"].(string),
			localPath,
			512*1024,
			oss.Routines(common.UploadThreads),
			oss.Checkpoint(true, ""),
		)
		if err != nil {
			logger.AppLog.Warn("Got error:", err)
			return
		}
		logger.AppLog.Info("Download to", localPath, "done.")
	} else {
		logger.AppLog.Warn("Bad signal type")
	}
}

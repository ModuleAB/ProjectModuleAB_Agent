package client

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"moduleab_server/models"
	"net/http"
)

func UploadRecord(r *models.Records) error {
	b, err := json.Marshal(r)
	if err != nil {
		return err
	}
	buf := bytes.NewReader(b)
	req, err := MakeRequest("POST", "/api/v1/records", buf)
	if err != nil {
		return err
	}
	resp, err := StdHttp.Do(req)
	if err != nil {
		return err
	}
	switch resp.StatusCode {
	case http.StatusForbidden:
		e := make(map[string]string)
		b, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return err
		}
		err = json.Unmarshal(b, &e)
		if err != nil {
			return err
		}
		return fmt.Errorf(e["error"])
	case http.StatusBadRequest:
		fallthrough
	case http.StatusInternalServerError:
		e := make(map[string]string)
		b, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return err
		}
		err = json.Unmarshal(b, &e)
		if err != nil {
			return err
		}
		return fmt.Errorf("%s %s", e["message"], e["error"])
	case http.StatusCreated:
		return nil
	default:
		return fmt.Errorf("Unknown error: %d", resp.StatusCode)
	}
	return fmt.Errorf("Unknown error.")
}

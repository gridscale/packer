package gridscale

import (
	"context"
	"errors"
	"io"

	"github.com/gridscale/gsclient-go/v3"
)

type StateBagMock struct {
	state map[string]interface{}
}

func (s StateBagMock) Get(s2 string) interface{} {
	return s.state[s2]
}

func (s StateBagMock) GetOk(s2 string) (interface{}, bool) {
	val, ok := s.state[s2]
	return val, ok
}

func (s StateBagMock) Put(s2 string, i interface{}) {
	s.state[s2] = i
}

func (s StateBagMock) Remove(s2 string) {
	delete(s.state, s2)
}

type StorageOperatorMock struct{}

func (s StorageOperatorMock) GetStorage(ctx context.Context, id string) (gsclient.Storage, error) {
	panic("implement me")
}

func (s StorageOperatorMock) GetStorageList(ctx context.Context) ([]gsclient.Storage, error) {
	panic("implement me")
}

func (s StorageOperatorMock) GetStoragesByLocation(ctx context.Context, id string) ([]gsclient.Storage, error) {
	panic("implement me")
}

func (s StorageOperatorMock) CreateStorage(ctx context.Context, body gsclient.StorageCreateRequest) (gsclient.CreateResponse, error) {
	if body.Name == "success" || body.Name == "success-secondary" {
		return gsclient.CreateResponse{
			ObjectUUID:  "test",
			RequestUUID: "test",
		}, nil
	}
	return gsclient.CreateResponse{}, errors.New("error")
}

func (s StorageOperatorMock) UpdateStorage(ctx context.Context, id string, body gsclient.StorageUpdateRequest) error {
	panic("implement me")
}

func (s StorageOperatorMock) CloneStorage(ctx context.Context, id string) (gsclient.CreateResponse, error) {
	panic("implement me")
}

func (s StorageOperatorMock) DeleteStorage(ctx context.Context, id string) error {
	if id == "success" {
		return nil
	}
	return errors.New("error")
}

func (s StorageOperatorMock) GetDeletedStorages(ctx context.Context) ([]gsclient.Storage, error) {
	panic("implement me")
}

func (s StorageOperatorMock) GetStorageEventList(ctx context.Context, id string) ([]gsclient.Event, error) {
	panic("implement me")
}

func (s StorageOperatorMock) CreateStorageFromBackup(ctx context.Context, backupID, storageName string) (gsclient.CreateResponse, error) {
	panic("implement me")
}

type uiMock struct {
	sayMessage   string
	errorMessage string
}

func (u *uiMock) Ask(s string) (string, error) {
	panic("implement me")
}

func (u *uiMock) Say(s string) {
	u.sayMessage = s
}

func (u *uiMock) Message(s string) {
	u.sayMessage = s
}

func (u *uiMock) Error(s string) {
	u.errorMessage = s
}

func (u *uiMock) Machine(s string, s2 ...string) {
	panic("implement me")
}

func (u *uiMock) TrackProgress(src string, currentSize, totalSize int64, stream io.ReadCloser) (body io.ReadCloser) {
	panic("implement me")
}

func produceTestConfig(raws map[string]interface{}) *Config {
	raws["ssh_username"] = "root"
	c, _, _ := NewConfig(raws)
	return c
}

package store

import "github.com/doganarif/govisual/internal/model"

type Store interface {
	Add(log *model.RequestLog)
	Get(id string) (*model.RequestLog, bool)
	GetAll() []*model.RequestLog
	GetLatest(n int) []*model.RequestLog
}

package repo

import (
	"context"
	"fmt"
)

type RepoFactoryImpl struct {
	SegmentCnt int
	DBConfig   DBConfig
}

type DBConfig struct {
	Host     string
	Port     int
	User     string
	Password string
	DBName   string
}

func NewRepoFactory(segmentCnt int, dbConfig DBConfig) RepositoryFactory {
	return RepoFactoryImpl{
		SegmentCnt: segmentCnt,
		DBConfig:   dbConfig,
	}
}

type RepoType string

var RepoTypeMem RepoType = "memory"
var RepoTypeMysql RepoType = "mysql"

func (r RepoFactoryImpl) GetRepository(ctx context.Context, repoType RepoType) (Repository, error) {
	switch repoType {
	case RepoTypeMem:
		return NewMemoryRepo(r.SegmentCnt), nil
	case RepoTypeMysql:
		return NewMySQLRepo(r.DBConfig)
	default:
		return nil, fmt.Errorf("unknown repo type: %s", repoType)
	}
}

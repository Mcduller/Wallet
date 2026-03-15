package repo

import (
	"context"
	"fmt"
)

type RepoFactoryImpl struct {
	SegmentCnt int
	DBUrl      string
}

func NewRepoFactory(segmentCnt int, DBUrl string) RepositoryFactory {
	return RepoFactoryImpl{
		SegmentCnt: segmentCnt,
		DBUrl:      DBUrl,
	}
}

type RepoType string

var RepoTypeMem RepoType = "memory"
var RepoTypeMysql RepoType = "mysql"

func (r RepoFactoryImpl) GetRepository(ctx context.Context, repoType RepoType) (Repository, error) {
	//TODO implement me
	switch repoType {
	case RepoTypeMem:
		return NewMemoryRepo(r.SegmentCnt), nil
	case RepoTypeMysql:
		// todo 替换为mysql实现
		return nil, nil
	default:
		return nil, fmt.Errorf("unknown repo type: %s", repoType)
	}
}

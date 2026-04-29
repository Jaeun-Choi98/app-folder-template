package service

import (
	"errors"
	"pjt/internal/config"
	dbhandler "pjt/internal/db/db-handler"
	"pjt/internal/infra/cache"
	"pjt/internal/service"
)

/**
 * Service layer responsibilities:
 * - Verifying referential integrity and unique key constraints
 * - Performing default value injection for PATCH API operations
 *
 * Examples:
 * - DELETE: If parent table, verify no referenced keys exist
 * - UPDATE/INSERT: If child table, verify reference keys exist and validate unique keys (logical or physical)
 *
 * The service layer acts as a view abstraction over both database and ram.Ram (in-memory database).
 *
 * Additionally, it processes nil values for PATCH API operations by injecting default values
 * retrieved from ram.Ram or database.
 */

var (
	errExistsNum  = errors.New("num already exists")
	errExistsID   = errors.New("ID already exists")
	errExistsName = errors.New("name already exists")

	// 아래 에러는 식별자가 아닌 값으로 검사를 진행할 때 사용
	errNotExistsName = errors.New("not exists name")

	errRefIntegrity = errors.New("referential integrity is not satisfied")

	errInvalidID = errors.New("invalid ID")
	errInvalidPW = errors.New("invalid PW")
)

type APIService struct {
	Dao    dbhandler.DBHandlerInterface
	Cache  *cache.Cache
	Config *config.Configuration
}

func NewAPIService(dao dbhandler.DBHandlerInterface, cache *cache.Cache, config *config.Configuration) service.APIServcieInterface {
	return &APIService{
		Dao:    dao,
		Cache:  cache,
		Config: config,
	}
}

func (a *APIService) Test() string {
	return a.Dao.Test()
}

func (a *APIService) Close() error {
	return a.Dao.Close()
}

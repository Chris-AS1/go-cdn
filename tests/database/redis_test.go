package database

import (
	"context"
	"fmt"
	"go-cdn/internal/config"
	"go-cdn/internal/database/repository"
	"go-cdn/internal/database/repository/redis"
	discovery "go-cdn/internal/discovery/controller"
	"go-cdn/pkg/model"
	"log"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
	"github.com/testcontainers/testcontainers-go"
	tc_rd "github.com/testcontainers/testcontainers-go/modules/redis"
)

const TEST_RD_PORT = 6379

type RedisContainer struct {
	*tc_rd.RedisContainer
}

func NewRedisContainer(ctx context.Context) (*RedisContainer, error) {
	redisContainer, err := tc_rd.RunContainer(ctx,
		testcontainers.WithImage("redis:7-alpine"),
	)
	if err != nil {
		return nil, err
	}

	return &RedisContainer{
		RedisContainer: redisContainer,
	}, nil
}

type RedisRepoTestSuite struct {
	suite.Suite
	redisContainer *RedisContainer
	repository     *redis.RedisRepository
	ctx            context.Context
}

func (suite *RedisRepoTestSuite) SetupSuite() {
	cfg, err := config.New()
	if err != nil {
		log.Fatal(err)
	}

	dcb, err := discovery.NewControllerBuilder().FromConfigs(cfg)
	if err != nil {
		log.Fatal(err)
	}
	dc := dcb.Build()

	suite.ctx = context.Background()

	redisContainer, err := NewRedisContainer(suite.ctx)
	if err != nil {
		log.Fatal(err)
	}

	suite.redisContainer = redisContainer
	// overrides the address read from configs
	ip, err := suite.redisContainer.ContainerIP(suite.ctx)
	if err != nil {
		log.Fatal(err)
	}
	cfg.Cache.RedisAddress = fmt.Sprintf("%s:%d", ip, TEST_RD_PORT)

	// skips the controller
	repository, err := redis.New(context.TODO(), dc, cfg)
	if err != nil {
		log.Fatal(err)
	}
	suite.repository = repository
}

func (suite *RedisRepoTestSuite) TearDownSuite() {
	if err := suite.redisContainer.Terminate(suite.ctx); err != nil {
		log.Fatalf("error terminating redis container: %s", err)
	}
}

func (suite *RedisRepoTestSuite) TestAddFile() {
	var err error
	t := suite.T()

	t.Run("TestAddFile", func(t *testing.T) {
		test_file := &model.StoredFile{
			IDHash:   "0001",
			Filename: "test",
			Content:  []byte{00, 10, 20},
		}
		err = suite.repository.AddFile(suite.ctx, test_file)
		assert.Nil(t, err)
	})

	t.Run("TestGetFile", func(t *testing.T) {
		stored_test_file, err := suite.repository.GetFile(suite.ctx, "0001")
		assert.Nil(t, err)
		assert.Equal(t, "0001", stored_test_file.IDHash)
		// filename is not stored
		assert.NotNil(t, stored_test_file.Content)
	})

	// Fetch a nonexistent file
	t.Run("TestGetFileNotFound", func(t *testing.T) {
		_, err := suite.repository.GetFile(suite.ctx, "0002")
		assert.ErrorIs(t, err, repository.ErrKeyDoesNotExist)
	})

	t.Run("TestRemoveFile", func(t *testing.T) {
		err = suite.repository.RemoveFile(suite.ctx, "0001")
		assert.Nil(t, err)
	})
}

func TestRedisRepoTestSuite(t *testing.T) {
	suite.Run(t, new(RedisRepoTestSuite))
}

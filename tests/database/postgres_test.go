package database

import (
	"context"
	"go-cdn/internal/config"
	"go-cdn/internal/database/repository"
	"go-cdn/internal/database/repository/postgres"
	discovery "go-cdn/internal/discovery/controller"
	"go-cdn/pkg/model"
	"log"
	"testing"
	"time"

	"github.com/stretchr/testify/suite"
	"github.com/testcontainers/testcontainers-go"
	tc_pg "github.com/testcontainers/testcontainers-go/modules/postgres"
	"github.com/testcontainers/testcontainers-go/wait"

	"github.com/stretchr/testify/assert"
)

const (
	TEST_DB_USERNAME = "pguser"
	TEST_DB_PASSWORD = "pgpassword"
	TEST_DB_DBNAME   = "test-db"
)

type PostgresContainer struct {
	*tc_pg.PostgresContainer
}

func CreatePostgresContainer(ctx context.Context) (*PostgresContainer, error) {
	pgContainer, err := tc_pg.RunContainer(ctx,
		testcontainers.WithImage("postgres:16-alpine"),

		tc_pg.WithDatabase(TEST_DB_DBNAME),
		tc_pg.WithUsername(TEST_DB_USERNAME),
		tc_pg.WithPassword(TEST_DB_PASSWORD),

		testcontainers.WithWaitStrategy(
			wait.ForLog("database system is ready to accept connections").
				WithOccurrence(2).WithStartupTimeout(5*time.Second)),
	)
	if err != nil {
		return nil, err
	}

	return &PostgresContainer{
		PostgresContainer: pgContainer,
	}, nil
}

type PostgresRepoTestSuite struct {
	suite.Suite
	pgContainer *PostgresContainer
	repository  *postgres.PostgresRepository
	ctx         context.Context
}

func (suite *PostgresRepoTestSuite) SetupSuite() {
	cfg, err := config.New()
	if err != nil {
		log.Println(err)
	}

	dcb, err := discovery.NewControllerBuilder().FromConfigs(cfg)
	if err != nil {
		log.Fatal(err)
	}
	dc := dcb.Build()

	suite.ctx = context.Background()

	pgContainer, err := CreatePostgresContainer(suite.ctx)
	if err != nil {
		log.Fatal(err)
	}

	suite.pgContainer = pgContainer
	// overrides the address read from configs
	cfg.Database.DatabaseAddress, err = suite.pgContainer.ContainerIP(suite.ctx)
	cfg.Database.DatabaseName = TEST_DB_DBNAME
	cfg.Database.DatabaseUsername = TEST_DB_USERNAME
	cfg.Database.DatabasePassword = TEST_DB_PASSWORD

	if err != nil {
		log.Fatal(err)
	}

	// skips the controller
	repository, err := postgres.New(context.TODO(), dc, cfg)
	if err != nil {
		log.Fatal(err)
	}
	suite.repository = repository
}

func (suite *PostgresRepoTestSuite) TearDownSuite() {
	if err := suite.pgContainer.Terminate(suite.ctx); err != nil {
		log.Fatalf("error terminating postgres container: %s", err)
	}
}

func (suite *PostgresRepoTestSuite) TestAddFile() {
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
		assert.Equal(t, "test", stored_test_file.Filename)
		assert.NotNil(t, stored_test_file.Content)
	})

	// Fetch a nonexistent file
	t.Run("TestGetFileNotFound", func(t *testing.T) {
		_, err = suite.repository.GetFile(suite.ctx, "0002")
		assert.ErrorIs(t, err, repository.ErrKeyDoesNotExist)
	})

	t.Run("TestRemoveFile", func(t *testing.T) {
		err = suite.repository.RemoveFile(suite.ctx, "0001")
		assert.Nil(t, err)
	})
}

func TestPostgresRepoTestSuite(t *testing.T) {
	suite.Run(t, new(PostgresRepoTestSuite))
}

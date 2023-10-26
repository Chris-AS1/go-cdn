package database

import (
	"fmt"
	"go-cdn/utils"
	"testing"
)

func TestDBConnection(t *testing.T) {
	utils.LoadEnv()
	r := dbConnection()
	fmt.Print(r)
}

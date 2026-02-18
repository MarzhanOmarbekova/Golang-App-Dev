package app

import (
	"Practice3/internal/_repository/_postgres"
	"Practice3/pkg/modules"
	"context"
	"fmt"
	"time"
)

func Run() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	dbConfig := initPosgreConfig()

	_postgre := _postgres.NewPGXDialect(ctx, dbConfig)

	fmt.Println(_postgre)
}

func initPosgreConfig() *modules.PostgreConfig {
	return &modules.PostgreConfig{
		Host:        "localhost",
		Port:        "5432",
		Username:    "postgres",
		Password:    "marzhan06",
		DBName:      "mydb",
		SSLMode:     "disable",
		ExecTimeout: 5 * time.Second,
	}
}

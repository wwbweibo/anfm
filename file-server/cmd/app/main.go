package main

import (
	"context"
	"fileserver/internal/controllers"
	"fileserver/internal/domain/file"
	"fileserver/internal/server"
	"fileserver/internal/tasks"
	"fileserver/utils"
	"log"

	domainFile "fileserver/internal/domain/file"

	"golang.org/x/sync/errgroup"
)

func main() {
	ctx, cancelF := context.WithCancel(context.Background())
	defer cancelF()
	config := &Config{}
	config.Load("config/config.yaml")

	dbconnection := utils.NewDbConnection()
	fileRepo := file.NewFileRepository(dbconnection)
	taskServer := initTaskServer(*config, fileRepo)
	ginServer := initGinServer(*config)
	errGroup := errgroup.Group{}
	errGroup.Go(func() error {
		return taskServer.Start(ctx)
	})
	errGroup.Go(func() error {
		return ginServer.Start(ctx)
	})
	if err := errGroup.Wait(); err != nil {
		log.Fatal(err)
	}
}

func initTaskServer(config Config,
	fileRepo domainFile.IFileRepository,
) *server.BackendTaskServer {
	taskServer := server.NewBackendTaskServer()
	fileScanOption := tasks.ScanOptions{}
	fileScanOption = fileScanOption.
		OptionRootPath(config.NasRootPath).
		OptionPlainPath(config.ScanOption.Path...).
		OptionRegexPath(config.ScanOption.RegexPath...).
		OptionExtensions(config.ScanOption.Extensions...)
	// register tasks here
	taskServer.RegisterTask(
		tasks.NewSysInitBackendTask(fileScanOption, fileRepo),
	)
	return taskServer
}

func initGinServer(config Config) *server.GinServer {
	server := server.NewGinServer()
	fileController := controllers.NewFileApiControllers()
	server.RegisterController(fileController)
	return server
}

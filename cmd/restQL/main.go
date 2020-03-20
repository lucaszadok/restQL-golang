package main

import (
	"context"
	"fmt"
	"github.com/b2wdigital/restQL-golang/internal/plataform/conf"
	"github.com/b2wdigital/restQL-golang/internal/plataform/logger"
	"github.com/b2wdigital/restQL-golang/internal/plataform/web"
	"github.com/pkg/errors"
	"github.com/valyala/fasthttp"
	"golang.org/x/sync/errgroup"
	"os"
	"os/signal"
	"syscall"
)

func main() {
	if err := start(); err != nil {
		fmt.Printf("[ERROR] failed to start api : %v", err)
		os.Exit(1)
	}
}

var build string

func start() error {
	//// =========================================================================
	//// Config
	//config := conf.New(build)
	cfg, err := conf.Load(build)
	if err != nil {
		return err
	}

	log := logger.New(os.Stdout, logger.LogOptions{
		Enable:    cfg.Logging.Enable,
		Timestamp: cfg.Logging.Timestamp,
		Level:     cfg.Logging.Level,
		Format:    cfg.Logging.Format,
	})
	//// =========================================================================
	//// Start API
	log.Info("initializing api")

	shutdownSignal := make(chan os.Signal, 1)
	signal.Notify(shutdownSignal, os.Interrupt, syscall.SIGTERM)

	api := &fasthttp.Server{
		Name:         "api",
		Handler:      web.API(log, cfg),
		TCPKeepalive: false,
		ReadTimeout:  cfg.Web.ReadTimeout,
	}
	health := &fasthttp.Server{
		Name:         "health",
		Handler:      web.Health(log, cfg),
		TCPKeepalive: false,
		ReadTimeout:  cfg.Web.ReadTimeout,
	}

	serverErrors := make(chan error, 1)
	go func() {
		log.Info("api listing", "port", cfg.Web.ApiAddr)
		serverErrors <- api.ListenAndServe(":" + cfg.Web.ApiAddr)
	}()

	go func() {
		defer log.Info("stopping health")
		log.Info("api health listing", "port", cfg.Web.ApiHealthAddr)
		serverErrors <- health.ListenAndServe(":" + cfg.Web.ApiHealthAddr)
	}()

	if cfg.Web.Env == "development" {
		debug := &fasthttp.Server{Name: "debug", Handler: web.Debug(log, cfg)}
		go func() {
			log.Info("api debug listing", "port", cfg.Web.DebugAddr)
			serverErrors <- debug.ListenAndServe(":" + cfg.Web.DebugAddr)
		}()
	}

	//// =========================================================================
	//// Shutdown
	select {
	case err := <-serverErrors:
		return errors.Wrap(err, "server error")
	case sig := <-shutdownSignal:
		log.Info("starting shutdown", "signal", sig)

		timeout, cancel := context.WithTimeout(context.Background(), cfg.Web.GracefulShutdownTimeout)
		defer cancel()
		err := shutdown(timeout, log, api, health)

		switch {
		case sig == syscall.SIGSTOP:
			return errors.New("integrity issue caused shutdown")
		case err != nil:
			return errors.Wrap(err, "could not stop server gracefully")
		}
	}

	return nil
}

func shutdown(ctx context.Context, log *logger.Logger, servers ...*fasthttp.Server) error {
	var groupErr error
	var g errgroup.Group
	done := make(chan struct{})

	go func() {
		groupErr = g.Wait()
		done <- struct{}{}
	}()

	for _, s := range servers {
		s := s
		g.Go(func() error {
			log.Debug("starting shutdown", "server", s.Name)
			err := s.Shutdown()
			if err != nil {
				log.Error(fmt.Sprintf("%s graceful shutdown did not complete", s.Name), err)
			}
			log.Debug("shutdown finished", "server", s.Name)
			return err
		})
	}

	select {
	case <-ctx.Done():
		return errors.New("graceful shutdown did not complete")
	case <-done:
		return groupErr
	}
}

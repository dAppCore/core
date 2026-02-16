package lab

import (
	"context"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"time"

	"forge.lthn.ai/core/cli/pkg/cli"
	"forge.lthn.ai/core/cli/pkg/lab"
	"forge.lthn.ai/core/cli/pkg/lab/collector"
	"forge.lthn.ai/core/cli/pkg/lab/handler"
)

func init() {
	cli.RegisterCommands(AddLabCommands)
}

var labCmd = &cli.Command{
	Use:   "lab",
	Short: "Homelab monitoring dashboard",
	Long:  "Lab dashboard with real-time monitoring of machines, training runs, models, and services.",
}

var (
	labBind string
)

var serveCmd = &cli.Command{
	Use:   "serve",
	Short: "Start the lab dashboard web server",
	Long:  "Starts the lab dashboard HTTP server with live-updating collectors for system stats, Docker, Forgejo, HuggingFace, InfluxDB, and more.",
	RunE:  runServe,
}

func init() {
	serveCmd.Flags().StringVar(&labBind, "bind", ":8080", "HTTP listen address")
}

// AddLabCommands registers the 'lab' command and subcommands.
func AddLabCommands(root *cli.Command) {
	labCmd.AddCommand(serveCmd)
	root.AddCommand(labCmd)
}

func runServe(cmd *cli.Command, args []string) error {
	cfg := lab.LoadConfig()
	cfg.Addr = labBind

	store := lab.NewStore()
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))

	// Setup collectors.
	reg := collector.NewRegistry(logger)
	reg.Register(collector.NewSystem(cfg, store), 60*time.Second)
	reg.Register(collector.NewPrometheus(cfg.PrometheusURL, store),
		time.Duration(cfg.PrometheusInterval)*time.Second)
	reg.Register(collector.NewHuggingFace(cfg.HFAuthor, store),
		time.Duration(cfg.HFInterval)*time.Second)
	reg.Register(collector.NewDocker(store),
		time.Duration(cfg.DockerInterval)*time.Second)

	if cfg.ForgeToken != "" {
		reg.Register(collector.NewForgejo(cfg.ForgeURL, cfg.ForgeToken, store),
			time.Duration(cfg.ForgeInterval)*time.Second)
	}

	reg.Register(collector.NewTraining(cfg, store),
		time.Duration(cfg.TrainingInterval)*time.Second)
	reg.Register(collector.NewServices(store), 60*time.Second)

	if cfg.InfluxToken != "" {
		reg.Register(collector.NewInfluxDB(cfg, store),
			time.Duration(cfg.InfluxInterval)*time.Second)
	}

	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt)
	defer cancel()
	reg.Start(ctx)
	defer reg.Stop()

	// Setup HTTP handlers.
	web := handler.NewWebHandler(store)
	api := handler.NewAPIHandler(store)

	mux := http.NewServeMux()

	// Web pages.
	mux.HandleFunc("GET /", web.Dashboard)
	mux.HandleFunc("GET /models", web.Models)
	mux.HandleFunc("GET /training", web.Training)
	mux.HandleFunc("GET /dataset", web.Dataset)
	mux.HandleFunc("GET /golden-set", func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, "/dataset", http.StatusMovedPermanently)
	})
	mux.HandleFunc("GET /runs", func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, "/training", http.StatusMovedPermanently)
	})
	mux.HandleFunc("GET /agents", web.Agents)
	mux.HandleFunc("GET /services", web.Services)

	// SSE for live updates.
	mux.HandleFunc("GET /events", web.Events)

	// JSON API.
	mux.HandleFunc("GET /api/status", api.Status)
	mux.HandleFunc("GET /api/models", api.Models)
	mux.HandleFunc("GET /api/training", api.Training)
	mux.HandleFunc("GET /api/dataset", api.GoldenSet)
	mux.HandleFunc("GET /api/golden-set", api.GoldenSet)
	mux.HandleFunc("GET /api/runs", api.Runs)
	mux.HandleFunc("GET /api/agents", api.Agents)
	mux.HandleFunc("GET /api/services", api.Services)
	mux.HandleFunc("GET /health", api.Health)

	srv := &http.Server{
		Addr:         cfg.Addr,
		Handler:      mux,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 10 * time.Second,
	}

	go func() {
		<-ctx.Done()
		logger.Info("shutting down")
		shutCtx, shutCancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer shutCancel()
		srv.Shutdown(shutCtx)
	}()

	logger.Info("lab dashboard starting", "addr", cfg.Addr)
	if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		return err
	}
	return nil
}

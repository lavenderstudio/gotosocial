// GoToSocial
// Copyright (C) GoToSocial Authors admin@gotosocial.org
// SPDX-License-Identifier: AGPL-3.0-or-later
//
// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU Affero General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// This program is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU Affero General Public License for more details.
//
// You should have received a copy of the GNU Affero General Public License
// along with this program.  If not, see <http://www.gnu.org/licenses/>.

//go:build !nootel

package observability

import (
	"context"
	"net/http"
	"os"
	"time"

	"code.superseriousbusiness.org/gopkg/log"
	"code.superseriousbusiness.org/gotosocial/internal/config"
	"code.superseriousbusiness.org/gotosocial/internal/gtserror"
	"code.superseriousbusiness.org/gotosocial/internal/state"

	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"go.opentelemetry.io/contrib/instrumentation/runtime"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/prometheus"
	"go.opentelemetry.io/otel/metric"
	sdk "go.opentelemetry.io/otel/sdk/metric"
	"go.opentelemetry.io/otel/sdk/metric/exemplar"
)

func InitializeMetrics(ctx context.Context, state *state.State) error {
	if !config.GetMetricsEnabled() {
		// Noop.
		return nil
	}

	r, err := Resource()
	if err != nil {
		// This can happen if semconv versioning is out-of-sync.
		return gtserror.Newf("error building tracing resource: %w", err)
	}

	// The exporter embeds a default OpenTelemetry Reader
	// and implements prometheus.Collector, allowing it
	// to be used as both a Reader and Collector.
	exporter, err := prometheus.New()
	if err != nil {
		return gtserror.Newf("error initializing prometheus: %w", err)
	}

	meterProvider := sdk.NewMeterProvider(
		sdk.WithExemplarFilter(exemplar.AlwaysOffFilter),
		sdk.WithResource(r),
		sdk.WithReader(exporter),
	)

	otel.SetMeterProvider(meterProvider)

	if err := runtime.Start(
		runtime.WithMeterProvider(meterProvider),
	); err != nil {
		return err
	}

	meter := meterProvider.Meter(serviceName)

	_, err = meter.Int64ObservableGauge(
		"gotosocial.instance.total_users",
		metric.WithDescription("Total number of users on this instance"),
		metric.WithInt64Callback(func(ctx context.Context, o metric.Int64Observer) error {
			userCount, err := state.DB.CountInstanceAccounts(ctx)
			if err != nil {
				return err
			}
			o.Observe(int64(userCount))
			return nil
		}),
	)
	if err != nil {
		return err
	}

	_, err = meter.Int64ObservableGauge(
		"gotosocial.instance.total_statuses",
		metric.WithDescription("Total number of statuses on this instance"),
		metric.WithInt64Callback(func(ctx context.Context, o metric.Int64Observer) error {
			statusCount, err := state.DB.CountInstanceStatuses(ctx)
			if err != nil {
				return err
			}
			o.Observe(int64(statusCount))
			return nil
		}),
	)
	if err != nil {
		return err
	}

	_, err = meter.Int64ObservableGauge(
		"gotosocial.instance.total_federating_instances",
		metric.WithDescription("Total number of other instances this instance is federating with"),
		metric.WithInt64Callback(func(ctx context.Context, o metric.Int64Observer) error {
			federatingCount, err := state.DB.CountInstancePeers(ctx)
			if err != nil {
				return err
			}
			o.Observe(int64(federatingCount))
			return nil
		}),
	)
	if err != nil {
		return err
	}

	_, err = meter.Int64ObservableGauge(
		"gotosocial.workers.delivery.count",
		metric.WithDescription("Current number of delivery workers"),
		metric.WithInt64Callback(func(ctx context.Context, o metric.Int64Observer) error {
			o.Observe(int64(state.Workers.Delivery.Len()))
			return nil
		}),
	)
	if err != nil {
		return err
	}

	_, err = meter.Int64ObservableUpDownCounter(
		"gotosocial.workers.delivery.queue",
		metric.WithDescription("Current number of queued delivery worker tasks"),
		metric.WithInt64Callback(func(ctx context.Context, o metric.Int64Observer) error {
			o.Observe(int64(state.Workers.Delivery.Queue.Len()))
			return nil
		}),
	)
	if err != nil {
		return err
	}

	_, err = meter.Int64ObservableGauge(
		"gotosocial.workers.dereference.count",
		metric.WithDescription("Current number of dereference workers"),
		metric.WithInt64Callback(func(ctx context.Context, o metric.Int64Observer) error {
			o.Observe(int64(state.Workers.Dereference.Len()))
			return nil
		}),
	)
	if err != nil {
		return err
	}

	_, err = meter.Int64ObservableUpDownCounter(
		"gotosocial.workers.dereference.queue",
		metric.WithDescription("Current number of queued dereference worker tasks"),
		metric.WithInt64Callback(func(ctx context.Context, o metric.Int64Observer) error {
			o.Observe(int64(state.Workers.Dereference.Queue.Len()))
			return nil
		}),
	)
	if err != nil {
		return err
	}

	_, err = meter.Int64ObservableGauge(
		"gotosocial.workers.client_api.count",
		metric.WithDescription("Current number of client API workers"),
		metric.WithInt64Callback(func(ctx context.Context, o metric.Int64Observer) error {
			o.Observe(int64(state.Workers.Client.Len()))
			return nil
		}),
	)
	if err != nil {
		return err
	}

	_, err = meter.Int64ObservableUpDownCounter(
		"gotosocial.workers.client_api.queue",
		metric.WithDescription("Current number of queued client API worker tasks"),
		metric.WithInt64Callback(func(ctx context.Context, o metric.Int64Observer) error {
			o.Observe(int64(state.Workers.Client.Queue.Len()))
			return nil
		}),
	)
	if err != nil {
		return err
	}

	_, err = meter.Int64ObservableGauge(
		"gotosocial.workers.fedi_api.count",
		metric.WithDescription("Current number of federator API workers"),
		metric.WithInt64Callback(func(ctx context.Context, o metric.Int64Observer) error {
			o.Observe(int64(state.Workers.Federator.Len()))
			return nil
		}),
	)
	if err != nil {
		return err
	}

	_, err = meter.Int64ObservableUpDownCounter(
		"gotosocial.workers.fedi_api.queue",
		metric.WithDescription("Current number of queued federator API worker tasks"),
		metric.WithInt64Callback(func(ctx context.Context, o metric.Int64Observer) error {
			o.Observe(int64(state.Workers.Federator.Queue.Len()))
			return nil
		}),
	)
	if err != nil {
		return err
	}

	_, err = meter.Int64ObservableGauge(
		"gotosocial.workers.processing.count",
		metric.WithDescription("Current number of processing workers"),
		metric.WithInt64Callback(func(ctx context.Context, o metric.Int64Observer) error {
			o.Observe(int64(state.Workers.Processing.Len()))
			return nil
		}),
	)
	if err != nil {
		return err
	}

	_, err = meter.Int64ObservableUpDownCounter(
		"gotosocial.workers.processing.queue",
		metric.WithDescription("Current number of queued processing worker tasks"),
		metric.WithInt64Callback(func(ctx context.Context, o metric.Int64Observer) error {
			o.Observe(int64(state.Workers.Processing.Queue.Len()))
			return nil
		}),
	)
	if err != nil {
		return err
	}

	_, err = meter.Int64ObservableGauge(
		"gotosocial.workers.webpush.count",
		metric.WithDescription("Current number of webpush workers"),
		metric.WithInt64Callback(func(ctx context.Context, o metric.Int64Observer) error {
			o.Observe(int64(state.Workers.WebPush.Len()))
			return nil
		}),
	)
	if err != nil {
		return err
	}

	_, err = meter.Int64ObservableUpDownCounter(
		"gotosocial.workers.webpush.queue",
		metric.WithDescription("Current number of queued webpush worker tasks"),
		metric.WithInt64Callback(func(ctx context.Context, o metric.Int64Observer) error {
			o.Observe(int64(state.Workers.WebPush.Queue.Len()))
			return nil
		}),
	)
	if err != nil {
		return err
	}

	// Create Prometheus metrics server.
	mux := http.NewServeMux()
	mux.Handle("/metrics", promhttp.Handler())

	// Nick these env vars from the autoexporter,
	// keeping the same defaults as are used there.
	host, ok := os.LookupEnv("OTEL_EXPORTER_PROMETHEUS_HOST")
	if !ok {
		host = "localhost"
	}
	port, ok := os.LookupEnv("OTEL_EXPORTER_PROMETHEUS_PORT")
	if !ok {
		port = "9464"
	}
	addr := host + ":" + port

	// Run Prometheus metrics server.
	server := &http.Server{
		Addr:              addr,
		Handler:           mux,
		ReadHeaderTimeout: 30 * time.Second,
	}
	go func() {
		log.Infof(ctx, "prometheus http server listening on %s", addr)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Errorf(ctx, "prometheus http server error: %v", err)
		}
	}()

	// Prepare graceful shutdown
	// of Prometheus metrics server.
	go func() {
		// Block until passed context
		// (server context) is done.
		_ = <-ctx.Done()
		timeout, cancel := context.WithTimeout(ctx, 30*time.Second)
		defer cancel()

		// Shut down the metrics server.
		if err := server.Shutdown(timeout); err != nil {
			log.Errorf(ctx, "error shutting down prometheus http server: %v", err)
		}
	}()

	return nil
}

func MetricsMiddleware() gin.HandlerFunc {
	return ginMiddleware()
}

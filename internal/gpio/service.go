	package gpio

	import (
	  "context"
	  "fmt"
	  "sync"
	  "time"

	  "github.com/prometheus/client_golang/prometheus"
	  "github.com/rs/zerolog"
	  "github.com/yourorg/gpio/internal/gpio/driver"
	)

	// Service represents the GPIO service interface
	type Service interface {
	  Configure(ctx context.Context, pin Pin) error
	  Read(ctx context.Context, pinNum int) (State, error)
	  Write(ctx context.Context, pinNum int, state State) error
	  Watch(ctx context.Context, pinNum int) (<-chan Event, error)
	  Close() error
	}

	type service struct {
	  driver     driver.Driver
	  logger     zerolog.Logger
	  metrics    *metrics
	  watchMu    sync.RWMutex
	  watchers   map[int][]chan Event
	  closeCh    chan struct{}
	  closeOnce  sync.Once
	}

	type metrics struct {
	  pinOps     *prometheus.CounterVec
	  pinErrors  *prometheus.CounterVec
	  pinStates  *prometheus.GaugeVec
	  watchCount prometheus.Gauge
	}

	func NewService(d driver.Driver, l zerolog.Logger) Service {
	  m := initMetrics()
	  return &service{
	    driver:   d,
	    logger:   l,
	    metrics:  m,
	    watchers: make(map[int][]chan Event),
	    closeCh:  make(chan struct{}),
	  }
	}

	func (s *service) Configure(ctx context.Context, pin Pin) error {
	  timer := prometheus.NewTimer(s.metrics.pinOps.WithLabelValues("configure"))
	  defer timer.ObserveDuration()

	  if err := s.validatePin(pin); err != nil {
	    s.metrics.pinErrors.WithLabelValues("configure").Inc()
	    return fmt.Errorf("pin validation failed: %w", err)
	  }

	  if err := s.driver.Configure(pin.Number, pin.Direction, pin.PullUp); err != nil {
	    s.metrics.pinErrors.WithLabelValues("configure").Inc()
	    return fmt.Errorf("driver configure failed: %w", err)
	  }

	  s.logger.Info().
	    Int("pin", pin.Number).
	    Interface("direction", pin.Direction).
	    Bool("pullup", pin.PullUp).
	    Msg("pin configured")

	  return nil
	}

	// Additional methods implemented similarly with proper error handling,
	// metrics, and logging


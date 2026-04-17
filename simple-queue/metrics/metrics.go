package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

type Metrics struct {
	EnqueueTotal prometheus.Counter
	DequeueTotal prometheus.Counter
	AckTotal     prometheus.Counter
	RequeueTotal prometheus.Counter
	DLQTotal     prometheus.Counter
	ErrorTotal   prometheus.Counter

	QueueSize    prometheus.Gauge
	InFlightSize prometheus.Gauge
	DLQSize      prometheus.Gauge
	ShardSize    *prometheus.GaugeVec
}

func New() *Metrics {
	return &Metrics{
		EnqueueTotal: promauto.NewCounter(prometheus.CounterOpts{
			Name: "sqs_enqueue_total",
			Help: "Total de mensagens enfileiradas",
		}),
		DequeueTotal: promauto.NewCounter(prometheus.CounterOpts{
			Name: "sqs_dequeue_total",
			Help: "Total de mensagens desenfileiradas",
		}),
		AckTotal: promauto.NewCounter(prometheus.CounterOpts{
			Name: "sqs_ack_total",
			Help: "Total de mensagens confirmadas (ack)",
		}),
		RequeueTotal: promauto.NewCounter(prometheus.CounterOpts{
			Name: "sqs_requeue_total",
			Help: "Total de mensagens reenfileiradas por timeout",
		}),
		DLQTotal: promauto.NewCounter(prometheus.CounterOpts{
			Name: "sqs_dlq_total",
			Help: "Total de mensagens enviadas para DLQ",
		}),
		ErrorTotal: promauto.NewCounter(prometheus.CounterOpts{
			Name: "sqs_error_total",
			Help: "Total de erros",
		}),
		QueueSize: promauto.NewGauge(prometheus.GaugeOpts{
			Name: "sqs_queue_size",
			Help: "Quantidade atual de mensagens na fila",
		}),
		InFlightSize: promauto.NewGauge(prometheus.GaugeOpts{
			Name: "sqs_inflight_size",
			Help: "Quantidade atual de mensagens em processamento",
		}),
		DLQSize: promauto.NewGauge(prometheus.GaugeOpts{
			Name: "sqs_dlq_size",
			Help: "Quantidade atual de mensagens na DLQ",
		}),
		ShardSize: promauto.NewGaugeVec(prometheus.GaugeOpts{
			Name: "sqs_shard_size",
			Help: "Quantidade de mensagens por shard",
		}, []string{"shard"}),
	}
}

func (m *Metrics) RecordEnqueue() {
	m.EnqueueTotal.Inc()
}

func (m *Metrics) RecordDequeue() {
	m.DequeueTotal.Inc()
}

func (m *Metrics) RecordAck() {
	m.AckTotal.Inc()
}

func (m *Metrics) RecordRequeue() {
	m.RequeueTotal.Inc()
}

func (m *Metrics) RecordDLQ() {
	m.DLQTotal.Inc()
}

func (m *Metrics) RecordError() {
	m.ErrorTotal.Inc()
}

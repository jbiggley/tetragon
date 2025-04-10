// SPDX-License-Identifier: Apache-2.0
// Copyright Authors of Tetragon

package eventmetrics

import (
	"github.com/cilium/tetragon/api/v1/tetragon"
	"github.com/cilium/tetragon/api/v1/tetragon/codegen/helpers"
	"github.com/cilium/tetragon/pkg/api/processapi"
	"github.com/cilium/tetragon/pkg/filters"
	"github.com/cilium/tetragon/pkg/logger"
	"github.com/cilium/tetragon/pkg/metrics"
	"github.com/cilium/tetragon/pkg/metrics/consts"
	"github.com/cilium/tetragon/pkg/metrics/errormetrics"
	"github.com/cilium/tetragon/pkg/metrics/syscallmetrics"
	v1 "github.com/cilium/tetragon/pkg/oldhubble/api/v1"
	"github.com/cilium/tetragon/pkg/reader/exec"
	"github.com/cilium/tetragon/pkg/tracingpolicy"
	"github.com/prometheus/client_golang/prometheus"
)

var (
	EventsProcessed = metrics.NewCounterVecWithPod(prometheus.CounterOpts{
		Namespace:   consts.MetricsNamespace,
		Name:        "events_total",
		Help:        "The total number of Tetragon events",
		ConstLabels: nil,
	}, []string{"type", "namespace", "workload", "pod", "binary"})
	FlagCount = prometheus.NewCounterVec(prometheus.CounterOpts{
		Namespace:   consts.MetricsNamespace,
		Name:        "flags_total",
		Help:        "The total number of Tetragon flags. For internal use only.",
		ConstLabels: nil,
	}, []string{"type"})
	NotifyOverflowedEvents = prometheus.NewCounter(prometheus.CounterOpts{
		Namespace:   consts.MetricsNamespace,
		Name:        "notify_overflowed_events_total",
		Help:        "The total number of events dropped because listener buffer was full",
		ConstLabels: nil,
	})

	policyStats = metrics.NewCounterVecWithPod(prometheus.CounterOpts{
		Namespace:   consts.MetricsNamespace,
		Name:        "policy_events_total",
		Help:        "Policy events calls observed.",
		ConstLabels: nil,
	}, []string{"policy", "hook", "namespace", "workload", "pod", "binary"})
)

func InitMetrics(registry *prometheus.Registry) {
	registry.MustRegister(EventsProcessed)
	registry.MustRegister(FlagCount)
	registry.MustRegister(NotifyOverflowedEvents)
	registry.MustRegister(policyStats)
}

func GetProcessInfo(process *tetragon.Process) (binary, pod, workload, namespace string) {
	if process != nil {
		binary = process.Binary
		if process.Pod != nil {
			namespace = process.Pod.Namespace
			workload = process.Pod.Workload
			pod = process.Pod.Name
		}
	} else {
		errormetrics.ErrorTotalInc(errormetrics.EventMissingProcessInfo)
	}
	return binary, pod, workload, namespace
}

func handleOriginalEvent(originalEvent interface{}) {
	var flags uint32
	switch msg := originalEvent.(type) {
	case *processapi.MsgExecveEventUnix:
		flags = msg.Process.Flags
	}
	for _, flag := range exec.DecodeCommonFlags(flags) {
		FlagCount.WithLabelValues(flag).Inc()
	}
}

func handleProcessedEvent(pInfo *tracingpolicy.PolicyInfo, processedEvent interface{}) {
	var eventType, namespace, workload, pod, binary string
	switch ev := processedEvent.(type) {
	case *tetragon.GetEventsResponse:
		binary, pod, workload, namespace = GetProcessInfo(filters.GetProcess(&v1.Event{Event: ev}))
		var err error
		eventType, err = helpers.ResponseTypeString(ev)
		if err != nil {
			logger.GetLogger().WithField("event", processedEvent).WithError(err).Warn("metrics: handleProcessedEvent: unhandled event")
			eventType = "unhandled"
		}
	default:
		eventType = "unknown"
	}
	EventsProcessed.WithLabelValues(eventType, namespace, workload, pod, binary).Inc()
	if pInfo != nil && pInfo.Name != "" {
		policyStats.
			WithLabelValues(pInfo.Name, pInfo.Hook, namespace, workload, pod, binary).
			Inc()
	}
}

func ProcessEvent(originalEvent interface{}, processedEvent interface{}) {
	handleOriginalEvent(originalEvent)

	var policyInfo tracingpolicy.PolicyInfo
	if policyEv, ok := originalEvent.(tracingpolicy.PolicyEvent); ok {
		policyInfo = policyEv.PolicyInfo()
	}

	handleProcessedEvent(&policyInfo, processedEvent)
	syscallmetrics.Handle(processedEvent)
}

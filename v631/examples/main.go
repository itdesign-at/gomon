package main

import (
	"log/slog"
	"os"
	"snmp_uptime/publisher"

	"snmp_uptime/check_value"
	"snmp_uptime/files"

	"github.com/gosnmp/gosnmp"

	"github.com/itdesign-at/golib/commandLine"
	"github.com/itdesign-at/golib/keyvalue"
)

const oidUptime = ".1.3.6.1.2.1.1.3.0"

var args keyvalue.Record

func main() {

	var hostProperties files.HostProperties
	var metricConfig *files.MetricConfig
	var uptimeInSeconds int64
	var err error

	args = commandLine.Parse(os.Args)

	keyword := args.String("k")
	host := args.String("h")
	service := args.String("s")
	node, _ := os.Hostname()
	args["n"] = node

	if host == "" || keyword == "" || service == "" {
		slog.Error("this program requires -k, -h and -s")
		os.Exit(1)
	}

	/**
	Array
	(
	    [$from] => /odin/watchit/V1/gauge/{{.NHS}}?fmt=graph&k={{.K}}
	    [$to] => nats://localhost
	    [$metric] => gauge
	    [k] => gauge
	    [h] => ea-pv-aca-1-dl01.m2m.energy-it.net
	    [s] => Temperature
	    [Exit] => 0
	    [Value] => 37
	    [Text] => Temperature: 37 Grad
	    [Host] => ea-pv-aca-1-dl01.m2m.energy-it.net
	    [Service] => Temperature
	    [Node] => uswesampol01
	    [State] => OK
	    [DSN] => /odin/watchit/V1/gauge/uswesampol01/eaQ2DpvQ2DacaQ2D1Q2Ddl01Q2Em2mQ2EenergyQ2DitQ2Enet/Temperature?fmt=graph&k=gauge
	)

	*/

	var cfg keyvalue.Record
	metricConfig = files.NewMetricConfig().WithMacros(args)
	cfg, err = metricConfig.Get(keyword)
	if err != nil {
		slog.Error(err.Error())
		os.Exit(1)
	}

	pub := publisher.NewNatsPublisher(cfg.String("$to"))
	_ = pub.WithSubject(cfg.String("$nats_subject")).PublishJson(args)
	slog.Info("Published", "keyword", keyword, "host", host, "service", service)
	os.Exit(0)

	cv := check_value.NewCheckValue(args)

	he := files.NewHostsExported()
	hostProperties, err = he.GetHostProperties(host)
	if err != nil {
		slog.Error(err.Error())
		os.Exit(1)
	}

	gosnmp.Default.Target = hostProperties.Host
	gosnmp.Default.Community = hostProperties.Community
	gosnmp.Default.Version = hostProperties.GoSnmpVersion

	err = gosnmp.Default.Connect()
	if err != nil {
		slog.Error(err.Error())
		os.Exit(1)
	}
	defer gosnmp.Default.Conn.Close()

	oids := []string{".1.3.6.1.2.1.1.3.0"}
	result, getError := gosnmp.Default.Get(oids)
	if getError != nil {
		slog.Error(getError.Error())
		os.Exit(1)
	}

	for _, variable := range result.Variables {
		switch variable.Name {
		case oidUptime:
			uptimeInSeconds = gosnmp.ToBigInt(variable.Value).Int64() / 100
		}
	}

	cv.Add(map[string]any{
		"k":     "gauge",
		"Value": uptimeInSeconds,
	})

	// p := publisher.NewNatsPublisher("x")

	cv.Bye()

	// fmt.Printf("uptime in minutes: %d\n", uptimeInSeconds/60)
}

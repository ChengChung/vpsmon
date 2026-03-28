package collector

import (
	"strconv"

	"github.com/chengchung/vpsmon/sdk"
	"github.com/chengchung/vpsmon/sdk/aliyun/cdt"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/sirupsen/logrus"
)

func init() {
	registerCollectorFactory(cdt.CDT_ALIYUN, newCDTCollector)
}

type cdtCollectors struct {
	cli        *cdt.Client
	collectors []Collector
}

func (c cdtCollectors) Update(ch chan<- prometheus.Metric) error {
	for _, collector := range c.collectors {
		if err := collector.Update(ch); err != nil {
			logrus.Errorf("failed to update cdt collector: %s", err)
			return err
		}
	}
	return nil
}

type cdtCollectorConverter interface {
	convert(base cdtCollectors) Collector
}

func newCDTCollector(client sdk.SDKClient, collectors []string) (Collector, error) {
	cli, ok := client.(*cdt.Client)
	if !ok {
		return nil, ErrUnexpectedClientType
	}

	cdtInstance := cdtCollectors{cli: cli}

	for _, collectorName := range collectors {
		converter, ok := cdtCollectorConvertersMap[collectorName]
		if !ok {
			logrus.Error("unknown cdt.aliyun collector: " + collectorName)
			return nil, ErrUnknownCollector
		}
		cdtInstance.collectors = append(cdtInstance.collectors, converter.convert(cdtInstance))
	}

	return &cdtInstance, nil
}

var cdtCollectorConvertersMap = map[string]cdtCollectorConverter{
	"traffic": cdtTrafficCollector{},
}

type cdtTrafficCollector struct {
	cli                   *cdt.Client
	totalTraffic          *prometheus.Desc
	productTraffic        *prometheus.Desc
	tierTrafficLowerBound *prometheus.Desc
	tierTrafficUpperBound *prometheus.Desc
	tierTrafficTraffic    *prometheus.Desc
}

func (c cdtTrafficCollector) Update(ch chan<- prometheus.Metric) error {
	resp, err := c.cli.ListCdtInternetTraffic(&cdt.ListCdtInternetTrafficRequest{})
	if err != nil {
		logrus.Errorf("failed to get cdt internet traffic: %s", err)
		return err
	}

	details := resp.Body.TrafficDetails
	for _, detail := range details {
		ISPType := detail.ISPType
		RegionId := detail.BusinessRegionId
		TotalTraffic := detail.Traffic

		ch <- prometheus.MustNewConstMetric(c.totalTraffic,
			prometheus.CounterValue,
			float64(*TotalTraffic),
			*ISPType, *RegionId)

		for _, productDetail := range detail.ProductTrafficDetails {
			Product := productDetail.Product
			ProductTraffic := productDetail.Traffic
			ch <- prometheus.MustNewConstMetric(c.productTraffic,
				prometheus.CounterValue,
				float64(*ProductTraffic),
				*ISPType, *RegionId, *Product)
		}

		for _, tierDetail := range detail.TrafficTierDetails {
			Tier := tierDetail.Tier
			TierStr := strconv.Itoa(int(*Tier))
			Traffic := tierDetail.Traffic
			LowestTraffic := tierDetail.LowestTraffic
			HighestTraffic := tierDetail.HighestTraffic
			ch <- prometheus.MustNewConstMetric(c.tierTrafficLowerBound,
				prometheus.GaugeValue,
				float64(*LowestTraffic),
				*ISPType, *RegionId, TierStr)
			ch <- prometheus.MustNewConstMetric(c.tierTrafficUpperBound,
				prometheus.GaugeValue,
				float64(*HighestTraffic),
				*ISPType, *RegionId, TierStr)
			ch <- prometheus.MustNewConstMetric(c.tierTrafficTraffic,
				prometheus.CounterValue,
				float64(*Traffic),
				*ISPType, *RegionId, TierStr)
		}
	}

	return nil
}

func (c cdtTrafficCollector) convert(base cdtCollectors) Collector {
	constLabels := prometheus.Labels{"client": base.cli.Name()}
	return cdtTrafficCollector{
		cli: base.cli,
		totalTraffic: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, "aliyun_cdt", "traffic_total"),
			"Total traffic by ISP type and bussiness region",
			[]string{"ISPType", "regionId"}, constLabels,
		),
		productTraffic: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, "aliyun_cdt", "product_traffic"),
			"Product traffic by ISP type and bussiness region",
			[]string{"ISPType", "regionId", "product"}, constLabels,
		),
		tierTrafficLowerBound: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, "aliyun_cdt", "tier_lowest_traffic"),
			"Total traffic by ISP type and bussiness region",
			[]string{"ISPType", "regionId", "tier"}, constLabels,
		),
		tierTrafficUpperBound: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, "aliyun_cdt", "tier_highest_traffic"),
			"Total traffic by ISP type and bussiness region",
			[]string{"ISPType", "regionId", "tier"}, constLabels,
		),
		tierTrafficTraffic: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, "aliyun_cdt", "tier_used_traffic"),
			"Total traffic by ISP type and bussiness region",
			[]string{"ISPType", "regionId", "tier"}, constLabels,
		),
	}
}

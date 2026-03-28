package cdt

import (
	"errors"

	cdt "github.com/alibabacloud-go/cdt-20210813/client"
	openapi "github.com/alibabacloud-go/darabonba-openapi/client"
	util "github.com/alibabacloud-go/tea-utils/service"
	"github.com/chengchung/vpsmon/sdk"

	"github.com/alibabacloud-go/tea/tea"
)

const (
	CDT_ALIYUN = "cdt.aliyun"
)

func init() {
	sdk.RegisterClientFactory(CDT_ALIYUN, &factory{})
}

type Client struct {
	cdt.Client
	name string
}

func (Client) Type() string {
	return CDT_ALIYUN
}

func (c Client) Name() string {
	return c.name
}

type factory struct{}

func (factory) New(name string, cfg map[string]string) (sdk.SDKClient, error) {
	if len(cfg) == 0 {
		return nil, errors.New("invalid openapi config")
	}

	accessKeyId := cfg["AccessKeyId"]
	accessKeySecret := cfg["AccessKeySecret"]
	endpoint := cfg["Endpoint"]
	if endpoint == "" {
		endpoint = "cdt.aliyuncs.com"
	}

	if accessKeyId == "" || accessKeySecret == "" {
		return nil, errors.New("missing required config fields: accessKeyId, accessKeySecret")
	}

	config := new(openapi.Config)
	config.SetAccessKeyId(accessKeyId).
		SetAccessKeySecret(accessKeySecret).
		SetEndpoint(endpoint)

	cli, err := cdt.NewClient(config)
	if err != nil {
		return nil, err
	}

	return &Client{*cli, name}, nil
}

type ListCdtInternetTrafficRequest struct {
	BusinessRegionId *string `json:"BusinessRegionId,omitempty" xml:"BusinessRegionId,omitempty"`
}

func (s ListCdtInternetTrafficRequest) String() string {
	return tea.Prettify(s)
}

func (s ListCdtInternetTrafficRequest) GoString() string {
	return s.String()
}

func (s *ListCdtInternetTrafficRequest) SetBusinessRegionId(v string) *ListCdtInternetTrafficRequest {
	s.BusinessRegionId = &v
	return s
}

type ListCdtInternetTrafficResponse struct {
	Headers map[string]*string                  `json:"headers,omitempty" xml:"headers,omitempty" require:"true"`
	Body    *ListCdtInternetTrafficResponseBody `json:"body,omitempty" xml:"body,omitempty" require:"true"`
}

func (s ListCdtInternetTrafficResponse) String() string {
	return tea.Prettify(s)
}

func (s ListCdtInternetTrafficResponse) GoString() string {
	return s.String()
}

type ListCdtInternetTrafficResponseBody struct {
	RequestId      *string              `json:"RequestId,omitempty" xml:"RequestId,omitempty"`
	TrafficDetails []TrafficDetailsItem `json:"TrafficDetails,omitempty" xml:"TrafficDetails,omitempty"`
}

func (s ListCdtInternetTrafficResponseBody) String() string {
	return tea.Prettify(s)
}

func (s ListCdtInternetTrafficResponseBody) GoString() string {
	return s.String()
}

type TrafficDetailsItem struct {
	ISPType               *string                     `json:"ISPType,omitempty" xml:"ISPType,omitempty"`
	BusinessRegionId      *string                     `json:"BusinessRegionId,omitempty" xml:"BusinessRegionId,omitempty"`
	Traffic               *int64                      `json:"Traffic,omitempty" xml:"Traffic,omitempty"`
	ProductTrafficDetails []ProductTrafficDetailsItem `json:"ProductTrafficDetails,omitempty" xml:"ProductTrafficDetails,omitempty"`
	TrafficTierDetails    []TrafficTierDetailsItem    `json:"TrafficTierDetails,omitempty" xml:"TrafficTierDetails,omitempty"`
}

func (s TrafficDetailsItem) String() string {
	return tea.Prettify(s)
}

func (s TrafficDetailsItem) GoString() string {
	return s.String()
}

type ProductTrafficDetailsItem struct {
	Product *string `json:"Product,omitempty" xml:"Product,omitempty"`
	Traffic *int64  `json:"Traffic,omitempty" xml:"Traffic,omitempty"`
}

func (s ProductTrafficDetailsItem) String() string {
	return tea.Prettify(s)
}

func (s ProductTrafficDetailsItem) GoString() string {
	return s.String()
}

type TrafficTierDetailsItem struct {
	Tier           *int64 `json:"Tier,omitempty" xml:"Tier,omitempty"`
	Traffic        *int64 `json:"Traffic,omitempty" xml:"Traffic,omitempty"`
	LowestTraffic  *int64 `json:"LowestTraffic,omitempty" xml:"LowestTraffic,omitempty"`
	HighestTraffic *int64 `json:"HighestTraffic,omitempty" xml:"HighestTraffic,omitempty"`
}

func (s TrafficTierDetailsItem) String() string {
	return tea.Prettify(s)
}

func (s TrafficTierDetailsItem) GoString() string {
	return s.String()
}

func (client *Client) ListCdtInternetTraffic(request *ListCdtInternetTrafficRequest) (_result *ListCdtInternetTrafficResponse, _err error) {
	runtime := &util.RuntimeOptions{}
	_result = &ListCdtInternetTrafficResponse{}
	_body, _err := client.ListCdtInternetTrafficWithOptions(request, runtime)
	if _err != nil {
		return _result, _err
	}
	_result = _body
	return _result, nil
}

func (client *Client) ListCdtInternetTrafficWithOptions(request *ListCdtInternetTrafficRequest, runtime *util.RuntimeOptions) (_result *ListCdtInternetTrafficResponse, _err error) {
	_err = util.ValidateModel(request)
	if _err != nil {
		return nil, _err
	}
	req := &openapi.OpenApiRequest{
		Body: util.ToMap(request),
	}
	_result = &ListCdtInternetTrafficResponse{}
	_body, _err := client.DoRPCRequest(tea.String("ListCdtInternetTraffic"), tea.String("2021-08-13"), tea.String("HTTPS"), tea.String("POST"), tea.String("AK"), tea.String("json"), req, runtime)
	if _err != nil {
		return _result, _err
	}
	_err = tea.Convert(_body, &_result)
	return _result, _err
}

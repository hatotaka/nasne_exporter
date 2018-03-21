package nasneclient

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"strconv"

	"github.com/golang/glog"
)

const loglevel = 10

type NasneClient struct {
	IPAddr string
}

func NewNasneClient(ipAddr string) (*NasneClient, error) {
	return &NasneClient{ipAddr}, nil
}

type BoxName struct {
	Errorcode int
	Name      string
}

func (nc *NasneClient) GetBoxName() (*BoxName, error) {
	bn := &BoxName{}
	if err := nc.getJson("boxNameGet", bn, nil); err != nil {
		return nil, err
	}

	return bn, nil
}

type SoftwareVersion struct {
	BackdatedVersion string
	SoftwareVersion  string
	Errcode          int
}

func (nc *NasneClient) GetSoftwareVersion() (*SoftwareVersion, error) {
	sv := &SoftwareVersion{}
	if err := nc.getJson("softwareVersionGet", sv, nil); err != nil {
		return nil, err
	}

	return sv, nil
}

type HardwareVersion struct {
	ProductName     string
	HardwareVersion int
	Errorcode       int
}

func (nc *NasneClient) GetHardwareVersion() (*HardwareVersion, error) {
	hv := &HardwareVersion{}
	if err := nc.getJson("hardwareVersionGet", hv, nil); err != nil {
		return nil, err
	}

	return hv, nil
}

type HDDInfo struct {
	HDD       HDDInfoHDD
	Errorcode int
}

type HDDInfoHDD struct {
	TotalVolumeSize float64
	FreeVolumeSize  float64
	UsedVolumeSize  float64
	SerialNumber    string
	ID              int
	InternalFlag    int
	MountStatus     int
	RegisterFlag    int
	Format          string
	Name            string
	VendorID        string
	ProductID       string
}

func (nc *NasneClient) GetHDDInfo(id int) (*HDDInfo, error) {
	hi := &HDDInfo{}
	param := url.Values{}
	param.Add("id", strconv.Itoa(id))
	if err := nc.getJson("HDDInfoGet", hi, &param); err != nil {
		return nil, err
	}
	return hi, nil
}

type HDDList struct {
	Errorcode int
	Number    int
	HDD       []*HDDListHDD
}

type HDDListHDD struct {
	ID           int
	InternalFlag int
	MountStatus  int
	RegisterFlag int
}

func (nc *NasneClient) GetHDDList() (*HDDList, error) {
	hl := &HDDList{}
	if err := nc.getJson("HDDListGet", hl, nil); err != nil {
		return nil, err
	}
	return hl, nil
}

type DTCPIPClientList struct {
	Errorcode int
	Number    int
	Client    []*DTCPIPClientListClient
}

type DTCPIPClientListClient struct {
	ID          int
	MacAddr     string
	IpAddr      string
	Name        string
	Purpose     int
	LiveInfo    *LiveInfo
	Content     *Content
	EncryptType int
}

type LiveInfo struct {
	BroadcastingType int
	ServiceID        int
}

type Content struct {
	ID string
}

func (nc *NasneClient) GetDTCPIPClientList() (*DTCPIPClientList, error) {
	dl := &DTCPIPClientList{}
	if err := nc.getJson("dtcpipClientListGet", dl, nil); err != nil {
		return nil, err
	}

	return dl, nil
}

func (nc *NasneClient) getJson(endpoint string, data interface{}, values *url.Values) error {
	var query string
	if values != nil {
		query = values.Encode()
	}

	url := fmt.Sprintf("http://%s:64210/status/%s?%s", nc.IPAddr, endpoint, query)

	glog.V(loglevel).Infof("url = %v", url)

	res, err := http.Get(url)
	if err != nil {
		return err
	}

	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return err
	}
	defer res.Body.Close()
	glog.V(loglevel).Infof("json=%v", string(body))

	if err := json.Unmarshal(body, data); err != nil {
		return err
	}

	glog.V(loglevel).Infof("data=%+v", data)

	return nil
}

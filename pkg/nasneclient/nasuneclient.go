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

const (
	portStatus   = 64210
	portRecorded = 64220
	portSchedule = 64220
)

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
	if err := nc.getJson("status/boxNameGet", portStatus, bn, nil); err != nil {
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
	if err := nc.getJson("status/softwareVersionGet", portStatus, sv, nil); err != nil {
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
	if err := nc.getJson("status/hardwareVersionGet", portStatus, hv, nil); err != nil {
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
	if err := nc.getJson("status/HDDInfoGet", portStatus, hi, &param); err != nil {
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
	if err := nc.getJson("status/HDDListGet", portStatus, hl, nil); err != nil {
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
	if err := nc.getJson("status/dtcpipClientListGet", portStatus, dl, nil); err != nil {
		return nil, err
	}

	return dl, nil
}

type RecordedTitleList struct {
	Errorcode      int
	Item           []*RecordedTitleListItem
	TotalMatches   int
	NumberReturned int
}

type RecordedTitleListItem struct {
	ID               string
	Title            string
	Description      string
	StartDateTime    string
	Duration         int
	ConditionID      string
	Quality          int
	ChannelName      string
	ChannelNumber    int
	BroadcastingType int
	ServiceID        int
	EventID          int
}

func (nc *NasneClient) GetRecordedTitleList() (*RecordedTitleList, error) {
	rtl := &RecordedTitleList{}

	param := url.Values{}
	param.Add("searchCriteria", "0")
	param.Add("filter", "0")
	param.Add("startingIndex", "0")
	param.Add("requestedCount", "0")
	param.Add("sortCriteria", "0")

	if err := nc.getJson("recorded/titleListGet", portRecorded, rtl, &param); err != nil {
		return nil, err
	}

	return rtl, nil
}

type ReservedList struct {
	Errorcode int
	Item      []*ReservedListItem
}

type ReservedListItem struct {
	ID          string
	Title       string
	Descritpion string

	ConflictID int
	EventID    int
}

func (nc *NasneClient) GetReservedList() (*ReservedList, error) {
	rl := &ReservedList{}

	param := url.Values{}
	param.Add("searchCriteria", "0")
	param.Add("&filter", "0")
	param.Add("startingIndex", "0")
	param.Add("requestedCount", "0")
	param.Add("sortCriteria", "0")
	param.Add("withDescriptionLong", "0")
	param.Add("withUserData", "1")

	if err := nc.getJson("schedule/reservedListGet", portSchedule, rl, &param); err != nil {
		return nil, err
	}

	return rl, nil
}

func (nc *NasneClient) getJson(endpoint string, port int, data interface{}, values *url.Values) error {
	var query string
	if values != nil {
		query = values.Encode()
	}

	url := fmt.Sprintf("http://%s:%d/%s?%s", nc.IPAddr, port, endpoint, query)

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

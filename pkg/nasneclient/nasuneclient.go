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

func (nc *NasneClient) GetBoxName() (*BoxName, error) {
	bn := &BoxName{}
	if err := nc.getJson("status/boxNameGet", portStatus, bn, nil); err != nil {
		return nil, err
	}

	return bn, nil
}

func (nc *NasneClient) GetSoftwareVersion() (*SoftwareVersion, error) {
	sv := &SoftwareVersion{}
	if err := nc.getJson("status/softwareVersionGet", portStatus, sv, nil); err != nil {
		return nil, err
	}

	return sv, nil
}

func (nc *NasneClient) GetHardwareVersion() (*HardwareVersion, error) {
	hv := &HardwareVersion{}
	if err := nc.getJson("status/hardwareVersionGet", portStatus, hv, nil); err != nil {
		return nil, err
	}

	return hv, nil
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

func (nc *NasneClient) GetHDDList() (*HDDList, error) {
	hl := &HDDList{}
	if err := nc.getJson("status/HDDListGet", portStatus, hl, nil); err != nil {
		return nil, err
	}
	return hl, nil
}

func (nc *NasneClient) GetDTCPIPClientList() (*DTCPIPClientList, error) {
	dl := &DTCPIPClientList{}
	if err := nc.getJson("status/dtcpipClientListGet", portStatus, dl, nil); err != nil {
		return nil, err
	}

	return dl, nil
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

func (nc *NasneClient) GetBoxStatusList() (*BoxStatusList, error) {
	bsl := &BoxStatusList{}
	if err := nc.getJson("status/boxStatusListGet", portStatus, bsl, nil); err != nil {
		return nil, err
	}
	return bsl, nil
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

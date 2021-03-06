package nasneclient

type BoxName struct {
	Errorcode int
	Name      string
}

type SoftwareVersion struct {
	BackdatedVersion string
	SoftwareVersion  string
	Errcode          int
}

type HardwareVersion struct {
	ProductName     string
	HardwareVersion int
	Errorcode       int
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

type ReservedList struct {
	Errorcode int
	Item      []*ReservedListItem

	TotalMatches   int
	NumberReturned int
}

type ReservedListItem struct {
	ID          string
	Title       string
	Descritpion string

	ConflictID int
	EventID    int
}

const (
	// If ConflictID is 1, channel reservation conflicts but can record.
	ConflictIDConflictOK = 1
	// If ConflictID is 2, channel reservation conflicts and can not record.
	ConflictIDConflictNG = 2
	// If EventID is 65536, channel reservation not found.
	EventIDNotFound = 65536
)

type BoxStatusList struct {
	Errorcode    int
	TuningStatus BoxStatusListTuningStatus
}

type BoxStatusListTuningStatus struct {
	Status            int
	NetworkId         int
	TransportStreamId int
	ServiceId         int
}

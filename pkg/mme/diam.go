package mme

import (
	"fmt"
	"net"
	"time"

	"github.com/fiorix/go-diameter/diam"
	"github.com/fiorix/go-diameter/diam/avp"
	"github.com/fiorix/go-diameter/diam/datatype"
	"github.com/fiorix/go-diameter/diam/dict"
	"github.com/fiorix/go-diameter/diam/sm"
)

// DiamClient is S6A diameter protocol client.
type DiamClient struct {
	cli *sm.Client
}

// DiamOpt is DiamClient options.
type DiamOpt struct {
	originHost  string
	originRealm string
	vendorID    uint32
	hostAddress string
}

// NewDiamClient create new Diameter session for HSS.
func NewDiamClient(opt *DiamOpt) *DiamClient {
	cfg := &sm.Settings{
		OriginHost:       datatype.DiameterIdentity(opt.originHost),
		OriginRealm:      datatype.DiameterIdentity(opt.originRealm),
		VendorID:         datatype.Unsigned32(opt.vendorID),
		ProductName:      "go-diameter-s6a",
		OriginStateID:    datatype.Unsigned32(time.Now().Unix()),
		FirmwareRevision: 1,
		HostIPAddresses: []datatype.Address{
			datatype.Address(net.ParseIP(opt.hostAddress)),
		},
	}

	mux := sm.New(cfg)

	cli := &sm.Client{
		Dict:               dict.Default,
		Handler:            mux,
		MaxRetransmits:     0,
		RetransmitInterval: time.Second,
		EnableWatchdog:     true,
		WatchdogInterval:   time.Duration(5) * time.Second,
		SupportedVendorID: []*diam.AVP{
			diam.NewAVP(avp.SupportedVendorID, avp.Mbit, 0, datatype.Unsigned32(10415)),
		},
		VendorSpecificApplicationID: []*diam.AVP{
			diam.NewAVP(avp.VendorSpecificApplicationID, avp.Mbit, 0, &diam.GroupedAVP{
				AVP: []*diam.AVP{
					diam.NewAVP(avp.AuthApplicationID, avp.Mbit, 0, datatype.Unsigned32(16777251)),
					diam.NewAVP(avp.VendorID, avp.Mbit, 0, datatype.Unsigned32(10415)),
				},
			}),
		},
	}

	return &DiamClient{
		cli: cli,
	}
}

// Start initiate diameter client start.
func (d *DiamClient) Start() {
	fmt.Println("Start")
	// m := diam.NewRequest(diam.AuthenticationInformation, diam.TGPP_S6A_APP_ID, dict.Default)
	// fmt.Println(m)

	go func() {
		conn, err := d.cli.DialNetwork("tcp4", "172.16.0.52:3868")
		if err != nil {
			fmt.Println(err)
		}
		fmt.Println(conn)
	}()
}

package inst

import (
	"comwrapper"
	"th3api/config"
	"time"
)

type InstAdaptor struct {
	Sid        string
	UsrTag     string
	MeterCount int
	Params     map[string]string
	Status     int
	Cb         comwrapper.CallBackPtr
	ReqNo      int
	TimeOut    time.Duration
	Model      string
	InDatas    []string
	IsFirstRet bool
	ServiceId  string
	Th3Api     Th3Api
}

func NewInstAdaptor(usrTag, sid string, params map[string]string, cb comwrapper.CallBackPtr) *InstAdaptor {
	inst := &InstAdaptor{
		Sid:     sid,
		UsrTag:  usrTag,
		Params:  params,
		Cb:      cb,
		Model:   config.GetBaseConf().Th3ApiModelName,
		TimeOut: time.Duration(config.GetBaseConf().Th3ApiTimeOut) * time.Second,
	}
	inst.Th3Api = Th3ApiFactory(inst)
	return inst
}

func (inst *InstAdaptor) WriteIn(data string, status comwrapper.DataStatus) error {
	inst.InDatas = append(inst.InDatas, data)
	return nil
}

func (inst *InstAdaptor) PushBack(cb comwrapper.CallBackPtr) error {
	return inst.Th3Api.PushBack(cb)
}

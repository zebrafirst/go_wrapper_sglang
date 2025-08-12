package inst

import (
	"comwrapper"
)

type Th3Api interface {
	PushBack(cb comwrapper.CallBackPtr) error
}

func Th3ApiFactory(inst *InstAdaptor) Th3Api {
	return NewOpenAITh3API(inst)
}

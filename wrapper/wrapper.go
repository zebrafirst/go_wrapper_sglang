package main

import (
	"comwrapper"
	"fmt"
	"runtime/debug"
	"th3api/common"
	"th3api/config"
	"th3api/inst"
	th3apiutils "th3api/th3api_utils"
	"unsafe"

	"go.uber.org/zap"
)

// WrapperInit 插件初始化, 全局只调用一次. 本地调试时, cfg参数由aiges.toml提供
func WrapperInit(cfg map[string]string) (err error) {
	fmt.Println("---- wrapper init ----")
	for k, v := range cfg {
		fmt.Printf("config param %s=%s\n", k, v)
	}
	if err := config.NewBaseConf(cfg); err != nil {
		fmt.Printf("New base conf failed, err: %v", err)
		return err
	}
	if err := th3apiutils.NewLocalLog(); err != nil {
		fmt.Printf("New local log failed, err: %v", err)
		return err
	}

	if err := th3apiutils.InitOtlp(); err != nil {
		fmt.Printf("New otlp event provider: %v \n", err)
		return err
	}
	inst.InitWebSearchTemplate()
	return nil
}

// WrapperCreate 插件会话实例创建, 每次建立会话请求时调用. 本地调试时, params参数由xtest.toml提供
// todo: prsIds又是什么？
func WrapperCreate(usrTag string, params map[string]string, prsIds []int, cb comwrapper.CallBackPtr) (hdl unsafe.Pointer, err error) {
	sid := params["sid"]
	paramStr := ""
	for k, v := range params {
		paramStr += fmt.Sprintf("%s=%s;", k, v)
	}
	th3apiutils.WLogger.Info("WrapperCreate params", zap.String("paramStr", paramStr), zap.String("sid", sid))

	inst := inst.NewInstAdaptor(usrTag, sid, params, cb)

	th3apiutils.WLogger.Info("WrapperCreate successful", zap.String("sid", sid))
	return unsafe.Pointer(inst), nil
}

// WrapperWrite 数据写入
func WrapperWrite(hdl unsafe.Pointer, req []comwrapper.WrapperData) (err error) {
	inst := (*inst.InstAdaptor)(hdl)

	if len(req) == 0 {
		th3apiutils.WLogger.Error("WrapperWrite data is nil", zap.String("sid", inst.Sid))
		return nil
	}

	for _, v := range req {
		inst.MeterCount += len(v.Data)

		th3apiutils.WLogger.Info("WrapperWrite data", zap.String("data", string(v.Data)), zap.Int("status", int(v.Status)), zap.String("sid", inst.Sid))

		// todo: 写入的status是什么？
		if err = inst.WriteIn(string(v.Data), v.Status); err != nil {
			th3apiutils.WLogger.Error("WrapperWrite inst.write", zap.Any("err", err), zap.String("sid", inst.Sid))
			return err
		}
	}

	if inst.Cb != nil { // 如果异步流式会话方式，这里直接开始回写数据
		th3apiutils.WLogger.Info("Session mode is async!", zap.String("sid", inst.Sid))
		go func() {
			defer func() {
				if r := recover(); r != nil {
					stack := debug.Stack()
					th3apiutils.WLogger.Error("Engine pusher crashed",
						zap.Any("err", r),
						zap.String("sid", inst.Sid),
						zap.String("stack", string(stack))) // 添加堆栈信息
				}
			}()
			if err := inst.PushBack(inst.Cb); err != nil { // 数据回写过程中出现了问题，该如何上报出去？
				th3apiutils.WLogger.Error("Engine push error", zap.Any("err", err), zap.String("sid", inst.Sid), zap.String("usrTag", inst.UsrTag))
			}
		}()
	}

	return err
}

// WrapperRead 数据结果读取, 服务配置[aiges]项配置asyncMode = false
func WrapperRead(hdl unsafe.Pointer) (respData []comwrapper.WrapperData, err error) {
	inst := (*inst.InstAdaptor)(hdl)
	th3apiutils.WLogger.Info("WrapperRead ", zap.String("sid", inst.Sid))
	return nil, nil
}

// WrapperDestroy 会话资源销毁
func WrapperDestroy(hdl interface{}) (err error) {
	inst := (*inst.InstAdaptor)(hdl.(unsafe.Pointer))
	th3apiutils.WLogger.Info("WrapperDestroy", zap.String("sid", inst.Sid))
	return
}

// WrapperExec 非流式请求-同步响应
func WrapperExec(usrTag string, params map[string]string, reqData []comwrapper.WrapperData) (respData []comwrapper.WrapperData, err error) {
	th3apiutils.WLogger.Info("WrapperExec", zap.Any("params", params), zap.Any("reqData", reqData))
	return nil, nil
}

func WrapperFini() (err error) {
	th3apiutils.FiniOtlp()
	return
}

func WrapperVersion() (version string) {
	return
}

func WrapperLoadRes(res comwrapper.WrapperData, resId int) (err error) {
	return
}

func WrapperUnloadRes(resId int) (err error) {
	return
}

func WrapperDebugInfo(hdl interface{}) (debug string) {
	return
}

func WrapperSetCtrl(fType comwrapper.CustomFuncType, f interface{}) (err error) {
	switch fType {
	case comwrapper.FuncMeter:
		common.MeterFunc = f.(func(usrTag string, key string, count int) (code int))
		fmt.Println("WrapperSetCtrl meterFunc set successful.")
	default:

	}
	return
}

func WrapperNotify(res comwrapper.WrapperData) (err error) {
	return nil
}

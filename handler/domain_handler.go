package handler

import (
	"net/http"
	"strings"

	"github.com/urlooker/web/g"
	"github.com/urlooker/web/http/errors"
	"github.com/urlooker/web/http/param"
	"github.com/urlooker/web/http/render"
	"github.com/urlooker/web/model"
	"github.com/urlooker/web/utils"
)

type Url struct {
	Ip         string              `json:"ip"`
	MonitorIdc string              `json:"monitor_idc"`
	Status     []*model.ItemStatus `json:"status"`
}

type Value struct {
	Data       []Url              `json:"data"`
	Time       []string           `json:"time"`
}

func UrlStatus(w http.ResponseWriter, r *http.Request) {
	sid := param.MustInt64(r, "id")

	sidIpIndex, err := model.RelSidIpRepo.GetBySid(sid)
	errors.MaybePanic(err)

	stra, err := model.GetStrategyById(sid)
	errors.MaybePanic(err)

	idcs := strings.Split(stra.MonitorIdc, ",")

	urlArr := make([]Url, 0)
	for _, index := range sidIpIndex {
		for _, idc := range idcs {
			status, err := model.ItemStatusRepo.GetByIpSidIdc(index.Ip, idc, index.Sid)
			if len(status) < 1 {
				continue
			}
			errors.MaybePanic(err)

			url := Url{
				Ip:         index.Ip,
				MonitorIdc: status[0].MonitorIdc,
			}
			url.Status = status

			urlArr = append(urlArr, url)
		}
	}

	//var DataMap map[string]Value
	DataMap := make(map[string]Value)
	for _, item := range urlArr {
		var timeList []string
		for _, item := range item.Status {
			t := utils.TimeFormat(item.PushTime)
			timeList = append(timeList, t)
		}
		value := DataMap[item.MonitorIdc]
		value.Time = timeList
		value.Data = append(value.Data, item)
		DataMap[item.MonitorIdc] = value
	}

	//绘图使用，时间轴
	var timeData []string
	if len(urlArr) > 0 {
		for _, item := range urlArr[0].Status {
			t := utils.TimeFormat(item.PushTime)
			timeData = append(timeData, t)
		}
	}

	events, err := model.EventRepo.GetByStrategyId(sid, g.Config.Past*60)
	errors.MaybePanic(err)

	strategy, err := model.GetStrategyById(sid)
	errors.MaybePanic(err)

	render.Data(r, "AlarmOn", g.Config.Alarm.Enable)
	render.Data(r, "TimeData", timeData)
	render.Data(r, "Id", sid)
	render.Data(r, "Url", strategy.Url)
	render.Data(r, "Events", events)
	render.Data(r, "Data", urlArr)
	render.Data(r, "DataMap", DataMap)
	render.HTML(r, w, "chart/index")
}

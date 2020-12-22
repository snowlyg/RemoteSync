package models

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/snowlyg/RemoteSync/logging"
	"github.com/snowlyg/RemoteSync/utils"
	"time"
)

type Loc struct {
	ID           uint   `gorm:"primarykey"`
	LocId        int64  `json:"loc_id"`
	LocDesc      string `json:"loc_desc"`
	LocWardFlag  int64  `json:"loc_ward_flag"`
	CtHospitalId int64  `json:"ct_hospital_id"`
	CreatedAt    time.Time
	UpdatedAt    time.Time
}

type RequestLoc struct {
	LocId           int64  `json:"loc_id"`
	LocDesc         string `json:"loc_desc"`
	LocWardFlag     int64  `json:"loc_ward_flag"`
	CtHospitalId    int64  `json:"ct_hospital_id"`
	ApplicationName string `json:"application_name"`
	ApplicationId   int64  `json:"application_id"`
}

func LocSync() error {

	appId := utils.GetAppInfoCache().Id
	appName := utils.GetAppInfoCache().Name

	if Sqlite == nil {
		logging.Err.Error("database is not init")
		return errors.New("database is not init")
	}

	query := "select loc_id,loc_desc,loc_ward_flag,loc_active_flag,ct_hospital_id from ct_loc where loc_active_flag = 1"

	rows, err := Mysql.Raw(query).Rows()
	if err != nil {
		logging.Err.Error("mysql raw error :", err)
		return err
	}
	defer rows.Close()

	var locs []Loc
	for rows.Next() {
		var loc Loc
		Sqlite.ScanRows(rows, &loc)
		locs = append(locs, loc)
	}

	if len(locs) == 0 {
		return nil
	}

	var oldLocs []Loc
	Sqlite.Find(&oldLocs)

	var delLocIds []int64
	var newLocs []Loc
	var requestLocs []*RequestLoc

	// 没有旧数据
	path := "common/v1/data_sync/loc"
	if len(oldLocs) == 0 {
		newLocs = locs
		for _, re := range locs {
			requestLoc := &RequestLoc{
				LocId:           re.LocId,
				LocDesc:         re.LocDesc,
				LocWardFlag:     re.LocWardFlag,
				CtHospitalId:    re.CtHospitalId,
				ApplicationId:   appId,
				ApplicationName: appName,
			}
			requestLocs = append(requestLocs, requestLoc)
		}
		Sqlite.Create(&newLocs)

		requestLocsJson, _ := json.Marshal(&requestLocs)
		var res interface{}
		res, err = utils.SyncServices(path, fmt.Sprintf("delLocIds=%s&requestLocs=%s", "", requestLocsJson))
		if err != nil {
			logging.Err.Error("post common/v1/sync_remote get error ", err)
		}

		logging.Norm.Infof("数据提交返回信息:", res)

		return nil

	}

	// not in new
	for _, ore := range oldLocs {
		in := false
		for _, re := range locs {
			if ore.LocId == re.LocId {
				in = true
			}
		}
		if !in {
			delLocIds = append(delLocIds, ore.LocId)
		}
	}

	// changed
	for _, re := range locs {
		in := false
		for _, ore := range oldLocs {
			if ore.LocId == re.LocId {
				if ore.LocWardFlag != re.LocWardFlag ||
					ore.LocDesc != re.LocDesc ||
					ore.CtHospitalId != re.CtHospitalId {
					requestLoc := &RequestLoc{
						LocId:           re.LocId,
						LocDesc:         re.LocDesc,
						LocWardFlag:     re.LocWardFlag,
						CtHospitalId:    re.CtHospitalId,
						ApplicationId:   appId,
						ApplicationName: appName,
					}
					requestLocs = append(requestLocs, requestLoc)
					newLocs = append(newLocs, re)
					delLocIds = append(delLocIds, ore.LocId)
				}
				in = true
			}
		}

		if !in {
			requestLoc := &RequestLoc{
				LocId:           re.LocId,
				LocDesc:         re.LocDesc,
				LocWardFlag:     re.LocWardFlag,
				CtHospitalId:    re.CtHospitalId,
				ApplicationId:   appId,
				ApplicationName: appName,
			}
			requestLocs = append(requestLocs, requestLoc)
			newLocs = append(newLocs, re)
		}
	}

	var delLocIdsJson []byte
	var requestLocsJson []byte
	if len(delLocIds) > 0 {
		Sqlite.Where("loc_id in ?", delLocIds).Delete(&Loc{})
		delLocIdsJson, _ = json.Marshal(&delLocIds)
	}

	if len(newLocs) > 0 {
		Sqlite.Create(&newLocs)
	}

	if len(requestLocs) > 0 {
		requestLocsJson, _ = json.Marshal(&requestLocs)
	}

	var res interface{}
	res, err = utils.SyncServices(path, fmt.Sprintf("delLocIds=%s&requestLocs=%s", string(delLocIdsJson), string(requestLocsJson)))
	if err != nil {
		logging.Err.Error(err)
	}

	logging.Norm.Infof("数据提交返回信息:", res)

	return nil
}

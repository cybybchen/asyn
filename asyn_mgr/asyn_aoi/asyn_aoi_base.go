package asyn_aoi

import (
	"bytes"
	"fmt"
	"gitlab.sunborngame.com/base/log"
	"twe/define"
	"twe/utils"
	"twe/utils/quadTree"
)

// 获取与相机范围相叠的实体
func (view *AsynAoiBase) GetVolModelNearTheCamera(camera *AsyncCamera, entities utils.GuidMap) {
	objects := view.QuadTree.RetrieveIntersections(camera.GetBounds(), false)
	for _, o := range objects {
		if !o.IsCamera() {
			entities[o.GetGuid()] = true
		}
	}
}

// 获取与实体范围相叠的相机
func (view *AsynAoiBase) GetCameraGuidNearTheVolModel(voltageModel quadTree.Object, cameras utils.GuidMap) {
	rect := voltageModel.GetBounds()
	objects := view.QuadTree.RetrieveIntersections(rect, true, define.QuadTreeSearchTypCamera)
	for _, o := range objects {
		cameras[o.GetGuid()] = true
	}
}

func (view *AsynAoiBase) AddCamera(userId uint64, bound quadTree.Bounds) {
	camera, ok := view.CameraMap[userId]
	if !ok {
		camera = NewAsyncCamera(userId, bound)
		view.CameraMap[userId] = camera
		view.QuadTree.Insert(camera)
		entities := view.GuidMapPool.Get() // put back
		view.GetVolModelNearTheCamera(camera, entities)
		if len(entities) > 0 {
			cameras := view.GuidMapPool.Get() // put back
			cameras[camera.GetGuid()] = true
			view.NotifyEnter(entities, cameras)
			view.GuidMapPool.Puts(cameras)
		}
		view.GuidMapPool.Puts(entities)
	}
}

func (view *AsynAoiBase) RemoveCamera(userId uint64) {
	camera := view.CameraMap[userId]
	if camera == nil {
		return
	}
	delete(view.CameraMap, userId)
	view.QuadTree.Delete(camera)
	oldEntityGuidMap := camera.GetSeeEntities()
	if len(oldEntityGuidMap) > 0 {
		cameras := view.GuidMapPool.Get()
		cameras[camera.GetGuid()] = true
		view.NotifyLeave(oldEntityGuidMap, cameras)
		view.GuidMapPool.Puts(cameras)
	}
}

func (view *AsynAoiBase) UpdateUserCamera(userId uint64, bound quadTree.Bounds) {
	camera := view.CameraMap[userId]
	if camera == nil {
		return
	}

	view.QuadTree.Delete(camera)
	camera.Bounds = bound
	view.QuadTree.Insert(camera)

	entities := view.GuidMapPool.Get() // put back
	view.GetVolModelNearTheCamera(camera, entities)
	oldEntityGuidMap := camera.GetSeeEntities()
	newEntityGuidMap := entities

	enterEntities := view.GuidMapPool.Get() // put back
	leaveEntities := view.GuidMapPool.Get() // put back
	utils.GuidMapDifference(newEntityGuidMap, oldEntityGuidMap, enterEntities)
	utils.GuidMapDifference(oldEntityGuidMap, newEntityGuidMap, leaveEntities)
	// notify entity enter or leave camera
	cameras := view.GuidMapPool.Get() // put back
	if len(enterEntities) > 0 {
		cameras[camera.GetGuid()] = true
		view.NotifyEnter(enterEntities, cameras)
	}
	if len(leaveEntities) > 0 {
		cameras[camera.GetGuid()] = true
		view.NotifyLeave(leaveEntities, cameras)
	}

	view.GuidMapPool.Puts(entities, enterEntities, leaveEntities, cameras)
}

func (view *AsynAoiBase) updateAsynAoiEntityModel(asynEntityModel IAsynEntityModel) {
	oldVoltageModel := view.AsynAoiModels[asynEntityModel.GetGuid()]
	if oldVoltageModel != nil {
		view.QuadTree.Delete(oldVoltageModel)
	}

	view.AsynAoiModels[asynEntityModel.GetGuid()] = asynEntityModel
	view.QuadTree.Insert(asynEntityModel)

	cameras := view.GuidMapPool.Get() // put back
	view.GetCameraGuidNearTheVolModel(asynEntityModel, cameras)
	oldCameraGuidMap := asynEntityModel.GetSeeMeCameras()
	newCameraGuidMap := cameras

	enterCameras := view.GuidMapPool.Get() // put back
	leaveCameras := view.GuidMapPool.Get() // put back
	utils.GuidMapDifference(newCameraGuidMap, oldCameraGuidMap, enterCameras)
	utils.GuidMapDifference(oldCameraGuidMap, newCameraGuidMap, leaveCameras)

	entities := view.GuidMapPool.Get() // put back
	if len(enterCameras) > 0 {
		entities[asynEntityModel.GetGuid()] = true
		view.NotifyEnter(entities, enterCameras)
	}
	if len(leaveCameras) > 0 {
		entities[asynEntityModel.GetGuid()] = true
		view.NotifyLeave(entities, leaveCameras)
	}
	view.GuidMapPool.Puts(cameras, enterCameras, leaveCameras, entities)
}

func (view *AsynAoiBase) addAsynAoiEntityModel(asynEntityModel IAsynEntityModel) {
	view.AsynAoiModels[asynEntityModel.GetGuid()] = asynEntityModel
	view.QuadTree.Insert(asynEntityModel)
	cameras := view.GuidMapPool.Get() // put back
	view.GetCameraGuidNearTheVolModel(asynEntityModel, cameras)
	if len(cameras) > 0 {
		entities := view.GuidMapPool.Get() // put back
		entities[asynEntityModel.GetGuid()] = true
		view.NotifyEnter(entities, cameras)
		view.GuidMapPool.Puts(entities)
	}
	view.GuidMapPool.Puts(cameras)
}

func (view *AsynAoiBase) RemoveModel(guid uint64) {
	asynEntityModel := view.AsynAoiModels[guid]
	if asynEntityModel == nil {
		return
	}

	view.QuadTree.Delete(asynEntityModel)
	cameras := asynEntityModel.GetSeeMeCameras()
	if len(cameras) > 0 {
		entities := view.GuidMapPool.Get() // put back
		entities[asynEntityModel.GetGuid()] = true
		view.NotifyLeave(entities, cameras)
		view.GuidMapPool.Puts(entities)
	}

	delete(view.AsynAoiModels, asynEntityModel.GetGuid())
}

func (view *AsynAoiBase) NotifyEnter(entities utils.GuidMap, cameras utils.GuidMap) {
	for guid := range entities {
		voltageModel := view.AsynAoiModels[guid]
		if voltageModel == nil {
			delete(entities, guid)
			continue
		}

		voltageModel.AddSeeMeCameras(cameras)
	}

	if len(entities) <= 0 {
		return
	}

	for cameraGuid := range cameras {
		camera := view.CameraMap[cameraGuid]
		if camera != nil {
			camera.AddSeeEntities(entities)
		} else {
			delete(cameras, cameraGuid)
		}
	}

	if view.enterNotifyFc != nil {
		view.enterNotifyFc(entities, cameras)
	}
}

func (view *AsynAoiBase) NotifyLeave(entities utils.GuidMap, cameras utils.GuidMap) {
	var copyCameras = view.GuidMapPool.Get()
	for cameraGuid := range cameras {
		camera := view.CameraMap[cameraGuid]
		if camera != nil {
			camera.DelSeeEntities(entities)
			copyCameras[cameraGuid] = true
		} else {
			delete(cameras, cameraGuid)
		}
	}

	for guid := range entities {
		voltageModel := view.AsynAoiModels[guid]
		if voltageModel != nil {
			voltageModel.RemoveSeeMeCameras(cameras)
		} else {
			delete(entities, guid)
		}
	}
	view.GuidMapPool.Puts(copyCameras)

	if view.leaveNotifyFc != nil {
		view.leaveNotifyFc(entities, copyCameras)
	}
}

func (view *AsynAoiBase) GetAysnEntityModel(guid uint64) IAsynEntityModel {
	return view.AsynAoiModels[guid]
}

func (view *AsynAoiBase) Log(asynAoiTyp int32) {
	var stringBuilder bytes.Buffer
	stringBuilder.WriteString(fmt.Sprintf("AsynAoi info typ【%d】 totalModels:%d ", asynAoiTyp, len(view.AsynAoiModels)))
	var tot int
	maxLevel := []int{20, 50, 100, 300, 6000}
	maxCnt := []int{0, 0, 0, 0, 0}
	maxTot := []int{0, 0, 0, 0, 0}
	for _, c := range view.CameraMap {
		cnt := len(c.seeEntities)
		tot += cnt
		for l, lc := range maxLevel {
			if cnt < lc {
				maxCnt[l]++
				maxTot[l] += cnt
				break
			}
		}
	}
	stringBuilder.WriteString(fmt.Sprintf("VolCamCnt:%d seeVolCntAvg:%.2f seeVolCntAvg_%d:%.2f|%d seeVolCntAvg_%d:%.2f|%d seeVolCntAvg_%d:%.2f|%d seeVolCntAvg_%d:%.2f|%d seeVolCntAvg_%d:%.2f|%d",
		len(view.CameraMap),
		div(tot, len(view.CameraMap)),
		maxLevel[0], div(maxTot[0], maxCnt[0]), maxCnt[0],
		maxLevel[1], div(maxTot[1], maxCnt[1]), maxCnt[1],
		maxLevel[2], div(maxTot[2], maxCnt[2]), maxCnt[2],
		maxLevel[3], div(maxTot[3], maxCnt[3]), maxCnt[3],
		maxLevel[4], div(maxTot[4], maxCnt[4]), maxCnt[4]),
	)
	log.Info(stringBuilder.String())
}

func div(x, y int) float32 {
	if y == 0 {
		return float32(x)
	} else {
		return float32(x) / float32(y)
	}
}

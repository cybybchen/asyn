package asyn_aoi

import (
	"twe/define"
	"twe/scene/scene_utils"
	"twe/utils"
	"twe/utils/lmath"
	"twe/utils/quadTree"
)

type IAsynAoi interface {
	GetAsynAoiTyp() int32
	GetAysnEntityModel(guid uint64) IAsynEntityModel

	AddCamera(userId uint64, bound quadTree.Bounds)
	RemoveCamera(userId uint64)
	UpdateUserCamera(userId uint64, bound quadTree.Bounds)

	AddModel(model ReqAoiInf)
	RemoveModel(guid uint64)
	UpdateModel(model ReqAoiInf)
	UpdateAndBroadcastModel(model ReqAoiInf)

	Log(asynAoiTyp int32)
}

type NotifyFc func(entities utils.GuidMap, cameras utils.GuidMap)

type AsynAoiBase struct {
	QuadTree      *quadTree.QuadTree
	CameraMap     map[uint64]*AsyncCamera
	GuidMapPool   *utils.GuidMapPool // avoid creating too much map object
	AsynAoiModels map[uint64]IAsynEntityModel

	enterNotifyFc NotifyFc
	leaveNotifyFc NotifyFc
}

func (asynAoiBase *AsynAoiBase) InitAsynAoiBase(enterNotifyFc NotifyFc, leaveNotifyFc NotifyFc) {
	width := float32(utils.RoundUpToPowerOfTwoMin(define.MapWidth, 8) * scene_utils.CellSize)
	height := float32(utils.RoundUpToPowerOfTwoMin(define.MapHeight, 8) * scene_utils.CellSize)
	asynAoiBase.QuadTree = quadTree.NewQuadTree(width, height, 20, 8)
	asynAoiBase.CameraMap = make(map[uint64]*AsyncCamera, 1024)
	asynAoiBase.GuidMapPool = utils.NewGuidMapPool(8)
	asynAoiBase.AsynAoiModels = make(map[uint64]IAsynEntityModel, 2048)
	asynAoiBase.enterNotifyFc = enterNotifyFc
	asynAoiBase.leaveNotifyFc = leaveNotifyFc
}

func (asynAoiBase *AsynAoiBase) SetEnterNotifyFc(fc func(entities utils.GuidMap, cameras utils.GuidMap)) {
	asynAoiBase.enterNotifyFc = fc
}

func (asynAoiBase *AsynAoiBase) SetLeaveNotifyFc(fc func(entities utils.GuidMap, cameras utils.GuidMap)) {
	asynAoiBase.leaveNotifyFc = fc
}

type AsyncCamera struct {
	quadTree.ObjectBase
	x           float32
	z           float32
	seeEntities utils.GuidMap
}

func NewAsyncCamera(guid uint64, bound quadTree.Bounds) *AsyncCamera {
	cb := &AsyncCamera{
		ObjectBase:  quadTree.NewObjectBase(guid),
		seeEntities: map[uint64]bool{},
	}
	cb.SetBounds(&bound)
	return cb
}

func (camera *AsyncCamera) updateAsyncCamera(x, z float32, cameraRange int32) {
	bound := quadTree.NewBounds(x-float32(cameraRange/2), z-float32(cameraRange/2), float32(cameraRange), float32(cameraRange))
	camera.SetBounds(bound)
}

func (camera *AsyncCamera) GetSearchType() define.QuadTreeSearchTyp {
	return define.QuadTreeSearchTypCamera
}

func (camera *AsyncCamera) AddSeeEntities(cameras utils.GuidMap) {
	for guid := range cameras {
		camera.seeEntities[guid] = true
	}
}

func (camera *AsyncCamera) DelSeeEntities(cameras utils.GuidMap) {
	for guid := range cameras {
		delete(camera.seeEntities, guid)
	}
}

func (camera *AsyncCamera) GetSeeEntities() utils.GuidMap {
	return camera.seeEntities
}

type IAsynEntityModel interface {
	quadTree.Object
	RemoveSeeMeCameras(cameras utils.GuidMap)
	AddSeeMeCameras(cameras utils.GuidMap)
	GetSeeMeCameras() utils.GuidMap
}

type AsynEntityModelBase struct {
	quadTree.ObjectBase
	seeMeCamerasMap map[uint64]bool
	pos             lmath.Vector3
}

func (base *AsynEntityModelBase) InitAsynEntityModelBase(guid uint64, bound quadTree.Bounds, pos lmath.Vector3) {
	base.ObjectBase = quadTree.NewObjectBase(guid)
	base.seeMeCamerasMap = map[uint64]bool{}
	base.pos = pos

	base.SetBounds(&bound)
	return
}

func (this *AsynEntityModelBase) GetSearchType() define.QuadTreeSearchTyp {
	return 0
}

func (this *AsynEntityModelBase) RemoveSeeMeCameras(cameras utils.GuidMap) {
	for guid, _ := range cameras {
		delete(this.seeMeCamerasMap, guid)
	}
}

func (this *AsynEntityModelBase) AddSeeMeCameras(cameras utils.GuidMap) {
	for guid, _ := range cameras {
		this.seeMeCamerasMap[guid] = true
	}
}

func (this *AsynEntityModelBase) GetSeeMeCameras() utils.GuidMap {
	return this.seeMeCamerasMap
}

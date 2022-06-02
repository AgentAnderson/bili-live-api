package api

import (
	"fmt"
	"time"

	jsoniter "github.com/json-iterator/go"

	"github.com/botplayerneo/bili-live-api/dto"
	"github.com/botplayerneo/bili-live-api/log"
	"github.com/botplayerneo/bili-live-api/resource"
	"github.com/botplayerneo/bili-live-api/websocket"
)

// Live 使用 NewLive() 来初始化
type Live struct {
	client *websocket.Client
	roomID int
}

// NewLive 构造函数
func NewLive(roomID int) *Live {
	return &Live{
		roomID: roomID,
	}
}

// Start 接收房间号，开始websocket心跳连接并阻塞
func (l *Live) Start() {
	for {
		l.client = websocket.New()
		if err := l.Listen(); err != nil {
			log.Warnf("连接失败, 重连中...:%v", err)
			time.Sleep(5 * time.Second)
			continue
		}
		break
	}
}

func (l *Live) Listen() error {
	id, err := resource.RealRoomID(l.roomID)
	if err != nil {
		return fmt.Errorf("获取房间号失败：%v", err)
	}

	if err := l.client.Connect(); err != nil {
		return fmt.Errorf("连接websocket失败：%v", err)
	}

	// TODO 发送进房包,可能有顺序问题
	go l.enterRoom(id)

	if err := l.client.Listening(); err != nil {
		return fmt.Errorf("监听websocket失败：%v", err)
	}
	return nil
}

// RegisterHandlers 注册不同的事件处理
// handler类型需要是定义在 websocket/handler_registration.go 中的类型，如:
// - websocket.DanmakuHandler
// - websocket.GiftHandler
// - websocket.GuardHandler
func (l *Live) RegisterHandlers(handlers ...interface{}) error {
	return websocket.RegisterHandlers(handlers...)
}

// 发送进入房间请求
func (l *Live) enterRoom(id int) {
	log.Infof("进入房间：%d", id)
	// 忽略错误
	var err error
	body, _ := jsoniter.Marshal(dto.WSEnterRoomBody{
		RoomID:    id, // 真实房间ID
		ProtoVer:  1,  // 填1
		Platform:  "web",
		ClientVer: "1.6.3",
	})
	if err = l.client.Write(&dto.WSPayload{
		ProtocolVersion: dto.JSON,
		Operation:       dto.RoomEnter,
		Body:            body,
	}); err != nil {
		log.Errorf("发送进入房间请求失败：%v", err)
		return
	}
}

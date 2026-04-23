package general_system_game

import (
	"context"
	"fmt"
	"g7/common/etcd"
	"g7/common/globals"
	"g7/common/logger"
	"g7/common/model_common"
	"g7/common/protocol"
	"g7/common/protos/pb"
	"g7/common/structs"
	"g7/common/utils"
	"g7/game/const_game"
	"g7/game/global_game"
	"g7/game/manager_game"
	"g7/game/model_game"
	"gorm.io/gorm"
	"log"
	"time"
)

var GMailSystem = &mailSystem{}

type mailSystem struct {
	curBaseId int64
}

func init() {
	manager_game.GISystemManager.Register(const_game.General_MailSystem, GMailSystem)
}

func (this *mailSystem) Init() {
	this.curBaseId = 20
}

func (this *mailSystem) GetName() string {
	return "general_mail_system"
}

func (this *mailSystem) LoadData(dao *model_game.PlayerDao, Player *model_game.Player) {
	Player.AllMailData = dao.GeneralD.MailData
	if Player.AllMailData.CurBaseMailId <= this.curBaseId {
		logger.Log.Info("base mail has changed")
		go this.syncBaseMailToPlayer(16, this.curBaseId, Player.PlayerId)
		Player.AllMailData.CurBaseMailId = this.curBaseId
		// 数据库批量获取然后生成
	}

}

func (this *mailSystem) syncBaseMailToPlayer(srcId, tarId, playerId int64) {
	//if tarId <= srcId {
	//	return
	//}

	// 构造过滤条件：双库通用（MySQL/Mongo都支持）
	where := "id > ? AND id <= ? AND end_time > ?"
	// 2. 参数
	params := []any{srcId, tarId, time.Now()}

	var baseMails []*model_common.BaseMail
	err := global_game.GGlobalDB.FindList(&baseMails, where, params...)
	if err != nil {
		logger.Log.Error(err.Error())
		return
	}
	var mailsToInsert []model_common.DBTableInterface
	for _, oneBaseMail := range baseMails {
		oneMailObj := &model_common.PlayerMail{
			Title:       oneBaseMail.Title,
			ServerID:    utils.StringToInit32(globals.ServerId),
			Content:     oneBaseMail.Content,
			Attachments: oneBaseMail.Attachments, // 附件JSON
			HasAttach:   oneBaseMail.HasAttach,
			BindBaseId:  &oneBaseMail.ID,
			Status:      0, // 未读
		}
		oneMailObj.PlayerID = playerId
		oneMailObj.CreatedAt = time.Now()
		oneMailObj.ExpireAt = oneBaseMail.EndTime.Unix()
		mailsToInsert = append(mailsToInsert, oneMailObj)
	}
	//fmt.Println("wait send len ", len(mailsToInsert))
	if len(mailsToInsert) == 0 {
		fmt.Println("mailsToInsert len == 0")
		return
	}
	err = global_game.GGameDB.BatchInsert(mailsToInsert)
	if err != nil {
		logger.Log.Info("batch insert mail error" + err.Error())
	}
	player := global_game.GPlayerMaps.GetPlayer(playerId)
	player.RunInActor(func() {
		player.AllMailData.WaitNotifyCount = int32(len(mailsToInsert))
	})

}

func (this *mailSystem) DailyReset(Player *model_game.Player) {}

func (this *mailSystem) OnEnterGame(Player *model_game.Player) {

}

/*
	玩家相关
*/

// ReqPlayerMailList 获取玩家邮件列表
func (this *mailSystem) ReqPlayerMailList(Player *model_game.Player) {
	var playerMailList []*model_common.PlayerMail
	err := global_game.GGameDB.FindListPro(playerMailList, map[string]any{"player_id": Player.PlayerId, "status": gorm.Expr("!= 2")}, "created_at DESC", 50, 1)
	if err != nil {
		log.Println(err)
	}
	for _, playerMail := range playerMailList {
		fmt.Println(playerMail)
	}
}

func (this *mailSystem) DeleteMail(mailIds []int64, Player *model_game.Player) error {
	if len(mailIds) == 0 {
		return nil
	}
	err := global_game.GGameDB.Update(&model_common.PlayerMail{}, map[string]any{"status": 2}, "player_id = ? AND id IN (?)", Player.PlayerId, mailIds)
	if err != nil {
		log.Println(err)
	}
	return err
}

// BatchReceiveAttach 批量领取邮件附件（支持多个邮件ID）
func (this *mailSystem) BatchReceiveAttach(mailIDs []int64, Player *model_game.Player) error {
	// 空邮件列表直接返回
	if len(mailIDs) == 0 {
		return nil
	}

	var rewards = make([]structs.KInt32VInt64Bind, 16)
	// ========================
	// 事务：保证不重复领取、不丢道具
	tx := global_game.GGameDB.TxBegin()
	defer func() {
		if err := recover(); err != nil {
			_ = tx.TxRollback()
		}
	}()

	// 2. 事务内查询邮件（走接口！）
	var mails []*model_common.PlayerMail
	// 构造过滤条件
	where := "player_id = ? AND status != ? AND id in(?)"
	// 2. 参数
	params := []any{Player.PlayerId, 2, mailIDs}
	err := tx.FindList(&mails, where, params...)
	if err != nil {
		_ = tx.TxRollback()
		return err
	}
	// 3. 发放道具
	for _, mail := range mails {
		for _, oneAttache := range mail.Attachments {
			rewards = append(rewards, structs.KInt32VInt64Bind{K: oneAttache.ItemID, V: oneAttache.Count, B: oneAttache.Bind})
		}
	}

	GBagSystem.GainAndConsumption(rewards, nil, "领取邮件", Player)

	updateData := map[string]any{"has_attach": 0}
	err = tx.Update(&model_common.PlayerMail{}, updateData, "player_id = ? AND id IN (?)", Player.PlayerId, mailIDs)
	if err != nil {
		tx.TxRollback()
		return err
	}

	// 5. 提交事务
	return tx.TxCommit()
}

/*
系统职能
*/

func (this *mailSystem) SendDefaultSystemTypeMail(title, content string, attaches []model_common.Attachment, presetSendTimeStamp int64, expireHours int32, sender string) error {
	deadlineTimeStamp := utils.FormatTimestamp(presetSendTimeStamp).Add(time.Hour * time.Duration(expireHours)).Unix()
	baseMail := this.makeSystemMailBase(title, content, attaches, presetSendTimeStamp, deadlineTimeStamp, sender)
	targetSeverId := utils.StringToInit64(globals.ServerId)
	baseMail.TargetServerID = &targetSeverId
	err := global_game.GGlobalDB.Insert(baseMail)
	if err != nil {
		log.Println(err)
		return err
	}
	serverList, err := etcd.GetGameServersByServerID(globals.ServerId)
	if err != nil {
		log.Println(err)
		return err
	}
	req := pb.Req_Node_NewBaseMail{MailId: baseMail.ID}
	for _, serverKV := range serverList {
		client, err := protocol.NewGameNodeClient(serverKV.V)
		if err != nil {
			log.Printf("%s send mail fail", serverKV.K)
			continue
		}
		ctxReq, cancelConn := context.WithTimeout(context.Background(), 3*time.Second)
		defer cancelConn()
		_, err = client.NotifyNewBaseMail(ctxReq, &req)
		if err != nil {
			log.Printf("%s send mail fail%s", serverKV.K, err.Error())
			continue
		}
	}
	return nil
}

func (this *mailSystem) RecvNode_NewBaseMail(nodeMsg *pb.Req_Node_NewBaseMail) {
	playerIds := global_game.GPlayerMaps.GetAllPlayerIds()
	for _, playerId := range playerIds {
		fmt.Println("onlinePlayer", playerId)
	}
	var mailsToInsert []model_common.DBTableInterface

	this.curBaseId = nodeMsg.GetMailId()

	baseMail := &model_common.BaseMail{}
	err := global_game.GGlobalDB.FindOne(baseMail, map[string]any{"id": nodeMsg.MailId})
	if err != nil {
		logger.Log.Info("find base mail error" + err.Error())
		return
	}

	// 遍历你本服的所有在线玩家

	for _, playerId := range playerIds {
		oneMailObj := &model_common.PlayerMail{
			Title:       baseMail.Title,
			ServerID:    utils.StringToInit32(globals.ServerId),
			Content:     baseMail.Content,
			Attachments: baseMail.Attachments, // 附件JSON
			HasAttach:   baseMail.HasAttach,
			BindBaseId:  &baseMail.ID,
			Status:      0, // 未读
		}
		oneMailObj.PlayerID = playerId
		oneMailObj.CreatedAt = time.Now()
		oneMailObj.ExpireAt = baseMail.EndTime.Unix()
		mailsToInsert = append(mailsToInsert, oneMailObj)
	}
	fmt.Println("wait send len ", len(mailsToInsert))
	if len(mailsToInsert) == 0 {
		fmt.Println("mailsToInsert len == 0")
		return
	}
	err = global_game.GGameDB.BatchInsert(mailsToInsert)
	if err != nil {
		logger.Log.Info("batch insert mail error" + err.Error())
	}

}

func (this *mailSystem) makeSystemMailBase(title, content string, attaches []model_common.Attachment, presetSendTimeStamp, deadlineTimeStamp int64, sender string) *model_common.BaseMail {
	attachments := model_common.Attachments(attaches)
	hasAttach := int32(0)
	if len(attaches) > 0 {
		hasAttach = 1
	}
	var creator string
	if len(sender) == 0 {
		creator = "system"
	}

	baseMail := &model_common.BaseMail{
		MailType:       globals.MailTypeSystem,
		Title:          title,
		Content:        content,
		HasAttach:      hasAttach,
		Attachments:    attachments,
		TargetServerID: nil,
		TargetGuildID:  nil,
		StartTime:      utils.FormatTimestamp(presetSendTimeStamp),
		EndTime:        utils.FormatTimestamp(deadlineTimeStamp),
		Status:         globals.MailStatusPending,
		SentCount:      0,
		CreatedAt:      time.Now(),
		CreatedBy:      creator,
		UpdatedAt:      time.Now(),
	}
	return baseMail
}

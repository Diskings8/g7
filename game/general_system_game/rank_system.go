package general_system_game

import (
	"fmt"
	"g7/common/globals"
	"g7/common/logger"
	"g7/common/protos/pb"
	"g7/common/redisx"
	"g7/common/utils"
	"g7/game/const_game"
	"g7/game/manager_game"
	"g7/game/model_game"
	"github.com/golang/protobuf/proto"
	"time"
)

var GRankSystem = &rankSystem{}

type rankSystem struct {
}

func init() {
	manager_game.GISystemManager.Register(const_game.General_RankSystem, GRankSystem)
}

func (this *rankSystem) Init() {

}

func (this *rankSystem) GetName() string {
	return "general_rank_system"
}

func (this *rankSystem) LoadData(dao *model_game.PlayerDao, Player *model_game.Player) {

}

func (this *rankSystem) DailyReset(Player *model_game.Player) {}

func (this *rankSystem) OnEnterGame(Player *model_game.Player) {

}

func (this *rankSystem) ReqGetRank(reqD any, player *model_game.Player) any {
	req := &pb.Req_RankList{}
	_ = proto.Unmarshal(reqD.([]byte), req)
	var rankType = req.GetRankType()
	var contentKey string
	switch rankType {
	case globals.RankTypeBattleScore:
		contentKey = utils.Int64ToString(player.PlayerId)
	}
	rsp := &pb.Rsp_RankList{}
	l, r, s, e := this.GetReqRankInfo(rankType, contentKey)
	if e != nil {
		logger.Log.Warn(e.Error())
		return rsp
	}
	rsp.RankList = l
	rsp.MyRank = r
	rsp.MyScore = s
	return rsp
}

func (this *rankSystem) getRankKeyByType(rankType int32) (rankKey string) {
	switch rankType {
	case globals.RankTypeBattleScore:
		rankKey = fmt.Sprintf("rank_battlescore_%s", globals.ServerId)

	default:

	}
	return
}

func (this *rankSystem) EnterAnyRank(rankType int32, score int64, Player *model_game.Player) {
	var rankKey, contentKey string
	var rankScore float64
	switch rankType {
	case globals.RankTypeBattleScore:
		contentKey = utils.Int64ToString(Player.PlayerId)
		rankScore = float64(score)
	}
	rankKey = this.getRankKeyByType(rankType)

	this.updatePlayerRank(rankKey, contentKey, rankScore)
}

func (this *rankSystem) updatePlayerRank(rankKey string, contentKey string, score float64) {
	// 扔异步，不卡玩家
	go func() {
		// 1. 获取分布式锁，防止并发乱序
		var retry, retryMax = 0, 5
		lockKey := rankKey + ":lock"
		for retry < retryMax {
			ok := redisx.TryLock(lockKey, 2*time.Second)
			if !ok {
				// 拿不到锁，延迟重试
				retry++
				time.Sleep(100 * time.Millisecond)
				continue
			}
			defer redisx.Unlock(lockKey)
			redisx.ZIncrBy(rankKey, score, contentKey)
		}
		logger.Log.Warn(fmt.Sprintf("lock key:%s, score:%f error", lockKey, score))
	}()
}

func (this *rankSystem) ClearRank(rankType int32) {
	rankKey := this.getRankKeyByType(rankType)
	redisx.Clear(rankKey)
}

func (this *rankSystem) GetRankList(rankType int32, from, to int64) ([]*pb.RankItemInfo, error) {
	rankKey := this.getRankKeyByType(rankType)
	res, err := redisx.ZRevRangeWithScores(rankKey, from, to)
	if err != nil {
		return nil, err
	}
	var rankList []*pb.RankItemInfo
	for rankIndex, rankRec := range res {
		one := &pb.RankItemInfo{
			Score:      rankRec.Score,
			Rank:       from + int64(rankIndex) + 1,
			ContentKey: rankRec.Member.(string),
		}
		rankList = append(rankList, one)
	}
	return rankList, nil
}

func (this *rankSystem) GetReqRankInfo(rankType int32, contentKey string) (rankList []*pb.RankItemInfo, rank int64, score float64, err error) {

	rankList, err = this.GetRankList(rankType, 0, 100)
	if err != nil {
		return nil, 0, 0, err
	}
	for _, one := range rankList {
		if one.ContentKey == contentKey {
			return rankList, one.GetRank(), one.GetScore(), nil
		}
	}
	return rankList, 0, 0, nil
}

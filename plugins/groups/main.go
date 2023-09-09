package groups

import (
	"time"

	"github.com/yqchilde/wxbot/engine/control"
	"github.com/yqchilde/wxbot/engine/pkg/log"
	"github.com/yqchilde/wxbot/engine/pkg/sqlite"
	"github.com/yqchilde/wxbot/engine/robot"
	"github.com/yqchilde/wxbot/plugins/daily"
)

var (
	db sqlite.DB
)

func init() {
	engine := control.Register("groups", &control.Options[*robot.Ctx]{
		Alias:      "虫洞",
		Help:       "通往其他群聊",
		DataFolder: "groups",
	})

	if err := sqlite.Open("data/plugins/daily/daily.db", &db); err != nil {
		log.Fatalf("open sqlite db failed: %v", err)
	}

	engine.OnFullMatch(`虫洞`, robot.OnlyPrivate).SetBlock(true).Handle(func(ctx *robot.Ctx) {
		ctx.ReplyText("请输入需要通往的空间！")
		recv, cancel := ctx.EventChannel(ctx.CheckUserSession()).Repeat()
		defer cancel()
		var msg string
		for {
			select {
			case <-time.After(30 * time.Second):
				ctx.ReplyText("操作时间太久了，请重新设置")
				return
			case ctx := <-recv:
				var groupInfo daily.GroupInfo
				msg = ctx.MessageString()
				result := db.Orm.Model(&daily.GroupInfo{}).Where("alias = ?", msg).Find(&groupInfo)
				if result.Error != nil {
				        ctx.ReplyText("创建空间虫洞失败！请稍候重试！")
					return
				}
				if err := ctx.InviteIntoGroup(groupInfo.GroupId, ctx.Event.FromWxId, 2); err != nil {
					ctx.ReplyText("创建空间虫洞失败！请稍候重试！")
				}
				return
			}
		}
	})
}


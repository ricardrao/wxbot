package message

import (
	"crypto/rc4"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/yqchilde/wxbot/engine/control"
	"github.com/yqchilde/wxbot/engine/robot"
)

var (
	key    = []byte("chiguaaihaozhe")
	cipher *rc4.Cipher
)

func init() {
	engine := control.Register("message", &control.Options[*robot.Ctx]{
		Alias:      "消息加密",
		Help:       "输入 {封装包裹，拆封包裹} => 将消息进行加密和对消息进行解密",
		DataFolder: "message",
	})

	engine.OnFullMatch(`加封包裹`, robot.OnlyPrivate).SetBlock(true).Handle(func(ctx *robot.Ctx) {
		ctx.ReplyText("收到！请输入内容我会将其加封。")
		recv, cancel := ctx.EventChannel(ctx.CheckUserSession()).Repeat()
		defer cancel()
		var msg string
		for {
			select {
			case <-time.After(30 * time.Second):
				ctx.ReplyText("操作时间太久了，请重新设置")
				return
			case ctx := <-recv:
				msg = ctx.MessageString()
				dest := make([]byte, len([]byte(msg)))
				cipher, _ = rc4.NewCipher(key)
				cipher.XORKeyStream(dest, []byte(msg))
				ctx.ReplyText("包裹加封中......请稍后。")
				ctx.ReplyText("加封完成。")
				tmp := ""
				for i := 0; i < len(dest); i++ {
					tmp += strconv.Itoa(int(dest[i])) + "/"
				}
				ctx.ReplyText(tmp)
				return
			}
		}
	})

	engine.OnFullMatch("拆封包裹", robot.OnlyPrivate).SetBlock(true).Handle(func(ctx *robot.Ctx) {
		ctx.ReplyText("收到！请输入内容我会将其拆封。")
		recv, cancel := ctx.EventChannel(ctx.CheckUserSession()).Repeat()
		defer cancel()
		var msg string
		for {
			select {
			case <-time.After(30 * time.Second):
				ctx.ReplyText("操作时间太久了，请重新设置")
				return
			case ctx := <-recv:
				msg = ctx.MessageString()
				s := strings.Split(msg, "/")
				fmt.Println(s)
				tmp := make([]byte, 0)
				for _, v := range s {
					if v != "" {
						num, _ := strconv.Atoi(v)
						tmp = append(tmp, byte(num))
					}
				}
				dest := make([]byte, len(tmp))
				ctx.ReplyText("包裹拆封中......请稍后。")
				cipher, _ = rc4.NewCipher(key)
				cipher.XORKeyStream(dest, tmp)
				ctx.ReplyText("拆封完成。")
				ctx.ReplyText(string(dest))
				return
			}
		}
	})
}

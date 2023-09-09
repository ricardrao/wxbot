package chatgpt

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/PullRequestInc/go-gpt3"

	"github.com/yqchilde/wxbot/engine/control"
	"github.com/yqchilde/wxbot/engine/pkg/log"
	"github.com/yqchilde/wxbot/engine/pkg/sqlite"
	"github.com/yqchilde/wxbot/engine/robot"
)

var (
	db         sqlite.DB
	chatGPT    ChatGPT
	gpt3Client gpt3.Client
	chatCTXMap sync.Map // 群号/私聊:消息上下文
	chatDone   = make(chan struct{})
)

func init() {
	engine := control.Register("chatgpt", &control.Options[*robot.Ctx]{
		Alias:      "问答",
		Help:       "输入 {开始问答} => 进行问答",
		DataFolder: "chatgpt",
		OnDisable: func(ctx *robot.Ctx) {
			ctx.ReplyText("禁用成功")
			chatDone <- struct{}{}
		},
	})

	if err := sqlite.Open(engine.GetDataFolder()+"/chatgpt.db", &db); err != nil {
		log.Fatalf("open sqlite db failed: %v", err)
	}
	if err := db.CreateAndFirstOrCreate("chatgpt", &chatGPT); err != nil {
		log.Fatalf("create chatgpt table failed: %v", err)
	}
	gpt3Client = gpt3.NewClient(chatGPT.ApiKey, gpt3.WithTimeout(time.Minute))

	engine.OnRegex("开始问答|来陪我唠嗑|陪我唠嗑|陪我说说话|来聊个五毛钱的天", robot.OnlyAtMe).SetBlock(true).Handle(func(ctx *robot.Ctx) {
		if chatGPT.ApiKey == "" {
			ctx.ReplyTextAndAt("请先私聊机器人配置apiKey\n指令：set chatgpt apiKey __\napiKey获取请到https://beta.openai.com获取")
			return
		}
		if _, ok := chatCTXMap.Load(ctx.Event.FromUniqueID); ok {
			ctx.ReplyTextAndAt("当前已经在会话中")
			return
		}

		recv, cancel := ctx.EventChannel(ctx.CheckGroupSession()).Repeat()
		defer cancel()
		ctx.ReplyTextAndAt("收到！已开始会话，输入\"结束问答\"结束问答，或5分钟后自动结束，请开始吧！")
		chatCTXMap.LoadOrStore(ctx.Event.FromUniqueID, "")
		for {
			select {
			case <-time.After(time.Minute * 5):
				chatCTXMap.LoadAndDelete(ctx.Event.FromUniqueID)
				ctx.ReplyTextAndAt("😊检测到您已有5分钟不再提问，那我先主动结束会话咯")
				return
			case <-chatDone:
				chatCTXMap.LoadAndDelete(ctx.Event.FromUniqueID)
				ctx.ReplyText("已退出会话")
				return
			case c := <-recv:
				if robot.OnlyAtMe(c) {
					if !robot.RegexRule(`\@[^ ]+\s(.+)`)(c) {
						continue
					}
					msg := c.State["regex_matched"].([]string)[1]
					if msg == "" {
						continue
					} else if msg == "结束问答" || msg == "就聊到这里" || msg == "我们下次再聊" || msg == "你说得很好，下次再说吧！" || msg == "再见" {
						chatCTXMap.LoadAndDelete(ctx.Event.FromUniqueID)
						ctx.ReplyText("已结束会话")
						return
					}
					question, answer := msg+"\n", ""
					if question == "" {
						continue
					}
					if c, ok := chatCTXMap.Load(ctx.Event.FromUniqueID); ok {
						question = c.(string) + question
					}
					time.Sleep(3 * time.Second)
					answer, err := askChatGPT(question)
					if err != nil {
						chatCTXMap.LoadAndDelete(ctx.Event.FromUniqueID)
						ctx.ReplyText("讨厌人家被你玩坏了！")
						continue
					}
					chatCTXMap.Store(ctx.Event.FromUniqueID, question+"\n"+answer)
					if r, need := filterReply(answer); need {
						answer, err := askChatGPT(question + "\n" + answer + r)
						if err != nil {
							chatCTXMap.LoadAndDelete(ctx.Event.FromUniqueID)
							ctx.ReplyText("讨厌人家被你玩坏了！")
							continue
						}
						chatCTXMap.Store(ctx.Event.FromUniqueID, question+"\n"+answer)
						ctx.ReplyText(answer)
					} else {
						ctx.ReplyText(r)
					}
				}
			}
		}
	})

	// 设置openai api key
	engine.OnRegex("set chatgpt apiKey (.*)", robot.OnlyPrivate, robot.AdminPermission).SetBlock(true).Handle(func(ctx *robot.Ctx) {
		apiKey := ctx.State["regex_matched"].([]string)[1]
		if err := db.Orm.Table("chatgpt").Where("1 = 1").Update("api_key", apiKey).Error; err != nil {
			ctx.ReplyTextAndAt("设置apiKey失败")
			return
		}
		chatGPT.ApiKey = apiKey
		gpt3Client = gpt3.NewClient(chatGPT.ApiKey, gpt3.WithTimeout(time.Minute))
		ctx.ReplyText("apiKey设置成功")
	})

	// 获取插件配置
	engine.OnFullMatch("get chatgpt info", robot.OnlyPrivate, robot.AdminPermission).SetBlock(true).Handle(func(ctx *robot.Ctx) {
		var chatGPT ChatGPT
		if err := db.Orm.Table("chatgpt").Limit(1).Find(&chatGPT).Error; err != nil {
			return
		}
		ctx.ReplyTextAndAt(fmt.Sprintf("插件 - ChatGPT\napiKey: %s", chatGPT.ApiKey))
	})
}

type ChatGPT struct {
	ApiKey string `gorm:"column:api_key"`
}

func askChatGPT(question string) (string, error) {
	resp, err := gpt3Client.CompletionWithEngine(context.Background(), gpt3.TextDavinci003Engine, gpt3.CompletionRequest{
		Prompt:           []string{question},
		MaxTokens:        gpt3.IntPtr(512),
		Temperature:      gpt3.Float32Ptr(0.7),
		TopP:             gpt3.Float32Ptr(1),
		Echo:             false,
		PresencePenalty:  0,
		FrequencyPenalty: 0,
	})
	if err != nil {
		return "", err
	}
	return resp.Choices[0].Text, nil
}

func filterReply(msg string) (string, bool) {
	punctuation := ",，!！?？"
	msg = strings.TrimSpace(msg)
	if len(msg) == 1 {
		return msg, true
	}
	if len(msg) == 3 && strings.ContainsAny(msg, punctuation) {
		return msg, true
	}
	msg = strings.TrimLeftFunc(msg, func(r rune) bool {
		if strings.ContainsAny(string(r), punctuation) {
			return true
		}
		return false
	})
	return msg, false
}

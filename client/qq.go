package client

import (
	"fmt"
	"math/rand"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/amtoaer/bing-bong/internal"
	"github.com/amtoaer/bing-bong/message"
	"github.com/amtoaer/bing-bong/model"
	log "github.com/sirupsen/logrus"
	zero "github.com/wdvxdr1123/ZeroBot"
	"github.com/wdvxdr1123/ZeroBot/driver"
	"github.com/wdvxdr1123/ZeroBot/extension"
)

type QQ struct {
	bot  *zero.Ctx
	Conf map[string]interface{}
}

func (q *QQ) Init() {
	var (
		account     = q.Conf["account"].(string)
		webSocket   = q.Conf["websocket"].(string)
		accessToken = q.Conf["accesstoken"].(string)
	)
	zero.Run(zero.Config{
		CommandPrefix: "/",
		Driver: []zero.Driver{
			driver.NewWebSocketClient(webSocket, accessToken),
		},
		SelfID: account,
	})
	accountNum, err := strconv.ParseInt(account, 10, 64)
	if err != nil {
		log.Fatalf("解析账号失败：%v", err)
	}
	time.Sleep(100 * time.Millisecond) // zero内部的机器人注册是异步的，这里暂且使用sleep等待0.1s，防止zero.GetBot为nil
	q.bot = zero.GetBot(accountNum)
	if q.bot != nil {
		log.Info("机器人成功启动")
	} else {
		log.Fatal("机器人启动失败")
	}
}

// 发送消息
func (q *QQ) SendMessage(userID int64, message string, isGroup bool) {
	rand.Seed(time.Now().Unix())                             //设置随机数种子
	time.Sleep(time.Duration(rand.Int63n(30)) * time.Second) //尝试随机sleep，或可降低风控风险
	if isGroup {
		q.bot.SendGroupMessage(userID, message)
	} else {
		q.bot.SendPrivateMessage(userID, message)
	}
}

// 监听事件并阻塞程序
func (q *QQ) HandleEvent(mm *message.Manager) {
	zero.OnCommandGroup([]string{"订阅", "subscribe"}).Handle(func(ctx *zero.Ctx) {
		var cmd extension.CommandModel
		err := ctx.Parse(&cmd)
		if err != nil {
			log.Errorf("处理命令失败：%v", err)
			return
		}
		if cmd.Args == "" {
			ctx.Send("请输入要订阅的链接！")
		} else {
			if _, err := url.ParseRequestURI(cmd.Args); err != nil {
				ctx.Send("请输入合法的链接！")
			} else {
				isGroup, userID := getCtxInfo(ctx)
				feeds, err := model.QueryFeed(userID, isGroup)
				if err != nil {
					log.Errorf("查询订阅出现错误：%v", err)
					ctx.Send("查询订阅出现未知错误，请联系管理员排查。")
					return
				}
				for _, existFeed := range feeds {
					if cmd.Args == existFeed.URL {
						ctx.Send("您已经订阅了该地址！")
						return
					}
				}
				ctx.Send("获取feeds信息中...")
				title, err := internal.ParseTitle(cmd.Args)
				if err != nil {
					log.Errorf("获取订阅标题失败：%v", err)
					ctx.Send("获取信息失败，请检查机器人网络并确保网址为rss地址。")
					return
				}
				mm.Subscribe(cmd.Args, &model.User{Account: userID, IsGroup: isGroup})
				err = model.InsertSubscription(cmd.Args, title, userID, isGroup)
				if err != nil {
					log.Errorf("插入订阅出现错误：%v", err)
					ctx.Send("订阅出现未知错误，请联系管理员排查。")
					return
				}
				ctx.Send(fmt.Sprintf("订阅《%s》成功！", title))
			}
		}
	})
	zero.OnCommandGroup([]string{"取消订阅", "unsubscribe"}).Handle(func(ctx *zero.Ctx) {
		isGroup, userID := getCtxInfo(ctx)
		feeds, err := model.QueryFeed(userID, isGroup)
		if err != nil {
			log.Errorf("查询订阅出现错误：%v", err)
			ctx.Send("查询订阅出现未知错误，请联系管理员排查。")
			return
		}
		hasUrl, messageToSend := buildMessage(feeds, false)
		if !hasUrl {
			ctx.Send("您还没有任何订阅内容！")
		} else {
			ctx.Send(messageToSend)
			next := ctx.FutureEvent("message", ctx.CheckSession())
			recv, cancel := next.Repeat()
			time.AfterFunc(10*time.Second, cancel)
			for event := range recv {
				numStr := event.Message.ExtractPlainText()
				if num, err := strconv.Atoi(numStr); err != nil {
					ctx.Send("请确保输入数字！")
				} else {
					num -= 1
					if num >= 0 && num < len(feeds) {
						err := model.DeleteSubscription(feeds[num].URL, userID, isGroup)
						if err != nil {
							log.Errorf("删除订阅出现错误：%v", err)
							return
						}
						mm.UnSubscribe(feeds[num].URL, &model.User{Account: userID, IsGroup: isGroup})
						ctx.Send("删除订阅成功！")
						break
					} else {
						ctx.Send("编号错误，请重试！")
					}
				}
			}
		}
	})
	zero.OnCommandGroup([]string{"查询订阅", "searchSubscription"}).Handle(func(ctx *zero.Ctx) {
		isGroup, userID := getCtxInfo(ctx)
		feeds, err := model.QueryFeed(userID, isGroup)
		if err != nil {
			log.Errorf("查询订阅出现错误：%v", err)
			ctx.Send("查询订阅出现未知错误，请联系管理员排查。")
			return
		}
		hasUrl, messageToSend := buildMessage(feeds, true)
		if !hasUrl {
			ctx.Send("您还没有任何订阅内容！")
		} else {
			ctx.Send(messageToSend)
		}
	})
	select {}
}

// 判断是否是group，获得userID
func getCtxInfo(ctx *zero.Ctx) (isGroup bool, userID int64) {
	if ctx.Event.GroupID != 0 {
		isGroup = true
		userID = ctx.Event.GroupID
	} else {
		isGroup = false
		userID = ctx.Event.UserID
	}
	return
}

func buildMessage(feeds []*model.Feed, isQuery bool) (hasFeed bool, message string) {
	var messages []string
	if len(feeds) == 0 {
		return
	}
	hasFeed = true
	for index, feed := range feeds {
		messages = append(messages, fmt.Sprintf("%d. %s", index+1, feed.Name))
	}
	if !isQuery {
		messages = append(messages, "请在十秒内输入您要取消的编号。")
	}
	return hasFeed, strings.Join(messages, "\n")
}

package main

import (
	"fmt"
	tgbot "github.com/yanzay/tbot"
	"gopkg.in/gomail.v2"
	"io/ioutil"
	"log"
	"os"
)

var kindlemails map[int64]string
var tempemail string = ""
var downloading chan string

func main() {
	token := os.Getenv("TELEGRAM_TOKEN")
	bot, err := tgbot.NewServer(token)
	kindlemails = make(map[int64]string)
	downloading = make(chan string)
	if err != nil {
		log.Fatal(err)
		return
	}

	bot.HandleFunc("/start", settingsHandle)
	bot.HandleFunc("/mail {text}", emailAddressHandle)
	bot.HandleFunc("Yes", replyYesHandle)
	bot.HandleFunc("No", replyNoHandle)
	bot.HandleFunc("Done", replyDoneHandler)
	bot.HandleFile(fileHandler)
	err = bot.ListenAndServe()
	log.Fatal(err)
}

func settingsHandle(message *tgbot.Message) {
	_, ok := kindlemails[message.ChatID]
	if !ok {
		message.Reply("请输入你的Kindle的邮箱,格式： /mail 邮箱")
	} else {
		message.Reply("请发给我你的书籍，大小不要超过20M，一次一本")
		buttons := [][]string{
			{"Done"},
		}
		message.ReplyKeyboard("发送你的图书，发送完毕点击Done", buttons)
	}
}

func emailAddressHandle(message *tgbot.Message) {
	username := message.ChatID
	tempemail = message.Vars["text"]
	_, isPresent := kindlemails[username]
	if !isPresent {
		kindlemails[username] = message.Vars["text"]
		message.Replyf("你的邮箱已设置为： %s", tempemail)
		tempemail = ""
		message.Reply("请发给我你的书籍，大小不要超过20M，一次一本")
		buttons := [][]string{
			{"Done"},
		}
		message.ReplyKeyboard("发送你的图书，发送完毕点击Done", buttons)
	} else {
		buttons := [][]string{
			{"Yes", "No"},
		}
		temp := fmt.Sprintf("你的名下有一个邮箱了，它是：%s， 你想要替换吗？\n yes or no", kindlemails[username])
		message.ReplyKeyboard(temp, buttons)
	}
}

func replyYesHandle(message *tgbot.Message) {
	kindlemails[message.ChatID] = tempemail
	message.Replyf("你的邮箱已设置为： %s", tempemail)
	message.Reply("请发给我你的书籍，大小不要超过20M，一次一本")
	buttons := [][]string{
		{"Done"},
	}
	message.ReplyKeyboard("发送你的图书，发送完毕点击Done", buttons)
}

func replyNoHandle(message *tgbot.Message) {
	message.Reply("好的，不会覆盖你以前的记录")
	message.Reply("请发给我你的书籍，大小不要超过20M，一次一本")
	buttons := [][]string{
		{"Done"},
	}
	message.ReplyKeyboard("发送你的图书，发送完毕点击Done", buttons)
}

func fileHandler(message *tgbot.Message) {
	err := message.Download("./uploads")
	if err != nil {
		message.Replyf("Error handling file: %q", err)
		return
	}
	downloading <- "done"
}

func replyDoneHandler(message *tgbot.Message) {
	<-downloading
	filesmap := make(map[int]string)
	files, _ := ioutil.ReadDir("./uploads")
	i := 0
	for _, file := range files {
		if file.IsDir() {
			continue
		} else {
			filesmap[i] = file.Name()
			i++
		}
	}
	_, ok := kindlemails[message.ChatID]
	if kindlemails[message.ChatID] == "" || !ok {
		message.Reply("你还没有设置邮箱，先设置邮箱吧！")
		return
	}
	message.Reply("开始发送邮件。。。")
	m := gomail.NewMessage()
	m.SetHeader("From", "your email")
	m.SetHeader("To", kindlemails[message.ChatID])
	m.SetHeader("Subject", "Arnold's Kindle bot 发送的邮件")
	m.SetBody("text/html", "Hello!")
	for i = 0; i < len(filesmap); i++ {
		m.Attach(fmt.Sprintf("./uploads/%s", filesmap[i]))
	}
	d := gomail.NewDialer("smtp.163.com", 465, "your email", "password")
	if err := d.DialAndSend(m); err != nil {
		panic(err)
	}
	message.Reply("发送成功，打开Kindle等待推送吧！")
	for i = 0; i < len(filesmap); i++ {
		err := os.Remove(fmt.Sprintf("./uploads/%s", filesmap[i]))
		if err != nil {
			fmt.Println("all deleted")
		}
	}
}

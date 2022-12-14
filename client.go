package main

import (
	"context"
	"flag"
	"log"
	"os"
	"os/signal"
	"syscall"

	pb "go-ping-discord/proto"

	"github.com/mileusna/crontab"
	"github.com/spf13/viper"
	"google.golang.org/grpc"
)

var (
	cronSchedule string
	subject      string
	content      string
)

func init() {
	viper.SetConfigName("config")
	viper.AddConfigPath("/conf/")
	viper.AddConfigPath("$HOME/.conf")
	viper.AddConfigPath(".")

	err := viper.ReadInConfig()
	if err != nil {
		log.Fatalf("error: couldn't read config file: %s", err)
	}

	flag.StringVar(&cronSchedule, "c", viper.GetString("cron"), "Runs the function at the interval specified by the cron")
	flag.StringVar(&subject, "subject", viper.GetString("subject"), "Message subject")
	flag.StringVar(&content, "msg", viper.GetString("msg"), "REST Port")

	flag.Parse()
}

func main() {
	ctab := crontab.New()

	conn, err := grpc.Dial("discord-bot:4444", grpc.WithInsecure())
	if err != nil {
		log.Fatalf("could not connect to server %v", err)
	}
	defer conn.Close()

	client := pb.NewDiscordMessageClient(conn)

	ctab.AddJob(cronSchedule, sendMsg, client, subject, content)

	sc := make(chan os.Signal, 1)
	signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM, os.Interrupt)
	<-sc
}

func sendMsg(client pb.DiscordMessageClient, subject, msg string) {
	resp, err := client.SendChanMessage(context.Background(), &pb.MessageChannel{
		Subject: subject,
		Content: msg,
	})
	if err != nil {
		log.Fatal(err)
	}
	log.Print(resp)
}

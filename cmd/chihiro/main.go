package main

import (
	"encoding/json"
	"log"
	"os"
	"strconv"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/sqs"
	"github.com/bettercallshao/kut/pkg/cmd"
	"github.com/bettercallshao/kut/pkg/menu"

	"github.com/bettercallshao/chihiro/pkg/version"
)

const NAME = "chihiro"

type Body struct {
	Message string
}

func main() {
	log.SetPrefix("[chihiro] ")
	log.Printf("starting chihiro %s ...\n", version.Version)

	sess := session.Must(session.NewSessionWithOptions(session.Options{
		SharedConfigState: session.SharedConfigEnable,
	}))
	svc := sqs.New(sess)
	queue := NAME
	urlResult, err := svc.GetQueueUrl(&sqs.GetQueueUrlInput{
		QueueName: &queue,
	})
	if err != nil {
		log.Printf("cannot find queue %s\n", NAME)
		log.Println(err)
		os.Exit(-1)
	}
	queueURL := urlResult.QueueUrl

	for {
		msgResult, err := svc.ReceiveMessage(&sqs.ReceiveMessageInput{
			AttributeNames: []*string{
				aws.String(sqs.MessageSystemAttributeNameSentTimestamp),
			},
			MessageAttributeNames: []*string{
				aws.String(sqs.QueueAttributeNameAll),
			},
			QueueUrl:            queueURL,
			MaxNumberOfMessages: aws.Int64(1),
			WaitTimeSeconds:     aws.Int64(20),
		})
		if err != nil {
			log.Printf("cannot receive message\n")
			log.Println(err)
			os.Exit(-1)
		}

		if len(msgResult.Messages) == 0 {
			continue
		}

		handleMsg(msgResult)

		_, err = svc.DeleteMessage(&sqs.DeleteMessageInput{
			QueueUrl:      queueURL,
			ReceiptHandle: msgResult.Messages[0].ReceiptHandle,
		})
		if err != nil {
			log.Printf("cannot delete message\n")
			log.Println(err)
			os.Exit(-1)
		}

		time.Sleep(100 * time.Millisecond)
	}
}

func handleMsg(msgResult *sqs.ReceiveMessageOutput) {

	loaded, err := menu.Load(NAME)
	if err != nil {
		log.Printf("cannot find menu %s\n", NAME)
		log.Println(err)
		return
	}

	var body Body
	json.Unmarshal([]byte(*msgResult.Messages[0].Body), &body)
	idx, err := strconv.Atoi(body.Message)
	if err != nil {
		log.Printf("cannot convert %s to index\n", body.Message)
		log.Println(err)
		return
	}

	if idx >= len(loaded.Actions) {
		log.Printf("cannot find action at %d\n", idx)
		return
	}

	input, err := menu.Render(loaded.Actions[idx])
	if err != nil {
		log.Printf("cannot render action at %d\n", idx)
		log.Println(err)
		return
	}

	log.Printf("running: %s\n", input)
	cmd.Run(input, nil)
	log.Printf("done")
}

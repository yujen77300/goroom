package models

import(
	"fmt"
	"log"
	"context"

	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/spf13/viper"
)

func ConnectToAWS() (string, string, *s3.Client) {
	viper.SetConfigName("config")
	viper.AddConfigPath(".")
	err := viper.ReadInConfig()
	if err != nil {
		log.Fatalln(err)
	}

	AWS_REGION := viper.GetString("AWS_REGION")
	AWS_ACCESS_KEY := viper.GetString("AWS_ACCESS_KEY_ID")
	AWS_SECRET_ACCESS_KEY := viper.GetString("AWS_SECRET_ACCESS_KEY")
	AWS_BUCKET_NAME := viper.GetString("AWS_BUCKET_NAME")

	staticProvider := credentials.NewStaticCredentialsProvider(
		AWS_ACCESS_KEY,
		AWS_SECRET_ACCESS_KEY,
		"",
	)

	cfg, err := config.LoadDefaultConfig(context.TODO(), config.WithRegion(AWS_REGION), config.WithCredentialsProvider(staticProvider))
	if err != nil {
		log.Fatalln(err)
	}

	client := s3.NewFromConfig(cfg)
	fmt.Printf("Datatype of client : %T\n", client)

	return AWS_REGION, AWS_BUCKET_NAME, client

}
package main

// s3 에 그냥 데이터 업로드 => 아테나 쿼리에서 테이블을 만드는데 s3 에서 데이터를 가져와서 컬럼을 만들 수 있다.
// 퀵사이트에서 s3 를 바로 볼 수 도 있음.

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
	"os"
	"strconv"
	"time"

	"github.com/DataDog/datadog-api-client-go/v2/api/datadog"
	"github.com/DataDog/datadog-api-client-go/v2/api/datadogV1"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
)

func FileUploadS3(responseContent []byte) error {
	sess, _ := session.NewSession(&aws.Config{
		Region: aws.String("ap-northeast-2"),
	})

	uploader := s3manager.NewUploader(sess)

	today := time.Now().Format("2006-01-02")
	s := strings.Split(today, "-")
	y := s[0]
	m := s[1]
	d := s[2]

	sloHistoryReader := bytes.NewReader(responseContent)

	_, err := uploader.Upload(&s3manager.UploadInput{
		Bucket: aws.String(os.Getenv("BUCKET")),
		Key:    aws.String(fmt.Sprintf("%s/%s/%s/%s/slo_history.json", os.Getenv("KEY"), y, m, d)),
		Body:   sloHistoryReader,
	})

	if err != nil {
		return err
	}

	return nil
}

func GetSloHistory(sloId string, day int) error {
	ctx := datadog.NewDefaultContext(context.Background())
	configuration := datadog.NewConfiguration()
	apiClient := datadog.NewAPIClient(configuration)
	api := datadogV1.NewServiceLevelObjectivesApi(apiClient)
	resp, r, err := api.GetSLOHistory(ctx, sloId, time.Now().AddDate(0, 0, day).Unix(), time.Now().Unix(), *datadogV1.NewGetSLOHistoryOptionalParameters())

	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `ServiceLevelObjectivesApi.GetSLOHistory`: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
		return err
	}

	responseContent, _ := json.MarshalIndent(resp, "", "  ")

	fmt.Println(string(responseContent))

	err = FileUploadS3(responseContent)
	if err != nil {
		return err
	}

	fmt.Println("파일 업로드 완료.")

	return nil
}

func main() {
	// SLO ID 환경 변수 처리
	sloId := os.Getenv("SLO_ID")
	day, _ := strconv.Atoi(os.Getenv("DAY"))
	err := GetSloHistory(sloId, day)
	if err != nil {
		fmt.Println(err)
	}
}

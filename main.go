package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
	"github.com/aws/aws-sdk-go/service/ses"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/DataDog/datadog-api-client-go/v2/api/datadog"
	"github.com/DataDog/datadog-api-client-go/v2/api/datadogV1"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
)

const (
	Sender  = "noreply@musinsa.com"
	CharSet = "UTF-8"
)

type SloHistory struct {
	toTs     int64
	sliValue float64
}

func getMonthRange(year int, month int) (time.Time, time.Time) {
	firstDay := time.Date(year, time.Month(month), 1, 0, 0, 0, 0, time.UTC)
	lastDay := time.Date(year, time.Month(month+1), 1, 0, 0, 0, 0, time.UTC)
	return firstDay, lastDay
}

func toMilliseconds(dt time.Time) int64 {
	return dt.UnixNano() / int64(time.Millisecond)
}

func getMilliseconds(firstDay, lastDay time.Time) (int64, int64) {
	firstDayMillis := toMilliseconds(firstDay)
	lastDayMillis := toMilliseconds(lastDay)
	return firstDayMillis, lastDayMillis
}

func GetYearMonthDay() (string, string, string) {
	today := time.Now().Format("2006-01-02")
	s := strings.Split(today, "-")
	y := s[0]
	m := s[1]
	d := s[2]

	return y, m, d
}

func SendMail() error {
	y, m, _ := GetYearMonthDay()
	iY, _ := strconv.Atoi(y)
	iM, _ := strconv.Atoi(m)
	firstDay, lastDay := getMonthRange(iY, iM)
	monthNumber := int(firstDay.Month())
	firstDayMillisecs, lastDayMillisecs := getMilliseconds(firstDay, lastDay)
	dashboardId := os.Getenv("DASHBOARD_ID")
	dashBoardURL := fmt.Sprintf("https://app.datadoghq.com/dashboard/%v/slo-slo?from_ts=%v&to_ts=%v&live=true", dashboardId, firstDayMillisecs, lastDayMillisecs)

	recipients := strings.Split(os.Getenv("RECIPIENT"), ",")
	subject := fmt.Sprintf("%v월 무신사 플랫폼서비스 SLO 요약 리포트", monthNumber)
	htmlBody := fmt.Sprintf("<h1>%v월 무신사 플랫폼 서비스 SLO 요약 리포트</h1>", monthNumber) + fmt.Sprintf("<h2><a href=%v>무신사 플랫폼 서비스 SLO 요약 대시보드입니다.</a></h2>", dashBoardURL)

	sess, _ := session.NewSession(&aws.Config{
		Region: aws.String("ap-northeast-2"),
	})

	svc := ses.New(sess)

	input := &ses.SendEmailInput{
		Destination: &ses.Destination{
			CcAddresses: []*string{},
			ToAddresses: aws.StringSlice(recipients),
		},
		Message: &ses.Message{
			Body: &ses.Body{
				Html: &ses.Content{
					Charset: aws.String(CharSet),
					Data:    aws.String(htmlBody),
				},
			},
			Subject: &ses.Content{
				Charset: aws.String(CharSet),
				Data:    aws.String(subject),
			},
		},
		Source: aws.String(Sender),
	}

	_, err := svc.SendEmail(input)
	if err != nil {
		return err
	}

	fmt.Println("send email success")

	return nil
}

func FileUploadS3(responseContent []byte) error {
	sess, _ := session.NewSession(&aws.Config{
		Region: aws.String("ap-northeast-2"),
	})

	uploader := s3manager.NewUploader(sess)

	y, m, d := GetYearMonthDay()

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

func GetSloHistory(sloId string, day int) (SloHistory, error) {
	var sloHistory SloHistory

	ctx := datadog.NewDefaultContext(context.Background())
	configuration := datadog.NewConfiguration()
	apiClient := datadog.NewAPIClient(configuration)
	api := datadogV1.NewServiceLevelObjectivesApi(apiClient)
	resp, r, err := api.GetSLOHistory(ctx, sloId, time.Now().AddDate(0, 0, day).Unix(), time.Now().Unix(), *datadogV1.NewGetSLOHistoryOptionalParameters())

	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `ServiceLevelObjectivesApi.GetSLOHistory`: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
		return sloHistory, err
	}

	responseContent, _ := json.MarshalIndent(resp, "", "  ")

	fmt.Println(string(responseContent))

	err = FileUploadS3(responseContent)
	if err != nil {
		return sloHistory, err
	}

	fmt.Println("파일 업로드 완료.")

	//overall := data["overall"].(map[string]interface{})
	//sliValue := overall["sli_value"].(float64)

	//sloHistory.toTs = responseContent[""]

	return sloHistory, err
}

func main() {
	// SLO ID 환경 변수 처리
	sloId := os.Getenv("SLO_ID")
	day, _ := strconv.Atoi(os.Getenv("DAY"))

	if _, err := GetSloHistory(sloId, day); err != nil {
		fmt.Println(err)
	}

	if err := SendMail(); err != nil {
		fmt.Println(err)
	}
}

package main

import (
    "fmt"
    "os"
    "strconv"
    "time"

    "github.com/aws/aws-sdk-go/aws"
    "github.com/aws/aws-sdk-go/aws/session"
    "github.com/aws/aws-sdk-go/service/costexplorer"
    "github.com/bradfitz/slice"
    "github.com/olekukonko/tablewriter"
)

// getDates returns a DateInterval for the last week
func getDates() *costexplorer.DateInterval {
    now := time.Now()
    then := now.AddDate(0, 0, -14)
    dateRange := costexplorer.DateInterval{}
    dateRange.SetEnd(now.Format("2006-01-02"))
    dateRange.SetStart(then.Format("2006-01-02"))
    return &dateRange
}

// covert string to float to string for formatting
// func formatNumber(s string) string {
//  f, _ := strconv.ParseFloat(s, 64)
//  // fmt.Printf("%.2f", f)
//  f = f / 100
//  return fmt.Sprintf("%.2f", f)
// }

// generate the date headers for the table
func dateHeaders() []string {
    now := time.Now()
    dates := []string{"AWS Service"}
    for i := 7; i > 0; i-- {
        n := now.AddDate(0, 0, -i)
        dates = append(dates, n.Format("01-02"))
    }
    return dates
}

func main() {
    sess, err := session.NewSession(&aws.Config{
        Region: aws.String("us-east-1")},
    )
    if err != nil {
        fmt.Println(err.Error())
    }

    svc := costexplorer.New(sess)

    resp, _ := svc.GetCostAndUsage((&costexplorer.GetCostAndUsageInput{
        Metrics:     []*string{aws.String("BlendedCost")},
        TimePeriod:  getDates(),
        Granularity: aws.String("MONTHLY"),
        GroupBy: []*costexplorer.GroupDefinition{
            &costexplorer.GroupDefinition{
                Key:  aws.String("SERVICE"),
                Type: aws.String("DIMENSION"),
            },
        },
    }))

    if err != nil {
        fmt.Println(err)

    } else {
        // fmt.Println(resp)
    }

    table := tablewriter.NewWriter(os.Stdout)

    table.SetHeader(dateHeaders())
    // append(table)

    // sorted := sort.Sort(resp.ResultsByTime[0].Groups)
    var data [][]string
    for i, group := range resp.ResultsByTime {
        // fmt.Println(group)
        // fmt.Println(len(group.Groups))
        slice.Sort(group.Groups[:], func(i, j int) bool {
            a, _ := strconv.ParseFloat(*group.Groups[i].Metrics["BlendedCost"].Amount, 64)
            b, _ := strconv.ParseFloat(*group.Groups[j].Metrics["BlendedCost"].Amount, 64)
            return a > b
        })

        fmt.Println(len(group.Groups))
        for j, key := range group.Groups {
            // fmt.Println(j, key)
            fmt.Println(aws.StringValue(key.Metrics["BlendedCost"].Amount))
            dollas := aws.StringValue(key.Metrics["BlendedCost"].Amount)
            // fmt.Println(dollas)
            // fmt.Println(len(data))
            if dollas != "" {
                if i == 0 {

                    data = append(data, []string{aws.StringValue(key.Keys[0]), dollas})
                } else {
                    if j < len(data) {
                        data[j] = append(data[j], dollas)
                    }
                }
            }
        }
    }

    // fmt.Println(data)

    table.AppendBulk(data)
    table.Render()
}
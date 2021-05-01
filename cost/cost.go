package cost

import (
	"fmt"
	"strconv"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/costexplorer"
	"github.com/psanford/aws-buddy/config"
	"github.com/spf13/cobra"
)

var (
	jsonOutput bool
	daysFlag   int
)

func Command() *cobra.Command {
	cmd := cobra.Command{
		Use:   "cost",
		Short: "COST Commands",
	}

	cmd.AddCommand(costDailyComparison())
	return &cmd
}

func costDailyComparison() *cobra.Command {
	cmd := cobra.Command{
		Use:   "daily",
		Short: "Show daily costs",
		Run:   dailyCostComparisonAction,
	}

	cmd.Flags().IntVarP(&daysFlag, "days", "", 14, "Number of days to fetch")

	return &cmd
}

func dailyCostComparisonAction(cmd *cobra.Command, args []string) {
	svc := costexplorer.New(config.Session())

	today := time.Now()
	start := today.AddDate(0, 0, -1*daysFlag)

	req := costexplorer.GetCostAndUsageInput{
		Granularity: aws.String("DAILY"),
		Metrics: []*string{
			aws.String("NetAmortizedCost"),
		},
		TimePeriod: &costexplorer.DateInterval{
			Start: aws.String(start.Format("2006-01-02")),
			End:   aws.String(today.Format("2006-01-02")),
		},
	}

	type dailyCost struct {
		date string
		amt  float64
	}

	var (
		maxCost float64
		costs   []dailyCost
	)

	for moreData := true; moreData; {
		output, err := svc.GetCostAndUsage(&req)
		if err != nil {
			panic(err)
		}

		if output.NextPageToken != nil {
			req.NextPageToken = output.NextPageToken
		} else {
			moreData = false
		}

		for _, result := range output.ResultsByTime {
			amtStr := result.Total["NetAmortizedCost"].Amount
			amt, err := strconv.ParseFloat(*amtStr, 64)
			if err != nil {
				panic(err)
			}

			if amt > maxCost {
				maxCost = amt
			}

			costs = append(costs, dailyCost{
				date: *result.TimePeriod.Start,
				amt:  amt,
			})
		}
	}

	starWidth := maxCost / 70.0

	for _, cost := range costs {

		stars := cost.amt / starWidth

		fmt.Printf("%s $%d ", cost.date, int(cost.amt))
		for i := 0.0; i < stars; i++ {
			fmt.Print("*")
		}
		fmt.Println()
	}
}

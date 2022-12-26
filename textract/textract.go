package textract

import (
	"bytes"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"path/filepath"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/textract"
	"github.com/psanford/aws-buddy/config"
	"github.com/spf13/cobra"
	"golang.org/x/tools/txtar"
)

func Command() *cobra.Command {
	cmd := cobra.Command{
		Use:   "textract",
		Short: "Textract commands",
	}

	cmd.AddCommand(analyzeDocCommand())

	return &cmd
}

var bucket string

func analyzeDocCommand() *cobra.Command {
	cmd := cobra.Command{
		Use:   "analyze <file>",
		Short: "Analyze document",
		Run:   analyzeDocAction,
	}

	cmd.Flags().StringVarP(&bucket, "bucket", "b", "", "S3 bucket used for storage")

	return &cmd
}

func analyzeDocAction(cmd *cobra.Command, args []string) {
	if len(args) < 1 {
		log.Fatalf("Usage: analyze <file>")
	}

	if bucket == "" {
		log.Fatalf("--bucket is required")
	}

	var aggregatedResult textract.AnalyzeDocumentOutput

	content, err := ioutil.ReadFile(args[0])
	if err != nil {
		log.Fatal(err)
	}

	s3svc := s3.New(config.Session())
	srcPath := fmt.Sprintf("input/%s", filepath.Base(args[0]))
	resultPath := fmt.Sprintf("output/%s", filepath.Base(args[0]))
	_, err = s3svc.PutObject(&s3.PutObjectInput{
		Bucket: &bucket,
		Key:    &srcPath,
		Body:   bytes.NewReader(content),
	})
	if err != nil {
		log.Fatalf("put object err: %s", err)
	}

	svc := textract.New(config.Session())

	startResult, err := svc.StartDocumentAnalysis(&textract.StartDocumentAnalysisInput{
		DocumentLocation: &textract.DocumentLocation{
			S3Object: &textract.S3Object{
				Bucket: &bucket,
				Name:   &srcPath,
			},
		},
		OutputConfig: &textract.OutputConfig{
			S3Bucket: &bucket,
			S3Prefix: &resultPath,
		},
		FeatureTypes: aws.StringSlice([]string{"TABLES"}),
	})

	if err != nil {
		log.Fatalf("start analysis err: %s", err)
	}

	log.Printf("job: %s started", *startResult.JobId)
	var lastStatus string
	for i := 0; i < 60; i++ {
		time.Sleep(5 * time.Second)

		result, err := svc.GetDocumentAnalysis(&textract.GetDocumentAnalysisInput{
			JobId:      startResult.JobId,
			MaxResults: aws.Int64(5),
		})
		if err != nil {
			log.Fatalf("GetDocumentAnalysis err: %s", err)
		}
		lastStatus = *result.JobStatus
		if lastStatus == textract.JobStatusSucceeded || lastStatus == textract.JobStatusFailed {
			break
		}
	}

	if lastStatus != textract.JobStatusSucceeded {
		log.Fatalf("Waiting for job timed out with status: %s", lastStatus)
	}

	var nextToken *string
	for {
		result, err := svc.GetDocumentAnalysis(&textract.GetDocumentAnalysisInput{
			JobId:     startResult.JobId,
			NextToken: nextToken,
		})
		if err != nil {
			log.Fatalf("GetDocumentAnalysis err: %s", err)
		}
		aggregatedResult.DocumentMetadata = result.DocumentMetadata
		for _, b := range result.Blocks {
			bb := *b
			aggregatedResult.Blocks = append(aggregatedResult.Blocks, &bb)
		}
		if result.NextToken != nil && *result.NextToken != "" {
			nextToken = result.NextToken
		} else {
			break
		}
	}

	out, err := json.Marshal(aggregatedResult)
	if err != nil {
		log.Fatal(err)
	}

	f, err := ioutil.TempFile("", "textract")
	if err != nil {
		log.Printf("err: %s", err)
	} else {
		f.Write(out)
		f.Close()
	}

	fmt.Printf("wrote: %s\n", f.Name())

	tableBlocks := make([]textract.Block, 0, 32)
	blocksByID := make(map[string]textract.Block)
	for _, block := range aggregatedResult.Blocks {
		blocksByID[*block.Id] = *block
		if *block.BlockType == "TABLE" {
			tableBlocks = append(tableBlocks, *block)
		}
	}

	w := worker{
		blocksByID: blocksByID,
	}

	ar := txtar.Archive{}

	for idx, table := range tableBlocks {
		_, data := w.generateTable(table)
		f := txtar.File{
			Name: fmt.Sprintf("%d.csv", idx),
			Data: data,
		}
		ar.Files = append(ar.Files, f)
	}

	f2, err := ioutil.TempFile("", "textract.tables")
	if err != nil {
		log.Printf("err: %s", err)
	} else {
		fmt.Fprintf(f2, "%s\n", txtar.Format(&ar))
		f2.Close()
	}

	fmt.Printf("wrote: %s\n", f2.Name())
}

type worker struct {
	blocksByID map[string]textract.Block
}

func (w *worker) generateTable(tableBlock textract.Block) ([][]string, []byte) {
	table := w.getRowsColumnsMap(tableBlock)
	var buf bytes.Buffer
	cw := csv.NewWriter(&buf)
	defer cw.Flush()

	var rows [][]string

	for r := int64(1); r <= table.MaxRow; r++ {
		row := make([]string, 0, table.MaxCol)
		for c := int64(1); c <= table.MaxCol; c++ {
			coord := Coord{r, c}
			cell := table.cells[coord]
			row = append(row, cell)
		}
		cw.Write(row)
		rows = append(rows, row)
	}
	cw.Flush()
	return rows, buf.Bytes()
}

type Coord struct {
	RowIdx int64
	ColIdx int64
}

type Table struct {
	cells  map[Coord]string
	MaxRow int64
	MaxCol int64
}

func (w *worker) getRowsColumnsMap(tableBlock textract.Block) *Table {

	table := Table{
		cells: make(map[Coord]string),
	}
	for _, relationship := range tableBlock.Relationships {
		if *relationship.Type == "CHILD" {
			for _, childID := range relationship.Ids {
				cell := w.blocksByID[*childID]
				if *cell.BlockType == "CELL" {
					coord := Coord{
						RowIdx: *cell.RowIndex,
						ColIdx: *cell.ColumnIndex,
					}

					if coord.RowIdx > table.MaxRow {
						table.MaxRow = coord.RowIdx
					}
					if coord.ColIdx > table.MaxCol {
						table.MaxCol = coord.ColIdx
					}
					table.cells[coord] = w.getText(cell)

				}
			}
		}
	}
	return &table
}

func (w *worker) getText(cell textract.Block) string {
	var textBuilder strings.Builder
	for _, relationship := range cell.Relationships {
		if *relationship.Type == "CHILD" {
			for _, childID := range relationship.Ids {
				word := w.blocksByID[*childID]
				if *word.BlockType == "WORD" {
					if textBuilder.Len() > 0 {
						textBuilder.Write([]byte(" "))
					}
					textBuilder.Write([]byte(*word.Text))
				}
			}
		}
	}
	return textBuilder.String()
}

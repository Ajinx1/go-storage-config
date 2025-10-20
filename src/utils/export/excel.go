package export

import (
	"fmt"
	"reflect"
	"strings"
	"time"

	"github.com/xuri/excelize/v2"
)

func ExportToExcel(data interface{}, sheetName, fileName string) ([]byte, string, string, error) {
	f := excelize.NewFile()
	f.SetSheetName("Sheet1", sheetName)

	dataValue := reflect.ValueOf(data)
	if dataValue.Kind() != reflect.Slice {
		return nil, "", "", fmt.Errorf("input data must be a slice")
	}
	if dataValue.Len() == 0 {
		return nil, "", "", fmt.Errorf("input slice is empty")
	}

	itemType := dataValue.Type().Elem()
	var headers []string
	for i := 0; i < itemType.NumField(); i++ {
		field := itemType.Field(i)

		jsonTag := field.Tag.Get("json")
		if jsonTag != "" && jsonTag != "-" {
			header := strings.Split(jsonTag, ",")[0]
			headers = append(headers, header)
		} else {
			headers = append(headers, field.Name)
		}
	}

	for i, header := range headers {
		cell := fmt.Sprintf("%c1", 'A'+i)
		f.SetCellValue(sheetName, cell, header)
	}

	style, err := f.NewStyle(&excelize.Style{
		Font: &excelize.Font{Bold: true, Size: 12, Color: "FFFFFF"},
		Fill: excelize.Fill{Type: "pattern", Color: []string{"4F81BD"}, Pattern: 1},
	})
	if err != nil {
		return nil, "", "", err
	}
	f.SetCellStyle(sheetName, "A1", fmt.Sprintf("%c1", 'A'+len(headers)-1), style)

	for rowIdx := 0; rowIdx < dataValue.Len(); rowIdx++ {
		row := rowIdx + 2
		item := dataValue.Index(rowIdx)
		for colIdx := 0; colIdx < itemType.NumField(); colIdx++ {
			col := fmt.Sprintf("%c%d", 'A'+colIdx, row)
			fieldValue := item.Field(colIdx)

			if fieldValue.Kind() == reflect.Ptr {
				if fieldValue.IsNil() {
					f.SetCellValue(sheetName, col, "")
					continue
				}
				fieldValue = fieldValue.Elem()
			}

			switch fieldValue.Kind() {
			case reflect.String:
				f.SetCellValue(sheetName, col, fieldValue.String())
			case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
				f.SetCellValue(sheetName, col, fieldValue.Int())
			case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
				f.SetCellValue(sheetName, col, fieldValue.Uint())
			case reflect.Float32, reflect.Float64:
				f.SetCellValue(sheetName, col, fieldValue.Float())
			case reflect.Bool:
				f.SetCellValue(sheetName, col, fieldValue.Bool())
			case reflect.Struct:
				if t, ok := fieldValue.Interface().(time.Time); ok {
					f.SetCellValue(sheetName, col, t.Format("2006-01-02 15:04:05"))
				} else {
					f.SetCellValue(sheetName, col, fmt.Sprintf("%v", fieldValue.Interface()))
				}
			default:
				f.SetCellValue(sheetName, col, fmt.Sprintf("%v", fieldValue.Interface()))
			}
		}
	}

	for i := 0; i < len(headers); i++ {
		col := fmt.Sprintf("%c", 'A'+i)
		f.SetColWidth(sheetName, col, col, 20)
	}

	dataBuffer, err := f.WriteToBuffer()
	if err != nil {
		return nil, "", "", err
	}

	return dataBuffer.Bytes(), fileName, "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet", nil
}

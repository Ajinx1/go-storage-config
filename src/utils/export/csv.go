package export

import (
	"bytes"
	"encoding/csv"
	"fmt"
	"reflect"
	"strings"
	"time"
)

func ExportToCSV(data interface{}, fileName string) ([]byte, string, string, error) {
	b := &bytes.Buffer{}
	writer := csv.NewWriter(b)

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

	if err := writer.Write(headers); err != nil {
		return nil, "", "", err
	}

	for rowIdx := 0; rowIdx < dataValue.Len(); rowIdx++ {
		item := dataValue.Index(rowIdx)
		var row []string

		for colIdx := 0; colIdx < itemType.NumField(); colIdx++ {
			fieldValue := item.Field(colIdx)

			if fieldValue.Kind() == reflect.Ptr {
				if fieldValue.IsNil() {
					row = append(row, "")
					continue
				}
				fieldValue = fieldValue.Elem()
			}

			switch fieldValue.Kind() {
			case reflect.String:
				row = append(row, fieldValue.String())
			case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
				row = append(row, fmt.Sprintf("%d", fieldValue.Int()))
			case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
				row = append(row, fmt.Sprintf("%d", fieldValue.Uint()))
			case reflect.Float32, reflect.Float64:
				row = append(row, fmt.Sprintf("%.2f", fieldValue.Float()))
			case reflect.Bool:
				row = append(row, fmt.Sprintf("%t", fieldValue.Bool()))
			case reflect.Struct:
				if t, ok := fieldValue.Interface().(time.Time); ok {
					row = append(row, t.Format("2006-01-02 15:04:05"))
				} else {
					row = append(row, fmt.Sprintf("%v", fieldValue.Interface()))
				}
			default:
				row = append(row, fmt.Sprintf("%v", fieldValue.Interface()))
			}
		}

		if err := writer.Write(row); err != nil {
			return nil, "", "", err
		}
	}

	writer.Flush()
	if err := writer.Error(); err != nil {
		return nil, "", "", err
	}

	return b.Bytes(), fileName, "text/csv", nil
}

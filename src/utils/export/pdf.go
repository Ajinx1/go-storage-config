package export

// func ExportToPDF(data interface{}, fileName, pdfHeader string) ([]byte, string, string, error) {

// 	dataValue := reflect.ValueOf(data)
// 	if dataValue.Kind() != reflect.Slice {
// 		return nil, "", "", fmt.Errorf("input data must be a slice")
// 	}
// 	if dataValue.Len() == 0 {
// 		return nil, "", "", fmt.Errorf("input slice is empty")
// 	}

// 	itemType := dataValue.Type().Elem()
// 	var headers []string
// 	for i := 0; i < itemType.NumField(); i++ {
// 		field := itemType.Field(i)

// 		jsonTag := field.Tag.Get("json")
// 		if jsonTag != "" && jsonTag != "-" {
// 			header := strings.Split(jsonTag, ",")[0]
// 			headers = append(headers, header)
// 		} else {
// 			headers = append(headers, field.Name)
// 		}
// 	}

// 	cfg := config.NewBuilder().
// 		WithPageSize(config.SizeA4).
// 		WithOrientation(config.Portrait).
// 		Build()
// 	m := core.NewMaroto(cfg)

// 	m.AddRow(12,
// 		text.NewCol(12, pdfHeader, props.Text{
// 			Style: fontstyle.Bold,
// 			Size:  14,
// 			Align: align.Center,
// 			Top:   2,
// 		}),
// 	)

// 	m.AddRow(5)

// 	colWidth := 12.0 / float64(len(headers))
// 	headerCols := make([]core.Col, len(headers))
// 	for i, header := range headers {
// 		headerCols[i] = col.New(int(colWidth)).Add(
// 			text.New(header, props.Text{
// 				Style:      fontstyle.Bold,
// 				Size:       10,
// 				Align:      align.Center,
// 				BorderType: border.Full,
// 				BackgroundColor: &props.Color{
// 					Red:   79,
// 					Green: 129,
// 					Blue:  189,
// 				},
// 				Color: &props.Color{
// 					Red:   255,
// 					Green: 255,
// 					Blue:  255,
// 				},
// 			}),
// 		)
// 	}
// 	m.AddRow(7, headerCols...)

// 	for rowIdx := 0; rowIdx < dataValue.Len(); rowIdx++ {
// 		item := dataValue.Index(rowIdx)
// 		dataCols := make([]core.Col, len(headers))
// 		for colIdx := 0; colIdx < itemType.NumField(); colIdx++ {
// 			fieldValue := item.Field(colIdx)
// 			var cellValue string

// 			switch fieldValue.Kind() {
// 			case reflect.String:
// 				cellValue = fieldValue.String()
// 			case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
// 				cellValue = fmt.Sprintf("%d", fieldValue.Int())
// 			case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
// 				cellValue = fmt.Sprintf("%d", fieldValue.Uint())
// 			case reflect.Float32, reflect.Float64:
// 				cellValue = fmt.Sprintf("%.2f", fieldValue.Float())
// 			case reflect.Bool:
// 				cellValue = fmt.Sprintf("%t", fieldValue.Bool())
// 			case reflect.Struct:
// 				if t, ok := fieldValue.Interface().(time.Time); ok {
// 					cellValue = t.Format("2006-01-02 15:04:05")
// 				} else {
// 					cellValue = fmt.Sprintf("%v", fieldValue.Interface())
// 				}
// 			default:
// 				cellValue = fmt.Sprintf("%v", fieldValue.Interface())
// 			}

// 			dataCols[colIdx] = col.New(int(colWidth)).Add(
// 				text.New(cellValue, props.Text{
// 					Size:       9,
// 					Align:      align.Left,
// 					BorderType: border.Full,
// 				}),
// 			)
// 		}
// 		m.AddRow(6, dataCols...)
// 	}

// 	buf := new(bytes.Buffer)
// 	err := m.Output(buf)
// 	if err != nil {
// 		return nil, "", "", err
// 	}

// 	return buf.Bytes(), fileName, "application/pdf", nil
// }

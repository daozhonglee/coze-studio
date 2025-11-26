/*
 * Copyright 2025 coze-dev Authors
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

// Package sheet 提供 Excel/CSV 文件解析和列类型推断功能
//
// 本包负责以下操作：
// - 从对象存储读取 Excel/CSV 文件
// - 解析文件内容提取表头和数据
// - 自动推断列的数据类型（数值/文本/日期/布尔）
// - 验证上传文件与表结构的兼容性
//
// 支持的文件格式：
// - xlsx: Excel 2007+ 格式（使用 excelize 库）
// - xls: Excel 97-2003 格式（使用 extrame/xls 库）
// - csv: 逗号分隔值文件
//
// 设计说明：
// 类型推断采用责任链模式，按优先级依次尝试：
// Number > Float > Boolean > Date > Text
// 如果同一列存在多种类型，则降级为 Text 类型。
package sheet

import (
	"bytes"
	"context"
	"encoding/csv"
	"fmt"
	"io"
	"math/rand"
	"os"
	"regexp"
	"runtime/debug"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/extrame/xls"
	"github.com/xuri/excelize/v2"

	"github.com/coze-dev/coze-studio/backend/api/model/data/knowledge"
	"github.com/coze-dev/coze-studio/backend/crossdomain/database/model"
	database "github.com/coze-dev/coze-studio/backend/crossdomain/database/model"
	"github.com/coze-dev/coze-studio/backend/domain/memory/database/entity"
	"github.com/coze-dev/coze-studio/backend/infra/storage"
	"github.com/coze-dev/coze-studio/backend/pkg/errorx"
	"github.com/coze-dev/coze-studio/backend/pkg/lang/ptr"
	"github.com/coze-dev/coze-studio/backend/pkg/logs"
	"github.com/coze-dev/coze-studio/backend/types/errno"
)

// CellTypeIdentifier 单元格类型识别器
//
// 用于责任链模式中的类型推断，按优先级尝试识别单元格值的类型。
type CellTypeIdentifier struct {
	// Priority 优先级，数值越大优先级越高
	Priority int64
	// IdentifyCellType 类型识别函数，成功返回类型指针，失败返回 nil
	IdentifyCellType func(cellValue string) *knowledge.ColumnType
	// TargetColumnType 目标列类型
	TargetColumnType knowledge.ColumnType
}

var (
	identifierChain    []CellTypeIdentifier
	initIdentifierOnce sync.Once
	dateTimePattern    string
)

// TosTableParser 对象存储表格解析器
//
// 从 TOS（对象存储）读取 Excel/CSV 文件并解析为结构化数据。
type TosTableParser struct {
	// UserID 操作用户 ID
	UserID int64
	// DocumentSource 文档来源类型
	DocumentSource model.DocumentSourceType
	// TosURI 对象存储中的文件路径
	TosURI string
	// TosServ 对象存储服务接口
	TosServ storage.Storage
}

// GetTableDataBySheetIDx 根据工作表索引获取表格数据
//
// 从 TOS 下载文件并解析指定工作表的内容。
//
// 参数：
//   - ctx: 上下文
//   - rMeta: 读取元数据，包含工作表索引、表头位置、读取方式等
//
// 返回值：
//   - TableReaderSheetData: 解析后的列定义和样本数据
//   - ExcelExtraInfo: Excel 文件的附加信息（工作表列表、文件大小等）
//   - error: 解析失败时的错误
func (t *TosTableParser) GetTableDataBySheetIDx(ctx context.Context, rMeta entity.TableReaderMeta) (*entity.TableReaderSheetData, *entity.ExcelExtraInfo, error) {
	meta, err := t.getLocalSheetMeta(ctx, rMeta.TosMaxLine)
	if err != nil {
		return nil, nil, err
	}

	defer func() {
		if meta == nil || meta.ExcelFile == nil {
			return
		}
		if err1 := meta.ExcelFile.Close(); err1 != nil {
			logs.CtxInfof(ctx, "[GetTableDataBySheetIdx] close excel file failed, err: %v", err1)
		}
	}()

	if len(meta.SheetsRowCount) == 0 {
		return nil, nil, errorx.New(errno.ErrMemoryDatabaseNoSheetFound)
	}

	if int(rMeta.SheetId) >= len(meta.SheetsRowCount) {
		return nil, nil, errorx.New(errno.ErrMemoryDatabaseSheetIndexExceed)
	}

	// get execl range: sheetIdx + headerIdx + startLineIdx
	if rMeta.StartLineIdx > int64(meta.SheetsRowCount[rMeta.SheetId]) {
		return nil, nil, errorx.New(errno.ErrMemoryDatabaseSheetIndexExceed)
	}
	extra := &entity.ExcelExtraInfo{
		ExtensionName: meta.ExtensionName,
		FileSize:      meta.FileSize,
		TosURI:        t.TosURI,
	}

	data, ocErr := t.getTableDataBySheetIdx(ctx, rMeta, meta, rMeta.SheetId)
	if ocErr != nil {
		return nil, nil, ocErr
	}
	for i := range meta.SheetsRowCount {
		extra.Sheets = append(extra.Sheets, &knowledge.DocTableSheet{
			ID:        int64(i),
			SheetName: meta.SheetsNameList[i],
			TotalRow:  int64(meta.SheetsRowCount[i]),
		})
	}
	return data, extra, ocErr
}

// getLocalSheetMeta 获取本地解析后的工作表元数据
//
// 从 TOS 下载文件并解析为本地可操作的数据结构。
// 支持 xlsx、xls、csv 三种格式。
func (t *TosTableParser) getLocalSheetMeta(ctx context.Context, maxLine int64) (*entity.LocalTableMeta, error) {
	documentExtension, object, ocErr := t.GetTosTableFile(ctx)
	if ocErr != nil {
		return nil, ocErr
	}

	reader := bytes.NewReader(object)
	if documentExtension == "csv" { // Processing csv files
		records, err := csv.NewReader(reader).ReadAll()
		if err != nil {
			return nil, err
		}

		if int64(len(records)) > maxLine {
			return nil, err
		}

		return &entity.LocalTableMeta{
			RawLines:       records,
			SheetsNameList: []string{"default"},
			SheetsRowCount: []int{len(records)},
			ExtensionName:  documentExtension,
			FileSize:       int64(len(object)),
		}, nil
	} else if documentExtension == "xls" {
		result, err := HandleTmpFile(ctx, reader, "xls", getXlsLocalSheetMetaWithTmpFileCallback, int64(len(object)))
		if err != nil {
			return nil, err
		}

		localTableMeta, ok := result.(*entity.LocalTableMeta)
		if !ok {
			return nil, err
		}

		return localTableMeta, nil
	}

	excelFile, err := excelize.OpenReader(reader)
	if err != nil {
		return nil, err
	}

	meta := &entity.LocalTableMeta{
		ExcelFile:     excelFile,
		ExtensionName: documentExtension,
		FileSize:      int64(len(object)),
	}

	for _, sheetName := range excelFile.GetSheetList() {
		total := 0
		rows, rowsErr := excelFile.Rows(sheetName)
		if rowsErr != nil {
			return nil, rowsErr
		}
		for rows.Next() {
			total++
			if int64(total) > maxLine {
				return nil, errorx.New(errno.ErrMemoryDatabaseSheetRowCountExceed, errorx.KVf("msg", "row count exceed limit when database importing excel, limit: %v", maxLine))
			}
		}
		meta.SheetsRowCount = append(meta.SheetsRowCount, total)
		meta.SheetsNameList = append(meta.SheetsNameList, sheetName)
	}

	return meta, nil
}

// GetTosTableFile 从对象存储下载表格文件
//
// 返回值：
//   - string: 文件扩展名（csv/xlsx）
//   - []byte: 文件内容
//   - error: 下载失败时的错误
func (t *TosTableParser) GetTosTableFile(ctx context.Context) (string, []byte, error) {
	names := strings.Split(t.TosURI, "/")
	objectName := strings.Split(names[len(names)-1], ".")
	documentExtension := objectName[len(objectName)-1]
	if documentExtension != "csv" && documentExtension != "xlsx" {
		return "", nil, errorx.New(errno.ErrMemoryDatabaseUnsupportedFileType)
	}

	object, err := t.TosServ.GetObject(ctx, t.TosURI)
	if err != nil {
		return "", nil, fmt.Errorf("get tos file failed:%+v", err)
	}

	if len(object) > 104857600 {
		return "", nil, errorx.New(errno.ErrMemoryDatabaseSheetSizeExceed)
	}

	return documentExtension, object, nil
}

func (t *TosTableParser) getTableDataBySheetIdx(ctx context.Context, rMeta entity.TableReaderMeta, localMeta *entity.LocalTableMeta, sheetIdx int64) (*entity.TableReaderSheetData, error) {
	if int(sheetIdx) >= len(localMeta.SheetsRowCount) {
		return nil, fmt.Errorf("start sheet index out of range")
	}

	res := &entity.TableReaderSheetData{
		Columns: make([]*knowledge.DocTableColumn, 0),
	}
	defer func() {
		if localMeta != nil && localMeta.ExcelFile != nil {
			if err := localMeta.ExcelFile.Close(); err != nil {
				logs.CtxErrorf(ctx, "[GetTableDataBySheetIdx] close excel file err: %v", err)
			}
		}
	}()

	sheetRowTotal := localMeta.SheetsRowCount[sheetIdx]
	headLine, sampleData, ocErr := t.parseSchemaInfo(ctx, localMeta, rMeta, int(sheetIdx), sheetRowTotal)
	if ocErr != nil {
		return nil, ocErr
	}

	if len(headLine) == 0 {
		return res, nil
	}

	for colIndex, cell := range headLine {
		res.Columns = append(res.Columns, &knowledge.DocTableColumn{
			ColumnName: cell,
			Sequence:   int64(colIndex),
		})
	}
	res.SampleData = sampleData

	return res, nil
}

func (t *TosTableParser) parseSchemaInfo(ctx context.Context, meta *entity.LocalTableMeta, rMeta entity.TableReaderMeta, sheetIdx, sheetRowTotal int) (header []string, sampleData [][]string, err error) {
	if len(meta.RawLines) != 0 {
		if len(meta.RawLines) == 0 {
			return []string{}, [][]string{}, nil
		}
		if rMeta.HeaderLineIdx >= int64(len(meta.RawLines)) {
			return nil, nil, fmt.Errorf("header idx exceed")
		}
		if rMeta.StartLineIdx > int64(len(meta.RawLines)) {
			return meta.RawLines[rMeta.HeaderLineIdx], [][]string{}, nil
		}

		switch rMeta.ReaderMethod {
		case database.TableReadDataMethodHead:
			end := int(rMeta.StartLineIdx + rMeta.ReadLineCnt)
			if end > len(meta.RawLines) {
				end = len(meta.RawLines)
			}
			return meta.RawLines[rMeta.HeaderLineIdx], meta.RawLines[rMeta.StartLineIdx:end], nil

		case database.TableReadDataMethodOnlyHeader:
			if len(meta.RawLines[rMeta.HeaderLineIdx]) >= 50 {
				return meta.RawLines[rMeta.HeaderLineIdx][:50], [][]string{}, nil
			}
			return meta.RawLines[rMeta.HeaderLineIdx], [][]string{}, nil

		case database.TableReadDataMethodAll:
			return meta.RawLines[rMeta.HeaderLineIdx], meta.RawLines[rMeta.StartLineIdx:], nil
		default:
			logs.CtxInfof(ctx, "[parseSchemaInfo] reader method is:%d", rMeta.ReaderMethod)
		}

		return nil, nil, fmt.Errorf("reader method is:%d", rMeta.ReaderMethod)
	}

	sampleData, header, err = t.GetExcelToParseTableRange(ctx, meta.ExcelFile, sheetIdx, int64(sheetRowTotal), rMeta)
	if err != nil {
		return nil, nil, err
	}

	return header, sampleData, nil
}

func (t *TosTableParser) GetExcelToParseTableRange(ctx context.Context, file *excelize.File, sheetIdx int, rowCount int64, rMeta entity.TableReaderMeta) ([][]string, []string, error) {
	selectFunc := func(idx int64) (isContinue, isBreak bool) {
		return idx < rMeta.StartLineIdx, false
	}
	if rMeta.ReaderMethod == database.TableReadDataMethodHead {
		selectFunc = func(idx int64) (isContinue, isBreak bool) {
			return idx < rMeta.StartLineIdx, idx-rMeta.StartLineIdx > rMeta.ReadLineCnt
		}
	}

	selectedExcelContent := make([][]string, 0)
	sheetName := file.GetSheetName(sheetIdx)
	rows, err := file.Rows(sheetName)
	if err != nil {
		return nil, nil, err
	}

	headLineContent := make([]string, 0)
	var (
		curRowIndex int64
		columnCount int
	)
	for rows.Next() {
		if curRowIndex < rMeta.HeaderLineIdx {
			curRowIndex++
			continue
		}

		if curRowIndex == rMeta.HeaderLineIdx {
			content, rowErr := rows.Columns()
			if rowErr != nil {
				return nil, nil, rowErr
			}
			if len(content) > 50 {
				content = content[:50]
			}
			headLineContent = content
			curRowIndex++
			columnCount = len(headLineContent)
			if rMeta.ReaderMethod == database.TableReadDataMethodOnlyHeader {
				return [][]string{}, headLineContent, nil
			}
			continue
		}

		isContinue, isBreak := selectFunc(curRowIndex)
		if isContinue {
			curRowIndex++
			continue
		}
		if isBreak {
			break
		}
		curRowIndex++

		columns, err := rows.Columns()
		if err != nil {
			continue
		}

		if len(columns) < columnCount {
			emptyStrs := make([]string, columnCount-len(columns))
			columns = append(columns, emptyStrs...)
		}
		if len(columns) > 50 {
			columns = columns[:50]
		}

		if len(columns) != 0 && getCharCount(columns) > 0 {
			selectedExcelContent = append(selectedExcelContent, columns)
		}
	}

	return selectedExcelContent, headLineContent, nil
}

func getCharCount(strs []string) int {
	var count int
	for _, str := range strs {
		count += len(str)
	}

	return count
}

type TmpFileCallback func(context.Context, string, ...interface{}) (interface{}, error)

func getTmpFile(extension string) string {
	return fmt.Sprintf("%s/%s", "/tmp", getTmpFileName(extension))
}

func getTmpFileName(extension string) string {
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	name := fmt.Sprintf("%s_%d_%d.%s", "dataset_", time.Now().UnixMilli(), r.Intn(10000), extension)
	return name
}

func HandleTmpFile(ctx context.Context, reader io.Reader, extension string, callback TmpFileCallback, args ...interface{}) (interface{}, error) {
	tmpFile := getTmpFile(extension)
	file, err := os.OpenFile(tmpFile, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0o755)
	if err != nil {
		return nil, err
	}
	if reader != nil {
		_, err = io.Copy(file, reader)
		if err != nil {
			return nil, err
		}
	}
	if err = file.Close(); err != nil {
		return nil, err
	}

	defer func() {
		if err = os.Remove(tmpFile); err != nil {
			logs.CtxErrorf(ctx, "[HandleTmpFile]Error removing tmpFile: %v", err)
			return
		}
	}()
	result, err := callback(ctx, tmpFile, args...)
	return result, err
}

func getXlsLocalSheetMetaWithTmpFileCallback(ctx context.Context, tmpFile string, args ...interface{}) (interface{}, error) {
	const prefix = "[getXlsLocalSheetMetaWithTmpFileCallback]"
	defer func() {
		if panicErr := recover(); panicErr != nil {
			s := string(debug.Stack())
			logs.CtxErrorf(ctx, "%s get recover, stack: %s", prefix, s)
			return
		}
	}()
	size, _ := args[0].(int64)
	file, err := os.Open(tmpFile)
	if err != nil {
		return nil, err
	}
	defer func() {
		if err = file.Close(); err != nil {
			logs.CtxErrorf(ctx, "%v close excel reader failed:%+v", prefix, err)
			return
		}
	}()
	xlFile, err := xls.OpenReader(file, "utf-8")
	if err != nil {
		return nil, err
	}

	sheet := xlFile.GetSheet(0)
	if sheet == nil {
		return nil, errorx.New(errno.ErrMemoryDatabaseXlsSheetEmpty)
	}

	records := make([][]string, 0)
	if sheet.MaxRow < 1 {
		return nil, errorx.New(errno.ErrMemoryDatabaseXlsSheetEmpty)
	}
	for i := 0; i <= int(sheet.MaxRow); i++ {
		row := sheet.Row(i)
		record := make([]string, row.LastCol()-row.FirstCol()+1)
		for j := row.FirstCol(); j <= row.LastCol(); j++ {
			record[j-row.FirstCol()] = row.Col(j)
		}
		records = append(records, record)
	}

	return &entity.LocalTableMeta{
		RawLines:       records,
		SheetsNameList: []string{"default"},
		SheetsRowCount: []int{len(records)},
		ExtensionName:  "xls",
		FileSize:       size,
	}, nil
}

// PredictColumnType 根据样本数据推断列类型
//
// 分析样本数据中每列的值，自动推断最合适的数据类型。
// 如果列中存在空值，会标记 ContainsEmptyValue 为 true。
//
// 参数：
//   - columns: 列定义列表（将被更新类型信息）
//   - sampleData: 样本数据
//   - sheetIdx: 工作表索引
//   - startLineIdx: 数据起始行索引
//
// 返回值：
//   - 更新后的列定义列表
//   - error: 推断失败时的错误
func (t *TosTableParser) PredictColumnType(columns []*knowledge.DocTableColumn, sampleData [][]string, sheetIdx, startLineIdx int64) ([]*knowledge.DocTableColumn, error) {
	if len(sampleData) == 0 {
		for _, column := range columns {
			column.ColumnType = knowledge.ColumnTypePtr(knowledge.ColumnType_Text) // bottom line display
			column.ContainsEmptyValue = ptr.Of(true)
		}
		return columns, nil
	}

	columnInfos, err := ParseDocTableColumnInfo(sampleData, len(columns))
	if err != nil {
		return nil, err
	}

	if len(columns) != len(columnInfos) {
		return nil, errorx.New(errno.ErrMemoryDatabaseColumnNotMatch, errorx.KVf("msg", "columnType length is not match with column count, column length:%d, predict column length:%d", len(columns), len(columnInfos)))
	}

	for i := range columns {
		columns[i].ContainsEmptyValue = &columnInfos[i].ContainsEmptyValue
		columns[i].ColumnType = &columnInfos[i].ColumnType
	}
	return columns, nil
}

// ParseDocTableColumnInfo 解析 Excel 内容并推断列类型信息
//
// 将行数据转置为列数据，然后对每列进行类型推断。
func ParseDocTableColumnInfo(selectedExcelContent [][]string, columnCount int) ([]entity.ColumnInfo, error) {
	columnList := transposeExcelContent(selectedExcelContent, columnCount)
	columInfoList := make([]entity.ColumnInfo, 0)
	for _, col := range columnList {
		columnType, containsEmptyValue := predictColumnType(col)
		columInfoList = append(columInfoList, entity.ColumnInfo{
			ColumnType:         columnType,
			ContainsEmptyValue: containsEmptyValue,
		})
	}

	return columInfoList, nil
}

// transposeExcelContent 转置 Excel 内容
//
// 将 [行][列] 格式转换为 [列][行] 格式，便于按列分析类型。
func transposeExcelContent(excelContent [][]string, columnCount int) [][]string {
	if len(excelContent) == 0 {
		return make([][]string, 0)
	}

	rowCount := len(excelContent)
	transposedExcelContent := make([][]string, columnCount)
	for i := range transposedExcelContent {
		transposedExcelContent[i] = make([]string, rowCount)
	}

	for i := range excelContent {
		for j := range excelContent[i] {
			if j >= columnCount {
				continue
			}
			transposedExcelContent[j][i] = excelContent[i][j]
		}
	}

	return transposedExcelContent
}

// predictColumnType 推断单列的数据类型
//
// 遍历列中的所有值，使用责任链模式识别类型。
// 如果发现不同类型的值，则降级为 Text 类型。
//
// 返回值：
//   - ColumnType: 推断出的列类型
//   - bool: 是否包含空值
func predictColumnType(columnContent []string) (knowledge.ColumnType, bool) {
	columnType := knowledge.ColumnType_Text
	containsEmptyValue := false
	for i, col := range columnContent {
		if col == "" {
			containsEmptyValue = true
			continue
		}

		cellType := GetCellType(col)
		if i == 0 {
			columnType = cellType
			continue
		}

		if GetColumnTypeCategory(columnType) != GetColumnTypeCategory(cellType) {
			return knowledge.ColumnType_Text, containsEmptyValue
		}

		if GetColumnTypePriority(cellType) < GetColumnTypePriority(columnType) {
			columnType = cellType
		}
	}
	return columnType, containsEmptyValue
}

// GetColumnTypePriority 获取列类型的优先级
//
// 用于类型降级判断：优先级低的类型可以被优先级高的类型覆盖。
func GetColumnTypePriority(columnType knowledge.ColumnType) int64 {
	for _, identifier := range identifierChain {
		if identifier.TargetColumnType == columnType {
			return identifier.Priority
		}
	}
	return 0
}

// GetColumnTypeCategory 获取列类型的分类
//
// 分类用于判断类型兼容性：Number 和 Float 属于数值类型，其他属于文本类型。
// 不同分类的类型不兼容，会降级为 Text。
func GetColumnTypeCategory(columnType knowledge.ColumnType) model.ColumnTypeCategory {
	if columnType == knowledge.ColumnType_Number || columnType == knowledge.ColumnType_Float {
		return model.ColumnTypeCategoryNumber
	}
	return model.ColumnTypeCategoryText
}

// GetCellType 获取单元格值的类型
//
// 使用责任链模式依次尝试识别：Number > Float > Boolean > Date > Text
func GetCellType(cellValue string) knowledge.ColumnType {
	InitIdentifier()
	cellTypeResult := knowledge.ColumnType_Text
	for _, identifier := range identifierChain {
		cellType := identifier.IdentifyCellType(cellValue)
		if cellType != nil {
			cellTypeResult = *cellType
			break
		}
	}
	return cellTypeResult
}

// IdentifyNumber 识别整数类型
func IdentifyNumber(cellValue string) *knowledge.ColumnType {
	_, err := strconv.ParseInt(cellValue, 10, 64)
	if err != nil {
		return nil
	}
	return knowledge.ColumnTypePtr(knowledge.ColumnType_Number)
}

// IdentifyFloat 识别浮点数类型
func IdentifyFloat(cellValue string) *knowledge.ColumnType {
	_, err := strconv.ParseFloat(cellValue, 64)
	if err != nil {
		return nil
	}
	return knowledge.ColumnTypePtr(knowledge.ColumnType_Float)
}

// IdentifyBoolean 识别布尔类型（true/false）
func IdentifyBoolean(cellValue string) *knowledge.ColumnType {
	lowerCellValue := strings.ToLower(cellValue)
	if lowerCellValue != "true" && lowerCellValue != "false" {
		return nil
	}
	return knowledge.ColumnTypePtr(knowledge.ColumnType_Boolean)
}

// IdentifyDate 识别日期类型
//
// 支持多种日期格式：
// - YYYY/MM/DD, YYYY-MM-DD, YYYY.MM.DD
// - YYYY年MM月DD日
// - 可选的时间部分
func IdentifyDate(cellValue string) *knowledge.ColumnType {
	matched, err := regexp.MatchString(dateTimePattern, cellValue)
	if err != nil || !matched {
		return nil
	}
	return knowledge.ColumnTypePtr(knowledge.ColumnType_Date)
}

// IdentifyText 识别文本类型（兜底类型）
func IdentifyText(cellValue string) *knowledge.ColumnType {
	return knowledge.ColumnTypePtr(knowledge.ColumnType_Text)
}

// InitIdentifier 初始化类型识别器链
//
// 使用 sync.Once 确保只初始化一次。
// 识别器按优先级排序：Number(5) > Float(4) > Boolean(3) > Date(2) > Text(1)
func InitIdentifier() {
	initIdentifierOnce.Do(func() {
		identifierChain = []CellTypeIdentifier{
			{
				Priority:         5,
				IdentifyCellType: IdentifyNumber,
				TargetColumnType: knowledge.ColumnType_Number,
			},
			{
				Priority:         4,
				IdentifyCellType: IdentifyFloat,
				TargetColumnType: knowledge.ColumnType_Float,
			},
			{
				Priority:         3,
				IdentifyCellType: IdentifyBoolean,
				TargetColumnType: knowledge.ColumnType_Boolean,
			},
			{
				Priority:         2,
				IdentifyCellType: IdentifyDate,
				TargetColumnType: knowledge.ColumnType_Date,
			},
			{
				Priority:         1,
				IdentifyCellType: IdentifyText,
				TargetColumnType: knowledge.ColumnType_Text,
			},
		}

		sort.SliceStable(identifierChain, func(i, j int) bool {
			return identifierChain[i].Priority > identifierChain[j].Priority
		})

		datePatternList := []string{
			"(\\d{2,4})\\/(0?[1-9]|1[0-2])\\/(0?[1-9]|[1-2]\\d|3[0-1])",
			"(\\d{2,4})-(0?[1-9]|1[0-2])-(0?[1-9]|[1-2]\\d|3[0-1])",
			"(\\d{2,4})\\.(0?[1-9]|1[0-2])\\.(0?[1-9]|[1-2]\\d|3[0-1])",
			"(\\d{2,4})年(0?[1-9]|1[0-2])月(0?[1-9]|[1-2]\\d|3[0-1])日",
		}

		timePatternList := []string{
			"([0-1]?\\d|2[0-3]):[0-5]\\d(:[0-5]\\d)?",
			"([0-1]?\\d|2[0-3])时[0-5]\\d分([0-5]\\d秒)?",
			"(0?\\d|1[0-2]):[0-5]\\d(:[0-5]\\d)? (am|AM|pm|PM)",
			"(0?\\d|1[0-2])时[0-5]\\d分([0-5]\\d秒)? (am|AM|pm|PM)",
		}

		dateTimePattern = getDateTimeRegExp(datePatternList, timePatternList)
	})
}

func getDateTimeRegExp(datePatternList []string, timePatterList []string) string {
	var (
		datePattern string
		timePattern string
	)

	for _, p := range datePatternList {
		p = fmt.Sprintf("(%s)", p)
		if datePattern == "" {
			datePattern = p
			continue
		}
		datePattern = fmt.Sprintf("%s|%s", datePattern, p)
	}

	for _, p := range timePatterList {
		p = fmt.Sprintf("(%s)", p)
		if timePattern == "" {
			timePattern = p
			continue
		}
		timePattern = fmt.Sprintf("%s|%s", timePattern, p)
	}

	return fmt.Sprintf("^(%s)( +(%s))?$", datePattern, timePattern)
}

// TransferPreviewData 转换预览数据格式
//
// 将二维字符串数组转换为以列索引为 key 的 map 格式，便于前端展示。
//
// 参数：
//   - ctx: 上下文
//   - columns: 列定义列表
//   - sampleData: 样本数据
//   - previewLine: 预览行数限制
//
// 返回值：
//   - 转换后的预览数据，格式为 []map[列索引]值
func (t *TosTableParser) TransferPreviewData(ctx context.Context, columns []*knowledge.DocTableColumn, sampleData [][]string, previewLine int) (previewData []map[int64]string, err error) {
	previewData = make([]map[int64]string, 0)

	for idx, line := range sampleData {
		if idx >= previewLine {
			break
		}
		lineValue := make(map[int64]string)
		if len(columns) != len(line) {
			logs.CtxWarnf(ctx, "[TransferPreviewData] sampleData's length:%d is not match with column count:%d", len(line), len(columns))
		}

		for i, value := range line {
			if i >= len(columns) {
				break
			}
			lineValue[int64(i)] = value
		}
		previewData = append(previewData, lineValue)
	}

	return previewData, nil
}

// CheckSheetIsValid 验证上传文件与表结构的兼容性
//
// 检查项包括：
// 1. 字段数量是否匹配
// 2. 字段数量是否超过限制（20列）
// 3. 数据行数是否超过限制（100000行）
// 4. 字段名称是否匹配
//
// 返回值：
//   - bool: 是否有效
//   - *string: 无效时的错误信息
func CheckSheetIsValid(fields []*model.FieldItem, parsedColumns []*knowledge.DocTableColumn, sheet *entity.ExcelExtraInfo) (bool, *string) {
	if len(fields) != len(parsedColumns) {
		return false, ptr.Of("field number not match")
	}
	if len(parsedColumns) > 20 {
		return false, ptr.Of("field number exceed 20")
	}
	if sheet.Sheets[0].TotalRow > 100000 {
		return false, ptr.Of("data rows exceed 100000")
	}

	equalColumns := map[string]int{}
	for _, field := range fields {
		for _, column := range parsedColumns {
			if field.Name == column.ColumnName {
				equalColumns[field.Name] = 1
				break
			}
		}
	}
	if len(equalColumns) != len(fields) {
		return false, ptr.Of("field name not match")
	}

	return true, nil
}

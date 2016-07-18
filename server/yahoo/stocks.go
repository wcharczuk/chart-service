package yahoo

import (
	"bufio"
	"errors"
	"fmt"
	"net/http"
	"reflect"
	"strconv"
	"strings"
	"sync"

	"github.com/blendlabs/go-exception"
	"github.com/blendlabs/go-util"
	"github.com/wcharczuk/chart-service/server/core"
)

var (
	_defaultStockInfoFormatLock sync.Mutex
	_defaultStockInfoFormat     string

	_indexMapLock sync.Mutex
	_indexMap     map[string]int

	_reverseIndexMapLock sync.Mutex
	_reverseIndexMap     map[int]string

	_fieldIndexMapLock sync.Mutex
	_fieldIndexMap     map[string]int
)

// StockInfo represents information about a stock.
type StockInfo struct {
	Format     string
	RawResults string

	Ticker string `csv:"s"`
	Name   string `csv:"n"`
	Notes  string `csv:"n4"`

	LastPrice     float64 `csv:"l1"`
	Change        float64 `csv:"c1"`
	ChangePercent string  `csv:"p2"`
	Volume        int64   `csv:"v"`

	LastTradeDate string `csv:"d1"`
	TradeDate     string `csv:"d2"`
	LastTradeTime string `csv:"t1"`

	PriceEarningsRatio float64 `csv:"r"`
}

// IsZero returns if the object has been set or not.
func (si StockInfo) IsZero() bool {
	return len(si.Name) == 0
}

// String returns a simple string representation of the object.
func (si StockInfo) String() string {
	return fmt.Sprintf("Ticker: %s Name: %s Last: %f", si.Ticker, si.Name, si.LastPrice)
}

// Parse reads a line into the stock info object.
func (si *StockInfo) Parse(line string) error {
	defer func() {
		if r := recover(); r != nil {
			fmt.Printf("Yahoo::StockInfo.Parse() panic: %v\n", r)
		}
	}()
	lookup := reverseIndexMap()
	fieldLookup := fieldIndexMap()
	parts := core.StringSplitQuoteAware(line, rune(','))
	if len(parts) != len(lookup) {
		fmt.Printf("%#v vs. %#v\n", parts, lookup)
		return errors.New("mismatched line components to lookup map, cannot continue")
	}

	siValue := util.ReflectValue(si)
	for index := 0; index < len(parts); index++ {
		rawValue := parts[index]
		if fieldName, hasField := lookup[index]; hasField {
			if fieldIndex, hasFieldIndex := fieldLookup[fieldName]; hasFieldIndex {
				field := siValue.Field(fieldIndex)
				finalValue, err := marshal(field.Kind(), rawValue)
				if err == nil && finalValue != nil {
					field.Set(reflect.ValueOf(finalValue))
				}
			}
		}
	}
	return nil
}

func marshal(fieldType reflect.Kind, rawValue string) (interface{}, error) {
	if rawValue == "N/A" {
		return nil, errors.New("field is unset")
	}
	switch fieldType {
	case reflect.String:
		return util.StripQuotes(rawValue), nil
	case reflect.Int:
		return strconv.Atoi(rawValue)
	case reflect.Int64:
		return strconv.ParseInt(rawValue, 10, 64)
	case reflect.Float32:
		return strconv.ParseFloat(rawValue, 32)
	case reflect.Float64:
		return strconv.ParseFloat(rawValue, 64)
	}
	return nil, errors.New("unknown field type; cannot marshal")
}

func stockInfoFormat() string {
	if len(_defaultStockInfoFormat) == 0 {
		_defaultStockInfoFormatLock.Lock()
		defer _defaultStockInfoFormatLock.Unlock()
		if len(_defaultStockInfoFormat) == 0 {
			_defaultStockInfoFormat = generateStockInfoFormat()
		}
	}
	return _defaultStockInfoFormat
}

func generateStockInfoFormat() string {
	var fields []string
	si := StockInfo{}
	siType := reflect.TypeOf(si)
	fieldCount := siType.NumField()
	for index := 0; index < fieldCount; index++ {
		csvTag := siType.Field(index).Tag.Get("csv")
		if len(csvTag) != 0 {
			fields = append(fields, csvTag)
		}
	}
	return strings.Join(fields, "")
}

func indexMap() map[string]int {
	if _indexMap == nil {
		_indexMapLock.Lock()
		defer _indexMapLock.Unlock()
		if _indexMap == nil {
			_indexMap = generateIndexMap()
		}
	}
	return _indexMap
}

func generateIndexMap() map[string]int {
	fields := map[string]int{}
	si := StockInfo{}
	siType := reflect.TypeOf(si)
	fieldCount := siType.NumField()
	fieldIndex := 0
	for index := 0; index < fieldCount; index++ {
		csvTag := siType.Field(index).Tag.Get("csv")
		if len(csvTag) != 0 {
			fields[csvTag] = fieldIndex
			fieldIndex++
		}
	}
	return fields
}

func reverseIndexMap() map[int]string {
	if _reverseIndexMap == nil {
		_reverseIndexMapLock.Lock()
		defer _reverseIndexMapLock.Unlock()
		if _reverseIndexMap == nil {
			_reverseIndexMap = generateReverseIndexMap()
		}
	}
	return _reverseIndexMap
}

func generateReverseIndexMap() map[int]string {
	fields := map[int]string{}

	lookup := indexMap()
	for key, value := range lookup {
		fields[value] = key
	}

	return fields
}

func fieldIndexMap() map[string]int {
	if _fieldIndexMap == nil {
		_fieldIndexMapLock.Lock()
		defer _fieldIndexMapLock.Unlock()
		if _fieldIndexMap == nil {
			_fieldIndexMap = generateFieldIndexMap()
		}
	}
	return _fieldIndexMap
}

func generateFieldIndexMap() map[string]int {
	fields := map[string]int{}
	si := StockInfo{}
	siType := reflect.TypeOf(si)
	fieldCount := siType.NumField()
	for index := 0; index < fieldCount; index++ {
		csvTag := siType.Field(index).Tag.Get("csv")
		if len(csvTag) != 0 {
			fields[csvTag] = index
		}
	}
	return fields
}

func generateInfoFormatFieldIndexMap() map[string]int {
	fields := map[string]int{}
	si := StockInfo{}
	siType := reflect.TypeOf(si)
	fieldCount := siType.NumField()
	for index := 0; index < fieldCount; index++ {
		csvTag := siType.Field(index).Tag.Get("csv")
		if len(csvTag) != 0 {
			fields[csvTag] = index

		}
	}
	return fields
}

// GetStockPrice returns stock price info from Yahoo for the given tickers.
func GetStockPrice(tickers []string) ([]StockInfo, error) {
	if len(tickers) == 0 {
		return []StockInfo{}, nil
	}

	rawResults, meta, resErr := core.NewRequest().AsGet().
		WithURL("http://download.finance.yahoo.com/d/quotes.csv").
		WithQueryString("s", strings.Join(tickers, "+")).
		WithQueryString("f", stockInfoFormat()).
		FetchStringWithMeta()

	if resErr != nil {
		return []StockInfo{}, resErr
	}

	if meta.StatusCode != http.StatusOK {
		return []StockInfo{}, exception.New("Non (200) response from pricing provider.")
	}
	var results []StockInfo

	scanner := bufio.NewScanner(strings.NewReader(rawResults))
	scanner.Split(bufio.ScanLines)

	var err error
	for scanner.Scan() {
		si := &StockInfo{}
		si.Format = stockInfoFormat()
		line := scanner.Text()
		si.RawResults = line

		err = si.Parse(line)
		if err != nil {
			return []StockInfo{}, err
		}
		if !si.IsZero() {
			results = append(results, *si)
		}
	}
	return results, nil
}

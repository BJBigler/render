package render

import (
	"encoding/json"
	"errors"
	"fmt"
	"html/template"
	"strings"
	"time"

	"golang.org/x/text/language"
	"golang.org/x/text/message"

	"cloud.google.com/go/datastore"
	"github.com/bjbigler/utils"
	"github.com/shopspring/decimal"
)

//Location returns New York location
func location() *time.Location {

	loc, err := time.LoadLocation("America/New_York")
	if err != nil {
		return time.UTC
	}

	return loc
}

//FullDateTimeET returns Eastern Time representation (2006-01-02 15:04)
func FullDateTimeET(val time.Time) string {
	return val.In(location()).Format("2006-01-02 15:04")
}

//ShortDateTime returns Eastern Time representation (Jan 01 15:04)
func ShortDateTime(val time.Time) string {
	return val.In(location()).Format("Jan 02 3:04pm")
}

//DateTimeFormal (January 2, 2006 at 3:04PM)
func DateTimeFormal(val time.Time) string {
	return val.In(location()).Format("January 02, 2006 at 3:04pm")
}

//TimeFormat returns AM/PM time format
func TimeFormat(val time.Time) string {
	return val.In(location()).Format("3:04pm")
}

//FullDisplayDate in NY
func FullDisplayDate(val time.Time) string {
	return FullDateFormat(val, location())
}

//FullDateFormat ...
func FullDateFormat(val time.Time, location *time.Location) string {
	return val.In(location).Format("Monday, January 2, 2006")
}

//TimeFormatAmPm ...
func TimeFormatAmPm(val time.Time, location *time.Location) string {
	return val.In(location).Format("3:04pm")
}

//FormatDate ...
func FormatDate(val time.Time, location *time.Location, format string) string {
	//fmt.Println(val, location, format)
	return val.In(location).Format(format)
}

//FormatDateUTC ...
func FormatDateUTC(val time.Time, format string) string {
	return val.In(time.UTC).Format(format)
}

//DisplayDate (01/02/2006)
func DisplayDate(val time.Time) string {
	minTime := time.Time{}

	if val.After(minTime) {
		return val.In(location()).Format("01/02/2006")
	}
	return ""
}

//DisplayDateTime (01/02/2006 03:04PM)
func DisplayDateTime(val time.Time) string {
	minTime := time.Time{}

	if val.After(minTime) {
		return val.In(location()).Format("01/02/2006 03:04PM")
	}
	return ""

}

//DateFormatDisplay (January 2006)
func DateFormatDisplay(val time.Time) string {

	return val.In(location()).Format("January 2006")
}

//DateMonth (Jan)
func DateMonth(val time.Time) string {
	return val.In(location()).Format("Jan")
}

//DateDay (2)
func DateDay(val time.Time) string {
	return val.In(location()).Format("2")
}

//DateYear (2006)
func DateYear(val time.Time) string {
	return val.In(location()).Format("2006")
}

//IntlDateDisplay (2006-01-02)
func IntlDateDisplay(val time.Time) string {
	return val.In(location()).Format("2006-01-02")
}

//WhenCompletedDisplay ...
func WhenCompletedDisplay(val time.Time) string {
	if val.IsZero() {
		return ""
	}

	return "completed " + DateFormatDisplay(val)
}

//WhenRevisedDisplay ...
func WhenRevisedDisplay(val time.Time) string {
	if val.IsZero() {
		return ""
	}

	return ", revised " + DateFormatDisplay(val)
}

//RenderSnippet utilizes GO's templating engine
//to render a template fragment into a string using
//*models*'s data.
func RenderFragment(tmpl string, model interface{}) (template.HTML, error) {

	result, err := ToStringFromString(tmpl, model)

	return template.HTML(result), err
}

//IssueDateFormatDisplay ...
func IssueDateFormatDisplay(val time.Time) string {
	if val.IsZero() {
		return ""
	}
	return DateFormatDisplay(val)
}

//Marshal ...
func Marshal(v interface{}) template.JS {
	a, _ := json.Marshal(v)
	return template.JS(a)
}

//URLSafeKey returns URL-safe representation of key.
func URLSafeKey(datastoreKey *datastore.Key) string {
	if datastoreKey == nil {
		return ""
	}

	return datastoreKey.Encode()
}

//KeyToStringID ...
func KeyToStringID(datastoreKey *datastore.Key) string {
	if datastoreKey == nil {
		return ""
	}

	return datastoreKey.String()
}

//Format2 returns a number with two decimal points
func Format2(d float64) string {
	return fmt.Sprintf("%.2f", d)
}

//FormatPhone ...
func FormatPhone(n string) string {
	if len(n) == 10 {
		return "(" + n[0:3] + ") " + n[3:6] + "-" + n[6:10]
	}

	return n
}

//Int64Display2FromPrecision10 ...
func Int64Display2FromPrecision10(number int64) string {
	//000012357000000 => 1.2347000000
	dec := decimal.New(number, -10)
	str := dec.StringFixed(2)
	return utils.FormatCommas(str)
}

//Int64Display0 formats an integer with precision 4 to 0 decimals
func Int64Display0(number int64) string {
	dec := decimal.New(number, -4)
	str := dec.StringFixed(0)
	return utils.FormatCommas(str)
}

//IntDisplay0 formats an integer with precision 0
func IntDisplay0(number int) string {

	return utils.RenderInteger("#,###.", number)

}

//Int64Display2 formats an integer with precision 4 to 2 decimals
func Int64Display2(number int64) string {
	dec := decimal.New(number, -4)
	str := dec.StringFixed(2)
	return utils.FormatCommas(str)
}

//Int64Display3 formats an integer with precision 4 to 3 decimals
func Int64Display3(number int64) string {
	dec := decimal.New(number, -4)
	str := dec.StringFixed(3)
	return utils.FormatCommas(str)

}

//Float64Display0 formats a float64 with precision 4 to 3 decimals
func Float64Display0(number float64) string {
	number = number / 10000
	p := message.NewPrinter(language.English)
	return p.Sprintf("%.0f", number)
}

//Float64Display2 formatsa float64 with precision 4 to 3 decimals
func Float64Display2(number float64) string {
	number = number / 10000
	p := message.NewPrinter(language.English)
	return p.Sprintf("%.2f", number)
}

//Float64Display3 formats a float64 with precision 4 to 3 decimals
func Float64Display3(number float64) string {
	number = number / 10000
	p := message.NewPrinter(language.English)
	return p.Sprintf("%.3f", number)
}

//DecimalDisplay0 ...
func DecimalDisplay0(dec decimal.Decimal) string {
	str := dec.StringFixed(0)
	return utils.FormatCommas(str)
}

//DecimalDisplay2 ...
func DecimalDisplay2(dec decimal.Decimal) string {
	str := dec.StringFixed(2)
	return utils.FormatCommas(str)
}

//DecimalDisplay3 ...
func DecimalDisplay3(dec decimal.Decimal) string {
	str := dec.StringFixed(3)
	return utils.FormatCommas(str)
}

//PlusOneZeroPad adds one to the value provided and pads with a zero
func PlusOneZeroPad(val int) string {
	newVal := val + 1
	return ZeroPad(newVal)
}

//PlusOne adds one to the value provided
func PlusOne(val int) int {
	return val + 1
}

//PlusOne64 adds one to the value provided
func PlusOne64(val int64) int64 {
	return val + 1
}

//Add ...
func Add(a, b int64) int64 {
	return a + b
}

//Subtract ...
func Subtract(a, b int64) int64 {
	return a - b
}

//Multiply ...
func Multiply(a, b int64) int64 {
	return a * b
}

//Divide ...
func Divide(a, b int64) float64 {
	return float64(a) / float64(b)
}

//Dashes returns dashes related to the level,
//e.g., level 1 = zero dashes, level 2 = two dashes, etc.
func Dashes(level int) string {
	level = level - 1

	result := ""

	for i := 0; i < level; i++ {
		result += "—"
	}

	return result
}

//ZeroPad adds a zero in front of the value
func ZeroPad(val int) string {
	return fmt.Sprintf("%02d", val)
}

//ZeroPad64 adds a zero in front of the value
func ZeroPad64(val int64) string {
	return fmt.Sprintf("%02d", val)
}

//FirstInitial returns first initial of name
func FirstInitial(name string) string {
	return name[0:1]
}

//CalcTabIndex ...
func CalcTabIndex(index int, num int, base int) int {
	return (index * base) + num
}

//IsToday ...
func IsToday(dte time.Time) bool {

	yearNow, monthNow, dayNow := time.Now().In(location()).Date()
	yearDte, monthDte, dayDte := dte.In(location()).Date()

	if yearNow == yearDte && monthNow == monthDte && dayNow == dayDte {
		return true
	}

	return false
}

//NewLineToBR ...
func NewLineToBR(s string) template.HTML {
	return template.HTML(strings.Replace(s, "\n", "<br />", -1))
}

//DictHelper ...
func DictHelper(values ...interface{}) (map[string]interface{}, error) {
	if len(values)%2 != 0 {
		return nil, errors.New("invalid dict call")
	}

	dict := make(map[string]interface{}, len(values)/2)

	for i := 0; i < len(values); i += 2 {
		key, ok := values[i].(string)

		if !ok {
			return nil, errors.New("dict keys must be strings")
		}

		dict[key] = values[i+1]
	}
	return dict, nil
}

//HTMLEscape ...
func HTMLEscape(input string) template.HTML {
	return template.HTML(input)
}

//ToUppercase ...
func ToUppercase(v string) string {
	return strings.ToUpper(v)
}

//ToLowercase ...
func ToLowercase(v string) string {
	return strings.ToLower(v)
}

//ToTitleCase ...
func ToTitleCase(v string) string {
	return strings.Title(v)
}

//PrepPhone replaces ")", "(", "-", " "
//to make sure it will show up in a "tel"
//tag with proper format.
func PrepPhone(v string) string {

	v = strings.Replace(v, "(", "", -1)
	v = strings.Replace(v, ")", "", -1)
	v = strings.Replace(v, " ", "", -1)
	v = strings.Replace(v, "-", "", -1)
	v = strings.Replace(v, ",", "", -1)

	return v
}

//Int64ToTime returns an HTML-time-formatted
//string from an int64. Example: 835 -> 08:35
func Int64ToTime(v int64) string {

	hours := fmt.Sprintf("%02d", v/100)
	minutes := fmt.Sprintf("%02d", v%100)
	return fmt.Sprintf("%s:%s", hours, minutes)

}

//ArrayToQS takes a key and string array
//and produces &key=v&key=v, etc
func ArrayToQS(key string, values []string) template.URL {

	sb := strings.Builder{}
	for _, v := range values {
		sb.WriteString(fmt.Sprintf("&%s=%s", key, v))
	}

	return template.URL(sb.String())
}

//PrecisionFormatter formats value to precision decimal places
func PrecisionFormatter(value, precision int64) string {
	amount := float64(value) / float64(intPow(10, precision))
	formatter := fmt.Sprintf("%%.%df", precision)
	return fmt.Sprintf(formatter, amount)

}

func PrecisionFormatterFloat64(value float64, precision int64) string {
	amount := value / float64(intPow(10, precision))
	formatter := fmt.Sprintf("%%.%df", precision)
	return fmt.Sprintf(formatter, amount)
}

func intPow(n, m int64) int64 {
	if m == 0 {
		return 1
	}
	result := n
	for i := int64(2); i <= m; i++ {
		result *= n
	}
	return result
}

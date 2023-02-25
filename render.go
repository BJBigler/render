package render

import (
	"bytes"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"html/template"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"

	"github.com/bjbigler/utils"
	"golang.org/x/net/context"
)

//ParseTemplateSets parses sets of files into templates, one per page needed.
//The *baseTemplate* should be a relative path, e.g., "views/master.html"
//In *sets*, the first string should be the template map lookup name.
//Ex: {"authenticators", "views/master.html", "views/authenticators.html"}
func ParseTemplateSets(baseTemplate string, sets [][]string) (templates map[string]*template.Template, err error) {

	templates = make(map[string]*template.Template)

	funcMap := GetFuncMap()

	templateName := baseTemplate
	templateParts := strings.Split(baseTemplate, "/")
	if len(templateParts) > 0 {
		templateName = templateParts[len(templateParts)-1]
	}

	for _, set := range sets {

		t := template.New(templateName).Funcs(funcMap)

		templateSet := []string{baseTemplate}
		templateSet = append(templateSet, set[1:]...)

		_, err = t.ParseFiles(templateSet...)

		if err != nil {
			return nil, err
		}

		//put the template in map using the lookup name
		lookupName := set[0]
		templates[lookupName] = t

	}

	return templates, err
}

//Template renders template from the template map produced by ParseTemplateSets
func Template(w http.ResponseWriter, templates map[string]*template.Template, model interface{}, templateIndex string) error {

	template, ok := templates[templateIndex]
	if ok {
		return template.Execute(w, &model)
	}
	return fmt.Errorf("template %s not found", templateIndex)

}

//FindAndParseTemplates parses all templates
func FindAndParseTemplates() (*template.Template, error) {
	funcMap := GetFuncMap()

	cleanRoot := filepath.Clean("views")
	pfx := len(cleanRoot) + 1
	root := template.New("")

	err := filepath.Walk(cleanRoot, func(path string, info os.FileInfo, e1 error) error {
		if !info.IsDir() && strings.HasSuffix(path, ".html") {
			if e1 != nil {
				return e1
			}

			b, e2 := ioutil.ReadFile(path)
			if e2 != nil {
				return e2
			}

			name := path[pfx:]
			t := root.New(name).Funcs(funcMap)
			_, e2 = t.ParseFiles(string(b))
			if e2 != nil {
				return e2
			}
		}

		return nil
	})

	return root, err
}

//ToBrowser pairs master.html template to whatever instanceTemplate gets
//passed in.
func ToBrowser(w http.ResponseWriter, model interface{}, templates ...string) error {

	templates = append([]string{"views/master.html"}, templates...)

	funcMap := GetFuncMap()                       //Sets up the funcMaps to be used on templates, e.g., formatters like displayDate
	tmpl := template.New("master.html")           //Initializes named template
	funcs := tmpl.Funcs(funcMap)                  //associates the funcMap with the template
	parsed, err := funcs.ParseFiles(templates...) //Parses the templates
	t := template.Must(parsed, err)               //Creates template
	err = t.Execute(w, model)                     //Merges template with data

	if err != nil {
		fmt.Println(err)
		return err
	}

	return nil
}

//RedirectTo ...
func RedirectTo(w http.ResponseWriter, redirectURL *url.URL) {
	jsonOut := fmt.Sprintf(`{"statusCode": %v, "redirectTo":"%s" }`, 6, redirectURL.String())
	JSONToBrowser(w, []byte(jsonOut))
}

//ToBrowserNoMaster prints out template with no master.
func ToBrowserNoMaster(w http.ResponseWriter, instanceTemplate string, model interface{}) error {

	funcMap := GetFuncMap()

	//Here, we grab the name by grabbing the text after the final slash
	positionOfLastSlash := strings.LastIndex(instanceTemplate, "/")
	templateName := string(instanceTemplate[positionOfLastSlash+1:])
	t := template.Must(template.New(templateName).Funcs(funcMap).ParseFiles(instanceTemplate))

	if err := t.Execute(w, model); err != nil {
		return err
	}

	return nil
}

//ToBrowserNoMasterNew ...
func ToBrowserNoMasterNew(w http.ResponseWriter, model interface{}, templates ...string) error {
	if len(templates) == 0 {
		return fmt.Errorf("no templates supplied")
	}

	funcMap := GetFuncMap()

	//Here, we grab the name by grabbing the text after the final slash
	positionOfLastSlash := strings.LastIndex(templates[0], "/")
	templateName := string(templates[0][positionOfLastSlash+1:])
	t := template.Must(template.New(templateName).Funcs(funcMap).ParseFiles(templates...))

	if err := t.Execute(w, model); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}

	return nil
}

//ToString renders the instanceTemplate alone, without any master supplied.
//It's used for html fragments, largely in AJAX.
func ToString(model interface{}, templates ...string) (string, error) {

	if len(templates) == 0 {
		return "", fmt.Errorf("no template(s) specified")
	}

	funcMap := GetFuncMap()

	instanceTemplate := templates[0]
	positionOfLastSlash := strings.LastIndex(instanceTemplate, "/")
	templateName := string(instanceTemplate[positionOfLastSlash+1:])

	tmpl := template.New(templateName)            //Initializes named template
	funcs := tmpl.Funcs(funcMap)                  //associates the funcMap with the template
	parsed, err := funcs.ParseFiles(templates...) //Parses the templates
	t := template.Must(parsed, err)               //Creates template

	var doc bytes.Buffer

	err = t.Execute(&doc, model)
	if err != nil {
		return "", err
	}

	return doc.String(), nil
}

//ToStringFromString renders a string HTML template
func ToStringFromString(html string, model interface{}) (result string, err error) {
	funcMap := GetFuncMap()                  //Sets up the funcMaps to be used on templates, e.g., formatters like displayDate
	bodyTemplate := template.New("template") //Initializes named template
	funcs := bodyTemplate.Funcs(funcMap)     //associates the funcMap with the template

	parsed, err := funcs.Parse(html) //Parses the template

	if err != nil {
		return "", err
	}

	t := template.Must(parsed, err) //Creates template
	if err != nil {
		return "", err
	}

	var b bytes.Buffer

	err = t.Execute(&b, model) //Executes
	if err != nil {
		return "", err
	}

	return b.String(), err
}

//ToHTMLOld renders a template as template.HTML to use as html fragments when compositing
//a page together.
func ToHTMLOld(instanceTemplate string, model interface{}) (template.HTML, error) {

	funcMap := GetFuncMap()

	//For some reason, the template name has to be filename, e.g., template.html.
	//The templates are supplied to the function in terms of their folder location,
	//e.g., /views/research/template.html.
	//Here, we grab the name by grabbing the text after the final slash
	positionOfLastSlash := strings.LastIndex(instanceTemplate, "/")
	templateName := string(instanceTemplate[positionOfLastSlash+1:])

	t := template.Must(template.New(templateName).Funcs(funcMap).ParseFiles(instanceTemplate))

	var doc bytes.Buffer

	err := t.Execute(&doc, model)
	if err != nil {
		return template.HTML(""), err
	}

	return template.HTML(doc.String()), nil
}

//ToHTML renders a template as template.HTML to use as html fragments when compositing
//a page together.
func ToHTML(model interface{}, templates ...string) (template.HTML, error) {

	if len(templates) == 0 {
		return template.HTML(""), fmt.Errorf("no templates passed into ToHTML()")
	}
	funcMap := GetFuncMap()

	//For some reason, the template name has to be filename, e.g., template.html.
	//The templates are supplied to the function in terms of their folder location,
	//e.g., /views/research/template.html.
	//Here, we grab the name by grabbing the text after the final slash

	positionOfLastSlash := strings.LastIndex(templates[0], "/")
	templateName := strings.TrimSpace(string(templates[0][positionOfLastSlash+1:]))

	tmpl := template.New(templateName) //Initializes named template

	funcs := tmpl.Funcs(funcMap) //associates the funcMap with the template

	parsed, err := funcs.ParseFiles(templates...) //Parses templates

	if err != nil {
		return template.HTML(""), err
	}

	t := template.Must(parsed, err) //Creates template

	var doc bytes.Buffer

	err = t.Execute(&doc, model)
	if err != nil {
		return template.HTML(""), err
	}

	return template.HTML(doc.String()), nil
}

//JSONToBrowser sends a JSON []byte to the browser
func JSONToBrowser(w http.ResponseWriter, json []byte) error {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Add("Access-Control-Allow-Credentials", "true")
	w.Header().Set("Access-Control-Allow-Methods", "POST, GET, OPTIONS, PUT, DELETE")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	_, err := w.Write(json)

	return err
}

//StringToBrowser takes *html* string and sends it to the browser.
func StringToBrowser(w http.ResponseWriter, html string) error {
	w.Header().Set("Content-Type", "text/html")
	w.Header().Add("Access-Control-Allow-Credentials", "true")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	_, err := w.Write([]byte(html))

	return err
}

//CsvToBrowser takes a [][]string and sends it to the browser as a CSV file.
func CsvToBrowser(w http.ResponseWriter, csvRecords [][]string, filename string) {
	w.Header().Set("Content-Type", "text/csv")
	w.Header().Set("Content-Disposition", "attachment;filename="+filename)
	w.Header().Add("Access-Control-Allow-Credentials", "true")

	b := &bytes.Buffer{}
	csvWriter := csv.NewWriter(b)
	csvWriter.WriteAll(csvRecords)

	w.Write(b.Bytes())
}

//WriteXlsToBrowser takes a string formatted as Office XML and outputs it to the browswer as a file.
func WriteXlsToBrowser(ctx context.Context, w http.ResponseWriter, xls string, filename string) {
	w.Header().Set("Content-Type", "application/vnd.ms-excel")
	w.Header().Set("Content-Disposition", "attachment;filename="+filename)
	w.Header().Add("Access-Control-Allow-Credentials", "true")

	b := bytes.NewBuffer([]byte(xls))

	if _, err := b.WriteTo(w); err != nil {
		output := fmt.Sprintf("%v", err)
		fmt.Fprint(w, output)
	}
}

//WriteIcsToBrowser delivers iCalendar files to the browser
func WriteIcsToBrowser(w http.ResponseWriter, calendar string, filename string) {
	w.Header().Set("Content-Type", "text/calendar")
	w.Header().Set("Content-Disposition", "attachment;filename="+filename)
	w.Header().Add("Access-Control-Allow-Credentials", "true")

	b := bytes.NewBuffer([]byte(calendar))

	if _, err := b.WriteTo(w); err != nil {
		output := fmt.Sprintf("%v", err)
		fmt.Fprint(w, output)
	}
}

//PDFToBrowser streams PDF file to browser. Its main purpose
//is security: instead of linking to the file system,
//code calling this func requires a login.
func PDFToBrowser(w http.ResponseWriter, path string) error {
	streamPDFbytes, err := ioutil.ReadFile(path)

	if err != nil {
		return err
	}

	lastSlash := strings.LastIndex(path, "/")
	filename := path[lastSlash+1:]

	w.Header().Set("Content-type", "application/pdf")
	w.Header().Set("Content-Disposition", "inline;filename="+filename)
	w.Header().Add("Access-Control-Allow-Credentials", "true")

	b := bytes.NewBuffer(streamPDFbytes)
	if _, err := b.WriteTo(w); err != nil {
		return err
	}

	return nil
}

//PDFBytesToBrowser streams PDF file to browser. Its main purpose
//is security: instead of linking to the file system,
//code calling this func requires a login.
func PDFBytesToBrowser(w http.ResponseWriter, fileName string, file []byte) error {

	w.Header().Set("Content-type", "application/pdf")
	w.Header().Set("Content-Disposition", "inline;filename="+fileName)
	w.Header().Add("Access-Control-Allow-Credentials", "true")

	b := bytes.NewBuffer(file)
	if _, err := b.WriteTo(w); err != nil {
		return err
	}

	return nil
}

//GenericBytesToBrowser streams a file without setting its content type
func GenericBytesToBrowser(w http.ResponseWriter, fileName string, file []byte) error {

	// w.Header().Set("Content-type", "application/pdf")
	// w.Header().Set("Content-Disposition", "inline;filename="+fileName)
	// w.Header().Add("Access-Control-Allow-Credentials", "true")

	b := bytes.NewBuffer(file)
	if _, err := b.WriteTo(w); err != nil {
		return err
	}

	return nil
}

//ReportError sends JSON message with "statusCode:0" and "error:" with the error specified
func ReportError(w http.ResponseWriter, message interface{}) {
	type helper struct {
		StatusCode int    `json:"statusCode"`
		Error      string `json:"error"`
	}

	h := helper{
		StatusCode: 0,
		Error:      fmt.Sprintf("%v", message),
	}

	jsonOut, _ := json.Marshal(h)
	JSONToBrowser(w, jsonOut)
}

//ReportMessage sends JSON messagae with "statusCode:0" and "msg:" *message
func ReportMessage(w http.ResponseWriter, message string) {

	type helper struct {
		StatusCode int    `json:"statusCode"`
		Message    string `json:"msg"`
	}

	h := helper{
		StatusCode: 0,
		Message:    message,
	}

	jsonOut, _ := json.Marshal(h)
	JSONToBrowser(w, jsonOut)

}

//ReportRedirect sends JSON message with "statusCode" of 1 "redirect" equal to the *redirect*
//provided. If  *redirectID* is also provided, the javascript will
//attempt to scroll into view any found element with that ID.
func ReportRedirect(w http.ResponseWriter, redirect string) {
	type helper struct {
		StatusCode int    `json:"statusCode"`
		Redirect   string `json:"redirect"`
	}

	h := helper{
		StatusCode: 1,
		Redirect:   redirect,
	}

	jsonOut, _ := json.Marshal(h)
	JSONToBrowser(w, jsonOut)
}

//ReportErrors sends JSON message with "statusCode:0" and the errors specified
func ReportErrors(w http.ResponseWriter, errors []error) {
	errorsJSON, _ := json.Marshal(errors)
	jsonOut := fmt.Sprintf(`{"statusCode": %v, "errors":%v }`, 0, string(errorsJSON))
	JSONToBrowser(w, []byte(jsonOut))
}

//ReportSuccess sends JSON message with "statusCode:1"
func ReportSuccess(w http.ResponseWriter) {
	jsonOut := fmt.Sprintf(`{"statusCode": %v }`, 1)
	JSONToBrowser(w, []byte(jsonOut))
}

//ReportReload sends JSON message with "statusCode:5", which doGetFetch/doPostFetch
//interpret as a reload
func ReportReload(w http.ResponseWriter) {
	jsonOut := fmt.Sprintf(`{"statusCode": %v }`, 5)
	JSONToBrowser(w, []byte(jsonOut))
}

//GetFuncMap provides a set of utility functions to help format data on an HTML output page.
func GetFuncMap() map[string]interface{} {
	return template.FuncMap{
		"formatDate":                     FormatDate,
		"formatDateUTC":                  FormatDateUTC,
		"displayDate":                    DisplayDate,
		"displayMorningAfternoonEvening": DisplayMorningAfternoonEvening, //
		"displayDateTime":                DisplayDateTime,
		"dateFormatDisplay":              DateFormatDisplay,
		"dateMonth":                      DateMonth,
		"dateDay":                        DateDay,
		"dateYear":                       DateYear,
		"dateTimeFormal":                 DateTimeFormal,
		"shortDateTime":                  ShortDateTime,
		"renderFragment":                 RenderFragment,
		"decimalDisplay0":                DecimalDisplay0, //Precision 6
		"decimalDisplay2":                DecimalDisplay2, //Precision 6
		"decimalDisplay3":                DecimalDisplay3, //Precision 6
		"intDisplay0":                    IntDisplay0,
		"int64Display0":                  Int64Display0,                //Precision 4
		"int64Display2":                  Int64Display2,                //Precision 4
		"int64Display3":                  Int64Display3,                //Precision 4
		"float64Display0":                Float64Display0,              //Precision 4
		"float64Display2":                Float64Display2,              //Precision 4
		"float64Display3":                Float64Display3,              //Precision 4
		"int64Display2FromPrecision10":   Int64Display2FromPrecision10, //Precision 10
		"fullDateTimeET":                 FullDateTimeET,               //
		"whenCompletedDisplay":           WhenCompletedDisplay,         //
		"whenRevisedDisplay":             WhenRevisedDisplay,           //
		"issueDateFormatDisplay":         IssueDateFormatDisplay,       //
		"marshal":                        Marshal,                      //
		"urlSafeKey":                     URLSafeKey,                   //
		"keyToStringID":                  KeyToStringID,                //
		"format2":                        Format2,                      //
		"formatPhone":                    FormatPhone,                  //
		"plusOne":                        PlusOne,                      //
		"plusOne64":                      PlusOne64,                    //
		"add":                            Add,                          //Add two numbers
		"subtract":                       Subtract,                     //Subtract two numbers
		"multiply":                       Multiply,
		"divide":                         Divide,
		"plusOneZeroPad":                 PlusOneZeroPad, //
		"zeroPad":                        ZeroPad,        //
		"zeroPad64":                      ZeroPad64,      //
		"dashes":                         Dashes,         //
		"fullDisplayDate":                FullDisplayDate,
		"fullDateFormat":                 FullDateFormat, //
		"timeFormatAmPm":                 TimeFormatAmPm,
		"intlDateDisplay":                IntlDateDisplay, //
		"firstInitial":                   FirstInitial,    //
		"calcTabIndex":                   CalcTabIndex,    //
		"isToday":                        IsToday,         //
		"newLineToBR":                    NewLineToBR,     //
		"timeFormat":                     TimeFormat,      //
		"dict":                           DictHelper,      //
		"htmlEscape":                     HTMLEscape,      //
		"toUppercase":                    ToUppercase,     //
		"toLowercase":                    ToLowercase,     //
		"toTitleCase":                    ToTitleCase,
		"int64ToTime":                    Int64ToTime, //Converts, e.g., 835 to 8:35
		"academicYearView":               utils.AcademicYearView,
		"prepPhone":                      PrepPhone, //Preps a phone number for use in an HTML tel tag
		"arrayToQS":                      ArrayToQS,
		"precisionFormatter":             PrecisionFormatter,
		"precisionFormatterFloat64":      PrecisionFormatterFloat64,
		"safe": func(s string) template.HTML {
			return template.HTML(s)
		},
	}
}

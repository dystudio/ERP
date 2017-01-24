package base

import (
	"encoding/json"
	md "goERP/models"
	"strconv"
	"strings"
)

type SequenceController struct {
	BaseController
}

func (ctl *SequenceController) Post() {
	action := ctl.Input().Get("action")
	switch action {
	case "validator":
		ctl.Validator()
	case "table": //bootstrap table的post请求
		ctl.PostList()
	case "create":
		ctl.PostCreate()
	default:
		ctl.PostList()
	}
}
func (ctl *SequenceController) Put() {
	id := ctl.Ctx.Input.Param(":id")
	ctl.URL = "/sequence/"
	if idInt64, e := strconv.ParseInt(id, 10, 64); e == nil {
		if sequence, err := md.GetSequenceByID(idInt64); err == nil {
			if err := ctl.ParseForm(&sequence); err == nil {

				if err := md.UpdateSequenceByID(sequence); err == nil {
					ctl.Redirect(ctl.URL+id+"?action=detail", 302)
				}
			}
		}
	}
	ctl.Redirect(ctl.URL+id+"?action=edit", 302)

}
func (ctl *SequenceController) Get() {
	ctl.PageName = "序号管理"
	action := ctl.Input().Get("action")
	switch action {
	case "create":
		ctl.Create()
	case "edit":
		ctl.Edit()
	case "detail":
		ctl.Detail()
	default:
		ctl.GetList()

	}
	ctl.Data["PageName"] = ctl.PageName + "\\" + ctl.PageAction
	ctl.URL = "/sequence/"
	ctl.Data["URL"] = ctl.URL

	ctl.Data["MenuSequenceActive"] = "active"
}
func (ctl *SequenceController) Edit() {
	id := ctl.Ctx.Input.Param(":id")
	if id != "" {
		if idInt64, e := strconv.ParseInt(id, 10, 64); e == nil {
			if sequence, err := md.GetSequenceByID(idInt64); err == nil {
				ctl.PageAction = sequence.Name
				ctl.Data["Sequence"] = sequence

			}
		}
	}
	ctl.Data["FormField"] = "form-edit"
	ctl.Data["Action"] = "edit"
	ctl.Data["RecordID"] = id
	ctl.Layout = "base/base.html"
	ctl.TplName = "config/sequence_form.html"
}
func (ctl *SequenceController) Create() {
	ctl.Data["Action"] = "create"
	ctl.Data["Readonly"] = false
	ctl.Data["FormField"] = "form-create"
	ctl.PageAction = "创建"
	ctl.Layout = "base/base.html"
	ctl.TplName = "config/sequence_form.html"
}
func (ctl *SequenceController) Detail() {
	//获取信息一样，直接调用Edit
	ctl.Edit()
	ctl.Data["Readonly"] = true
	ctl.Data["Action"] = "detail"
}
func (ctl *SequenceController) PostCreate() {
	result := make(map[string]interface{})
	postData := ctl.GetString("postData")
	sequence := new(md.Sequence)
	var (
		err  error
		id   int64
		errs []error
	)
	if err = json.Unmarshal([]byte(postData), sequence); err == nil {
		// 获得struct表名
		// structName := reflect.Indirect(reflect.ValueOf(category)).Type().Name()
		if id, errs = md.AddSequence(sequence, &ctl.User); len(errs) == 0 {
			result["code"] = "success"
			result["location"] = ctl.URL + strconv.FormatInt(id, 10) + "?action=detail"
		} else {
			result["code"] = "failed"
			result["message"] = "数据创建失败"
			var debugs []string
			for _, item := range errs {
				debugs = append(debugs, item.Error())
			}
			result["debug"] = debugs
		}
	} else {
		result["code"] = "failed"
		result["message"] = "请求数据解析失败"
		result["debug"] = err.Error()
	}
	ctl.Data["json"] = result
	ctl.ServeJSON()
}
func (ctl *SequenceController) Validator() {
	name := ctl.GetString("name")
	name = strings.TrimSpace(name)
	recordID, _ := ctl.GetInt64("recordID")
	result := make(map[string]bool)
	obj, err := md.GetSequenceByName(name)
	if err != nil {
		result["valid"] = true
	} else {
		if obj.Name == name {
			if recordID == obj.ID {
				result["valid"] = true
			} else {
				result["valid"] = false
			}
		} else {
			result["valid"] = true
		}
	}
	ctl.Data["json"] = result
	ctl.ServeJSON()
}

// 获得符合要求的数据
func (ctl *SequenceController) sequenceList(query map[string]interface{}, exclude map[string]interface{}, fields []string, sortby []string, order []string, offset int64, limit int64) (map[string]interface{}, error) {

	var arrs []md.Sequence
	paginator, arrs, err := md.GetAllSequence(query, exclude, fields, sortby, order, offset, limit)
	result := make(map[string]interface{})
	if err == nil {

		//使用多线程来处理数据，待修改
		tableLines := make([]interface{}, 0, 4)
		for _, line := range arrs {
			oneLine := make(map[string]interface{})
			oneLine["Name"] = line.Name
			oneLine["ID"] = line.ID
			oneLine["id"] = line.ID
			oneLine["Prefix"] = line.Prefix
			oneLine["Current"] = line.Current
			oneLine["Padding"] = line.Padding
			oneLine["StructName"] = line.StructName
			oneLine["Active"] = line.Active
			oneLine["IsDefault"] = line.IsDefault
			tableLines = append(tableLines, oneLine)
		}
		result["data"] = tableLines
		if jsonResult, er := json.Marshal(&paginator); er == nil {
			result["paginator"] = string(jsonResult)
			result["total"] = paginator.TotalCount
		}
	}
	return result, err
}
func (ctl *SequenceController) PostList() {
	query := make(map[string]interface{})
	exclude := make(map[string]interface{})
	fields := make([]string, 0, 0)
	sortby := make([]string, 0, 0)
	order := make([]string, 0, 0)
	offset, _ := ctl.GetInt64("offset")
	limit, _ := ctl.GetInt64("limit")
	name := strings.TrimSpace(ctl.GetString("Name"))
	if name != "" {
		query["Name"] = name
	}

	if result, err := ctl.sequenceList(query, exclude, fields, sortby, order, offset, limit); err == nil {
		ctl.Data["json"] = result
	}
	ctl.ServeJSON()

}

func (ctl *SequenceController) GetList() {
	viewType := ctl.Input().Get("view")
	if viewType == "" || viewType == "table" {
		ctl.Data["ViewType"] = "table"
	}
	ctl.PageAction = "列表"
	ctl.Data["tableId"] = "table-sequence"
	ctl.Layout = "base/base_list_view.html"
	ctl.TplName = "config/sequence_list_search.html"

}

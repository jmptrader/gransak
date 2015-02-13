package core

import (
	"reflect"
	"strings"
)

func NewGransak() *GransakCore {
	return &GransakCore{
		separator:   "_",
		placeholder: "{{.}}",
		valueholder: "{{v}}",
	}
}

type GransakCore struct {
	toEvaluate      []string
	evaluatedTokens []string
	template        string
	separator       string
	placeholder     string
	valueholder     string
	pos             int
	param           *gransakParam
	tableName       string
}

func (this *GransakCore) Table(tableName string) *GransakCore {
	this.tableName = tableName

	return this
}

func (this *GransakCore) ToSql(input string, param interface{}) string {
	this.reset()

	this.tokenize(input)

	this.param = newGransakParam(param, reflect.TypeOf(param).Kind())

	for this.pos = 0; this.pos < len(this.toEvaluate); this.pos++ {
		token := this.toEvaluate[this.pos]

		if node, isCandidate := isCandidateToOperator(token); isCandidate {
			if foundNode, found := this.find(node, this.pos); found {

				foundNode.Apply(this)

			} else {

				this.evaluated(token)

			}
		} else {

			this.evaluated(token)

		}
	}

	this.replaceValue()

	this.appendSelectStatement()

	this.tableName = ""

	return strings.Trim(this.template, " ")
}

func (this *GransakCore) reset() {
	this.toEvaluate = []string{}
	this.evaluatedTokens = []string{}
	this.template = ""
	this.param = nil
}

func (this *GransakCore) tokenize(input string) {
	this.toEvaluate = strings.Split(input, this.separator)
}

func (this *GransakCore) find(nodeParam *Node, pos int) (*Node, bool) {
	if pos >= len(this.toEvaluate) {
		return nil, false
	}

	next := this.toEvaluate[pos]

	if nodeParam.Name != next {
		return nil, false
	}

	if len(nodeParam.Nodes) > 0 {
		for _, node := range nodeParam.Nodes {
			if foundNode, found := this.find(node, pos+1); found {

				return foundNode, true

			}
		}

		//none of its children nodes matched, check if is itself an operator
		if nodeParam.IsOperator == true {
			this.pos = pos
			return nodeParam, true
		}

	} else {
		this.pos = pos
		return nodeParam, true
	}

	return nil, false
}

func (this *GransakCore) appendField() string {
	field := this.getLastField()
	if field != "" {
		this.template += field + " " + this.placeholder + " "
	}
	return field
}

func (this *GransakCore) getLastField() string {
	field := strings.Join(this.evaluatedTokens, this.separator)
	this.evaluatedTokens = []string{}
	return field
}

func (this *GransakCore) evaluated(token string) {
	this.evaluatedTokens = append(this.evaluatedTokens, token)
}

func (this *GransakCore) replace(replace, replaceFor string) {
	this.template = strings.Replace(
		this.template,
		replace,
		replaceFor,
		-1,
	)
}

func (this *GransakCore) replacePlaceholder(replaceFor string) {
	this.replace(this.placeholder, replaceFor)
}

func (this *GransakCore) replaceValueHolders(replaceFor string) {
	this.replace(this.valueholder, replaceFor)
}

func (this *GransakCore) replaceValue() {
	if len(this.param.parts) == 0 {
		this.replaceValueHolders(this.param.StrRepresentation)
	} else {
		for index, value := range this.param.parts {
			if isLast(index, this.param.parts) {
				this.replaceValueHolders(value)
			} else {
				this.template = strings.Replace(this.template, this.valueholder, value, 1)
			}
		}
	}
}

func (this *GransakCore) getCorrectSqlFormat(value string) string {
	if this.param.kind == reflect.String {
		return "'" + value + "'"
	}
	return value
}

func (this *GransakCore) appendSelectStatement() {
	if strings.Trim(this.tableName, " ") != "" {
		this.template = "SELECT * FROM " + this.tableName + " WHERE " + this.template
	}
}

func isCandidateToOperator(item string) (*Node, bool) {
	for _, node := range Tree.Nodes {
		if node.Name == item {
			return node, true
		}
	}
	return &Node{}, false
}

func isLast(index int, slice []string) bool {
	return index == (len(slice) - 1)
}
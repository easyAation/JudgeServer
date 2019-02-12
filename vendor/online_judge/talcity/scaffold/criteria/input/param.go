package input

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"reflect"
	"strconv"
	"strings"
	"unsafe"

	"github.com/beego/mux"
)

var (
	// for ParseMultipartForm
	multipartFormMemoryLimit int64 = 50 * 1000 * 1000
)

type SearchSuit struct {
	Filters  map[string]interface{}
	Querys   map[string]interface{}
	OrderBys [][]string
	Page     int
	Size     int
}

func NewSearchSuit() SearchSuit {
	return SearchSuit{
		Filters: make(map[string]interface{}),
		Querys:  make(map[string]interface{}),
	}
}

type Param struct {
	Req            *http.Request
	vars           map[string]string
	query          url.Values
	errs           []string
	multipartPared bool
}

func NewParam(req *http.Request) *Param {
	vars := mux.Params(req)
	return &Param{
		Req:  req,
		vars: vars,
	}
}

func (p *Param) AddError(msg string) *Param {
	p.errs = append(p.errs, msg)
	return p
}

func (p *Param) AddErrf(format string, a ...interface{}) *Param {
	p.errs = append(p.errs, fmt.Sprintf(format, a...))
	return p
}

func (p *Param) Error() error {
	if len(p.errs) == 0 {
		return nil
	}
	return errors.New("param err: " + strings.Join(p.errs, "\n"))
}

// Var get the router params
func (p *Param) Var(key string, result *string) *Param {
	ret, ok := p.vars[key]
	if !ok {
		return p.AddError(fmt.Sprintf("path var %s not set", key))
	}
	*result = ret
	return p
}

func (p *Param) VarInt64(key string, ret *int64) *Param {
	val, ok := p.vars[key]
	if !ok {
		return p.AddError(fmt.Sprintf("path var %s not set", key))
	}

	var (
		err error
	)
	*ret, err = strconv.ParseInt(val, 10, 64)
	if err != nil {
		return p.AddError(err.Error())
	}
	return p
}

func (p *Param) Query(key string, ret *string) *Param {
	if p.query == nil {
		p.query = p.Req.URL.Query()
	}
	*ret = p.query.Get(key)
	return p
}

func (p *Param) Required(key string, ret *string) *Param {
	p.Query(key, ret)
	if *ret == "" {
		p.AddError(key + " not found")
	}
	return p
}

func (p *Param) Int(key string, ret *int) *Param {
	var (
		val string
		err error
	)
	p.Query(key, &val)
	*ret, err = strconv.Atoi(val)
	if err != nil {
		return p.AddError(err.Error())
	}
	return p
}

func (p *Param) Int64(key string, ret *int64) *Param {
	var (
		val string
		err error
	)

	p.Query(key, &val)
	*ret, err = strconv.ParseInt(val, 10, 64)
	if err != nil {
		return p.AddError(err.Error())
	}

	return p
}

func (p *Param) Bool(key string, ret *bool) *Param {
	var val string

	p.Query(key, &val)
	*ret, _ = strconv.ParseBool(val)
	return p
}

func (p *Param) IntWithDefault(key string, ret *int, defaultVale int) *Param {
	var (
		val string
		err error
	)
	p.Query(key, &val)
	if val == "" {
		*ret = defaultVale
		return p
	}
	*ret, err = strconv.Atoi(val)
	if err != nil {
		p.AddError(err.Error())
	}
	return p
}

func (p *Param) FormData(key string, ret *string) *Param {
	if !p.multipartPared {
		err := p.Req.ParseMultipartForm(multipartFormMemoryLimit)
		if err != nil {
			return p.AddError(err.Error())
		}
	}
	*ret = p.Req.PostForm.Get(key)
	return p
}

func (p *Param) FormDataInt64(key string, ret *int64) *Param {
	var (
		err error
		val string
	)
	p.FormData(key, &val)
	if len(p.errs) > 0 {
		return p
	}

	*ret, err = strconv.ParseInt(val, 10, 64)
	if err != nil {
		return p.AddErrf("for key %s: %v", key, err)
	}
	return p
}

func (p *Param) FormDataInt(key string, ret *int) *Param {
	var (
		valInt64 int64
	)
	p.FormDataInt64(key, &valInt64)
	if len(p.errs) > 0 {
		return p
	}
	*ret = int(valInt64)
	return p
}

func (p *Param) FormDataRequired(key string, ret *string) *Param {
	p.FormData(key, ret)
	if len(p.errs) > 0 {
		return p
	}
	if *ret == "" {
		return p.AddError(key + " not found")
	}
	return p
}

func (p *Param) JSONBody(obj interface{}) *Param {
	b, err := ioutil.ReadAll(p.Req.Body)
	if err != nil {
		return p.AddError(err.Error())
	}
	defer p.Req.Body.Close()

	err = json.Unmarshal(b, obj)
	if err != nil {
		return p.AddErrf("invalid body: %v", err.Error())
	}
	if err = Validate(obj); err != nil {
		return p.AddError(err.Error())
	}
	return p
}

// Nest 返回解析Request之后的值，并填充到所给的struct中
// struct中根据值所在的位置设置tag：header、var、query、body
// body中的比较特殊，body的tag为body:body，他会调用json.Marshal填充值
func (p *Param) Nest(obj interface{}) *Param {
	t := reflect.TypeOf(obj)
	if t == nil {
		return p.AddErrf("null nest value")
	}
	if t.Kind() == reflect.Ptr {
		t = t.Elem()
	}
	if t.Kind() != reflect.Struct {
		return p.AddErrf("invalid nest type %d", t.Kind())
	}

	var setFieldValue = func(fieldType reflect.Type, fieldValue reflect.Value, s string) error {
		var err error
		isPtr := fieldType.Kind() == reflect.Ptr
		emptyValue := s == ""
		var k reflect.Kind
		if isPtr {
			k = fieldType.Elem().Kind()
		} else {
			k = fieldType.Kind()
		}
		switch k {
		case reflect.String:
			if isPtr {
				fieldValue.Set(reflect.NewAt(fieldType.Elem(), (unsafe.Pointer)(&s)))
			} else {
				fieldValue.SetString(s)
			}
		case reflect.Int, reflect.Int64:
			if emptyValue {
				if !isPtr {
					fieldValue.SetInt(0)
				}
				break
			}
			// not empty
			i, err := strconv.ParseInt(s, 10, 64)
			if err != nil {
				return err
			}
			if isPtr {
				fieldValue.Set(reflect.NewAt(fieldType.Elem(), (unsafe.Pointer)(&i)))
			} else {
				fieldValue.SetInt(i)
			}
		case reflect.Bool:
			if emptyValue {
				if !isPtr {
					fieldValue.SetBool(false)
				}
				break
			}
			b, err := strconv.ParseBool(s)
			if err != nil {
				return err
			}
			if isPtr {
				fieldValue.Set(reflect.NewAt(fieldType.Elem(), (unsafe.Pointer)(&b)))
			} else {
				fieldValue.SetBool(b)
			}
		case reflect.Float64, reflect.Float32:
			if emptyValue {
				if !isPtr {
					fieldValue.SetFloat(0)
				}
				break
			}
			fl, err := strconv.ParseFloat(s, 64)
			if err != nil {
				return err
			}
			if isPtr {
				fieldValue.Set(reflect.NewAt(fieldType.Elem(), (unsafe.Pointer)(&fl)))
			} else {
				fieldValue.SetFloat(fl)
			}
		case reflect.Struct, reflect.Map:
			ptr := reflect.New(fieldType)
			if emptyValue {
				if !isPtr {
					fieldValue.Set(ptr.Elem())
				}
				if k == reflect.Struct {
					if err := checker.Struct(ptr.Elem().Interface()); err != nil {
						return err
					}
				}
				break
			}
			// not empty value
			err = json.Unmarshal([]byte(s), ptr.Interface())
			if err != nil {
				return err
			}
			if k == reflect.Struct {
				if err := checker.Struct(ptr.Elem().Interface()); err != nil {
					return err
				}
			}
			fieldValue.Set(ptr.Elem())
		default:
			return fmt.Errorf("unrecongnized struct field type %s", fieldType.Kind())
		}
		return err
	}

	value := reflect.ValueOf(obj).Elem()
	for i := 0; i < t.NumField(); i++ {
		f := t.Field(i)
		if tv := f.Tag.Get("header"); tv != "" {
			if err := setFieldValue(f.Type, value.Field(i), p.Req.Header.Get(tv)); err != nil {
				return p.AddError(err.Error())
			}
			continue
		}
		if tv := f.Tag.Get("query"); tv != "" {
			var s string
			if err := p.Query(tv, &s).Error(); err != nil {
				return p
			}
			if err := setFieldValue(f.Type, value.Field(i), s); err != nil {
				return p.AddError(err.Error())
			}
			continue
		}
		if tv := f.Tag.Get("var"); tv != "" {
			var s string
			if err := p.Var(":"+tv, &s).Error(); err != nil {
				return p
			}
			if err := setFieldValue(f.Type, value.Field(i), s); err != nil {
				return p.AddError(err.Error())
			}
			continue
		}
		if tv := f.Tag.Get("body"); tv != "" {
			b, err := ioutil.ReadAll(p.Req.Body)
			if err != nil {
				return p.AddError(err.Error())
			}
			defer p.Req.Body.Close()

			if err := setFieldValue(f.Type, value.Field(i), string(b)); err != nil {
				return p.AddError(err.Error())
			}
			continue
		}
		return p.AddErrf("unrecognized struct field %#v", f)
	}
	if err := Validate(obj); err != nil {
		return p.AddError(err.Error())
	}
	return p
}

func (p *Param) SearchSuit(unit *SearchSuit, defaultPage, defaultSize int) *Param {
	var (
		rawFilters string
		rawQuerys  string
		rawOrderBy string
		fparts     []string
		qparts     []string
		oparts     []string
	)

	p.IntWithDefault("page", &unit.Page, defaultPage).IntWithDefault("size", &unit.Size, defaultSize)
	if len(p.errs) > 0 {
		return p
	}

	p.Query("filters", &rawFilters).Query("querys", &rawQuerys).Query("order_by", &rawOrderBy)
	if rawFilters != "" {
		fparts = strings.Split(rawFilters, ";")
	}
	if rawQuerys != "" {
		qparts = strings.Split(rawQuerys, ";")
	}
	if rawOrderBy != "" {
		oparts = strings.Split(rawOrderBy, ";")
	}

	for _, f := range fparts {
		if f == "" {
			continue
		}
		fv := strings.Split(f, ",")
		if len(fv) != 2 {
			return p.AddError("filter not valid")
		}
		if unit.Filters == nil {
			unit.Filters = make(map[string]interface{}, 1)
		}
		unit.Filters[fv[0]] = fv[1]
	}
	for _, q := range qparts {
		if q == "" {
			continue
		}
		qv := strings.Split(q, ",")
		if len(qv) != 2 {
			return p.AddError("querys not valid")
		}
		if unit.Querys == nil {
			unit.Querys = make(map[string]interface{}, 1)
		}
		unit.Querys[qv[0]] = qv[1]
	}
	for _, o := range oparts {
		if o == "" {
			continue
		}
		ov := strings.Split(o, ",")
		if len(ov) != 2 {
			return p.AddError("order_by not valid")
		}
		sort := strings.ToUpper(ov[1])
		if sort != "DESC" && sort != "ASC" {
			p.AddError("order_by sort method not valid")
		}
		unit.OrderBys = append(unit.OrderBys, ov)
	}

	return p
}

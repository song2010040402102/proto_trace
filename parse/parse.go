package parse

import (
	"bufio"
	"bytes"
	"encoding/json"
	"io"
	"os"
	"path/filepath"
	"reflect"
	"strconv"
	"strings"

	"github.com/astaxie/beego/logs"
	"github.com/golang/protobuf/jsonpb"
	"github.com/golang/protobuf/proto"

	_ "proto_trace/Lyingdragon_Protocol"
)

func removeLeftBlank(s string) string {
	for i := 0; i < len(s); i++ {
		if s[i] != ' ' && s[i] != '	' {
			s = s[i:]
			break
		}
	}
	return s
}

func removeRightBlank(s string) string {
	for i := len(s) - 1; i >= 0; i-- {
		if s[i] != ' ' && s[i] != '	' {
			s = s[:i+1]
			break
		}
	}
	return s
}

func removeSideBlank(s string) string {
	return removeRightBlank(removeLeftBlank(s))
}

func walkDir(dirPth, suffix string) (files []string, err error) {
	suffix = strings.ToUpper(suffix)
	err = filepath.Walk(dirPth, func(filename string, fi os.FileInfo, err error) error {
		if err != nil || fi.IsDir() {
			return err
		}
		if strings.HasSuffix(strings.ToUpper(fi.Name()), suffix) {
			files = append(files, filename)
		}
		return nil
	})
	return files, err
}

func readFile(filename string) ([]string, error) {
	var lines []string
	f, err := os.Open(filename)
	if err != nil {
		logs.Error(filename, "Open error!", err)
		return lines, err
	}
	defer f.Close()

	count := 0
	rd := bufio.NewReader(f)
	for {
		line, err := rd.ReadString('\n')
		if len(line) > 0 && line[len(line)-1] == '\n' {
			line = line[:len(line)-1]
			if len(line) > 0 && line[len(line)-1] == '\r' {
				line = line[:len(line)-1]
			}
		}
		if index := strings.Index(line, "//"); index >= 0 {
			line = line[:index]
		} else if index = strings.Index(line, "/*"); index >= 0 {
			count++
			line = line[:index]
		} else if index = strings.Index(line, "*/"); index >= 0 && count > 0 {
			count--
			if index < len(line)-2 {
				line = line[index+2:]
			} else {
				line = ""
			}
		}
		line = removeSideBlank(line)
		if count == 0 && len(line) > 0 {
			lines = append(lines, line)
		}
		if err == io.EOF {
			break
		}
		if err != nil {
			logs.Error(filename, "Read error!", err)
			return []string{}, err
		}
	}
	return lines, nil
}

func initProto(msg proto.Message) {
	if msg == nil {
		return
	}
	rType := reflect.TypeOf(msg)
	rVal := reflect.ValueOf(msg)
	if rType.Kind() == reflect.Ptr {
		rType = rType.Elem()
		rVal = rVal.Elem()
	} else {
		return
	}
	for i := 0; i < rType.NumField(); i++ {
		t := rType.Field(i)
		v := rVal.Field(i)
		if t.Name == "Type" {
			continue
		}
		if t.Type.Kind() == reflect.Slice {
			//数组默认填充一个元素，因为通常请求参数中若含数组的话，一般长度不为0，若长度必须为0，则手动删除
			if t.Type.Elem().Kind() == reflect.Ptr {
				e := reflect.New(t.Type.Elem().Elem())
				initProto(e.Interface().(proto.Message))
				v.Set(reflect.Append(v, e))
			} else {
				v.Set(reflect.MakeSlice(t.Type, 1, 1))
			}
		} else if t.Type.Kind() == reflect.Ptr {
			switch t.Type.Elem().Kind() {
			case reflect.Bool:
				v.Set(reflect.ValueOf(new(bool)))
			case reflect.Int:
				v.Set(reflect.ValueOf(new(int)))
			case reflect.Int8:
				v.Set(reflect.ValueOf(new(int8)))
			case reflect.Int16:
				v.Set(reflect.ValueOf(new(int16)))
			case reflect.Int32:
				v.Set(reflect.ValueOf(new(int32)))
			case reflect.Int64:
				v.Set(reflect.ValueOf(new(int64)))
			case reflect.Uint:
				v.Set(reflect.ValueOf(new(uint)))
			case reflect.Uint8:
				v.Set(reflect.ValueOf(new(uint8)))
			case reflect.Uint16:
				v.Set(reflect.ValueOf(new(uint16)))
			case reflect.Uint32:
				v.Set(reflect.ValueOf(new(uint32)))
			case reflect.Uint64:
				v.Set(reflect.ValueOf(new(uint64)))
			case reflect.Float32:
				v.Set(reflect.ValueOf(new(float32)))
			case reflect.Float64:
				v.Set(reflect.ValueOf(new(float64)))
			case reflect.String:
				v.Set(reflect.ValueOf(new(string)))
			case reflect.Struct:
				{
					msg = reflect.New(t.Type.Elem()).Interface().(proto.Message)
					initProto(msg)
					v.Set(reflect.ValueOf(msg))
				}
			}
		}
	}
}

func ParseProto(dir string) bool {
	g_id2type = make(map[int32]string)
	g_type2name = make(map[string]string)
	filenames, _ := walkDir(dir, ".proto")
	for _, filename := range filenames {
		lines, err := readFile(filename)
		if err != nil {
			logs.Error(filename, "read error!")
			return false
		}
		for i := 0; i < len(lines); i++ {
			if len(lines[i]) >= 3 && (lines[i][:3] == "C2S" || lines[i][:3] == "S2C") ||
				len(lines[i]) >= 5 && (lines[i][:5] == "C_2_S" || lines[i][:5] == "S_2_C") {
				ss := strings.Split(lines[i], "=")
				if len(ss) == 2 {
					s := removeLeftBlank(ss[1])
					s = s[:len(s)-1]
					i, err := strconv.ParseInt(s, 10, 32)
					if err == nil {
						g_id2type[int32(i)] = removeRightBlank(ss[0])
					}
				}
			} else if len(lines[i]) >= 7 && lines[i][:7] == "message" {
				count := 1
				s := removeLeftBlank(lines[i][7:])
				s = removeRightBlank(s[:len(s)-1])
				for i++; i < len(lines); i++ {
					if strings.Contains(lines[i], "type") {
						str := strings.Replace(lines[i], " ", "", -1)
						str = strings.Replace(str, "	", "", -1)
						ss := strings.Split(str, "=")
						if len(ss) == 3 {
							g_type2name[ss[2][:len(ss[2])-2]] = s
						}
					} else if lines[i][len(lines[i])-1] == '{' {
						count++
					} else if lines[i][0] == '}' {
						count--
						if count == 0 {
							break
						}
					}
				}
			}
		}
	}
	return true
}

func GetMessageNameById(id int32) string {
	return g_type2name[g_id2type[id]]
}

func GetMessageObjectById(id int32) proto.Message {
	msgName := GetMessageNameById(id)
	msgType := proto.MessageType("Lyingdragon.Protocol." + msgName)
	if msgType != nil {
		return reflect.New(msgType.Elem()).Interface().(proto.Message)
	}
	return nil
}

func GetJsonByMsgId(id int32) string {
	msgObj := GetMessageObjectById(id)
	initProto(msgObj)
	strJson, _ := new(jsonpb.Marshaler).MarshalToString(msgObj)
	var buff bytes.Buffer
	json.Indent(&buff, []byte(strJson), "", "    ")
	return buff.String()
}

func Json2Message(id int32, js string) proto.Message {
	var buff bytes.Buffer
	err := json.Compact(&buff, []byte(js))
	if err != nil {
		logs.Error("Compact", err)
		return nil
	}
	msgObj := GetMessageObjectById(id)
	err = jsonpb.UnmarshalString(buff.String(), msgObj)
	if err != nil {
		logs.Error("UnmarshalString", err)
		return nil
	}
	return msgObj
}

func GetAllMsgType() map[int32]string {
	return g_id2type
}

func GetAllMsgName() map[string]string {
	return g_type2name
}

var g_id2type map[int32]string
var g_type2name map[string]string

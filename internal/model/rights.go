package model

import (
	"fmt"
	"strings"

	"go.mongodb.org/mongo-driver/bson/bsontype"
	"go.mongodb.org/mongo-driver/x/bsonx/bsoncore"
)

// Rights 是角色的权限集合（一组 authority.key）。
// 旧库里 rights 字段有两种存法：逗号分隔的字符串，或字符串数组；
// 本类型在解码时统一归一化为 []string，写库时序列化为数组（对齐旧 PHP 的 array_values）。
type Rights []string

// UnmarshalBSONValue 兼容 string / array / null 三种存法。
func (r *Rights) UnmarshalBSONValue(t bsontype.Type, b []byte) error {
	switch t {
	case bsontype.String:
		s, _, ok := bsoncore.ReadString(b)
		if !ok {
			return fmt.Errorf("rights: 非法字符串")
		}
		*r = SplitCSV(s)
	case bsontype.Array:
		arr, _, ok := bsoncore.ReadArray(b)
		if !ok {
			return fmt.Errorf("rights: 非法数组")
		}
		vals, err := arr.Values()
		if err != nil {
			return err
		}
		out := make(Rights, 0, len(vals))
		for _, v := range vals {
			if s, ok := v.StringValueOK(); ok {
				s = strings.TrimSpace(s)
				if s != "" {
					out = append(out, s)
				}
			}
		}
		*r = out
	case bsontype.Null, bsontype.Undefined:
		*r = nil
	default:
		*r = nil
	}
	return nil
}

// String 以逗号连接（前端 permissions 字段为逗号分隔字符串）。
func (r Rights) String() string { return strings.Join(r, ",") }

// Set 返回便于查找的集合。
func (r Rights) Set() map[string]struct{} {
	m := make(map[string]struct{}, len(r))
	for _, k := range r {
		if k != "" {
			m[k] = struct{}{}
		}
	}
	return m
}

// SplitCSV 把逗号分隔字符串拆成去空的字符串切片。
func SplitCSV(s string) Rights {
	parts := strings.Split(s, ",")
	out := make(Rights, 0, len(parts))
	for _, p := range parts {
		p = strings.TrimSpace(p)
		if p != "" {
			out = append(out, p)
		}
	}
	return out
}

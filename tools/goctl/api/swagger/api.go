package swagger

import (
	"fmt"
	"strings"

	"github.com/zeromicro/go-zero/tools/goctl/api/spec"
)

func fillAllStructs(api *spec.ApiSpec) {
	var (
		tps         []spec.Type
		structTypes = make(map[string]spec.DefineStruct)
		groups      []spec.Group
	)

	// bytes2, _ := json.Marshal(api.Types)
	// fmt.Printf("bytes2 %+v\n", string(bytes2))
	for _, tp := range api.Types {
		structTypes[tp.Name()] = tp.(spec.DefineStruct)
	}

	for _, tp := range api.Types {
		filledTP := fillStruct(tp, structTypes, "")
		tps = append(tps, filledTP)
		structTypes[filledTP.Name()] = filledTP.(spec.DefineStruct)
	}

	// bytes1, _ := json.Marshal(structTypes)
	// fmt.Printf("bytes1 %+v\n", string(bytes1))

	for _, group := range api.Service.Groups {
		routes := make([]spec.Route, 0, len(group.Routes))
		for _, route := range group.Routes {
			route.RequestType = fillStruct(route.RequestType, structTypes, "")
			route.ResponseType = fillStruct(route.ResponseType, structTypes, "")
			routes = append(routes, route)
		}
		group.Routes = routes
		groups = append(groups, group)
	}
	api.Service.Groups = groups
	api.Types = tps
}

func contains(list []string, s string) bool {
	for _, v := range list {
		if v == s {
			return true
		}
	}
	return false
}
func CheckContains(parents []string, name string) {
	if contains(parents, name) {
		// 过滤掉空字符串
		var filtered []string
		for _, p := range parents {
			if p != "" {
				filtered = append(filtered, p)
			}
		}
		panic(fmt.Sprintf("结构体存在递归引用 %+v %v", strings.Join(filtered, "->"), name))
	}
}

func fillStruct(tp spec.Type, allTypes map[string]spec.DefineStruct, parents ...string) spec.Type {
	switch val := tp.(type) {
	case spec.DefineStruct:
		var members []spec.Member
		// 先把空结构的成员填充一下
		if len(val.Members) == 0 {
			st, ok := allTypes[val.RawName]
			if ok {
				val.Members = st.Members
			}
		}
		// 再递归填充其子成员
		for _, member := range val.Members {
			switch memberType := member.Type.(type) {
			case spec.PointerType:
				member.Type = spec.PointerType{
					RawName: memberType.RawName,
					Type:    fillStruct(memberType.Type, allTypes, append(parents, val.Name())...),
				}
			case spec.ArrayType:
				member.Type = spec.ArrayType{
					RawName: memberType.RawName,
					Value:   fillStruct(memberType.Value, allTypes, append(parents, val.Name())...),
				}
			case spec.MapType:
				member.Type = spec.MapType{
					RawName: memberType.RawName,
					Key:     memberType.Key,
					Value:   fillStruct(memberType.Value, allTypes, append(parents, val.Name())...),
				}
			case spec.DefineStruct:
				CheckContains(append(parents, val.Name()), memberType.Name())
				if st, ok := allTypes[memberType.Name()]; ok {
					member.Type = fillStruct(st, allTypes, append(parents, val.Name())...)
				}
			case spec.NestedStruct:
				member.Type = fillStruct(member.Type, allTypes, append(parents, val.Name())...)
			}
			members = append(members, member)
		}
		val.Members = members
		return val
	case spec.NestedStruct:
		members := make([]spec.Member, 0, len(val.Members))
		for _, member := range val.Members {
			switch memberType := member.Type.(type) {
			case spec.PointerType:
				member.Type = spec.PointerType{
					RawName: memberType.RawName,
					Type:    fillStruct(memberType.Type, allTypes, append(parents, val.Name())...),
				}
			case spec.ArrayType:
				member.Type = spec.ArrayType{
					RawName: memberType.RawName,
					Value:   fillStruct(memberType.Value, allTypes, append(parents, val.Name())...),
				}
			case spec.MapType:
				member.Type = spec.MapType{
					RawName: memberType.RawName,
					Key:     memberType.Key,
					Value:   fillStruct(memberType.Value, allTypes, append(parents, val.Name())...),
				}
			case spec.DefineStruct:
				CheckContains(append(parents, val.Name()), memberType.Name())
				if st, ok := allTypes[memberType.Name()]; ok {
					member.Type = fillStruct(st, allTypes, append(parents, val.Name())...)
				}
			case spec.NestedStruct:
				member.Type = fillStruct(member.Type, allTypes, append(parents, val.Name())...)
			}
			members = append(members, member)
		}
		if len(members) == 0 {
			st, ok := allTypes[val.RawName]
			if ok {
				members = st.Members
			}
		}
		val.Members = members
		return val
	case spec.PointerType:
		return spec.PointerType{
			RawName: val.RawName,
			Type:    fillStruct(val.Type, allTypes, parents...),
		}
	case spec.ArrayType:
		return spec.ArrayType{
			RawName: val.RawName,
			Value:   fillStruct(val.Value, allTypes, parents...),
		}
	case spec.MapType:
		return spec.MapType{
			RawName: val.RawName,
			Key:     val.Key,
			Value:   fillStruct(val.Value, allTypes, parents...),
		}
	default:
		return tp
	}
}

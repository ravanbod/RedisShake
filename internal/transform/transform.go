package transform

import (
	"RedisShake/internal/config"
	"RedisShake/internal/entry"
	"RedisShake/internal/log"
	lua "github.com/yuin/gopher-lua"
	"strings"
)

const (
	Allow    = 0
	Disallow = 1
	Error    = 2
)

var luaInstance *lua.LState

func Init() {
	luaString := config.Opt.Transform
	luaString = strings.TrimSpace(luaString)
	if len(luaString) == 0 {
		log.Infof("no transform script")
		return
	}
	luaInstance = lua.NewState()
	err := luaInstance.DoString(luaString)
	if err != nil {
		log.Panicf("load transform script failed: %v", err)
	}
	log.Infof("load transform script success")
}

func Transform(e *entry.Entry) int {
	if luaInstance == nil {
		return Allow
	}

	keys := luaInstance.NewTable()
	for _, key := range e.Keys {
		keys.Append(lua.LString(key))
	}

	slots := luaInstance.NewTable()
	for _, slot := range e.Slots {
		slots.Append(lua.LNumber(slot))
	}

	f := luaInstance.GetGlobal("filter")
	luaInstance.Push(f)
	luaInstance.Push(lua.LString(e.Group))   // group
	luaInstance.Push(lua.LString(e.CmdName)) // cmd name
	luaInstance.Push(keys)                   // keys
	luaInstance.Push(slots)                  // slots
	luaInstance.Push(lua.LNumber(e.DbId))    // dbid

	luaInstance.Call(8, 2)

	code := int(luaInstance.Get(1).(lua.LNumber))
	e.DbId = int(luaInstance.Get(2).(lua.LNumber))
	luaInstance.Pop(2)
	return code
}
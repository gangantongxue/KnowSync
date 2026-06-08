package service

import (
	"strconv"
	"strings"
)

// GetPrivateConversationID 生成私聊会话ID（排序后拼接）.
func GetPrivateConversationID(id1, id2 string) string {
	id1Int, err1 := strconv.ParseUint(id1, 10, 64)
	id2Int, err2 := strconv.ParseUint(id2, 10, 64)
	if err1 == nil && err2 == nil {
		if id1Int < id2Int {
			return id1 + "_" + id2
		}
		return id2 + "_" + id1
	}
	if strings.Compare(id1, id2) < 0 {
		return id1 + "_" + id2
	}
	return id2 + "_" + id1
}

package info

import (
	"sync"
)

var blogMapInfoInstance *BlogMapInfo = nil
var blogMapInfoOnce sync.Once

type BlogInfo struct {
	Name string
	Id   int
	Sort string
}

type BlogMapInfo struct {
	netBlogMap   map[int]*BlogInfo
	localBlogMap map[int]*BlogInfo
}

func ShareBlogMapInstance() *BlogMapInfo {
	blogMapInfoOnce.Do(func() {
		blogMapInfoInstance = &BlogMapInfo{}
	})
	return blogMapInfoInstance
}

func (b *BlogMapInfo) AddNetBlog(info *BlogInfo) {
	if b.netBlogMap == nil {
		b.netBlogMap = make(map[int]*BlogInfo)
	}
	b.netBlogMap[info.Id] = info
}

func (b *BlogMapInfo) ClearNetBlogMap() {
	b.netBlogMap = nil
}

func (b *BlogMapInfo) QueryNetBlogInfo(blogId int) (string, string) {
	if v, ok := b.netBlogMap[blogId]; ok {
		return v.Name, v.Sort
	}
	return "", ""
}

func (b *BlogMapInfo) AddLocalBlog(info *BlogInfo) {
	if b.localBlogMap == nil {
		b.localBlogMap = make(map[int]*BlogInfo)
	}
	b.localBlogMap[info.Id] = info
}

func (b *BlogMapInfo) ClearLocalBlogMap() {
	b.localBlogMap = nil
}

func (b *BlogMapInfo) QueryLocalBlogInfo(index int) (string, string) {
	if v, ok := b.localBlogMap[index]; ok {
		return v.Name, v.Sort
	}
	return "", ""
}

func (b *BlogMapInfo) DeleteLocalBlogInfo(index int) {
	if _, ok := b.localBlogMap[index]; ok {
		delete(b.localBlogMap, index)
	}
}

package service

import (
	"encoding/json"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"sync"
	"time"

	"kineticgo/internal/model"
	"kineticgo/internal/repository"
)

// cookieEntry 用于 JSON 持久化的 cookie 精简结构.
// 只保留重建 http.Cookie 所必需的关键字段.
type cookieEntry struct {
	Name     string    `json:"name"`
	Value    string    `json:"value"`
	Domain   string    `json:"domain"`
	Path     string    `json:"path"`
	Expires  time.Time `json:"expires,omitempty"`
	Secure   bool      `json:"secure"`
	HttpOnly bool      `json:"httpOnly"`
}

type PersistentJar struct {
	inner http.CookieJar
	// raw 按 "domain|path|name" 做主键去重,保留完整 cookie 信息
	raw map[string]*http.Cookie
	mu  sync.Mutex
}

func NewPersistentJar() *PersistentJar {
	inner, _ := cookiejar.New(nil)
	return &PersistentJar{
		inner: inner,
		raw:   make(map[string]*http.Cookie),
	}
}

// SetCookies 委托给内部 jar 处理实际的 cookie 管理,同时把完整 cookie 信息记到 raw map.
func (p *PersistentJar) SetCookies(u *url.URL, cookies []*http.Cookie) {
	p.inner.SetCookies(u, cookies)

	p.mu.Lock()
	defer p.mu.Unlock()
	for _, c := range cookies {
		// cookie 里没显式 domain/path 时, 按 RFC 6265 用请求 URL 的对应字段
		domain := c.Domain
		if domain == "" {
			domain = u.Host
		}
		path := c.Path
		if path == "" {
			path = "/"
		}
		// Clone 一份避免外部后续修改污染内部状态
		clone := *c
		clone.Domain = domain
		clone.Path = path
		key := domain + "|" + path + "|" + c.Name
		p.raw[key] = &clone
	}
}

// Cookies 委托给内部 jar,行为完全和 cookiejar.Jar 一致.
func (p *PersistentJar) Cookies(u *url.URL) []*http.Cookie {
	return p.inner.Cookies(u)
}

// Snapshot 导出所有未过期 cookie 的精简快照,可直接 JSON 序列化.
func (p *PersistentJar) Snapshot() []cookieEntry {
	p.mu.Lock()
	defer p.mu.Unlock()

	now := time.Now()
	result := make([]cookieEntry, 0, len(p.raw))
	for _, c := range p.raw {
		// 过滤已过期的 cookie (Expires 为零值表示 session cookie, 也保留)
		if !c.Expires.IsZero() && c.Expires.Before(now) {
			continue
		}
		result = append(result, cookieEntry{
			Name:     c.Name,
			Value:    c.Value,
			Domain:   c.Domain,
			Path:     c.Path,
			Expires:  c.Expires,
			Secure:   c.Secure,
			HttpOnly: c.HttpOnly,
		})
	}
	return result
}

// Load 从快照恢复 cookie 到 jar (同时刷新 raw map 和 inner jar).
func (p *PersistentJar) Load(entries []cookieEntry) {
	for _, e := range entries {
		path := e.Path
		if path == "" {
			path = "/"
		}
		u := &url.URL{Scheme: "https", Host: e.Domain, Path: path}
		c := &http.Cookie{
			Name:     e.Name,
			Value:    e.Value,
			Domain:   e.Domain,
			Path:     path,
			Expires:  e.Expires,
			Secure:   e.Secure,
			HttpOnly: e.HttpOnly,
		}
		// 走 SetCookies 路径, raw map 会被自动更新
		p.SetCookies(u, []*http.Cookie{c})
	}
}

// parseCookiesFromConfig 从 schedule.Config JSON 中解析 cookies 字段.
func parseCookiesFromConfig(raw []byte) []cookieEntry {
	if len(raw) == 0 {
		return nil
	}
	var wrapper struct {
		Cookies []cookieEntry `json:"cookies"`
	}
	if err := json.Unmarshal(raw, &wrapper); err != nil {
		return nil
	}
	return wrapper.Cookies
}

// buildConfigWithCookies 将 cookies 合并回 config map,返回新的 JSON 字节流.
func buildConfigWithCookies(raw []byte, entries []cookieEntry) ([]byte, error) {
	var config map[string]any
	if len(raw) > 0 {
		if err := json.Unmarshal(raw, &config); err != nil {
			return nil, err
		}
	} else {
		config = make(map[string]any)
	}
	config["cookies"] = entries
	return json.Marshal(config)
}

// saveCookiesToSchedule 把 cookie 快照持久化到 schedule 的 Config 字段.
func saveCookiesToSchedule(
	repo *repository.TaskRepository,
	schedule *model.TaskSchedule,
	entries []cookieEntry,
) error {
	newConfig, err := buildConfigWithCookies([]byte(schedule.Config), entries)
	if err != nil {
		return err
	}
	schedule.Config = model.TempleConfig(newConfig)
	return repo.UpdateTaskSchedule(schedule)
}

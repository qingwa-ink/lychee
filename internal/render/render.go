// Package render 封装页面渲染：html/template 渲染页面骨架，Jet 渲染复杂片段，
// 统一注入当前 locale 的 i18n 文案（见 doc/项目设计文档.md §7）。
package render

import (
	"bytes"
	"encoding/json"
	"fmt"
	"html/template"
	"io"
	"path/filepath"

	"github.com/CloudyKit/jet"
	"github.com/qingwa-ink/lychee/internal/pkg/i18n"
)

// Renderer 持有解析好的模板与 Jet 引擎。
type Renderer struct {
	templates     *template.Template
	jetSet        *jet.Set
	i18n          *i18n.Store
	assetsVersion string // 静态资源版本号，拼到 /static/* 的查询串上做缓存击穿
}

// New 解析 templatesDir 下的 *.html，并初始化 jetDir 的 Jet 引擎。
// assetsVersion 用于静态资源 URL 缓存击穿（每次发版/重启变化即可）。
func New(templatesDir, jetDir string, store *i18n.Store, assetsVersion string) (*Renderer, error) {
	tmpl, err := template.ParseGlob(filepath.Join(templatesDir, "*.html"))
	if err != nil {
		return nil, fmt.Errorf("parse templates: %w", err)
	}
	return &Renderer{
		templates:     tmpl,
		jetSet:        jet.NewHTMLSet(jetDir),
		i18n:          store,
		assetsVersion: assetsVersion,
	}, nil
}

// PageData 传给页面模板的数据。
type PageData struct {
	Locale  string            // 当前语种 zh/en
	I18n    map[string]string // 当前语种全量文案
	Page    string            // 当前页标识（nav 高亮、前端路由用）
	Title   string            // 浏览器标题文案 key（模板用 index .I18n .Title）
	Content template.HTML     // 已渲染的内容 HTML（两段式：先渲染内容，再套 layout）
	ShowNav bool              // 是否显示导航
	Scripts []string          // 页面 JS 路径
	Data    any               // 各页自定义数据

	// 预序列化 JSON，注入到布局供前端 JS 复用（locale cookie 为 HttpOnly，JS 读不到）
	LocaleJSON template.JS
	I18NJSON   template.JS

	// AssetsVersion 静态资源版本号，layout 中拼到 /static/* 的 ?v= 上做缓存击穿。
	AssetsVersion string
}

// Render 渲染整页到 w。contentTmpl 为内容模板名（如 page_login），先单独渲染成 HTML，
// 再注入 layout。html/template 的 {{template}} 只接受字面量名字，故采用两段式。
func (r *Renderer) Render(w io.Writer, locale, contentTmpl string, data *PageData) error {
	data.Locale = locale
	data.I18n = r.i18n.Messages(locale)
	data.AssetsVersion = r.assetsVersion
	if b, err := json.Marshal(locale); err == nil {
		data.LocaleJSON = template.JS(b)
	}
	if b, err := json.Marshal(data.I18n); err == nil {
		data.I18NJSON = template.JS(b)
	}

	// 第一段：渲染内容模板（如 page_login / page_app）
	var buf bytes.Buffer
	if err := r.templates.ExecuteTemplate(&buf, contentTmpl, data); err != nil {
		return fmt.Errorf("render content %s: %w", contentTmpl, err)
	}
	data.Content = template.HTML(buf.String())

	// 第二段：套用布局
	if err := r.templates.ExecuteTemplate(w, "layout", data); err != nil {
		return fmt.Errorf("render layout: %w", err)
	}
	return nil
}

// Fragment 渲染 Jet 片段为 HTML，供注入页面（任务看板/图表等复杂区块）。
func (r *Renderer) Fragment(name, locale string, data any) (template.HTML, error) {
	t, err := r.jetSet.GetTemplate(name)
	if err != nil {
		return "", fmt.Errorf("jet get %s: %w", name, err)
	}
	vars := jet.VarMap{}
	vars.Set("i18n", r.i18n.Messages(locale))
	vars.Set("locale", locale)
	var buf bytes.Buffer
	if err := t.Execute(&buf, vars, data); err != nil {
		return "", fmt.Errorf("jet exec %s: %w", name, err)
	}
	return template.HTML(buf.String()), nil
}

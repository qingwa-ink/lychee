// Package i18n 提供多语言文案库与语言解析（见 doc/项目设计文档.md §3.1）。
package i18n

import (
	"strings"
)

// 内置语种
const (
	LangZH = "zh"
	LangEN = "en"
)

// Store 文案库。
type Store struct {
	defaultLang string
	catalogs    map[string]map[string]string
}

// New 构造文案库，默认语种由传入值决定。
func New(defaultLang string) *Store {
	s := &Store{
		defaultLang: normalize(defaultLang),
		catalogs: map[string]map[string]string{
			LangZH: zhCatalog,
			LangEN: enCatalog,
		},
	}
	if _, ok := s.catalogs[s.defaultLang]; !ok {
		s.defaultLang = LangZH
	}
	return s
}

// Default 返回默认语种。
func (s *Store) Default() string { return s.defaultLang }

// Languages 返回所有可用语种。
func (s *Store) Languages() []string {
	out := make([]string, 0, len(s.catalogs))
	for lang := range s.catalogs {
		out = append(out, lang)
	}
	return out
}

// Exists 判断语种是否存在。
func (s *Store) Exists(lang string) bool {
	_, ok := s.catalogs[normalize(lang)]
	return ok
}

// Messages 返回某语种的全量文案（不存在时回退到默认语种）。
func (s *Store) Messages(lang string) map[string]string {
	if m, ok := s.catalogs[normalize(lang)]; ok {
		return m
	}
	return s.catalogs[s.defaultLang]
}

// T 翻译单个 key（不存在时回退到默认语种，再找不到则原样返回 key）。
func (s *Store) T(lang, key string) string {
	if m, ok := s.catalogs[normalize(lang)]; ok {
		if v, ok := m[key]; ok {
			return v
		}
	}
	if v, ok := s.catalogs[s.defaultLang][key]; ok {
		return v
	}
	return key
}

// Resolve 按优先级解析语种：显式 cookie > Accept-Language > 默认。
func (s *Store) Resolve(cookieVal, acceptLang string) string {
	if cookieVal != "" && s.Exists(cookieVal) {
		return normalize(cookieVal)
	}
	for _, l := range parseAcceptLanguage(acceptLang) {
		if s.Exists(l) {
			return normalize(l)
		}
	}
	return s.defaultLang
}

func normalize(lang string) string {
	return strings.ToLower(strings.TrimSpace(lang))
}

// parseAcceptLanguage 简单解析 Accept-Language，返回主语种列表（忽略 q 值顺序细节）。
// 例如 "zh-CN,zh;q=0.9,en;q=0.8" -> ["zh","zh","en"]。
func parseAcceptLanguage(s string) []string {
	var out []string
	for _, part := range strings.Split(s, ",") {
		part = strings.TrimSpace(part)
		if part == "" {
			continue
		}
		if i := strings.Index(part, ";"); i >= 0 {
			part = part[:i]
		}
		part = strings.TrimSpace(part)
		if i := strings.Index(part, "-"); i >= 0 {
			part = part[:i]
		}
		part = strings.ToLower(part)
		if part != "" {
			out = append(out, part)
		}
	}
	return out
}

var zhCatalog = map[string]string{
	"app.name":            "荔枝小秘书",
	"common.save":         "保存",
	"common.cancel":       "取消",
	"common.delete":       "删除",
	"common.edit":         "编辑",
	"common.copy":         "复制",
	"common.confirm":      "确认",
	"common.back":         "返回",
	"common.search":       "搜索",
	"common.create":       "新建",
	"common.created_at":   "创建时间",
	"common.updated_at":   "修改时间",
	"auth.login":          "登录",
	"auth.register":       "注册",
	"auth.logout":         "登出",
	"auth.email":          "邮箱",
	"auth.password":       "密码",
	"auth.forgot_password": "忘记密码",
	"auth.send_code":      "发送验证码",
	"nav.tasks":           "任务",
	"nav.phrases":         "常用语",
	"nav.checkin":         "打卡",
	"nav.logs":            "日志",
	"nav.settings":        "设置",
	"task.priority":       "优先级",
	"task.status":         "状态",
	"task.due_date":       "截止日期",
	"task.status_editing": "编辑中",
	"task.status_pending": "待执行",
	"task.status_done":    "已完成",

	// 页面与通用 UI（F1 前端）
	"app.tagline":           "您的第一个 AI 编程助手",
	"common.loading":        "加载中…",
	"common.error":          "出错了",
	"common.retry":          "重试",
	"common.optional":       "可选",
	"common.operation":      "操作",
	"common.empty":          "暂无数据",
	"nav.locale_label":      "语言",
	"nav.home":              "首页",
	"auth.code":             "邮箱验证码",
	"auth.new_password":     "新密码",
	"auth.old_password":     "旧密码",
	"auth.confirm_password": "确认密码",
	"auth.login_title":      "欢迎回来",
	"auth.register_title":   "创建账号",
	"auth.forgot_title":     "重置密码",
	"auth.no_account":       "还没有账号？立即注册",
	"auth.have_account":     "已有账号？去登录",
	"auth.back_to_login":    "返回登录",
	"auth.reset_done":       "密码已重置，请重新登录",
	"auth.register_done":    "注册成功，请登录",
	"auth.session_expired":  "登录已过期，请重新登录",
	"auth.ratelimited":      "操作过于频繁，请稍后再试",

	// 设置与常用语页（F1.2）
	"settings.account":         "账号",
	"settings.language":        "语言",
	"settings.change_password": "修改密码",
	"settings.use_old_password": "用旧密码验证",
	"settings.use_code":         "用验证码验证",
	"settings.pw_changed":      "密码修改成功，请重新登录",
	"phrases.new":              "新建常用语",
	"phrases.content":          "内容",
	"phrases.placeholder":      "输入常用语内容…",
	"phrases.empty":            "暂无常用语",
	"phrases.saved":            "已保存",
	"phrases.deleted":          "已删除",
	"common.prev":              "上一页",
	"common.next":              "下一页",
	"common.confirm_delete":    "确认删除？",

	// 任务页（F1.3）
	"task.add_group":           "新建分组",
	"task.add_subgroup":        "新建子分组",
	"task.all":                 "全部任务",
	"task.sort":                "排序",
	"task.sort_created":        "按创建时间",
	"task.sort_updated":        "按修改时间",
	"task.status_completed":    "已完成",
	"task.order_asc":           "升序",
	"task.order_desc":          "降序",
	"task.new_task":            "新建任务",
	"task.edit_task":           "编辑任务",
	"task.content":             "内容",
	"task.insert_phrase":       "插入常用语…",
	"task.no_group":            "请先选择或新建分组",
	"task.copy_done":           "复制成功",
	"task.confirm_delete_group": "删除分组将级联删除其下所有子分组与任务，确认？",
	"task.confirm_delete_task": "确认删除该任务？",

	// 打卡与健康（F1.4）
	"checkin.record":      "记一笔",
	"checkin.type":        "类型",
	"checkin.value":       "数值",
	"checkin.date":        "日期",
	"checkin.report":      "每日报告",
	"checkin.target":      "每日目标",
	"checkin.set_target":  "设置目标",
	"checkin.achievement": "完成度",
	"checkin.records":     "打卡记录",
	"checkin.no_goal":     "未设目标",
	"checkin.recorded":    "已打卡",
	"checkin.water":       "喝水",
	"checkin.exercise":    "起身运动",
	"checkin.nap":         "午休",

	// 操作日志与图表（F1.5）
	"logs.chart_title":  "操作频率",
	"logs.by_day":       "按天",
	"logs.by_hour":      "按小时",
	"logs.operations":   "操作历史",
	"logs.logins":       "登录日志",
	"logs.category":     "类别",
	"logs.cat_login":    "登录",
	"logs.cat_operation": "操作",
	"logs.action":       "动作",
	"logs.path":         "路径",
	"logs.ip":           "IP",
	"logs.ua":           "客户端",
}

var enCatalog = map[string]string{
	"app.name":            "Lychee Assistant",
	"common.save":         "Save",
	"common.cancel":       "Cancel",
	"common.delete":       "Delete",
	"common.edit":         "Edit",
	"common.copy":         "Copy",
	"common.confirm":      "Confirm",
	"common.back":         "Back",
	"common.search":       "Search",
	"common.create":       "New",
	"common.created_at":   "Created",
	"common.updated_at":   "Updated",
	"auth.login":          "Login",
	"auth.register":       "Register",
	"auth.logout":         "Logout",
	"auth.email":          "Email",
	"auth.password":       "Password",
	"auth.forgot_password": "Forgot Password",
	"auth.send_code":      "Send Code",
	"nav.tasks":           "Tasks",
	"nav.phrases":         "Phrases",
	"nav.checkin":         "Check-in",
	"nav.logs":            "Logs",
	"nav.settings":        "Settings",
	"task.priority":       "Priority",
	"task.status":         "Status",
	"task.due_date":       "Due Date",
	"task.status_editing": "Editing",
	"task.status_pending": "Pending",
	"task.status_done":    "Completed",

	// Page & common UI (F1 frontend)
	"app.tagline":           "Your first AI coding assistant",
	"common.loading":        "Loading…",
	"common.error":          "Error",
	"common.retry":          "Retry",
	"common.optional":       "optional",
	"common.operation":      "Actions",
	"common.empty":          "No data",
	"nav.locale_label":      "Language",
	"nav.home":              "Home",
	"auth.code":             "Email Code",
	"auth.new_password":     "New Password",
	"auth.old_password":     "Old Password",
	"auth.confirm_password": "Confirm Password",
	"auth.login_title":      "Welcome Back",
	"auth.register_title":   "Create Account",
	"auth.forgot_title":     "Reset Password",
	"auth.no_account":       "No account? Register",
	"auth.have_account":     "Have an account? Login",
	"auth.back_to_login":    "Back to Login",
	"auth.reset_done":       "Password reset, please login",
	"auth.register_done":    "Registered, please login",
	"auth.session_expired":  "Session expired, please login again",
	"auth.ratelimited":      "Too many requests, please try later",

	// Settings & phrases pages (F1.2)
	"settings.account":         "Account",
	"settings.language":        "Language",
	"settings.change_password": "Change Password",
	"settings.use_old_password": "Verify with old password",
	"settings.use_code":         "Verify with code",
	"settings.pw_changed":      "Password changed, please login again",
	"phrases.new":              "New Phrase",
	"phrases.content":          "Content",
	"phrases.placeholder":      "Enter phrase content…",
	"phrases.empty":            "No phrases",
	"phrases.saved":            "Saved",
	"phrases.deleted":          "Deleted",
	"common.prev":              "Prev",
	"common.next":              "Next",
	"common.confirm_delete":    "Confirm delete?",

	// Tasks page (F1.3)
	"task.add_group":           "New Group",
	"task.add_subgroup":        "New Subgroup",
	"task.all":                 "All Tasks",
	"task.sort":                "Sort",
	"task.sort_created":        "By Created",
	"task.sort_updated":        "By Updated",
	"task.status_completed":    "Completed",
	"task.order_asc":           "Asc",
	"task.order_desc":          "Desc",
	"task.new_task":            "New Task",
	"task.edit_task":           "Edit Task",
	"task.content":             "Content",
	"task.insert_phrase":       "Insert phrase…",
	"task.no_group":            "Select or create a group first",
	"task.copy_done":           "Copied",
	"task.confirm_delete_group": "Deleting a group cascades to all subgroups and tasks. Confirm?",
	"task.confirm_delete_task": "Delete this task?",

	// Check-in & wellness (F1.4)
	"checkin.record":      "Add Record",
	"checkin.type":        "Type",
	"checkin.value":       "Value",
	"checkin.date":        "Date",
	"checkin.report":      "Daily Report",
	"checkin.target":      "Daily Target",
	"checkin.set_target":  "Set Target",
	"checkin.achievement": "Achievement",
	"checkin.records":     "Records",
	"checkin.no_goal":     "No target",
	"checkin.recorded":    "Recorded",
	"checkin.water":       "Water",
	"checkin.exercise":    "Exercise",
	"checkin.nap":         "Nap",

	// Operation logs & charts (F1.5)
	"logs.chart_title":  "Activity Frequency",
	"logs.by_day":       "By Day",
	"logs.by_hour":      "By Hour",
	"logs.operations":   "Operations",
	"logs.logins":       "Logins",
	"logs.category":     "Category",
	"logs.cat_login":    "Login",
	"logs.cat_operation": "Operation",
	"logs.action":       "Action",
	"logs.path":         "Path",
	"logs.ip":           "IP",
	"logs.ua":           "Client",
}

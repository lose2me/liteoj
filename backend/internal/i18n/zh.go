// Package i18n centralizes user-facing messages returned by the HTTP layer.
// Only Chinese is supported; the package is effectively a string constant pool
// so translations are not mixed with control flow inside handlers.
package i18n

import "fmt"

// Auth / user
const (
	ErrBadCredentials = "用户名或密码错误"
	ErrBadPassword    = "密码不合法"
	ErrOldPassword    = "原密码错误"
	ErrIssueToken     = "颁发令牌失败"
	ErrHashFailed     = "密码哈希失败"
	ErrUpdateFailed   = "更新失败"
	ErrUserNotFound   = "用户不存在"
	ErrMissingToken   = "缺少登录令牌"
	ErrBadAuthHeader  = "认证头格式错误"
	ErrInvalidToken   = "登录令牌无效"
	ErrAdminOnly      = "仅管理员可访问"
)

// Common
const (
	ErrBadRequest      = "请求参数有误"
	ErrNotFound        = "资源不存在"
	ErrForbidden       = "权限不足"
	ErrProblemNotFound = "题目不存在"
)

// Submission
const (
	ErrLangNotAllowed      = "语言不在允许列表"
	ErrNoTestData          = "题目尚未配置测试数据"
	ErrProblemsetLangBlock = "该题单不允许 " // + language
	ErrSubmissionANotFound = "A 提交不存在"
	ErrSubmissionBNotFound = "B 提交不存在"
	ErrUnsupportedLanguage = "不支持的语言"
)

// AI
const (
	ErrAIAcNoAnalyze             = "AC 提交无需解析"
	ErrAIOptNonAC                = "仅 AC 提交可生成优化建议"
	ErrAIQueueFull               = "AI 任务队列已满，请稍后再试"
	ErrAIEnabledDisabled         = "AI 未启用（设置 ai.enabled=true 并配置 bifrost_*）"
	ErrAIConfigMissing           = "AI 未配置（bifrost_base_url / bifrost_api_key 为空）"
	ErrAINeedRaw                 = "请先在「详细」里粘贴题目原文"
	ErrAIEmptyResponse           = "AI 返回空响应"
)

// Admin bulk-create users
const (
	ErrBulkRowEmpty = "账号或密码为空"
)

// Admin dangerous operations
const (
	ErrDangerSecondaryPasswordMissing = "未配置二级密码：请先在 config.toml 的 [admin_danger] 下设置 secondary_password"
	ErrDangerSecondaryPasswordWrong   = "二级密码错误"
	ErrDangerRefuseClearRoot          = "拒绝清理根目录"
)

const DefaultHomeMarkdown = "# LiteOJ\n\n" +
	"> 一个面向课堂的轻量 OJ，支持题库、题单、AI 解析。\n\n" +
	"登录后可以开始刷题，或加入老师的题单。\n\n" +
	"## 当前环境\n\n" +
	"- C（gcc）：`gcc (Debian 14.2.0-19) 14.2.0`\n" +
	"- C++（g++）：`g++ (Debian 14.2.0-19) 14.2.0`\n" +
	"- Java（javac）：`javac 17.0.13`\n" +
	"- Python（python3）：`Python 3.11.11`\n" +
	"- 当前调用的大模型：`deepseek-v4-flash`\n"

// BulkSummary formats the final summary line shown on the batch-import result.
func BulkSummary(created, updated, failed int) string {
	return fmt.Sprintf("新增 %d / 更新 %d / 失败 %d", created, updated, failed)
}

func WarnDangerResetUploadCleanupFailed(detail string) string {
	return fmt.Sprintf("业务数据已清空，但上传目录清理失败：%s", detail)
}

func ErrDangerRefuseClearSuspiciousDir(abs string) string {
	return fmt.Sprintf("拒绝清理可疑目录 %s", abs)
}

func AITaskSubjectSubmission(id uint) string {
	return fmt.Sprintf("提交 #%d", id)
}

func AITaskSubjectProblem(id uint) string {
	return fmt.Sprintf("题目 #%d", id)
}

func ParseAITaskSubjectSubmission(subject string) (uint, bool) {
	for _, pattern := range []string{"提交 #%d", "submission #%d"} {
		var id uint
		if _, err := fmt.Sscanf(subject, pattern, &id); err == nil {
			return id, true
		}
	}
	return 0, false
}

func ParseAITaskSubjectProblem(subject string) (uint, bool) {
	for _, pattern := range []string{"题目 #%d", "problem #%d"} {
		var id uint
		if _, err := fmt.Sscanf(subject, pattern, &id); err == nil {
			return id, true
		}
	}
	return 0, false
}

func ErrAIPromptMissing(key string) string {
	return fmt.Sprintf("AI 提示词未配置：请在 config.toml 的 [ai] 下设置 %s", key)
}

func ErrAISubmissionNotFound(err error) string {
	return fmt.Sprintf("提交不存在：%v", err)
}

func ErrAIProblemNotFound(err error) string {
	return fmt.Sprintf("题目不存在：%v", err)
}

func ErrAIUnknownKind(kind string) string {
	return fmt.Sprintf("未知的 AI 任务类型：%s", kind)
}

func ErrAIReadResponse(err error) string {
	return fmt.Sprintf("AI 读取响应失败：%v", err)
}

func ErrAINon2xxResponse(status int, raw string) string {
	return fmt.Sprintf("AI 返回非成功响应（%d）：%s", status, raw)
}

func ErrAIParseResponse(err error) string {
	return fmt.Sprintf("AI 响应解析失败：%v", err)
}

func ErrAIOutputNotJSON(out string) string {
	return fmt.Sprintf("模型输出无法解析为 JSON：%s", out)
}

func ErrAIOutputJSONParse(err error) string {
	return fmt.Sprintf("模型输出 JSON 解析失败：%v", err)
}

func ProblemSetCopyTitle(title string) string {
	return title + "（副本）"
}

func ErrJudgeCompileUnexpectedCount(n int) string {
	return fmt.Sprintf("编译阶段返回结果数量异常：%d", n)
}

func ErrGoJudgeStatus(status int, body string) string {
	return fmt.Sprintf("go-judge 返回异常状态（%d）：%s", status, body)
}

func ErrGoJudgeDecode(err error, body string) string {
	return fmt.Sprintf("go-judge 响应解析失败：%v；原始内容：%s", err, body)
}

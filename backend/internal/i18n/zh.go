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
)

// AI
const (
	ErrAIAcNoAnalyze = "AC 提交无需解析"
	ErrAIOptNonAC    = "仅 AC 提交可生成优化建议"
)

// Admin bulk-create users
const (
	ErrBulkRowEmpty = "账号或密码为空"
)

// BulkSummary formats the final summary line shown on the batch-import result.
func BulkSummary(created, updated, failed int) string {
	return fmt.Sprintf("新增 %d / 更新 %d / 失败 %d", created, updated, failed)
}

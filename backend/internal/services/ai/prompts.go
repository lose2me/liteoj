package ai

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	"github.com/liteoj/liteoj/backend/internal/config"
	"github.com/liteoj/liteoj/backend/internal/i18n"
	"github.com/liteoj/liteoj/backend/internal/models"
)

// Prompts bundles every AI flow. Each flow composes a system prompt from
// config with task-specific data, calls the OpenAI-compatible client, and
// post-processes the model output.
//
// All problem-authoring flows (GenTitle/GenDesc/GenIO/GenIdea/GenExplain/
// GenAll + TagProblem) consume the admin-pasted 详细 (raw) string rather
// than the already-structured problem fields — there's a single source of
// truth so generations stay consistent.
//
// Prompt text is owned by config.toml (no code-level defaults); the admin
// enforces format/style/heading conventions there.
type Prompts struct {
	Cfg    *config.Config
	Client *Client
}

func NewPrompts(cfg *config.Config, client *Client) *Prompts {
	return &Prompts{Cfg: cfg, Client: client}
}

func (p *Prompts) ensureEnabled() error {
	if !p.Cfg.AIEnabled {
		return errors.New(i18n.ErrAIEnabledDisabled)
	}
	if p.Cfg.BifrostBaseURL == "" || p.Cfg.BifrostAPIKey == "" {
		return errors.New(i18n.ErrAIConfigMissing)
	}
	return nil
}

func (p *Prompts) ensurePrompt(sys, key string) error {
	if strings.TrimSpace(sys) == "" {
		return errors.New(i18n.ErrAIPromptMissing(key))
	}
	return nil
}

func requireRaw(raw string) error {
	if strings.TrimSpace(raw) == "" {
		return errors.New(i18n.ErrAINeedRaw)
	}
	return nil
}

// 以下三段是"JSON 输出格式"约束，硬编码在代码里。admin 在 config.toml
// 调的是"判断/输出**内容**"（答疑风格、乱写规则、节点命名、选标签原则等），
// "输出**结构**"（必须是 JSON、schema 字段名、不要 ``` 包裹）在这儿统一控制，
// 防止 admin 改坏导致 parser 炸掉。
//
// analyzeOutputSuffix 只保留 JSON 结构包裹部分——具体什么算"乱写"由 admin
// 在 prompt_wrong_answer 里写规则，模型自己判定 ok=true/false；代码只负责
// 约束返回格式，不约束判定口径。
const analyzeOutputSuffix = "\n\n严格以 JSON（不要包裹 ```）输出，不要除 JSON 外的任何文字：\n" +
	"{\"ok\": true|false, \"reason\": \"未认真作答的类别\", \"explanation\": \"markdown 解释或空字符串\"}\n" +
	"若 ok=true 则在 explanation 填写你原本要给出的 Markdown 解析；若 ok=false 则 explanation 留空字符串，reason 填简短类别名。"

const analyzeForceAdminSuffix = "\n\n【管理员强制解析模式】\n" +
	"当前触发者是管理员，本次任务用于教学演示或兜底排查。你不得以“乱写 / 未认真作答”为由拒绝，" +
	"必须输出 `ok=true`，`reason` 置空字符串，并在 `explanation` 中继续完成解析。\n" +
	"即使代码明显是占位代码、随机字符、空模板或与题目不匹配，也要直接说明：当前代码哪里与题意不符、" +
	"缺少哪些关键步骤、下一步应如何补齐输入、核心逻辑与输出。严禁输出 `ok=false`。"

// tagOutputSuffix 强制打标签返回严格 JSON，schema 与 parser 对齐。admin 可以
// 在 prompt_tag 里改选标签原则 / 难度判定标准，但不能破坏这份 schema。
const tagOutputSuffix = "\n\n严格只输出 JSON（不要包裹 ```，不要前后缀，不要任何解释）：\n" +
	"{\"tags\":[{\"group\":\"一级\",\"tag\":\"二级\"}],\"difficulty\":\"入门|简单|中等|困难\"}"

// genAllOutputSuffix 强制"一键填充"返回严格 JSON，字段与 GenAllResult 对齐。
// admin 可以改每个字段的节点 / 风格 / 长度，但不能改 schema。
const genAllOutputSuffix = "\n\n输出必须是一个 JSON 对象，且**只**包含以下 4 个字段（键名逐字一致，顺序不限，不得新增额外键）：\n" +
	"- `title` (string)\n" +
	"- `description` (string)\n" +
	"- `solution_idea_md` (string)\n" +
	"- `solution_md` (string)\n" +
	"JSON 字符串内的换行一律用 `\\n` 转义；JSON 对象外**严禁**任何额外文字、解释、前后缀、或 Markdown 代码块包裹。"

// AnalyzeWrongAnswer explains why a non-AC submission failed. Unlike the
// authoring flows this one operates on the actual submission + problem, not
// on raw-pasted content — it runs on the student submission detail page.
// 模型被要求按 {ok, reason, explanation} JSON 输出——见 analyzeOutputSuffix。
// 解析由 runner 完成；这里原样返回 text 即可。
func (p *Prompts) AnalyzeWrongAnswer(
	ctx context.Context, prob *models.Problem, sub *models.Submission, cases string, forceAnalyze bool,
) (string, string, string, error) {
	if err := p.ensureEnabled(); err != nil {
		return "", "", "", err
	}
	if err := p.ensurePrompt(p.Cfg.AIPromptWA, "prompt_wrong_answer"); err != nil {
		return "", "", "", err
	}
	user := fmt.Sprintf(
		"题目标题：%s\n\n题目描述：\n%s\n\n提交语言：%s\n\n提交代码：\n```%s\n%s\n```\n\n判题结果：%s\n%s\n\n测试用例结果（JSON）：\n%s",
		prob.Title, prob.Description, sub.Language, sub.Language, sub.Code,
		sub.Verdict, sub.Message, cases,
	)
	sys := p.Cfg.AIPromptWA
	if forceAnalyze {
		sys += analyzeForceAdminSuffix
	}
	messages := []Message{
		{Role: "system", Content: sys + analyzeOutputSuffix},
		{Role: "user", Content: user},
	}
	prompt := formatMessages(messages)
	text, raw, err := p.Client.Chat(ctx, messages)
	return text, prompt, raw, err
}

// OptimizeAC gives polishing / performance / style suggestions for an AC
// submission. Similar shape to AnalyzeWrongAnswer but without the "乱写"
// detection suffix — an AC code passed all cases, so the optimization advice
// is plain Markdown and goes straight into ai_explanation.
func (p *Prompts) OptimizeAC(
	ctx context.Context, prob *models.Problem, sub *models.Submission,
) (string, string, string, error) {
	if err := p.ensureEnabled(); err != nil {
		return "", "", "", err
	}
	if err := p.ensurePrompt(p.Cfg.AIPromptOpt, "prompt_optimize"); err != nil {
		return "", "", "", err
	}
	user := fmt.Sprintf(
		"题目标题：%s\n\n题目描述：\n%s\n\n提交语言：%s\n\n提交代码（已 AC，用时 %d ms / 内存 %d KB）：\n```%s\n%s\n```",
		prob.Title, prob.Description, sub.Language, sub.TimeUsedMS, sub.MemoryUsedKB,
		sub.Language, sub.Code,
	)
	messages := []Message{
		{Role: "system", Content: p.Cfg.AIPromptOpt},
		{Role: "user", Content: user},
	}
	prompt := formatMessages(messages)
	text, raw, err := p.Client.Chat(ctx, messages)
	return stripCodeFence(text), prompt, raw, err
}

// TagSuggestion is the parsed output of TagProblem.
type TagSuggestion struct {
	GroupIDs   []uint   `json:"group_ids"`
	TagIDs     []uint   `json:"tag_ids"`
	Matched    []string `json:"matched"`   // "group::tag" pairs matched to dictionary
	Unmatched  []string `json:"unmatched"` // raw names the model proposed but not in dict
	Difficulty string   `json:"difficulty,omitempty"`
	RawText    string   `json:"raw_text"`
}

// validDifficulties is the closed set the Problem.difficulty column accepts.
// The model often hallucinates synonyms ("简单题"/"easy"/"基础") — we only
// honor exact matches and drop everything else, leaving the field empty so
// the admin notices and either accepts a tighter suggestion or sets it
// manually.
var validDifficulties = map[string]struct{}{
	"入门": {}, "简单": {}, "中等": {}, "困难": {},
}

// TagProblem asks the model to pick tags from the predefined dictionary
// using the admin-pasted raw content as the sole source material.
func (p *Prompts) TagProblem(
	ctx context.Context, raw string, groups []models.TagGroup, tagsByGroup map[uint][]models.Tag,
) (*TagSuggestion, string, string, error) {
	if err := p.ensureEnabled(); err != nil {
		return nil, "", "", err
	}
	if err := p.ensurePrompt(p.Cfg.AIPromptTag, "prompt_tag"); err != nil {
		return nil, "", "", err
	}
	if err := requireRaw(raw); err != nil {
		return nil, "", "", err
	}

	// Build a readable dictionary for the model and a lookup for post-processing.
	var b strings.Builder
	type key struct{ g, t string }
	known := map[key]struct{ gid, tid uint }{}
	for _, g := range groups {
		b.WriteString("- ")
		b.WriteString(g.Name)
		b.WriteString(": ")
		ts := tagsByGroup[g.ID]
		for i, t := range ts {
			if i > 0 {
				b.WriteString(", ")
			}
			b.WriteString(t.Name)
			known[key{g.Name, t.Name}] = struct{ gid, tid uint }{g.ID, t.ID}
		}
		b.WriteByte('\n')
	}

	user := fmt.Sprintf(
		"可选一级标签与二级标签字典：\n%s\n\n题目原文（详细）：\n%s\n\n请严格输出 JSON：{\"tags\":[{\"group\":\"一级\",\"tag\":\"二级\"}],\"difficulty\":\"入门|简单|中等|困难\"}，不要其他文字。",
		b.String(), raw,
	)
	messages := []Message{
		{Role: "system", Content: p.Cfg.AIPromptTag + tagOutputSuffix},
		{Role: "user", Content: user},
	}
	prompt := formatMessages(messages)
	rawOut, rawBody, err := p.Client.Chat(ctx, messages)
	if err != nil {
		return nil, prompt, rawBody, err
	}

	sug := &TagSuggestion{RawText: rawOut}
	jsonText := extractJSONObject(rawOut)
	if jsonText == "" {
		return sug, prompt, rawBody, nil
	}
	var parsed struct {
		Tags []struct {
			Group string `json:"group"`
			Tag   string `json:"tag"`
		} `json:"tags"`
		Difficulty string `json:"difficulty"`
	}
	if err := json.Unmarshal([]byte(jsonText), &parsed); err != nil {
		return sug, prompt, rawBody, nil
	}
	if d := strings.TrimSpace(parsed.Difficulty); d != "" {
		if _, ok := validDifficulties[d]; ok {
			sug.Difficulty = d
		}
	}
	seenGroup := map[uint]bool{}
	seenTag := map[uint]bool{}
	for _, pair := range parsed.Tags {
		k := key{strings.TrimSpace(pair.Group), strings.TrimSpace(pair.Tag)}
		if ids, ok := known[k]; ok {
			if !seenGroup[ids.gid] {
				sug.GroupIDs = append(sug.GroupIDs, ids.gid)
				seenGroup[ids.gid] = true
			}
			if !seenTag[ids.tid] {
				sug.TagIDs = append(sug.TagIDs, ids.tid)
				seenTag[ids.tid] = true
			}
			sug.Matched = append(sug.Matched, k.g+"::"+k.t)
		} else {
			sug.Unmatched = append(sug.Unmatched, k.g+"::"+k.t)
		}
	}
	return sug, prompt, rawBody, nil
}

// singleTextGen is a one-shot "system prompt + raw content → markdown"
// helper. Trims any accidental code-fence wrapper the model emits.
func (p *Prompts) singleTextGen(
	ctx context.Context, raw, sysPrompt, cfgKey string,
) (string, string, string, error) {
	if err := p.ensureEnabled(); err != nil {
		return "", "", "", err
	}
	if err := p.ensurePrompt(sysPrompt, cfgKey); err != nil {
		return "", "", "", err
	}
	if err := requireRaw(raw); err != nil {
		return "", "", "", err
	}
	messages := []Message{
		{Role: "system", Content: sysPrompt},
		{Role: "user", Content: raw},
	}
	prompt := formatMessages(messages)
	text, rawBody, err := p.Client.Chat(ctx, messages)
	return stripCodeFence(text), prompt, rawBody, err
}

// GenTitle produces a single-line problem title. Model should obey the
// config prompt re: length / format; we just trim whitespace and drop
// trailing punctuation that would look odd in a list.
func (p *Prompts) GenTitle(ctx context.Context, raw string) (string, string, string, error) {
	text, prompt, rawBody, err := p.singleTextGen(ctx, raw, p.Cfg.AIPromptGenTitle, "prompt_gen_title")
	if err != nil {
		return "", prompt, rawBody, err
	}
	title := strings.TrimSpace(text)
	// Take the first line only — if the model spilled over, we keep the
	// most likely title and discard the rest.
	if nl := strings.IndexByte(title, '\n'); nl > 0 {
		title = strings.TrimSpace(title[:nl])
	}
	return title, prompt, rawBody, nil
}

func (p *Prompts) GenDesc(ctx context.Context, raw string) (string, string, string, error) {
	return p.singleTextGen(ctx, raw, p.Cfg.AIPromptGenDesc, "prompt_gen_desc")
}

func (p *Prompts) GenIdea(ctx context.Context, raw string) (string, string, string, error) {
	return p.singleTextGen(ctx, raw, p.Cfg.AIPromptGenIdea, "prompt_gen_idea")
}

func (p *Prompts) GenExplain(ctx context.Context, raw string) (string, string, string, error) {
	return p.singleTextGen(ctx, raw, p.Cfg.AIPromptGenExplain, "prompt_gen_explain")
}

// GenAllResult is the structured payload returned by GenAll. Fields match
// the problem columns the admin will paste into via 一键填充. Input/output
// format and samples live inside `description` (as `## 输入格式` /
// `## 输出格式` / `## 样例` sections), not as separate keys.
type GenAllResult struct {
	Title          string `json:"title"`
	Description    string `json:"description"`
	SolutionIdeaMD string `json:"solution_idea_md"`
	SolutionMD     string `json:"solution_md"`
}

// GenAll executes the merged "一键填充" prompt: a single model call that
// produces every authoring field at once. The config prompt must instruct
// the model to emit strict JSON matching GenAllResult.
func (p *Prompts) GenAll(ctx context.Context, raw string) (*GenAllResult, string, string, error) {
	if err := p.ensureEnabled(); err != nil {
		return nil, "", "", err
	}
	if err := p.ensurePrompt(p.Cfg.AIPromptGenAll, "prompt_gen_all"); err != nil {
		return nil, "", "", err
	}
	if err := requireRaw(raw); err != nil {
		return nil, "", "", err
	}
	messages := []Message{
		{Role: "system", Content: p.Cfg.AIPromptGenAll + genAllOutputSuffix},
		{Role: "user", Content: raw},
	}
	prompt := formatMessages(messages)
	out, rawBody, err := p.Client.Chat(ctx, messages)
	if err != nil {
		return nil, prompt, rawBody, err
	}
	jsonText := extractJSONObject(out)
	if jsonText == "" {
		return nil, prompt, rawBody, errors.New(i18n.ErrAIOutputNotJSON(out))
	}
	var r GenAllResult
	if err := json.Unmarshal([]byte(jsonText), &r); err != nil {
		return nil, prompt, rawBody, errors.New(i18n.ErrAIOutputJSONParse(err))
	}
	return &r, prompt, rawBody, nil
}

// stripCodeFence drops a ```...``` wrapper if the model over-eagerly wrapped
// the output despite instructions.
func stripCodeFence(s string) string {
	s = strings.TrimSpace(s)
	if !strings.HasPrefix(s, "```") {
		return s
	}
	if nl := strings.IndexByte(s, '\n'); nl > 0 {
		s = s[nl+1:]
	}
	s = strings.TrimRight(s, " \n")
	s = strings.TrimSuffix(s, "```")
	return strings.TrimSpace(s)
}

// formatMessages joins chat messages into a single human-readable string for
// audit logging. Not used as wire format — the Client still sends them
// structured — but gives admins a single blob to read in the UI.
func formatMessages(msgs []Message) string {
	var b strings.Builder
	for i, m := range msgs {
		if i > 0 {
			b.WriteString("\n\n")
		}
		b.WriteString("## ")
		b.WriteString(m.Role)
		b.WriteString("\n")
		b.WriteString(m.Content)
	}
	return b.String()
}

// extractJSONObject finds the first balanced {...} object in s. Handles simple
// Markdown code fences around the JSON.
func extractJSONObject(s string) string {
	start := strings.IndexByte(s, '{')
	if start < 0 {
		return ""
	}
	depth := 0
	for i := start; i < len(s); i++ {
		switch s[i] {
		case '{':
			depth++
		case '}':
			depth--
			if depth == 0 {
				return s[start : i+1]
			}
		}
	}
	return ""
}

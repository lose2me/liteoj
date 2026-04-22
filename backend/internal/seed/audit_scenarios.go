package seed

import (
	"log"
	"time"

	"gorm.io/gorm"

	"github.com/liteoj/liteoj/backend/internal/auth"
	"github.com/liteoj/liteoj/backend/internal/models"
)

// EnsureAuditScenarios seeds data specifically engineered for the Phase E
// regression tests: penalty-formula verification rows, the "wrong-code
// exemption" corpus (Hello-World-style entry problems), and a "受限题单" that
// toggles all three disable_* flags at once. Idempotent — checks for the
// sentinel Hello World problem before creating anything.
//
// Students added here (stu7..stu10) are dedicated to audit scenarios:
//   - stu7 : in-set WA×2 → Analyze(ok=true) → AC (expected penalty = 60 min)
//   - stu8 : in-set Analyze(rejected=true) → AC (expected penalty = 0 min)
//   - stu9 : 独立页 multiple AI analyses, no in-set AC (must not bleed into penalty)
//   - stu10: banned from 综合练习 · 30 题 (historical submissions preserved, not in rank)
func EnsureAuditScenarios(db *gorm.DB) error {
	// Sentinel: re-runs are no-ops once the Hello World entry problem exists.
	var count int64
	db.Model(&models.Problem{}).Where("title = ?", "Hello World").Count(&count)
	if count > 0 {
		return nil
	}

	pwd, _ := auth.HashPassword("123456")
	seenBase := time.Now()
	newStudents := []models.User{
		{Username: "stu7", Name: "学生庚", PasswordHash: pwd, Role: models.RoleStudent,
			LastSeenAt: ptrTime(seenBase.Add(-30 * time.Minute))},
		{Username: "stu8", Name: "学生辛", PasswordHash: pwd, Role: models.RoleStudent,
			LastSeenAt: ptrTime(seenBase.Add(-45 * time.Minute))},
		{Username: "stu9", Name: "学生壬", PasswordHash: pwd, Role: models.RoleStudent,
			LastSeenAt: ptrTime(seenBase.Add(-60 * time.Minute))},
		{Username: "stu10", Name: "学生癸", PasswordHash: pwd, Role: models.RoleStudent,
			LastSeenAt: ptrTime(seenBase.Add(-2 * time.Hour))},
	}
	for i := range newStudents {
		if err := db.Where("username = ?", newStudents[i].Username).FirstOrCreate(&newStudents[i]).Error; err != nil {
			return err
		}
	}

	// HelloWorld 题：验证"乱写判定豁免" —— `int main(){return 0;}` 这类模板
	// 对普通题应判 rejected，对这道题应该 ok=true 并提示缺少 cout。
	hello := models.Problem{
		Title:      "Hello World",
		Difficulty: "入门",
		Description: "输出一行 `Hello World`。\n\n" +
			"## 输入格式\n\n无输入。\n\n" +
			"## 输出格式\n\n一行字符串 `Hello World`。\n\n" +
			"## 样例\n\n### 样例输入 1\n\n```\n```\n\n### 样例输出 1\n\n```\nHello World\n```\n\n" +
			"## 说明 / 提示\n\n本题演示最小可执行的 I/O。",
		SolutionIdeaMD: "## 算法分析\n\n直接输出常量字符串，无任何计算。\n\n" +
			"## 实现要点\n\n- C++ 用 `std::cout << \"Hello World\\n\";`。\n- Python 用 `print(\"Hello World\")`。\n\n" +
			"## 复杂度分析\n\n时间复杂度 $O(1)$，空间复杂度 $O(1)$。",
		SolutionMD: "## 题目分析\n\n最小 I/O 示例。\n\n" +
			"## 算法与做法\n\n直接输出字符串。\n\n" +
			"## 参考实现\n\n```cpp\n#include<iostream>\nint main(){std::cout<<\"Hello World\\n\";}\n```\n\n" +
			"## 复杂度与易错点\n\n$O(1)$。\n\n- 末尾是否换行因 checker 而异，本题允许或不允许两可。",
		TimeLimitMS:   1000,
		MemoryLimitMB: 256,
		Visible:       true,
		CreatedBy:     1,
	}
	// 原样输出题：等价"最小语法示例"，同样命中豁免范围。
	echo := models.Problem{
		Title:      "原样输出",
		Difficulty: "入门",
		Description: "读入一行字符串 $s$，原样输出。\n\n" +
			"## 输入格式\n\n一行字符串 $s$，$1 \\le |s| \\le 100$，仅含可见 ASCII。\n\n" +
			"## 输出格式\n\n一行，即 $s$ 本身。\n\n" +
			"## 样例\n\n### 样例输入 1\n\n```\nhello\n```\n\n### 样例输出 1\n\n```\nhello\n```\n",
		SolutionIdeaMD: "## 算法分析\n\nI/O 练习，直接回显。\n\n" +
			"## 实现要点\n\n- `getline(cin, s)` + `cout << s`。\n\n" +
			"## 复杂度分析\n\n时间复杂度 $O(|s|)$，空间复杂度 $O(|s|)$。",
		SolutionMD: "## 题目分析\n\n读入字符串并回显。\n\n" +
			"## 算法与做法\n\n调用 `getline` 吃掉整行并原样输出。\n\n" +
			"## 参考实现\n\n```cpp\n#include<iostream>\n#include<string>\nint main(){std::string s;std::getline(std::cin,s);std::cout<<s<<'\\n';}\n```\n\n" +
			"## 复杂度与易错点\n\n$O(|s|)$。\n\n- 换行符是否保留由 checker 决定。",
		TimeLimitMS:   1000,
		MemoryLimitMB: 256,
		Visible:       true,
		CreatedBy:     1,
	}
	if err := db.Create(&hello).Error; err != nil {
		return err
	}
	if err := db.Create(&echo).Error; err != nil {
		return err
	}
	// 测试用例：Hello World 无输入 + 固定输出；原样输出三组代表性样例。
	db.Create(&[]models.Testcase{
		{ProblemID: hello.ID, Input: "", ExpectedOutput: "Hello World\n", OrderIndex: 1},
	})
	db.Create(&[]models.Testcase{
		{ProblemID: echo.ID, Input: "hello\n", ExpectedOutput: "hello\n", OrderIndex: 1},
		{ProblemID: echo.ID, Input: "12345\n", ExpectedOutput: "12345\n", OrderIndex: 2},
		{ProblemID: echo.ID, Input: "ok!@#\n", ExpectedOutput: "ok!@#\n", OrderIndex: 3},
	})

	// 受限题单：三个 disable_* 开关全开，用于覆盖 ProblemDetail chip + AI 拒绝路径。
	restricted := models.ProblemSet{
		Title:           "受限题单",
		DisableIdea:     true,
		DisableSolution: true,
		DisableAI:       true,
		CreatedBy:       1,
	}
	if err := db.Create(&restricted).Error; err != nil {
		return err
	}
	db.Create(&[]models.ProblemSetItem{
		{ProblemSetID: restricted.ID, ProblemID: hello.ID, OrderIndex: 0},
		{ProblemSetID: restricted.ID, ProblemID: echo.ID, OrderIndex: 1},
	})
	// 成员：admin + stu7 便于验证"chip 显示 / AI 按钮不出现"。
	db.Create(&[]models.ProblemSetMember{
		{ProblemSetID: restricted.ID, UserID: 1, JoinedAt: time.Now()},
		{ProblemSetID: restricted.ID, UserID: newStudents[0].ID, JoinedAt: time.Now()},
	})

	// 目标题单：penalty 场景要挂在"入门练习"上（第一个练习题单），假设 id=1。
	// 更稳的做法：按 title 反查以防 id 漂移。
	var practice models.ProblemSet
	if err := db.Where("title = ?", "入门练习").First(&practice).Error; err != nil {
		return err
	}
	// 把 stu7 / stu8 加入入门练习作为成员，penalty 计算才会包含他们。
	db.Create(&[]models.ProblemSetMember{
		{ProblemSetID: practice.ID, UserID: newStudents[0].ID, JoinedAt: time.Now()},
		{ProblemSetID: practice.ID, UserID: newStudents[1].ID, JoinedAt: time.Now()},
		{ProblemSetID: practice.ID, UserID: newStudents[2].ID, JoinedAt: time.Now()},
	})

	// 要让 stu7 的罚时 = 60，必须针对"入门练习"里的某道题做 2 WA + 1 成功 AI + 1 AC。
	// 选 A+B Problem（title 反查拿 id）——首个题单首题，A 编号。
	var abProblem models.Problem
	if err := db.Where("title = ?", "A + B Problem").First(&abProblem).Error; err != nil {
		return err
	}

	practiceID := practice.ID
	now := time.Now()

	// stu7：题单内 2 条 WA（后一条带成功 AI 解析）+ 1 条 AC。
	// ranking 聚合逻辑（handlers/ranking.go）：每条非 AC 计 WABeforeAC++；若该
	// 条同时 ai_explanation != '' && !ai_rejected 则 AIUsedBeforeAC++。
	// WA=2 + AI=1 → (2+1)*20 = 60。
	stu7 := newStudents[0].ID
	db.Create(&models.Submission{
		UserID: stu7, ProblemID: abProblem.ID, ProblemSetID: &practiceID,
		Language: "cpp", Verdict: models.VerdictWA,
		Code:               "#include<iostream>\nint main(){long long a,b;std::cin>>a>>b;std::cout<<a;}\n",
		TestcaseResultJSON: `[{"index":1,"verdict":"WA","time_ms":2,"memory_kb":1024,"message":"expected 3 got 1"}]`,
		TimeUsedMS:         2, MemoryUsedKB: 1024,
		Message:   "第 1 个用例期望 3，实际 1",
		CreatedAt: now.Add(-90 * time.Minute),
	})
	db.Create(&models.Submission{
		UserID: stu7, ProblemID: abProblem.ID, ProblemSetID: &practiceID,
		Language: "cpp", Verdict: models.VerdictWA,
		Code:               "#include<iostream>\nint main(){long long a,b;std::cin>>a>>b;std::cout<<a*b;}\n",
		TestcaseResultJSON: `[{"index":1,"verdict":"WA","time_ms":2,"memory_kb":1024,"message":"expected 3 got 2"}]`,
		TimeUsedMS:         2, MemoryUsedKB: 1024,
		Message:       "第 1 个用例期望 3，实际 2",
		AIExplanation: "代码把 `a + b` 写成了 `a * b`。修正为 `std::cout << a + b;` 即可。",
		CreatedAt:     now.Add(-60 * time.Minute),
	})
	db.Create(&models.Submission{
		UserID: stu7, ProblemID: abProblem.ID, ProblemSetID: &practiceID,
		Language: "cpp", Verdict: models.VerdictAC,
		Code:               "#include<iostream>\nint main(){long long a,b;std::cin>>a>>b;std::cout<<a+b;}\n",
		TestcaseResultJSON: `[{"index":1,"verdict":"AC","time_ms":2,"memory_kb":1024,"message":""},{"index":2,"verdict":"AC","time_ms":2,"memory_kb":1024,"message":""},{"index":3,"verdict":"AC","time_ms":2,"memory_kb":1024,"message":""}]`,
		TimeUsedMS:         2, MemoryUsedKB: 1024,
		CreatedAt: now.Add(-30 * time.Minute),
	})

	// stu8：题单内直接 AC（无 WA 前缀），penalty=0。另在独立页交一条带 AI rejected
	// 的 submission，展示"AI 拒绝"能力但不影响题单罚时（独立页不计入）。
	stu8 := newStudents[1].ID
	var gcdProblem models.Problem
	_ = db.Where("title = ?", "最大公约数").First(&gcdProblem).Error
	if gcdProblem.ID > 0 {
		db.Create(&models.Submission{
			UserID: stu8, ProblemID: gcdProblem.ID, ProblemSetID: &practiceID,
			Language: "cpp", Verdict: models.VerdictAC,
			Code:               "#include<iostream>\nusing namespace std;\nlong long g(long long a,long long b){return b?g(b,a%b):a;}\nint main(){long long a,b;cin>>a>>b;cout<<g(a,b);}\n",
			TestcaseResultJSON: `[{"index":1,"verdict":"AC","time_ms":1,"memory_kb":900,"message":""},{"index":2,"verdict":"AC","time_ms":1,"memory_kb":900,"message":""}]`,
			TimeUsedMS:         1, MemoryUsedKB: 900,
			CreatedAt: now.Add(-25 * time.Minute),
		})
		db.Create(&models.Submission{
			UserID: stu8, ProblemID: gcdProblem.ID, // 独立页
			Language: "cpp", Verdict: models.VerdictWA,
			Code:               "int main(){return 0;}\n",
			TestcaseResultJSON: `[{"index":1,"verdict":"WA","time_ms":0,"memory_kb":0,"message":"expected 6 got nothing"}]`,
			TimeUsedMS:         0, MemoryUsedKB: 0,
			AIRejected:     true,
			AIRejectReason: "代码与题目无关（空模板）",
			Message:        "独立页：第 1 个用例期望 6，实际无输出",
			CreatedAt:      now.Add(-120 * time.Minute),
		})
	}

	// stu9 场景：独立页 AI 解析 ×3 + 独立页 AC；题单内完全没做题。
	// 期望：题单排名中 stu9 不出现或 penalty=0，AC 不影响罚时聚合。
	stu9 := newStudents[2].ID
	for i := 0; i < 3; i++ {
		db.Create(&models.Submission{
			UserID: stu9, ProblemID: abProblem.ID, // 独立页
			Language: "cpp", Verdict: models.VerdictWA,
			Code:               "#include<iostream>\nint main(){int a,b;std::cin>>a>>b;std::cout<<a-b;}\n",
			TestcaseResultJSON: `[{"index":1,"verdict":"WA","time_ms":2,"memory_kb":900,"message":"expected 3 got -1"}]`,
			TimeUsedMS:         2, MemoryUsedKB: 900,
			AIExplanation: "代码把 `+` 写成了 `-`。仔细检查运算符。",
			CreatedAt:     now.Add(time.Duration(-(200 + i*40)) * time.Minute),
		})
	}
	db.Create(&models.Submission{
		UserID: stu9, ProblemID: abProblem.ID, // 独立页
		Language: "cpp", Verdict: models.VerdictAC,
		Code:               "#include<iostream>\nint main(){int a,b;std::cin>>a>>b;std::cout<<a+b;}\n",
		TestcaseResultJSON: `[{"index":1,"verdict":"AC","time_ms":2,"memory_kb":900,"message":""}]`,
		TimeUsedMS:         2, MemoryUsedKB: 900,
		CreatedAt: now.Add(-20 * time.Minute),
	})

	// stu10 场景：加入"综合练习·30 题"、做过若干题、然后被 ban。被 ban 后 submissions
	// 必须保留，排名里消失。
	var bigSet models.ProblemSet
	_ = db.Where("title = ?", "综合练习 · 30 题").First(&bigSet).Error
	stu10 := newStudents[3].ID
	if bigSet.ID > 0 {
		// 先作为 member 提交几条 AC（用 padding 题目）
		var padItems []models.ProblemSetItem
		db.Where("problem_set_id = ?", bigSet.ID).Order("order_index ASC").Limit(5).Find(&padItems)
		for i, it := range padItems {
			db.Create(&models.Submission{
				UserID: stu10, ProblemID: it.ProblemID, ProblemSetID: &bigSet.ID,
				Language: "cpp", Verdict: models.VerdictAC,
				Code:               "#include<iostream>\nint main(){long long a;std::cin>>a;std::cout<<a;}\n",
				TestcaseResultJSON: `[{"index":1,"verdict":"AC","time_ms":2,"memory_kb":800,"message":""}]`,
				TimeUsedMS:         2, MemoryUsedKB: 800,
				CreatedAt: now.Add(time.Duration(-(72*60 + i*20)) * time.Minute),
			})
		}
		// 然后写 ban 表 + 删 member 行（不删 submissions——核心语义）
		db.Create(&models.ProblemSetBan{
			ProblemSetID: bigSet.ID,
			UserID:       stu10,
			BannedAt:     now.Add(-24 * time.Hour),
			BannedBy:     1,
		})
		// 保守删成员（若原本未加入，delete 无操作）
		db.Where("problem_set_id = ? AND user_id = ?", bigSet.ID, stu10).
			Delete(&models.ProblemSetMember{})
	}

	log.Printf("seed: audit scenarios installed (stu7..10, Hello World, 受限题单)")
	return nil
}

func ptrTime(t time.Time) *time.Time { return &t }

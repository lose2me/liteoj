package seed

import (
	"fmt"
	"log"
	"time"

	"gorm.io/gorm"

	"github.com/liteoj/liteoj/backend/internal/auth"
	"github.com/liteoj/liteoj/backend/internal/models"
)

// EnsureTestData seeds tag dictionary + a minimal dataset when the db is empty.
// Safe to call on every boot — idempotent via per-table "is empty?" checks.
func EnsureTestData(db *gorm.DB) error {
	if err := EnsureTaxonomy(db); err != nil {
		return err
	}
	if err := ensureSampleProblems(db); err != nil {
		return err
	}
	// EnsureAuditScenarios 按 Hello World 题是否存在自我去重；在已有数据上也可以
	// 安全反复跑（无新建 → 直接返回）。
	return EnsureAuditScenarios(db)
}

// EnsureTaxonomy installs the default tag dictionary when TagGroup is empty.
func EnsureTaxonomy(db *gorm.DB) error {
	var count int64
	db.Model(&models.TagGroup{}).Count(&count)
	if count > 0 {
		return nil
	}
	log.Println("seed: installing knowledge taxonomy (22 groups)")
	for gi, g := range defaultTaxonomy {
		gname := g[0].(string)
		group := &models.TagGroup{Name: gname, OrderIndex: gi + 1}
		if err := db.Create(group).Error; err != nil {
			return err
		}
		tagNames, _ := g[1].([]string)
		for ti, tname := range tagNames {
			if err := db.Create(&models.Tag{
				GroupID: group.ID, Name: tname, OrderIndex: ti + 1,
			}).Error; err != nil {
				return err
			}
		}
	}
	log.Printf("seed: taxonomy inserted")
	return nil
}

// ensureSampleProblems adds a few demo problems + a couple of problem sets
// and student accounts, only when there are no problems yet. Each sample
// problem has every non-metadata column populated (description + solution
// idea + full solution MD + difficulty + tags), so the UI surfaces the full
// layout and the LaTeX renderer can be smoke-tested against realistic input:
// $...$, $$...$$, \gcd, \sum, 分数、矩阵、代码块等。
func ensureSampleProblems(db *gorm.DB) error {
	var n int64
	db.Model(&models.Problem{}).Count(&n)
	if n > 0 {
		return nil
	}
	log.Println("seed: installing sample problems")

	studentPwd, _ := auth.HashPassword("123456")
	// Offset LastSeenAt so the admin dashboard's "recent activity" widget has
	// varied data. Pointer-to-time because LastSeenAt is nullable.
	seenA := time.Now().Add(-10 * time.Minute)
	seenB := time.Now().Add(-2 * time.Hour)
	seenC := time.Now().Add(-25 * time.Hour)
	seenD := time.Now().Add(-3 * 24 * time.Hour)
	seenE := time.Now().Add(-7 * 24 * time.Hour)
	seenF := time.Now().Add(-15 * 24 * time.Hour)
	students := []models.User{
		{Username: "stu1", Name: "学生甲", PasswordHash: studentPwd, Role: models.RoleStudent, LastSeenAt: &seenA},
		{Username: "stu2", Name: "学生乙", PasswordHash: studentPwd, Role: models.RoleStudent, LastSeenAt: &seenB},
		{Username: "stu3", Name: "学生丙", PasswordHash: studentPwd, Role: models.RoleStudent, LastSeenAt: &seenC},
		{Username: "stu4", Name: "学生丁", PasswordHash: studentPwd, Role: models.RoleStudent, LastSeenAt: &seenD},
		{Username: "stu5", Name: "学生戊", PasswordHash: studentPwd, Role: models.RoleStudent, LastSeenAt: &seenE},
		{Username: "stu6", Name: "学生己", PasswordHash: studentPwd, Role: models.RoleStudent, LastSeenAt: &seenF},
	}
	for i := range students {
		if err := db.Where("username = ?", students[i].Username).FirstOrCreate(&students[i]).Error; err != nil {
			return err
		}
	}

	// 批量学生 bulk01..bulk44：补齐 bigSet 的"50 人题单"测试场景，同时为用户管理
	// 页"最后活跃"列提供跨度从 5h 到 ~5 天的分布数据。没有提交记录——排行页
	// 里呈现"大部分 0 AC"的真实分布。
	extras := make([]models.User, 0, 44)
	for i := 0; i < 44; i++ {
		seen := time.Now().Add(-time.Duration(i*3+5) * time.Hour)
		extras = append(extras, models.User{
			Username:     fmt.Sprintf("bulk%02d", i+1),
			Name:         fmt.Sprintf("批量学生%02d", i+1),
			PasswordHash: studentPwd,
			Role:         models.RoleStudent,
			LastSeenAt:   &seen,
		})
	}
	for i := range extras {
		if err := db.Where("username = ?", extras[i].Username).FirstOrCreate(&extras[i]).Error; err != nil {
			return err
		}
	}

	tagID := func(name string) uint {
		var t models.Tag
		db.Where("name = ?", name).First(&t)
		return t.ID
	}

	// --- Problem 1: A + B (入门，示范基础 LaTeX + 代码块) ---
	p1 := models.Problem{
		Title:      "A + B Problem",
		Difficulty: "入门",
		Description: "给定两个整数 $a, b$，求 $a + b$。\n\n" +
			"## 输入格式\n\n一行两个整数 $a, b$，满足 $-10^9 \\le a, b \\le 10^9$。\n\n" +
			"## 输出格式\n\n一行一个整数，表示 $a + b$。\n\n" +
			"## 样例\n\n### 样例输入 1\n\n```\n1 2\n```\n\n### 样例输出 1\n\n```\n3\n```\n\n" +
			"### 样例输入 2\n\n```\n-5 7\n```\n\n### 样例输出 2\n\n```\n2\n```\n\n" +
			"## 说明 / 提示\n\n本题用于演示平台基本功能，注意结果可能超出 `int` 范围，应当使用 `long long`。",
		SolutionIdeaMD: "## 算法分析\n\n直接读入两个整数并相加，属于 IO 练习题。\n\n" +
			"## 实现要点\n\n- 使用 `long long` 以防止溢出。\n- C++ 可通过 `std::ios::sync_with_stdio(false)` 加速。\n\n" +
			"## 复杂度分析\n\n时间复杂度 $O(1)$，空间复杂度 $O(1)$。",
		SolutionMD: "## 题目分析\n\n给定 $a, b \\in [-10^9, 10^9]$，求 $a + b$。和至多为 $2 \\times 10^9$，已超出 32 位有符号整型范围，需要 64 位整型。\n\n" +
			"## 算法与做法\n\n读入两个整数，直接输出其和。\n\n" +
			"## 参考实现\n\n```cpp\n#include <iostream>\nusing namespace std;\nint main() {\n    ios::sync_with_stdio(false);\n    cin.tie(nullptr);\n    long long a, b;\n    cin >> a >> b;\n    cout << a + b << '\\n';\n    return 0;\n}\n```\n\n" +
			"## 复杂度与易错点\n\n时间复杂度 $O(1)$。\n\n- 忘了使用 `long long` 会导致溢出。\n- 输出换行最好显式写 `'\\n'` 而非 `endl`，减少缓冲刷新。",
		TimeLimitMS:   1000,
		MemoryLimitMB: 256,
		Visible:       true,
	}

	// --- Problem 2: 最大公约数 (入门，验证 \gcd 与中文标点边界) ---
	p2 := models.Problem{
		Title:      "最大公约数",
		Difficulty: "入门",
		Description: "输入两个正整数 $a, b$，输出它们的最大公约数 $\\gcd(a, b)$。\n\n" +
			"形式化地，求满足 $d \\mid a$ 且 $d \\mid b$ 的最大正整数 $d$。\n\n" +
			"## 输入格式\n\n一行两个正整数 $a, b$，满足 $1 \\le a, b \\le 10^9$。\n\n" +
			"## 输出格式\n\n一行一个正整数，表示 $\\gcd(a, b)$。\n\n" +
			"## 样例\n\n### 样例输入 1\n\n```\n12 18\n```\n\n### 样例输出 1\n\n```\n6\n```\n\n" +
			"### 样例输入 2\n\n```\n17 31\n```\n\n### 样例输出 2\n\n```\n1\n```\n\n" +
			"## 说明 / 提示\n\n辗转相除法满足 $\\gcd(a, b) = \\gcd(b, a \\bmod b)$，终止于 $b = 0$ 时返回 $a$。",
		SolutionIdeaMD: "## 算法分析\n\n辗转相除法（欧几里得算法）：若 $b = 0$ 则 $\\gcd = a$，否则 $\\gcd(a, b) = \\gcd(b, a \\bmod b)$。\n\n" +
			"## 实现要点\n\n- 递归或循环均可，短平快。\n- 标准库 `std::gcd`（C++17）可直接用。\n\n" +
			"## 复杂度分析\n\n时间复杂度 $O(\\log(\\min(a, b)))$，空间复杂度 $O(1)$。",
		SolutionMD: "## 题目分析\n\n求两个正整数的最大公约数，经典欧几里得算法模板题。\n\n" +
			"## 算法与做法\n\n由 $\\gcd(a, b) = \\gcd(b, a \\bmod b)$ 递归，当 $b = 0$ 时返回 $a$。\n\n" +
			"## 参考实现\n\n```cpp\n#include <iostream>\nusing namespace std;\nlong long gcd(long long a, long long b) {\n    return b == 0 ? a : gcd(b, a % b);\n}\nint main() {\n    long long a, b;\n    cin >> a >> b;\n    cout << gcd(a, b) << '\\n';\n    return 0;\n}\n```\n\n" +
			"## 复杂度与易错点\n\n时间复杂度 $O(\\log(\\min(a, b)))$。\n\n- 保证 $a, b$ 非负；若含负数需先取绝对值。\n- 若数据范围更大，应使用 `__int128` 或高精度。",
		TimeLimitMS:   1000,
		MemoryLimitMB: 256,
		Visible:       true,
	}

	// --- Problem 3: 斐波那契数 (简单，演示块公式 + 组合数学 tag) ---
	p3 := models.Problem{
		Title:      "斐波那契数",
		Difficulty: "简单",
		Description: "定义 Fibonacci 数列：\n\n$$F_0 = 0,\\quad F_1 = 1,\\quad F_n = F_{n-1} + F_{n-2}\\ (n \\ge 2).$$\n\n" +
			"给定 $n$，求 $F_n$。\n\n" +
			"## 输入格式\n\n一行一个非负整数 $n$，满足 $0 \\le n \\le 40$。\n\n" +
			"## 输出格式\n\n一行一个整数，表示 $F_n$。\n\n" +
			"## 样例\n\n### 样例输入 1\n\n```\n10\n```\n\n### 样例输出 1\n\n```\n55\n```\n\n" +
			"### 样例输入 2\n\n```\n0\n```\n\n### 样例输出 2\n\n```\n0\n```\n\n" +
			"## 说明 / 提示\n\n$n \\le 40$ 时结果在 `int` 范围内；若 $n$ 更大，可使用矩阵快速幂或通项公式 $F_n = \\dfrac{1}{\\sqrt{5}}\\left(\\phi^n - \\psi^n\\right)$，其中 $\\phi = \\dfrac{1+\\sqrt{5}}{2}$。",
		SolutionIdeaMD: "## 算法分析\n\n一维线性 DP / 递推：仅需 $O(1)$ 个滚动变量保存最近两项。\n\n" +
			"## 实现要点\n\n- 处理 $n = 0, 1$ 的边界。\n- 用两个变量滚动即可，无需数组。\n\n" +
			"## 复杂度分析\n\n时间复杂度 $O(n)$，空间复杂度 $O(1)$。",
		SolutionMD: "## 题目分析\n\n标准 Fibonacci 递推，$n \\le 40$ 直接线性扫一遍即可。\n\n" +
			"## 算法与做法\n\n用 $a, b$ 分别表示 $F_{i-2}, F_{i-1}$，每轮更新为 $(b, a + b)$。\n\n" +
			"## 参考实现\n\n```cpp\n#include <iostream>\nusing namespace std;\nint main() {\n    int n;\n    cin >> n;\n    long long a = 0, b = 1;\n    for (int i = 0; i < n; ++i) {\n        long long c = a + b;\n        a = b; b = c;\n    }\n    cout << a << '\\n';\n    return 0;\n}\n```\n\n" +
			"## 复杂度与易错点\n\n时间复杂度 $O(n)$。\n\n- 注意 `n = 0` 时应输出 $0$。\n- 更大的 $n$ 请用矩阵快速幂 $O(\\log n)$。",
		TimeLimitMS:   1000,
		MemoryLimitMB: 256,
		Visible:       true,
	}

	// --- Problem 4: 最短路径 (中等，验证矩阵 / 求和公式 / 图论 tag) ---
	p4 := models.Problem{
		Title:      "单源最短路径",
		Difficulty: "中等",
		Description: "给定一个 $n$ 个点、$m$ 条边的有向加权图，边权均为正整数。指定源点 $s$，对每个点 $v$ 求从 $s$ 到 $v$ 的最短路径长度 $d(v)$；若不可达输出 $-1$。\n\n" +
			"## 输入格式\n\n第一行三个整数 $n, m, s$，满足 $1 \\le n \\le 10^5,\\ 0 \\le m \\le 2 \\times 10^5,\\ 1 \\le s \\le n$。\n\n" +
			"接下来 $m$ 行，每行三个整数 $u, v, w$，表示一条 $u \\to v$ 权值 $w$ 的有向边，$1 \\le u, v \\le n$，$1 \\le w \\le 10^4$。\n\n" +
			"## 输出格式\n\n一行 $n$ 个整数 $d(1), d(2), \\ldots, d(n)$，空格分隔。\n\n" +
			"## 样例\n\n### 样例输入 1\n\n```\n4 4 1\n1 2 2\n1 3 5\n2 3 2\n3 4 1\n```\n\n### 样例输出 1\n\n```\n0 2 4 5\n```\n\n" +
			"## 说明 / 提示\n\n对 Dijkstra 算法，堆优化版本的复杂度为 $O((n + m)\\log n)$。当 $m = \\Theta(n^2)$ 时可考虑朴素 $O(n^2)$ 实现。",
		SolutionIdeaMD: "## 算法分析\n\n边权非负，使用 Dijkstra 算法。用优先队列（小根堆）扩展当前距离最小的未确定点。\n\n" +
			"## 实现要点\n\n- 邻接表存图，`vector<pair<int,int>>`。\n- 距离数组初始化为 $+\\infty$，$d(s) = 0$。\n- 弹出堆顶时若 `d[u]` 已被更短距离覆盖则跳过（懒惰删除）。\n\n" +
			"## 复杂度分析\n\n时间复杂度 $O((n + m)\\log n)$，空间复杂度 $O(n + m)$。",
		SolutionMD: "## 题目分析\n\n经典单源最短路，边权非负 $\\Rightarrow$ Dijkstra。\n\n" +
			"## 算法与做法\n\n设 $d$ 为最短距离数组，初值 $+\\infty$，$d[s] = 0$。用小根堆维护 $(d[u], u)$，每次取出最小者尝试松弛其所有出边：若 $d[v] > d[u] + w$ 则更新并入堆。\n\n" +
			"## 参考实现\n\n```cpp\n#include <bits/stdc++.h>\nusing namespace std;\nconst long long INF = 1e18;\nint main() {\n    ios::sync_with_stdio(false);\n    cin.tie(nullptr);\n    int n, m, s;\n    cin >> n >> m >> s;\n    vector<vector<pair<int,int>>> g(n + 1);\n    for (int i = 0; i < m; ++i) {\n        int u, v, w; cin >> u >> v >> w;\n        g[u].push_back({v, w});\n    }\n    vector<long long> d(n + 1, INF);\n    d[s] = 0;\n    priority_queue<pair<long long,int>, vector<pair<long long,int>>, greater<>> pq;\n    pq.push({0, s});\n    while (!pq.empty()) {\n        auto [du, u] = pq.top(); pq.pop();\n        if (du > d[u]) continue;\n        for (auto [v, w] : g[u]) {\n            if (d[u] + w < d[v]) {\n                d[v] = d[u] + w;\n                pq.push({d[v], v});\n            }\n        }\n    }\n    for (int i = 1; i <= n; ++i) {\n        cout << (d[i] == INF ? -1 : d[i]) << \" \\n\"[i == n];\n    }\n    return 0;\n}\n```\n\n" +
			"## 复杂度与易错点\n\n时间复杂度 $O((n + m)\\log n)$。\n\n- 忘记懒惰删除会退化为 $O(nm \\log n)$。\n- 无向图需双向加边。\n- 长整型防止累加溢出。",
		TimeLimitMS:   1500,
		MemoryLimitMB: 256,
		Visible:       true,
	}

	// --- Problem 5: 表达式求值 (困难，演示多块公式 + 栈 tag) ---
	p5 := models.Problem{
		Title:      "四则运算表达式求值",
		Difficulty: "困难",
		Description: "给定一个合法的中缀四则运算表达式，仅含非负整数、运算符 `+ - * /` 和小括号 `()`。求其值，除法为向下取整。\n\n" +
			"形式化文法：\n\n$$\\begin{aligned}\\text{Expr} &\\to \\text{Term}\\ ((+|-)\\ \\text{Term})^*\\\\ \\text{Term} &\\to \\text{Factor}\\ ((*|/)\\ \\text{Factor})^*\\\\ \\text{Factor} &\\to \\text{Integer}\\ |\\ (\\ \\text{Expr}\\ )\\end{aligned}$$\n\n" +
			"## 输入格式\n\n一行一个表达式，长度不超过 $10^5$，中间可能含空格。保证过程与结果绝对值均在 $[-10^{18}, 10^{18}]$ 内。\n\n" +
			"## 输出格式\n\n一行一个整数，表示表达式的值。\n\n" +
			"## 样例\n\n### 样例输入 1\n\n```\n1 + 2 * (3 + 4)\n```\n\n### 样例输出 1\n\n```\n15\n```\n\n" +
			"### 样例输入 2\n\n```\n(10 - 3) * 4 / 2\n```\n\n### 样例输出 2\n\n```\n14\n```\n\n" +
			"## 说明 / 提示\n\n除法保证除数非零；采用调度场算法（Shunting-yard）或递归下降均可。",
		SolutionIdeaMD: "## 算法分析\n\n可选两种思路：\n\n1. 调度场算法：用运算符栈 + 数栈，遇到运算符时按优先级弹栈运算。\n2. 递归下降：按文法 `Expr → Term → Factor` 直接递归。\n\n" +
			"## 实现要点\n\n- 先把字符串里的空格滤掉。\n- 运算符优先级：`* /` 高于 `+ -`，括号内先算。\n- 除法向下取整，C++ 的负数 `/` 与数学除法行为不同，需手工处理。\n\n" +
			"## 复杂度分析\n\n时间复杂度 $O(|s|)$，空间复杂度 $O(|s|)$。",
		SolutionMD: "## 题目分析\n\n标准表达式求值。直接套调度场算法。\n\n" +
			"## 算法与做法\n\n1. 预处理：去掉所有空白。\n2. 用两个栈：数栈 `nums`、运算符栈 `ops`。\n3. 遇到数字 → 压 `nums`。\n4. 遇到 `(` → 压 `ops`；遇到 `)` → 不断弹 `ops` 并执行，直到遇到 `(` 弹出。\n5. 遇到 `+ - * /` → 当栈顶优先级 $\\ge$ 当前时弹出并执行，然后压入当前。\n6. 结束后依次弹栈执行。\n\n" +
			"## 参考实现\n\n```cpp\n#include <bits/stdc++.h>\nusing namespace std;\nint prio(char c) { return c == '+' || c == '-' ? 1 : c == '*' || c == '/' ? 2 : 0; }\nlong long apply(long long a, long long b, char op) {\n    if (op == '+') return a + b;\n    if (op == '-') return a - b;\n    if (op == '*') return a * b;\n    long long q = a / b;\n    if ((a % b != 0) && ((a < 0) ^ (b < 0))) --q;\n    return q;\n}\nint main() {\n    string s; getline(cin, s);\n    string t; for (char c : s) if (!isspace((unsigned char)c)) t += c;\n    stack<long long> nums; stack<char> ops;\n    for (int i = 0; i < (int)t.size();) {\n        if (isdigit(t[i])) {\n            long long x = 0;\n            while (i < (int)t.size() && isdigit(t[i])) x = x * 10 + (t[i++] - '0');\n            nums.push(x);\n        } else if (t[i] == '(') { ops.push(t[i++]); }\n        else if (t[i] == ')') {\n            while (!ops.empty() && ops.top() != '(') {\n                long long b = nums.top(); nums.pop();\n                long long a = nums.top(); nums.pop();\n                nums.push(apply(a, b, ops.top())); ops.pop();\n            }\n            ops.pop(); ++i;\n        } else {\n            while (!ops.empty() && ops.top() != '(' && prio(ops.top()) >= prio(t[i])) {\n                long long b = nums.top(); nums.pop();\n                long long a = nums.top(); nums.pop();\n                nums.push(apply(a, b, ops.top())); ops.pop();\n            }\n            ops.push(t[i++]);\n        }\n    }\n    while (!ops.empty()) {\n        long long b = nums.top(); nums.pop();\n        long long a = nums.top(); nums.pop();\n        nums.push(apply(a, b, ops.top())); ops.pop();\n    }\n    cout << nums.top() << '\\n';\n    return 0;\n}\n```\n\n" +
			"## 复杂度与易错点\n\n时间复杂度 $O(|s|)$。\n\n- C++ 的负数除法默认向 0 取整，需要手工向下取整。\n- 括号优先级必须正确处理。\n- 多位数字要连续读完再压栈。",
		TimeLimitMS:   2000,
		MemoryLimitMB: 256,
		Visible:       true,
	}

	type seededProblem struct {
		Problem *models.Problem
		TagIDs  []uint
	}
	seededProblems := []seededProblem{
		{Problem: &p1, TagIDs: nonZero([]uint{tagID("顺序结构")})},
		{Problem: &p2, TagIDs: nonZero([]uint{tagID("最大公约数 gcd")})},
		{Problem: &p3, TagIDs: nonZero([]uint{tagID("线性 DP"), tagID("递推"), tagID("Fibonacci 数列")})},
		{Problem: &p4, TagIDs: nonZero([]uint{tagID("最短路"), tagID("优先队列"), tagID("图遍历")})},
		{Problem: &p5, TagIDs: nonZero([]uint{tagID("栈"), tagID("模拟"), tagID("分类讨论")})},
	}
	for _, item := range seededProblems {
		item.Problem.CreatedBy = 1 // admin seeded by EnsureAdmin is id=1
		if err := db.Create(item.Problem).Error; err != nil {
			return err
		}
		if err := attachProblemTags(db, item.Problem.ID, item.TagIDs); err != nil {
			return err
		}
	}
	probs := []*models.Problem{&p1, &p2, &p3, &p4, &p5}

	// Padding problems: 30 入门级变种，支撑"综合练习 · 30 题"题单，让 UI 能看到
	// 超过 30 道题时的翻页、表格与进度展示。描述简短、tag 统一挂到"顺序结构"。
	seqTag := tagID("顺序结构")
	padding := make([]*models.Problem, 0, 30)
	padTCs := make([]models.Testcase, 0, 60)
	for i := 0; i < 30; i++ {
		op, k, symbol, verb := paddingOp(i)
		title := fmt.Sprintf("%s练习 %d", verb, k)
		desc := fmt.Sprintf(
			"给定整数 $a$，输出 $a %s %d$ 的结果。\n\n"+
				"## 输入格式\n\n一行一个整数 $a$，满足 $-10^9 \\le a \\le 10^9$。\n\n"+
				"## 输出格式\n\n一行一个整数。\n\n"+
				"## 样例\n\n### 样例输入 1\n\n```\n10\n```\n\n### 样例输出 1\n\n```\n%d\n```\n\n"+
				"### 样例输入 2\n\n```\n-5\n```\n\n### 样例输出 2\n\n```\n%d\n```\n",
			symbol, k, paddingApply(10, op, k), paddingApply(-5, op, k))
		p := models.Problem{
			Title:         title,
			Difficulty:    "入门",
			Description:   desc,
			TimeLimitMS:   1000,
			MemoryLimitMB: 256,
			Visible:       true,
			CreatedBy:     1,
		}
		if err := db.Create(&p).Error; err != nil {
			return err
		}
		if err := attachProblemTags(db, p.ID, nonZero([]uint{seqTag})); err != nil {
			return err
		}
		padding = append(padding, &p)
		padTCs = append(padTCs,
			models.Testcase{ProblemID: p.ID, Input: "10\n", ExpectedOutput: fmt.Sprintf("%d\n", paddingApply(10, op, k)), OrderIndex: 1},
			models.Testcase{ProblemID: p.ID, Input: "-5\n", ExpectedOutput: fmt.Sprintf("%d\n", paddingApply(-5, op, k)), OrderIndex: 2},
			models.Testcase{ProblemID: p.ID, Input: "0\n", ExpectedOutput: fmt.Sprintf("%d\n", paddingApply(0, op, k)), OrderIndex: 3},
		)
	}
	if len(padTCs) > 0 {
		if err := db.Create(&padTCs).Error; err != nil {
			return err
		}
	}

	db.Create(&[]models.Testcase{
		{ProblemID: p1.ID, Input: "1 2\n", ExpectedOutput: "3\n", OrderIndex: 1},
		{ProblemID: p1.ID, Input: "100 200\n", ExpectedOutput: "300\n", OrderIndex: 2},
		{ProblemID: p1.ID, Input: "-5 7\n", ExpectedOutput: "2\n", OrderIndex: 3},
	})
	db.Create(&[]models.Testcase{
		{ProblemID: p2.ID, Input: "12 18\n", ExpectedOutput: "6\n", OrderIndex: 1},
		{ProblemID: p2.ID, Input: "17 31\n", ExpectedOutput: "1\n", OrderIndex: 2},
	})
	db.Create(&[]models.Testcase{
		{ProblemID: p3.ID, Input: "0\n", ExpectedOutput: "0\n", OrderIndex: 1},
		{ProblemID: p3.ID, Input: "1\n", ExpectedOutput: "1\n", OrderIndex: 2},
		{ProblemID: p3.ID, Input: "10\n", ExpectedOutput: "55\n", OrderIndex: 3},
		{ProblemID: p3.ID, Input: "40\n", ExpectedOutput: "102334155\n", OrderIndex: 4},
	})
	db.Create(&[]models.Testcase{
		{ProblemID: p4.ID, Input: "4 4 1\n1 2 2\n1 3 5\n2 3 2\n3 4 1\n", ExpectedOutput: "0 2 4 5\n", OrderIndex: 1},
		{ProblemID: p4.ID, Input: "3 1 1\n2 3 5\n", ExpectedOutput: "0 -1 -1\n", OrderIndex: 2},
	})
	db.Create(&[]models.Testcase{
		{ProblemID: p5.ID, Input: "1 + 2 * (3 + 4)\n", ExpectedOutput: "15\n", OrderIndex: 1},
		{ProblemID: p5.ID, Input: "(10 - 3) * 4 / 2\n", ExpectedOutput: "14\n", OrderIndex: 2},
		{ProblemID: p5.ID, Input: "100 / 3\n", ExpectedOutput: "33\n", OrderIndex: 3},
	})

	now := time.Now()
	later := now.Add(7 * 24 * time.Hour)
	// "入门练习" is a plain practice set — no time window, no password, no
	// language restriction. Covers the defaults branch of every gate.
	practice := models.ProblemSet{
		Title:     "入门练习",
		CreatedBy: 1,
	}
	// "周赛 #1" exercises every non-default ProblemSet column: a time window,
	// an admin-set access password, and a narrowed language allow-list. The
	// landing page surfaces the password gate and the submit path narrows the
	// language picker from JudgeLangs ∩ AllowedLangs.
	contest := models.ProblemSet{
		Title:            "周赛 #1",
		Password:         "weekly",
		AllowedLangsJSON: `["cpp","python"]`,
		StartTime:        &now,
		EndTime:          &later,
		CreatedBy:        1,
	}
	db.Create(&practice)
	db.Create(&contest)
	db.Create(&[]models.ProblemSetItem{
		{ProblemSetID: practice.ID, ProblemID: p1.ID, OrderIndex: 0},
		{ProblemSetID: practice.ID, ProblemID: p2.ID, OrderIndex: 1},
		{ProblemSetID: practice.ID, ProblemID: p3.ID, OrderIndex: 2},
		{ProblemSetID: contest.ID, ProblemID: p3.ID, OrderIndex: 0},
		{ProblemSetID: contest.ID, ProblemID: p4.ID, OrderIndex: 1},
		{ProblemSetID: contest.ID, ProblemID: p5.ID, OrderIndex: 2},
	})

	// 第三个题单：覆盖"题单内题目超过 30 题"的 UI 场景（表格滚动、榜首列、进度条）。
	// 不设置密码与语言限制，走默认分支。
	bigSet := models.ProblemSet{
		Title:     "综合练习 · 30 题",
		CreatedBy: 1,
	}
	db.Create(&bigSet)
	bigItems := make([]models.ProblemSetItem, 0, len(padding))
	for i, p := range padding {
		bigItems = append(bigItems, models.ProblemSetItem{
			ProblemSetID: bigSet.ID, ProblemID: p.ID, OrderIndex: i,
		})
	}
	db.Create(&bigItems)

	// 成员关系预置：让 seed 里所有 inSet=... 的提交能合法存在。
	// - practice/contest：stu1..stu3 + admin
	// - bigSet：stu1..stu6 + bulk01..44 共 50 学生 + admin，凑"50 人题单"测试场景
	memberRows := []models.ProblemSetMember{
		{ProblemSetID: practice.ID, UserID: 1, JoinedAt: time.Now()},
		{ProblemSetID: contest.ID, UserID: 1, JoinedAt: time.Now()},
		{ProblemSetID: bigSet.ID, UserID: 1, JoinedAt: time.Now()},
	}
	for _, s := range students[:3] {
		memberRows = append(memberRows,
			models.ProblemSetMember{ProblemSetID: practice.ID, UserID: s.ID, JoinedAt: time.Now()},
			models.ProblemSetMember{ProblemSetID: contest.ID, UserID: s.ID, JoinedAt: time.Now()},
		)
	}
	for _, s := range students {
		memberRows = append(memberRows,
			models.ProblemSetMember{ProblemSetID: bigSet.ID, UserID: s.ID, JoinedAt: time.Now()},
		)
	}
	for _, s := range extras {
		memberRows = append(memberRows,
			models.ProblemSetMember{ProblemSetID: bigSet.ID, UserID: s.ID, JoinedAt: time.Now()},
		)
	}
	if err := db.Create(&memberRows).Error; err != nil {
		return err
	}

	// --- Submissions: hand-crafted to cover every column and verdict.
	// Without these the /submissions table, ranking, verdict-distribution pie,
	// and AI "analyze" cache would all be empty on first boot — and the
	// `Submission.TestcaseResultJSON`, `Message`, `AIExplanation`,
	// `ProblemSetID` columns would never see data.
	if err := seedSubmissions(db, students, probs, &practice, &contest); err != nil {
		return err
	}

	// Padding-set submissions: let stu4/5/6 partially clear "综合练习 · 30 题"
	// so 后台用户管理表格、排行榜 AK 列、个人中心 AC 数/AC 率 都能看到非零数据。
	// 也同时包含 OLE / PE / UKE 新枚举的样例行，保证 VerdictPie 能渲染所有颜色。
	if err := seedPaddingSubmissions(db, students, padding, &bigSet); err != nil {
		return err
	}

	// --- AI tasks: covers every `Kind` and every `Status` so the audit page
	// renders realistic rows + opens non-empty prompt/output modals.
	if err := seedAITasks(db, students); err != nil {
		return err
	}

	log.Printf("seed: %d sample problems + 3 problem sets inserted", len(probs)+len(padding))
	return nil
}

// seedSubmissions writes one submission per {user, problem, verdict} combo
// chosen to touch every Submission column at least once. Kept in a helper so
// ensureSampleProblems stays legible. Admin rows are spread across the past
// two-plus weeks so the heatmap / ranking / 个人中心 stats all render with
// real depth on first boot. The last block seeds full in-set AC sweeps so
// AK rendering (ranking AK column + 个人中心 AK count) has non-zero data.
func seedSubmissions(
	db *gorm.DB, students []models.User, probs []*models.Problem, practice, contest *models.ProblemSet,
) error {
	var admin models.User
	if err := db.Where("role = ?", models.RoleAdmin).First(&admin).Error; err != nil {
		return err
	}
	type subSeed struct {
		user      uint
		prob      uint
		inSet     *uint
		lang      string
		code      string
		verdict   string
		message   string
		cases     string
		timeMS    int
		memKB     int
		aiExplain string
		ageMin    int
	}
	setID := contest.ID
	practiceID := practice.ID
	// Canned good-run testcase JSONs by problem, reused below to avoid
	// re-typing the full JSON blob for every AC row.
	acCasesBy := map[uint]string{
		probs[0].ID: `[{"index":1,"verdict":"AC","time_ms":2,"memory_kb":1024,"message":""},{"index":2,"verdict":"AC","time_ms":3,"memory_kb":1024,"message":""},{"index":3,"verdict":"AC","time_ms":2,"memory_kb":1024,"message":""}]`,
		probs[1].ID: `[{"index":1,"verdict":"AC","time_ms":1,"memory_kb":900,"message":""},{"index":2,"verdict":"AC","time_ms":1,"memory_kb":900,"message":""}]`,
		probs[2].ID: `[{"index":1,"verdict":"AC","time_ms":1,"memory_kb":900,"message":""},{"index":2,"verdict":"AC","time_ms":1,"memory_kb":900,"message":""},{"index":3,"verdict":"AC","time_ms":2,"memory_kb":900,"message":""},{"index":4,"verdict":"AC","time_ms":3,"memory_kb":900,"message":""}]`,
		probs[3].ID: `[{"index":1,"verdict":"AC","time_ms":60,"memory_kb":8000,"message":""},{"index":2,"verdict":"AC","time_ms":40,"memory_kb":8000,"message":""}]`,
		probs[4].ID: `[{"index":1,"verdict":"AC","time_ms":30,"memory_kb":5000,"message":""},{"index":2,"verdict":"AC","time_ms":28,"memory_kb":5000,"message":""},{"index":3,"verdict":"AC","time_ms":35,"memory_kb":5000,"message":""}]`,
	}
	day := 1440 // minutes
	rows := []subSeed{
		// --- students (existing fixtures) ---
		{
			user: students[0].ID, prob: probs[0].ID, lang: "cpp", verdict: models.VerdictAC,
			code:   "#include<iostream>\nusing namespace std;\nint main(){long long a,b;cin>>a>>b;cout<<a+b;}\n",
			cases:  acCasesBy[probs[0].ID],
			timeMS: 3, memKB: 1024, ageMin: 120,
		},
		{
			user: students[0].ID, prob: probs[2].ID, inSet: &setID, lang: "cpp", verdict: models.VerdictAC,
			code:   "#include<iostream>\nint main(){int n;std::cin>>n;long long a=0,b=1;while(n--){long long c=a+b;a=b;b=c;}std::cout<<a;}\n",
			cases:  acCasesBy[probs[2].ID],
			timeMS: 3, memKB: 900, ageMin: 95,
		},
		{
			user: students[0].ID, prob: probs[1].ID, lang: "cpp", verdict: models.VerdictWA,
			code:    "#include<iostream>\nint main(){long long a,b;std::cin>>a>>b;std::cout<<a;}\n",
			message: "第 1 个用例期望 6，实际输出 12",
			cases:   `[{"index":1,"verdict":"WA","time_ms":2,"memory_kb":1000,"message":"expected 6"},{"index":2,"verdict":"AC","time_ms":2,"memory_kb":1000,"message":""}]`,
			timeMS:  2, memKB: 1000,
			aiExplain: "代码只输出了 `a` 而非 `gcd(a, b)`。应当实现辗转相除：`while (b) { a %= b; swap(a, b); }` 然后输出 `a`。\n\n```cpp\n#include<iostream>\nusing namespace std;\nint main(){long long a,b;cin>>a>>b;while(b){a%=b;swap(a,b);}cout<<a;}\n```",
			ageMin:    50,
		},
		{
			user: students[1].ID, prob: probs[4].ID, lang: "python", verdict: models.VerdictTLE,
			code:    "s = input()\nresult = eval(s)  # naive, slow for large inputs\nprint(result)\n",
			message: "用例 1 超过时间限制 2000 ms",
			cases:   `[{"index":1,"verdict":"TLE","time_ms":2000,"memory_kb":12000,"message":""},{"index":2,"verdict":"TLE","time_ms":2000,"memory_kb":12000,"message":""},{"index":3,"verdict":"TLE","time_ms":2000,"memory_kb":12000,"message":""}]`,
			timeMS:  2000, memKB: 12000, ageMin: 30,
		},
		{
			user: students[1].ID, prob: probs[3].ID, lang: "cpp", verdict: models.VerdictRE,
			code:    "int main(){int*p=0;*p=1;}\n",
			message: "运行时错误：SIGSEGV",
			cases:   `[{"index":1,"verdict":"RE","time_ms":5,"memory_kb":2048,"message":"SIGSEGV"}]`,
			timeMS:  5, memKB: 2048, ageMin: 24,
			aiExplain: "代码故意解引用了空指针，导致段错误。最短路题应实现 Dijkstra；参考 solution_md 中的模板。",
		},
		{
			user: students[1].ID, prob: probs[0].ID, lang: "cpp", verdict: models.VerdictCE,
			code:    "#include <iostream>\nint main() { std::cout << a + b; }\n",
			message: "编译错误：'a' was not declared in this scope",
			cases:   `[]`,
			timeMS:  0, memKB: 0, ageMin: 18,
			aiExplain: "变量 `a`、`b` 未声明也未读入。需要 `long long a, b; std::cin >> a >> b;`。",
		},
		{
			user: students[2].ID, prob: probs[3].ID, lang: "cpp", verdict: models.VerdictMLE,
			code:    "#include<bits/stdc++.h>\nusing namespace std;\nint main(){vector<vector<int>> huge(100000,vector<int>(100000));cout<<0;}\n",
			message: "内存超限",
			cases:   `[{"index":1,"verdict":"MLE","time_ms":100,"memory_kb":300000,"message":""}]`,
			timeMS:  100, memKB: 300000, ageMin: 10,
		},
		{
			user: students[2].ID, prob: probs[0].ID, lang: "python", verdict: models.VerdictAC,
			code:   "a, b = map(int, input().split())\nprint(a + b)\n",
			cases:  `[{"index":1,"verdict":"AC","time_ms":30,"memory_kb":8000,"message":""},{"index":2,"verdict":"AC","time_ms":28,"memory_kb":8000,"message":""},{"index":3,"verdict":"AC","time_ms":29,"memory_kb":8000,"message":""}]`,
			timeMS: 30, memKB: 8000, ageMin: 5,
		},
		{
			user: students[2].ID, prob: probs[1].ID, lang: "cpp", verdict: models.VerdictSE,
			code:    "int main(){}\n",
			message: "sandbox launcher returned non-zero",
			cases:   `[]`,
			timeMS:  0, memKB: 0, ageMin: 2,
		},
		{
			user: students[0].ID, prob: probs[3].ID, lang: "cpp", verdict: models.VerdictPending,
			code:   "int main(){}\n",
			cases:  ``,
			timeMS: 0, memKB: 0, ageMin: 0,
		},

		// --- admin: dogfood every problem across the current month so
		// ranking / heatmap / 个人中心 distribution all render with depth ---
		{
			user: admin.ID, prob: probs[0].ID, lang: "cpp", verdict: models.VerdictAC,
			code:  "#include<iostream>\nusing namespace std;\nint main(){long long a,b;cin>>a>>b;cout<<a+b<<'\\n';}\n",
			cases: acCasesBy[probs[0].ID], timeMS: 2, memKB: 1024,
			ageMin: 18 * day,
		},
		{
			user: admin.ID, prob: probs[0].ID, lang: "python", verdict: models.VerdictAC,
			code:  "a, b = map(int, input().split())\nprint(a + b)\n",
			cases: acCasesBy[probs[0].ID], timeMS: 25, memKB: 7500,
			ageMin: 17 * day,
		},
		{
			user: admin.ID, prob: probs[1].ID, lang: "cpp", verdict: models.VerdictWA,
			code:    "#include<iostream>\nint main(){long long a,b;std::cin>>a>>b;std::cout<<(a+b);}\n",
			message: "用例 1 期望 6，实际 30",
			cases:   `[{"index":1,"verdict":"WA","time_ms":2,"memory_kb":900,"message":"expected 6"}]`,
			timeMS:  2, memKB: 900, ageMin: 15*day + 120,
		},
		{
			user: admin.ID, prob: probs[1].ID, lang: "cpp", verdict: models.VerdictAC,
			code:  "#include<iostream>\nusing namespace std;\nlong long g(long long a,long long b){return b?g(b,a%b):a;}\nint main(){long long a,b;cin>>a>>b;cout<<g(a,b)<<'\\n';}\n",
			cases: acCasesBy[probs[1].ID], timeMS: 1, memKB: 900,
			ageMin: 15*day + 60,
		},
		{
			user: admin.ID, prob: probs[2].ID, lang: "cpp", verdict: models.VerdictAC,
			code:  "#include<iostream>\nint main(){int n;std::cin>>n;long long a=0,b=1;while(n--){long long c=a+b;a=b;b=c;}std::cout<<a;}\n",
			cases: acCasesBy[probs[2].ID], timeMS: 2, memKB: 900,
			ageMin: 13 * day,
		},
		{
			user: admin.ID, prob: probs[3].ID, lang: "python", verdict: models.VerdictTLE,
			code:    "# naive Bellman-Ford, too slow\nimport sys\n...\n",
			message: "所有用例超时",
			cases:   `[{"index":1,"verdict":"TLE","time_ms":1500,"memory_kb":11000,"message":""},{"index":2,"verdict":"TLE","time_ms":1500,"memory_kb":11000,"message":""}]`,
			timeMS:  1500, memKB: 11000, ageMin: 11*day + 300,
		},
		{
			user: admin.ID, prob: probs[3].ID, lang: "cpp", verdict: models.VerdictAC,
			code:  "// Dijkstra + 堆优化，复杂度 O((n+m) log n)\n#include<bits/stdc++.h>\nusing namespace std;\nint main(){ /* ... */ return 0; }\n",
			cases: acCasesBy[probs[3].ID], timeMS: 45, memKB: 9000,
			ageMin: 11 * day,
		},
		{
			user: admin.ID, prob: probs[4].ID, lang: "cpp", verdict: models.VerdictAC,
			code:  "// 调度场算法\n#include<bits/stdc++.h>\nusing namespace std;\nint main(){ /* ... */ return 0; }\n",
			cases: acCasesBy[probs[4].ID], timeMS: 12, memKB: 3800,
			ageMin: 9 * day,
		},
		{
			user: admin.ID, prob: probs[0].ID, inSet: &setID, lang: "cpp", verdict: models.VerdictAC,
			code:  "#include<iostream>\nint main(){long long a,b;std::cin>>a>>b;std::cout<<a+b;}\n",
			cases: acCasesBy[probs[0].ID], timeMS: 2, memKB: 1024,
			ageMin: 7 * day,
		},
		{
			user: admin.ID, prob: probs[2].ID, inSet: &setID, lang: "python", verdict: models.VerdictAC,
			code:  "n = int(input())\na, b = 0, 1\nfor _ in range(n):\n    a, b = b, a + b\nprint(a)\n",
			cases: acCasesBy[probs[2].ID], timeMS: 28, memKB: 7800,
			ageMin: 5 * day,
		},
		{
			user: admin.ID, prob: probs[4].ID, inSet: &setID, lang: "cpp", verdict: models.VerdictAC,
			code:  "// same as standalone submission\n#include<bits/stdc++.h>\nint main(){}\n",
			cases: acCasesBy[probs[4].ID], timeMS: 14, memKB: 3900,
			ageMin: 3 * day,
		},
		{
			user: admin.ID, prob: probs[1].ID, lang: "python", verdict: models.VerdictAC,
			code:  "import math\na, b = map(int, input().split())\nprint(math.gcd(a, b))\n",
			cases: acCasesBy[probs[1].ID], timeMS: 26, memKB: 7400,
			ageMin: 1 * day,
		},
		{
			user: admin.ID, prob: probs[3].ID, lang: "cpp", verdict: models.VerdictAC,
			code:  "// 再来一遍巩固\n#include<bits/stdc++.h>\nint main(){}\n",
			cases: acCasesBy[probs[3].ID], timeMS: 42, memKB: 8800,
			ageMin: 360,
		},

		// --- AK demo data: every in-set AC below carries an explicit inSet so
		// the AK aggregator (which JOINs on problem_set_id + problem_id) can
		// recognise the full sweep. Admin completes both practice (p1+p2+p3)
		// and contest (p3+p4+p5); student[0] completes contest with realistic
		// WA-before-AC attempts so the ranking penalty column has non-zero data.
		{
			user: admin.ID, prob: probs[0].ID, inSet: &practiceID, lang: "cpp", verdict: models.VerdictAC,
			code:  "#include<iostream>\nint main(){long long a,b;std::cin>>a>>b;std::cout<<a+b;}\n",
			cases: acCasesBy[probs[0].ID], timeMS: 2, memKB: 1024,
			ageMin: 8 * day,
		},
		{
			user: admin.ID, prob: probs[1].ID, inSet: &practiceID, lang: "cpp", verdict: models.VerdictAC,
			code:  "#include<iostream>\nusing namespace std;\nlong long g(long long a,long long b){return b?g(b,a%b):a;}\nint main(){long long a,b;cin>>a>>b;cout<<g(a,b);}\n",
			cases: acCasesBy[probs[1].ID], timeMS: 1, memKB: 900,
			ageMin: 8*day - 60,
		},
		{
			user: admin.ID, prob: probs[2].ID, inSet: &practiceID, lang: "cpp", verdict: models.VerdictAC,
			code:  "#include<iostream>\nint main(){int n;std::cin>>n;long long a=0,b=1;while(n--){long long c=a+b;a=b;b=c;}std::cout<<a;}\n",
			cases: acCasesBy[probs[2].ID], timeMS: 2, memKB: 900,
			ageMin: 8*day - 120,
		},
		{
			user: admin.ID, prob: probs[3].ID, inSet: &setID, lang: "cpp", verdict: models.VerdictAC,
			code:  "// Dijkstra 堆优化（题单内）\n#include<bits/stdc++.h>\nint main(){}\n",
			cases: acCasesBy[probs[3].ID], timeMS: 50, memKB: 9200,
			ageMin: 4 * day,
		},
		{
			user: students[0].ID, prob: probs[3].ID, inSet: &setID, lang: "cpp", verdict: models.VerdictWA,
			code:    "// student 第一次交错了\n#include<bits/stdc++.h>\nint main(){}\n",
			message: "用例 1 输出不匹配",
			cases:   `[{"index":1,"verdict":"WA","time_ms":50,"memory_kb":9000,"message":"expected 0 2 4 5"}]`,
			timeMS:  50, memKB: 9000, ageMin: 2*day + 300,
		},
		{
			user: students[0].ID, prob: probs[3].ID, inSet: &setID, lang: "cpp", verdict: models.VerdictAC,
			code:  "// student 改对了\n#include<bits/stdc++.h>\nint main(){}\n",
			cases: acCasesBy[probs[3].ID], timeMS: 55, memKB: 9200,
			ageMin: 2 * day,
		},
		{
			user: students[0].ID, prob: probs[4].ID, inSet: &setID, lang: "cpp", verdict: models.VerdictTLE,
			code:    "// student 第一次太慢\n#include<bits/stdc++.h>\nint main(){}\n",
			message: "超时",
			cases:   `[{"index":1,"verdict":"TLE","time_ms":2000,"memory_kb":4000,"message":""}]`,
			timeMS:  2000, memKB: 4000, ageMin: day + 600,
		},
		{
			user: students[0].ID, prob: probs[4].ID, inSet: &setID, lang: "cpp", verdict: models.VerdictAC,
			code:  "// student 调度场 AC\n#include<bits/stdc++.h>\nint main(){}\n",
			cases: acCasesBy[probs[4].ID], timeMS: 20, memKB: 4200,
			ageMin: day + 60,
		},
	}
	for _, r := range rows {
		sub := &models.Submission{
			UserID: r.user, ProblemID: r.prob, ProblemSetID: r.inSet,
			Language: r.lang, Code: r.code, Verdict: r.verdict, Message: r.message,
			TimeUsedMS: r.timeMS, MemoryUsedKB: r.memKB,
			TestcaseResultJSON: r.cases, AIExplanation: r.aiExplain,
			CreatedAt: time.Now().Add(-time.Duration(r.ageMin) * time.Minute),
		}
		if err := db.Create(sub).Error; err != nil {
			return err
		}
	}
	log.Printf("seed: %d sample submissions inserted", len(rows))
	return nil
}

// seedAITasks writes one row per task Kind × representative Status so the
// AI 队列 page opens with realistic data and every AITask column gets hit.
func seedAITasks(db *gorm.DB, students []models.User) error {
	base := time.Now()
	finished := func(min int) *time.Time {
		t := base.Add(-time.Duration(min) * time.Minute)
		return &t
	}
	rows := []models.AITask{
		{
			Kind: models.AITaskKindAnalyze, UserID: students[0].ID, Username: students[0].Name,
			Subject: "submission #3", Status: models.AITaskStatusDone,
			StartedAt:  base.Add(-45 * time.Minute),
			FinishedAt: finished(44),
			DurationMS: 38_000,
			Prompt:     "## system\n你是 OJ 平台的答疑助手。...\n\n## user\n题目标题：最大公约数\n提交代码：...\n判题结果：WA",
			Output:     "代码只输出了 `a` 而非 `gcd(a, b)`。应当实现辗转相除...",
		},
		{
			Kind: models.AITaskKindTag, UserID: 1, Username: "超级管理员",
			Subject: "problem #2", Status: models.AITaskStatusDone,
			StartedAt:  base.Add(-40 * time.Minute),
			FinishedAt: finished(39),
			DurationMS: 12_300,
			Prompt:     "## system\n你是 OJ 题目标签 + 难度助手。...\n\n## user\n可选一级标签与二级标签字典：...\n题目原文（详细）：输入两个正整数 a, b，输出 gcd(a, b)。",
			Output:     `{"tags":[{"group":"数论","tag":"最大公约数 gcd"}],"difficulty":"入门"}`,
		},
		{
			Kind: models.AITaskKindGenTitle, UserID: 1, Username: "超级管理员",
			Subject: "problem #3", Status: models.AITaskStatusDone,
			StartedAt:  base.Add(-35 * time.Minute),
			FinishedAt: finished(34),
			DurationMS: 3_200,
			Prompt:     "## system\n你是 OJ 题目命名助手。...\n\n## user\n输入 n，输出 Fibonacci 第 n 项 ...",
			Output:     "斐波那契数",
		},
		{
			Kind: models.AITaskKindGenDesc, UserID: 1, Username: "超级管理员",
			Subject: "problem #4", Status: models.AITaskStatusDone,
			StartedAt:  base.Add(-30 * time.Minute),
			FinishedAt: finished(28),
			DurationMS: 82_000,
			Prompt:     "## system\n你是 OJ 题目描述编辑。...\n\n## user\nDijkstra 最短路题干 ...",
			Output:     "## 题目描述\n\n给定一个带权有向图 ...\n\n## 输入格式\n\n第一行 $n, m, s$ ...",
		},
		{
			Kind: models.AITaskKindGenAll, UserID: 1, Username: "超级管理员",
			Subject: "problem #5", Status: models.AITaskStatusFailed,
			StartedAt:  base.Add(-15 * time.Minute),
			FinishedAt: finished(14),
			DurationMS: 60_000,
			Error:      "ai: read response body: context deadline exceeded",
			Prompt:     "## system\n你是 OJ 题目生成流水线。...\n\n## user\n四则运算表达式求值原文 ...",
			Output:     `{"choices":[{"message":{"content":"{\"title\":\"表达式求值\",\"description\":\"## 题目描述\\n\\n给定 ...`,
		},
		{
			Kind: models.AITaskKindGenIdea, UserID: 1, Username: "超级管理员",
			Subject: "problem #4", Status: models.AITaskStatusRunning,
			StartedAt:  base.Add(-1 * time.Minute),
			DurationMS: 0,
			Prompt:     "## system\n你是 OJ 题目思路编辑。...\n\n## user\n最短路题干 ...",
		},
	}
	for i := range rows {
		if err := db.Create(&rows[i]).Error; err != nil {
			return err
		}
	}
	log.Printf("seed: %d sample AI tasks inserted", len(rows))
	return nil
}

func nonZero(ids []uint) []uint {
	out := make([]uint, 0, len(ids))
	for _, id := range ids {
		if id != 0 {
			out = append(out, id)
		}
	}
	return out
}

func attachProblemTags(db *gorm.DB, problemID uint, tagIDs []uint) error {
	tagIDs = nonZero(tagIDs)
	if len(tagIDs) == 0 {
		return nil
	}
	seen := make(map[uint]bool, len(tagIDs))
	rows := make([]models.ProblemTag, 0, len(tagIDs))
	for _, tagID := range tagIDs {
		if seen[tagID] {
			continue
		}
		seen[tagID] = true
		rows = append(rows, models.ProblemTag{ProblemID: problemID, TagID: tagID})
	}
	return db.Create(&rows).Error
}

// paddingOp returns the arithmetic spec for the i-th padding problem. i in
// 0..29 covers three buckets so the generated题单 has a mix of 加 / 乘 / 减：
//
//	0..9  → a + (i+1),     LaTeX symbol "+"
//	10..19 → a × (i-8),    LaTeX symbol "\times", k ∈ [2, 11]
//	20..29 → a - (i-19),   LaTeX symbol "-"
func paddingOp(i int) (op string, k int, symbol, verb string) {
	switch {
	case i < 10:
		return "+", i + 1, "+", "加法"
	case i < 20:
		return "*", i - 8, "\\times", "乘法"
	default:
		return "-", i - 19, "-", "减法"
	}
}

// paddingApply evaluates the arithmetic op so testcase expected-output can be
// generated at seed time rather than hand-written.
func paddingApply(a int, op string, k int) int {
	switch op {
	case "+":
		return a + k
	case "*":
		return a * k
	case "-":
		return a - k
	}
	return a
}

// seedPaddingSubmissions fills the 30-problem set with partial sweeps + a small
// cross-section of the new OLE / PE / UKE verdicts. Without this:
//
//	· 后台用户管理页 stu4/5/6 的统计列全是 0
//	· VerdictPie 永远看不到 OLE / PE / UKE 色块
//	· 个人中心 AK 数不会被新题单触发
func seedPaddingSubmissions(
	db *gorm.DB, students []models.User, padding []*models.Problem, bigSet *models.ProblemSet,
) error {
	if len(padding) == 0 {
		return nil
	}
	setID := bigSet.ID
	// stu4: AK 整个 30 题题单 → 排行榜 AK=1、个人中心 AK=1
	// stu5: 完成前 20 题，夹杂 WA/OLE/PE → AC 率中等、VerdictPie 多样
	// stu6: 完成前 10 题 + 若干错题 + 一次 UKE → 最低 AC 率、UKE 出现
	day := 1440
	rows := make([]models.Submission, 0, 120)
	addSub := func(uid uint, pid uint, verdict string, ageMin int, code string, cases string, timeMS, memKB int, message string) {
		rows = append(rows, models.Submission{
			UserID: uid, ProblemID: pid, ProblemSetID: &setID,
			Language: "cpp", Code: code, Verdict: verdict, Message: message,
			TimeUsedMS: timeMS, MemoryUsedKB: memKB,
			TestcaseResultJSON: cases,
			CreatedAt:          time.Now().Add(-time.Duration(ageMin) * time.Minute),
		})
	}
	acCases := `[{"index":1,"verdict":"AC","time_ms":2,"memory_kb":800,"message":""},{"index":2,"verdict":"AC","time_ms":2,"memory_kb":800,"message":""},{"index":3,"verdict":"AC","time_ms":2,"memory_kb":800,"message":""}]`
	waCases := `[{"index":1,"verdict":"WA","time_ms":2,"memory_kb":800,"message":"expected X got Y"}]`
	oleCases := `[{"index":1,"verdict":"OLE","time_ms":5,"memory_kb":1500,"message":"stdout exceeded 1MB"}]`
	peCases := `[{"index":1,"verdict":"PE","time_ms":2,"memory_kb":800,"message":"trailing whitespace differs"}]`
	ukeCases := `[]`

	okCode := "#include<iostream>\nint main(){long long a;std::cin>>a;std::cout<<a;}\n"
	badCode := "#include<iostream>\nint main(){long long a;std::cin>>a;std::cout<<(a+1);}\n"
	oleCode := "#include<iostream>\nint main(){for(int i=0;i<2000000;i++) std::cout<<'x';}\n"
	peCode := "#include<iostream>\nint main(){long long a;std::cin>>a;std::cout<<a<<\"   \\n\\n\";}\n"

	// stu4 — full 30-problem AK，时间分布在最近两周
	for i, p := range padding {
		addSub(students[3].ID, p.ID, models.VerdictAC, (14-i/3)*day+i*30, okCode, acCases, 2, 800, "")
	}
	// stu5 — 20 个 AC + 5 WA + 3 OLE + 2 PE
	for i := 0; i < 20; i++ {
		addSub(students[4].ID, padding[i].ID, models.VerdictAC, (10-i/4)*day+i*20, okCode, acCases, 2, 800, "")
	}
	for i := 20; i < 25; i++ {
		addSub(students[4].ID, padding[i].ID, models.VerdictWA, 3*day+(i-20)*60, badCode, waCases, 2, 800, "输出不匹配")
	}
	for i := 25; i < 28; i++ {
		addSub(students[4].ID, padding[i].ID, models.VerdictOLE, 2*day+(i-25)*60, oleCode, oleCases, 5, 1500, "输出超限")
	}
	for i := 28; i < 30; i++ {
		addSub(students[4].ID, padding[i].ID, models.VerdictPE, day+(i-28)*60, peCode, peCases, 2, 800, "格式错误")
	}
	// stu6 — 10 AC + 6 WA + 2 OLE + 1 PE + 1 UKE
	for i := 0; i < 10; i++ {
		addSub(students[5].ID, padding[i].ID, models.VerdictAC, (8-i/3)*day+i*30, okCode, acCases, 2, 800, "")
	}
	for i := 10; i < 16; i++ {
		addSub(students[5].ID, padding[i].ID, models.VerdictWA, 5*day+(i-10)*60, badCode, waCases, 2, 800, "输出不匹配")
	}
	addSub(students[5].ID, padding[16].ID, models.VerdictOLE, 4*day+10, oleCode, oleCases, 5, 1500, "输出超限")
	addSub(students[5].ID, padding[17].ID, models.VerdictOLE, 4*day+70, oleCode, oleCases, 5, 1500, "输出超限")
	addSub(students[5].ID, padding[18].ID, models.VerdictPE, 3*day+30, peCode, peCases, 2, 800, "格式错误")
	addSub(students[5].ID, padding[19].ID, models.VerdictUKE, 2*day+30, okCode, ukeCases, 0, 0, "判题机内部错误")

	if err := db.Create(&rows).Error; err != nil {
		return err
	}
	log.Printf("seed: %d padding-set submissions inserted", len(rows))
	return nil
}

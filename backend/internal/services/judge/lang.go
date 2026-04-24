package judge

// Language defines how to compile and run source for one language.
type Language struct {
	ID         string
	Src        string   // source filename inside sandbox
	Compile    []string // nil if interpreted
	CompileOut string   // artifact name copied out (for compiled langs)
	Run        []string // args to execute; use {{bin}} placeholder resolved by runner
	Env        []string
}

// Languages is indexed by the id used in the /api JudgeLangs list.
var Languages = map[string]Language{
	"c": {
		ID:         "c",
		Src:        "main.c",
		Compile:    []string{"/usr/bin/gcc", "-O2", "-w", "-o", "main", "main.c"},
		CompileOut: "main",
		Run:        []string{"./main"},
	},
	"cpp": {
		ID:         "cpp",
		Src:        "main.cpp",
		Compile:    []string{"/usr/bin/g++", "-O2", "-w", "-o", "main", "main.cpp"},
		CompileOut: "main",
		Run:        []string{"./main"},
	},
	"java": {
		ID:         "java",
		Src:        "Main.java",
		Compile:    []string{"/usr/bin/javac", "Main.java"},
		CompileOut: "Main.class",
		Run:        []string{"/usr/bin/java", "Main"},
	},
	"python": {
		ID:  "python",
		Src: "main.py",
		Run: []string{"/usr/bin/python3", "main.py"},
	},
}

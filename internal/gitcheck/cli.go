// Package gitcheck implements the gitcheck command, shared by the
// cmd/gitcheck shim and by "hall gitcheck".
package gitcheck

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/QuietSugar/hall-kit/internal/i18n"
)

type checkResult struct {
	name    string
	safe    bool
	details []string
	// 汇总表中显示的短标签，如 "[未推送3]"；safe 时为空
	tag string
	// 该项检查的耗时，详情模式 (-v) 下展示
	elapsed time.Duration
}

var verifyRemote bool

// strict 为 true 时，被 gitignore 忽略的文件也计入「未跟踪/被忽略文件」检查
var strict bool

// netTimeout 是联网 git 调用（ls-remote、lfs push 等）的快速失败超时。
// 访问不可达的远程（如被墙的 github.com）时，由操作系统默认 TCP 超时
// 可能要等几分钟，这里统一限制为 15 秒。
const netTimeout = 15 * time.Second

// netTimeoutError 标记一次联网 git 调用超时。
type netTimeoutError struct {
	remote  string
	timeout time.Duration
}

func (e *netTimeoutError) Error() string {
	return fmt.Sprintf("git call to remote %q exceeded %v", e.remote, e.timeout)
}

// netErrPatterns 用于判断 git 报错是否属于网络不可达类，
// 命中时提示用户可能需要配置代理。
var netErrPatterns = []string{
	"timed out",
	"timeout",
	"deadline exceeded",
	"could not resolve",
	"temporary failure in name resolution",
	"failed to connect",
	"connection refused",
	"connection reset",
	"no route to host",
	"network is unreachable",
	"proxyconnect",
	"connect tunnel failed",
	"returned error: 502",
	"returned error: 503",
	"returned error: 504",
}

func isNetError(err error) bool {
	s := strings.ToLower(err.Error())
	for _, p := range netErrPatterns {
		if strings.Contains(s, p) {
			return true
		}
	}
	return false
}

// netErrorDetail 把网络类错误归类为「超时」或「不可达」（附代理提示由
// 调用方决定），非网络类错误（如认证失败、仓库不存在）返回 fallback。
func netErrorDetail(remote string, err error, fallback string) string {
	var nt *netTimeoutError
	if errors.As(err, &nt) {
		return i18n.T("gitcheck.net.timeout", remote, int(netTimeout.Seconds()))
	}
	if isNetError(err) {
		return i18n.T("gitcheck.net.unreachable", remote, err)
	}
	return fallback
}

// runGit 执行 git 命令并合并 stderr 到错误信息中。
func runGit(ctx context.Context, dir string, args ...string) (string, error) {
	cmd := exec.CommandContext(ctx, "git", args...)
	cmd.Dir = dir
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	err := cmd.Run()
	if errors.Is(ctx.Err(), context.DeadlineExceeded) {
		return "", &netTimeoutError{remote: dir, timeout: netTimeout}
	}
	if err != nil {
		msg := strings.TrimSpace(stderr.String())
		if msg != "" {
			return "", fmt.Errorf("%w: %s", err, msg)
		}
		return "", err
	}
	return stdout.String(), nil
}

// git 执行本地 git 调用（不限时）。
func git(dir string, args ...string) (string, error) {
	return runGit(context.Background(), dir, args...)
}

// gitNet 执行联网 git 调用，超过 netTimeout 快速失败。
func gitNet(dir string, args ...string) (string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), netTimeout)
	defer cancel()
	return runGit(ctx, dir, args...)
}

func must(s string, _ error) string { return s }

func lines(s string) []string {
	var out []string
	for _, l := range strings.Split(strings.TrimRight(s, "\n"), "\n") {
		if strings.TrimSpace(l) != "" {
			out = append(out, l)
		}
	}
	return out
}

func remoteRefLines(dir string) []string {
	out, _ := git(dir, "for-each-ref", "--format=%(refname)", "refs/remotes/")
	return lines(out)
}

// ---------- 单项检查 ----------

// 1. 未跟踪 / 被忽略的文件
//
// 默认只检查未跟踪文件（??）；被忽略文件（!!）多为可重建产物，
// 仅在 --strict 模式下计入。
func checkUntracked(dir string) checkResult {
	r := checkResult{name: i18n.T("gitcheck.check.untracked.name"), safe: true}
	args := []string{"status", "--porcelain"}
	if strict {
		args = append(args, "--ignored")
	}
	out, err := git(dir, args...)
	if err != nil {
		r.safe = false
		r.details = append(r.details, i18n.T("gitcheck.error.gitstatus", err))
		return r
	}
	for _, l := range lines(out) {
		if len(l) < 2 {
			continue
		}
		if code := l[:2]; code == "??" || (strict && code == "!!") {
			r.safe = false
			r.details = append(r.details, l)
		}
	}
	if !r.safe {
		r.tag = i18n.T("gitcheck.untracked.tag", len(r.details))
	}
	return r
}

// 2. 未提交的改动（工作区 + 暂存区）
func checkUncommitted(dir string) checkResult {
	r := checkResult{name: i18n.T("gitcheck.check.uncommitted.name"), safe: true}
	out, err := git(dir, "status", "--porcelain")
	if err != nil {
		r.safe = false
		r.details = append(r.details, i18n.T("gitcheck.error.gitstatus", err))
		return r
	}
	for _, l := range lines(out) {
		if len(l) >= 2 && l[:2] != "??" {
			r.safe = false
			r.details = append(r.details, l)
		}
	}
	if !r.safe {
		r.tag = i18n.T("gitcheck.uncommitted.tag", len(r.details))
	}
	return r
}

// 3. stash（stash 只存在于本地 .git，从不推送，有即不可删）
func checkStash(dir string) checkResult {
	r := checkResult{name: i18n.T("gitcheck.check.stash.name"), safe: true}
	out, err := git(dir, "stash", "list")
	if err != nil {
		return r
	}
	if ls := lines(out); len(ls) > 0 {
		r.safe = false
		r.details = ls
		r.tag = i18n.T("gitcheck.stash.tag", len(ls))
	}
	return r
}

// 4. 本地分支是否有未推送的提交（对照所有远程跟踪分支，而非仅上游）
func checkBranches(dir string) checkResult {
	r := checkResult{name: i18n.T("gitcheck.check.branches.name"), safe: true}

	branchesOut, err := git(dir, "for-each-ref", "--format=%(refname:short)", "refs/heads/")
	if err != nil {
		return r
	}
	remotes := remoteRefLines(dir)
	total := 0

	for _, br := range lines(branchesOut) {
		args := []string{"rev-list", br, "--not"}
		if len(remotes) > 0 {
			args = append(args, "--remotes")
		}
		out, err := git(dir, args...)
		if err != nil {
			continue
		}
		if unpushed := lines(out); len(unpushed) > 0 {
			r.safe = false
			total += len(unpushed)
			if len(remotes) == 0 {
				r.details = append(r.details, i18n.T("gitcheck.branches.noremote", br, len(unpushed)))
			} else {
				r.details = append(r.details, i18n.T("gitcheck.branches.unpushed", br, len(unpushed)))
			}
		}
	}
	if !r.safe {
		r.tag = i18n.T("gitcheck.branches.tag", total)
	}
	return r
}

// 5. 游离提交：detached HEAD 上不属于任何分支/远程的提交
//
// 只检查 detached HEAD（用 rev-list，毫秒级）。不做 git fsck 全量
// 可达性分析：它要遍历整个对象库（大仓库秒级），而查出的不可达提交
// 几乎都是 amend/rebase/reset 的正常残骸，保护价值极低。
func checkDangling(dir string) checkResult {
	r := checkResult{name: i18n.T("gitcheck.check.dangling.name"), safe: true}

	head, err := git(dir, "symbolic-ref", "-q", "HEAD")
	if err != nil || strings.TrimSpace(head) == "" {
		headSHA, err := git(dir, "rev-parse", "HEAD")
		if err == nil {
			sha := strings.TrimSpace(headSHA)
			// 注意不能用 --all：git 的 --all 包含 HEAD 本身，会把结果清空
			args := []string{"rev-list", "HEAD", "--not", "--branches"}
			if len(remoteRefLines(dir)) > 0 {
				args = append(args, "--remotes")
			}
			if out, err := git(dir, args...); err == nil {
				if u := lines(out); len(u) > 0 {
					r.safe = false
					r.details = append(r.details, i18n.T("gitcheck.dangling.detail", sha[:8], len(u)))
					r.tag = i18n.T("gitcheck.dangling.tag", len(u))
				}
			}
		}
	}
	return r
}

// 6. tag 推送状态（仅 --verify-remote 时检查）
//
// 离线模式下本地无法区分 tag 是否已推送：fetch 会把远程 tag 合并进本地
// refs/tags，而"tag 指向的提交不在任何远程分支上"是发版后分支被删的
// 常态，用它做判据几乎必然误报。因此只在 --verify-remote 模式下
// 与远程 tag 列表精确比对。
func checkTags(dir string) checkResult {
	r := checkResult{name: i18n.T("gitcheck.check.tags.name"), safe: true}
	if !verifyRemote {
		return r
	}
	tags := lines(must(git(dir, "tag", "--list")))
	if len(tags) == 0 {
		return r
	}

	netFailed := false
	for _, remote := range lines(must(git(dir, "remote"))) {
		out, err := gitNet(dir, "ls-remote", "--tags", remote)
		if err != nil {
			r.safe = false
			netFailed = true
			r.details = append(r.details, netErrorDetail(remote, err, i18n.T("gitcheck.tags.remoteerr", remote, err)))
			continue
		}
		remoteTags := map[string]bool{}
		for _, l := range lines(out) {
			parts := strings.Fields(l)
			if len(parts) == 2 {
				name := strings.TrimPrefix(parts[1], "refs/tags/")
				name = strings.TrimSuffix(name, "^{}")
				remoteTags[name] = true
			}
		}
		for _, t := range tags {
			if !remoteTags[t] {
				r.safe = false
				r.details = append(r.details, i18n.T("gitcheck.tags.missing", t, remote))
			}
		}
	}
	if !r.safe {
		r.tag = i18n.T("gitcheck.tags.tag", len(r.details))
	}
	if netFailed {
		r.details = append(r.details, i18n.T("gitcheck.net.proxyhint"))
	}
	return r
}

// 7. 远程仓库配置
func checkRemote(dir string) checkResult {
	r := checkResult{name: i18n.T("gitcheck.check.remote.name"), safe: true}
	out, err := git(dir, "remote")
	if err != nil || len(lines(out)) == 0 {
		r.safe = false
		r.details = append(r.details, i18n.T("gitcheck.remote.none"))
		r.tag = i18n.T("gitcheck.remote.tag")
	}
	return r
}

// 8. （--verify-remote）联网验证远程真实存在且可达
func checkRemoteAlive(dir string) checkResult {
	r := checkResult{name: i18n.T("gitcheck.check.remotealive.name"), safe: true}
	netFailed := false
	for _, remote := range lines(must(git(dir, "remote"))) {
		if _, err := gitNet(dir, "ls-remote", "--heads", remote); err != nil {
			r.safe = false
			netFailed = true
			r.details = append(r.details, netErrorDetail(remote, err, i18n.T("gitcheck.remotealive.unreachable", remote, err)))
		}
	}
	if !r.safe {
		r.tag = i18n.T("gitcheck.remotealive.tag")
	}
	if netFailed {
		r.details = append(r.details, i18n.T("gitcheck.net.proxyhint"))
	}
	return r
}

// 9. 子模块内部状态（未提交/未跟踪/未推送的内容）
func checkSubmodules(dir string) checkResult {
	r := checkResult{name: i18n.T("gitcheck.check.submodules.name"), safe: true}
	out, err := git(dir, "submodule", "status")
	if err != nil {
		return r
	}
	for _, l := range lines(out) {
		if strings.HasPrefix(l, "-") {
			continue // 未初始化的子模块，本地没有内容可丢
		}
		fields := strings.Fields(strings.TrimLeft(l, "+ "))
		if len(fields) < 2 {
			continue
		}
		sub := filepath.Join(dir, fields[1])
		if _, err := os.Stat(sub); err != nil {
			continue
		}
		var probs []string
		if s := lines(must(git(sub, "status", "--porcelain"))); len(s) > 0 {
			probs = append(probs, i18n.T("gitcheck.submodules.dirty", len(s)))
		}
		subRemotes := remoteRefLines(sub)
		if len(subRemotes) == 0 {
			if len(lines(must(git(sub, "remote")))) == 0 {
				probs = append(probs, i18n.T("gitcheck.submodules.noremote"))
			}
		} else {
			if u := lines(must(git(sub, "rev-list", "--all", "--not", "--remotes"))); len(u) > 0 {
				probs = append(probs, i18n.T("gitcheck.submodules.unpushed", len(u)))
			}
		}
		if len(probs) > 0 {
			r.safe = false
			r.details = append(r.details, i18n.T("gitcheck.submodules.detail", fields[1], strings.Join(probs, i18n.T("gitcheck.submodules.joiner"))))
		}
	}
	if !r.safe {
		r.tag = i18n.T("gitcheck.submodules.tag", len(r.details))
	}
	return r
}

// 10. Git LFS：提交推送了不代表 LFS 大对象已上传
//
// 离线模式无法确认，只提示；--verify-remote 模式下用
// `git lfs push --dry-run --all` 探测远程缺失的对象。
func checkLFS(dir string) checkResult {
	r := checkResult{name: i18n.T("gitcheck.check.lfs.name"), safe: true}

	// 判断是否使用 LFS：git lfs ls-files 有结果，或 .gitattributes 声明了 filter=lfs
	usesLFS := false
	if _, err := exec.LookPath("git-lfs"); err == nil {
		if out, err := git(dir, "lfs", "ls-files", "-n"); err == nil && len(lines(out)) > 0 {
			usesLFS = true
		}
	}
	if !usesLFS {
		if data, err := os.ReadFile(filepath.Join(dir, ".gitattributes")); err == nil {
			if strings.Contains(string(data), "filter=lfs") {
				usesLFS = true
			}
		}
	}
	if !usesLFS {
		return r
	}

	if !verifyRemote {
		// 离线无法确认对象是否已上传，作为提示而非失败项
		r.details = append(r.details, i18n.T("gitcheck.lfs.offlinehint"))
		return r
	}
	if _, err := exec.LookPath("git-lfs"); err != nil {
		r.safe = false
		r.details = append(r.details, i18n.T("gitcheck.lfs.notool"))
		r.tag = i18n.T("gitcheck.lfs.notool.tag")
		return r
	}
	netFailed := false
	for _, remote := range lines(must(git(dir, "remote"))) {
		out, err := gitNet(dir, "lfs", "push", "--dry-run", "--all", remote)
		if err != nil {
			r.safe = false
			netFailed = true
			r.details = append(r.details, netErrorDetail(remote, err, i18n.T("gitcheck.lfs.verifyerr", remote, err)))
			continue
		}
		if pending := lines(out); len(pending) > 0 {
			r.safe = false
			r.details = append(r.details, i18n.T("gitcheck.lfs.pending", len(pending), remote))
		}
	}
	if !r.safe {
		r.tag = i18n.T("gitcheck.lfs.tag")
	}
	if netFailed {
		r.details = append(r.details, i18n.T("gitcheck.net.proxyhint"))
	}
	return r
}

// ---------- 检查执行 ----------

// timed 执行一项检查并记录耗时（详情模式 -v 下展示）。
func timed(f func(string) checkResult, dir string) checkResult {
	start := time.Now()
	r := f(dir)
	r.elapsed = time.Since(start)
	return r
}

// formatDur 把耗时格式化为 "850ms" / "1.5s"。
func formatDur(d time.Duration) string {
	if d < time.Second {
		return fmt.Sprintf("%dms", d.Milliseconds())
	}
	return fmt.Sprintf("%.1fs", d.Seconds())
}

func runChecks(dir string) []checkResult {
	checks := []checkResult{
		timed(checkRemote, dir),
		timed(checkUntracked, dir),
		timed(checkUncommitted, dir),
		timed(checkStash, dir),
		timed(checkBranches, dir),
		timed(checkDangling, dir),
		timed(checkTags, dir),
		timed(checkSubmodules, dir),
		timed(checkLFS, dir),
	}
	if verifyRemote {
		checks = append(checks, timed(checkRemoteAlive, dir))
	}
	return checks
}

func branchName(dir string) string {
	if b, err := git(dir, "symbolic-ref", "--short", "-q", "HEAD"); err == nil && strings.TrimSpace(b) != "" {
		return strings.TrimSpace(b)
	}
	return "detached"
}

func allSafe(checks []checkResult) bool {
	for _, c := range checks {
		if !c.safe {
			return false
		}
	}
	return true
}

// ---------- 目录扫描 ----------

func summaryTags(checks []checkResult) string {
	var tags []string
	for _, c := range checks {
		if !c.safe && c.tag != "" {
			tags = append(tags, c.tag)
		}
	}
	return strings.Join(tags, "")
}

// 查找所有最外层 git 仓库根（发现仓库后不再深入其内部）
func findRepos(root string) []string {
	var repos []string
	_ = filepath.WalkDir(root, func(path string, d fs.DirEntry, err error) error {
		if err != nil || !d.IsDir() {
			return nil
		}
		if st, err := os.Stat(filepath.Join(path, ".git")); err == nil && st.IsDir() {
			repos = append(repos, path)
			return filepath.SkipDir
		}
		return nil
	})
	return repos
}

func scanDir(root string, printClean, quiet, verbose bool) int {
	repos := findRepos(root)
	if len(repos) == 0 {
		fmt.Println(i18n.T("gitcheck.scan.norepos"))
		return 0
	}

	dirty, clean, exitCode := 0, 0, 0
	headerPrinted := false
	printHeader := func() {
		if !headerPrinted {
			fmt.Printf("%-8s  %-14s  %s\n", i18n.T("gitcheck.scan.col.status"), i18n.T("gitcheck.scan.col.branch"), i18n.T("gitcheck.scan.col.path"))
			headerPrinted = true
		}
	}

	for _, repo := range repos {
		checks := runChecks(repo)
		rel, _ := filepath.Rel(root, repo)
		if rel == "" {
			rel = "."
		}
		branch := branchName(repo)

		if !allSafe(checks) {
			dirty++
			exitCode = 1
			printHeader()
			fmt.Printf("%-8s  %-14s  %s  %s\n", i18n.T("gitcheck.scan.dirty"), "["+branch+"]", rel, summaryTags(checks))
			if verbose {
				for _, c := range checks {
					status := i18n.T("gitcheck.scan.ok")
					if !c.safe {
						status = i18n.T("gitcheck.scan.fail")
					}
					fmt.Printf("  %-6s %-24s %s\n", status, c.name, formatDur(c.elapsed))
					for _, d := range c.details {
						fmt.Printf("         - %s\n", d)
					}
				}
			}
		} else if printClean {
			clean++
			printHeader()
			fmt.Printf("%-8s  %-14s  %s\n", i18n.T("gitcheck.scan.clean"), "["+branch+"]", rel)
		}
	}

	if !quiet {
		fmt.Println(i18n.T("gitcheck.scan.summary", dirty, len(repos)-dirty, len(repos)))
	}
	_ = clean
	return exitCode
}

// ---------- main ----------

func usage() {
	fmt.Println(i18n.T("gitcheck.usage"))
}

// langFlag 预扫描 --lang（支持 "--lang zh" 和 "--lang=zh" 两种形式），
// 确保 usage/错误消息也能用对语言。
func langFlag(args []string) string {
	for i, a := range args {
		if a == "--lang" && i+1 < len(args) {
			return args[i+1]
		}
		if v, ok := strings.CutPrefix(a, "--lang="); ok {
			return v
		}
	}
	return ""
}

// Main runs the gitcheck command with args, the arguments after the
// program name. It is shared by cmd/gitcheck and "hall gitcheck".
func Main(args []string) {
	i18n.SetLang(i18n.Detect(langFlag(args)))

	dir := "."
	printClean, quiet, verbose := false, false, false

	for i := 0; i < len(args); i++ {
		a := args[i]
		switch a {
		case "--verify-remote":
			verifyRemote = true
		case "--strict":
			strict = true
		case "-a":
			printClean = true
		case "-q":
			quiet = true
		case "-v":
			verbose = true
		case "--lang":
			i++ // 值已在校验语言时消费
		case "-h", "--help":
			usage()
			os.Exit(0)
		default:
			if strings.HasPrefix(a, "--lang=") {
				continue // 值已在校验语言时消费
			}
			if strings.HasPrefix(a, "-") {
				fmt.Fprintln(os.Stderr, i18n.T("gitcheck.error.unknownflag", a))
				os.Exit(2)
			}
			dir = a
		}
	}

	if _, err := exec.LookPath("git"); err != nil {
		fmt.Fprintln(os.Stderr, i18n.T("gitcheck.error.nogit"))
		os.Exit(2)
	}
	abs, err := filepath.Abs(dir)
	if err != nil {
		fmt.Fprintln(os.Stderr, i18n.T("gitcheck.error.badpath", dir))
		os.Exit(2)
	}
	st, err := os.Stat(abs)
	if err != nil || !st.IsDir() {
		fmt.Fprintln(os.Stderr, i18n.T("gitcheck.error.nodir", dir))
		os.Exit(2)
	}
	if resolved, err := filepath.EvalSymlinks(abs); err == nil {
		abs = resolved
	}

	if !quiet {
		fmt.Print(i18n.T("gitcheck.scan.target", abs))
	}
	os.Exit(scanDir(abs, printClean, quiet, verbose))
}

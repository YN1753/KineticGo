package service

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"kineticgo/pkg/location"
	"math/big"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"strings"
	"time"

	"kineticgo/internal/model"
	"kineticgo/internal/ocr"
	"kineticgo/internal/repository"

	"github.com/PuerkitoBio/goquery"
)

const (
	signInUserAgent = "Mozilla/5.0 (Linux; Android 10; K) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Mobile Safari/537.36"
	casLoginURL     = "https://uias.suse.edu.cn/sso/login?service=https%3A%2F%2Fqfhy.suse.edu.cn%2Fsite%2Fappware%2Fsystem%2Fsso%2Flogin%3Ftarget%3Dhttps%3A%2F%2Fqfhy.suse.edu.cn%2Fxg%2Fapp%2Fqddk%2Fadmin%2Fqddkdk"

	rsaModulusHex  = "008aed7e057fe8f14c73550b0e6467b023616ddc8fa91846d2613cdb7f7621e3cada4cd5d812d627af6b87727ade4e26d26208b7326815941492b2204c3167ab2d53df1e3a2c9153bdb7c8c2e968df97a5e7e01cc410f92c4c2c2fba529b3ee988ebc1fca99ff5119e036d732c368acf8beba01aa2fdafa45b21e4de4928d0d403"
	rsaExponentHex = "010001"
	rsaChunkSize   = 126

	// 652 签到相关接口
	qfhyBase    = "https://qfhy.suse.edu.cn"
	uiasBase    = "https://uias.suse.edu.cn"
	qddkEntry   = "https://qfhy.suse.edu.cn/xg/app/qddk/admin/qddkdk"
	captchaURL  = "https://uias.suse.edu.cn/sso/captcha.jpg"
	checkInURL  = "https://qfhy.suse.edu.cn/site/qddk/qdrw/api/checkSignLocationWithPhoto.rst"
	taskListURL = "https://qfhy.suse.edu.cn/site/qddk/qdrw/api/myList.rst"
)

type signInConfig struct {
	Account  string `json:"account"`
	Password string `json:"password"`
	Local    string `json:"local"`
}

type SignInService struct {
	TaskRepo *repository.TaskRepository
	client   *http.Client
}

func NewSignInService(taskRepo *repository.TaskRepository) model.TaskInstance {
	jar, _ := cookiejar.New(nil)
	return &SignInService{
		TaskRepo: taskRepo,
		client: &http.Client{
			Jar:     jar,
			Timeout: 15 * time.Second,
		},
	}
}

func (s SignInService) Run(ctx context.Context, scheduleId uint) error {
	info := func(msg string) { TaskLog(ctx, LogInfo, msg) }

	cfg, err := s.loadConfig(scheduleId)
	if err != nil {
		return err
	}

	if cfg.Local == "" {
		return errors.New("未选择校区")
	}
	loc, ok := location.GetRandom(cfg.Local)
	if !ok {
		return fmt.Errorf("未知校区: %s", cfg.Local)
	}

	info("开始 652 自动签到 — " + cfg.Local + "校区")

	// 每次 Run 重置 CookieJar,避免上次失败残留污染本次会话
	jar, _ := cookiejar.New(nil)
	s.client.Jar = jar

	// 登录(含 OCR 验证码 + 1 次重试)
	if err := s.loginWithRetry(ctx, cfg); err != nil {
		return fmt.Errorf("登录失败: %w", err)
	}
	info("统一身份认证登录成功")

	// 执行签到
	if err := s.doCheckIn(ctx, loc); err != nil {
		return fmt.Errorf("签到失败: %w", err)
	}

	info("签到完成")
	return nil
}

func (s SignInService) Stop(scheduleId uint) error {
	return nil
}

func (s SignInService) loadConfig(scheduleId uint) (cfg signInConfig, err error) {
	schedule, err := s.TaskRepo.GetTaskScheduleById(scheduleId)
	if err != nil {
		return cfg, errors.New("获取任务配置失败")
	}
	if err = json.Unmarshal([]byte(schedule.Config), &cfg); err != nil {
		return cfg, errors.New("解析任务配置失败")
	}
	cfg.Account = strings.TrimSpace(cfg.Account)
	cfg.Password = strings.TrimSpace(cfg.Password)
	cfg.Local = strings.TrimSpace(cfg.Local)
	return cfg, nil
}

func (s SignInService) loginWithRetry(ctx context.Context, cfg signInConfig) error {
	warn := func(msg string) { TaskLog(ctx, LogWarn, msg) }

	for attempt := 1; attempt <= 2; attempt++ {
		if attempt > 1 {
			warn("首次登录尝试失败,准备重试")
			time.Sleep(time.Second)
		}
		if err := s.tryLogin(ctx, cfg); err == nil {
			return nil
		} else {
			warn(fmt.Sprintf("第%d次登录尝试失败: %s", attempt, err.Error()))
		}
	}
	return errors.New("两次登录尝试均失败")
}

// tryLogin 单次 CAS 登录流程.
func (s SignInService) tryLogin(ctx context.Context, cfg signInConfig) error {
	info := func(msg string) { TaskLog(ctx, LogInfo, msg) }

	// Step 1: 预访问获取 execution 令牌
	req1, _ := http.NewRequestWithContext(ctx, http.MethodGet, casLoginURL, nil)
	req1.Header.Set("Accept", "text/html,application/xhtml+xml,application/xml;q=0.9,*/*;q=0.8")
	req1.Header.Set("User-Agent", signInUserAgent)

	resp1, err := s.client.Do(req1)
	if err != nil {
		return fmt.Errorf("连接 CAS 登录页失败: %w", err)
	}
	defer resp1.Body.Close()

	doc, err := goquery.NewDocumentFromReader(resp1.Body)
	if err != nil {
		return fmt.Errorf("解析登录页 HTML 失败: %w", err)
	}
	execution, exists := doc.Find("form#fm1 input[name='execution']").Attr("value")
	if !exists {
		return errors.New("登录页未找到 execution 令牌")
	}

	// Step 2: 获取验证码
	req2, _ := http.NewRequestWithContext(ctx, http.MethodGet, captchaURL, nil)
	req2.Header.Set("Referer", casLoginURL)
	req2.Header.Set("User-Agent", signInUserAgent)

	resp2, err := s.client.Do(req2)
	if err != nil {
		return fmt.Errorf("获取验证码失败: %w", err)
	}
	defer resp2.Body.Close()

	imgBytes, err := io.ReadAll(resp2.Body)
	if err != nil {
		return fmt.Errorf("读取验证码图片失败: %w", err)
	}

	// Step 3: OCR 识别
	eng, err := ocr.GetGlobalEngine()
	if err != nil {
		return fmt.Errorf("初始化 OCR 引擎失败: %w", err)
	}
	captchaCode, err := eng.Recognize(imgBytes)
	if err != nil {
		return fmt.Errorf("OCR 识别失败: %w", err)
	}
	captchaCode = strings.TrimSpace(captchaCode)
	captchaCode = strings.Trim(captchaCode, "\x00")
	info("验证码识别结果: " + captchaCode)

	// Step 4: 加密密码
	encryptedPwd, err := encryptPassword(cfg.Password)
	if err != nil {
		return fmt.Errorf("密码加密失败: %w", err)
	}

	// Step 5: 提交登录表单
	formData := url.Values{}
	formData.Set("username", cfg.Account)
	formData.Set("password", encryptedPwd)
	formData.Set("authcode", captchaCode)
	formData.Set("execution", execution)
	formData.Set("encrypted", "true")
	formData.Set("_eventId", "submit")
	formData.Set("loginType", "1")
	formData.Set("rememberMe", "true")

	req3, _ := http.NewRequestWithContext(ctx, http.MethodPost, casLoginURL, strings.NewReader(formData.Encode()))
	req3.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req3.Header.Set("Accept", "text/html,application/xhtml+xml,application/xml;q=0.9,*/*;q=0.8")
	req3.Header.Set("Origin", uiasBase)
	req3.Header.Set("Referer", casLoginURL)
	req3.Header.Set("User-Agent", signInUserAgent)

	resp3, err := s.client.Do(req3)
	if err != nil {
		return fmt.Errorf("提交登录请求失败: %w", err)
	}
	defer resp3.Body.Close()

	// 判定: client 自动跟随 302,若最终落地在 qfhy 则成功
	if resp3.Request.URL.Host != "uias.suse.edu.cn" {
		return nil
	}

	// 仍在 uias,登录被拒
	body, _ := io.ReadAll(resp3.Body)
	return fmt.Errorf("CAS 拒绝登录: %s", string(body))
}

// doCheckIn 执行签到定位提交.
func (s SignInService) doCheckIn(ctx context.Context, loc location.SignInLocation) error {
	info := func(msg string) { TaskLog(ctx, LogInfo, msg) }

	// 会话预热
	warmupURLs := []string{
		qfhyBase + "/site/app/base/common/api/user/current.rst",
		qfhyBase + "/site/app/base/common/api/user/groups.rst?appCode=qddk",
		qfhyBase + "/site/app/base/common/api/group/qddk/identity.rst",
	}
	for _, u := range warmupURLs {
		req, _ := http.NewRequestWithContext(ctx, http.MethodGet, u, nil)
		setQfhyHeaders(req)
		if resp, err := s.client.Do(req); err == nil {
			resp.Body.Close()
		}
	}

	// 拉取今日任务
	qdrwId, qdxxId, err := s.fetchTaskIds(ctx)
	if err != nil {
		return err
	}
	info(fmt.Sprintf("获取任务: qdrwId=%v, qdxxId=%v", qdrwId, qdxxId))

	// 组装 Payload: 从 location 包获取基础数据,再补充任务相关字段
	payload := loc.ToPayload()
	payload["id"] = qdxxId
	payload["qdzt"] = 1
	payload["qdsj"] = time.Now().Format("2006-01-02 15:04:05")
	payload["isOuted"] = 0
	payload["isLated"] = 0

	jsonBytes, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("序列化签到数据失败: %w", err)
	}

	req, _ := http.NewRequestWithContext(ctx, http.MethodPost, checkInURL, bytes.NewReader(jsonBytes))
	setQfhyHeaders(req)
	req.Header.Set("Content-Type", "application/json;charset=UTF-8")
	req.Header.Set("Origin", qfhyBase)
	req.Header.Set("X-Requested-With", "com.tencent.mm")

	resp, err := s.client.Do(req)
	if err != nil {
		return fmt.Errorf("提交签到请求失败: %w", err)
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("签到接口返回 %d: %s", resp.StatusCode, string(body))
	}

	info("签到提交成功: " + string(body))
	return nil
}

// fetchTaskIds 从任务列表接口提取 qdrwId 和 qdxxId.
func (s SignInService) fetchTaskIds(ctx context.Context) (qdrwId, qdxxId any, err error) {
	find := func(status int) (any, any, bool) {
		urlStr := fmt.Sprintf("%s?status=%d&pageSize=1&pageNumber=1", taskListURL, status)
		req, _ := http.NewRequestWithContext(ctx, http.MethodGet, urlStr, nil)
		setQfhyHeaders(req)

		resp, e := s.client.Do(req)
		if e != nil {
			return nil, nil, false
		}
		defer resp.Body.Close()

		var result struct {
			Success bool `json:"success"`
			Result  struct {
				Data []map[string]any `json:"data"`
			} `json:"result"`
		}
		if json.NewDecoder(resp.Body).Decode(&result) != nil || !result.Success || len(result.Result.Data) == 0 {
			return nil, nil, false
		}
		item := result.Result.Data[0]

		var rw, xx any
		if v, ok := item["qdrwId"]; ok {
			rw = v
		} else if v, ok := item["rwId"]; ok {
			rw = v
		} else {
			rw = item["id"]
		}
		if v, ok := item["qdxxId"]; ok {
			xx = v
		} else if v, ok := item["id"]; ok {
			xx = v
		}
		return rw, xx, rw != nil && xx != nil
	}

	// 先查待签到(status=1)
	if rw, xx, ok := find(1); ok {
		return rw, xx, nil
	}
	//  fallback 查已签到(status=2)做时间覆盖
	if rw, xx, ok := find(2); ok {
		return rw, xx, nil
	}
	return nil, nil, errors.New("未找到今日签到任务")
}

// setQfhyHeaders 设置 qfhy 后端接口的通用请求头.
func setQfhyHeaders(req *http.Request) {
	req.Header.Set("User-Agent", signInUserAgent)
	req.Header.Set("Accept", "application/json, text/plain, */*")
	req.Header.Set("Referer", qddkEntry)
	req.Header.Set("appcode", "qddk")
}

// encryptPassword 使用学校硬编码 RSA 公钥分块加密密码.
func encryptPassword(password string) (string, error) {
	N := new(big.Int)
	if _, ok := N.SetString(rsaModulusHex, 16); !ok {
		return "", errors.New("解析 RSA 模数失败")
	}
	e := new(big.Int)
	if _, ok := e.SetString(rsaExponentHex, 16); !ok {
		return "", errors.New("解析 RSA 指数失败")
	}

	plainBytes := []byte(password)
	for len(plainBytes)%rsaChunkSize != 0 {
		plainBytes = append(plainBytes, 0)
	}

	var blocks []string
	for i := 0; i < len(plainBytes); i += rsaChunkSize {
		chunk := plainBytes[i : i+rsaChunkSize]
		reversed := make([]byte, len(chunk))
		for m := 0; m < len(chunk); m++ {
			reversed[m] = chunk[len(chunk)-1-m]
		}
		M := new(big.Int).SetBytes(reversed)
		C := new(big.Int).Exp(M, e, N)
		hexStr := fmt.Sprintf("%x", C)
		if len(hexStr) < 256 {
			hexStr = strings.Repeat("0", 256-len(hexStr)) + hexStr
		}
		blocks = append(blocks, hexStr)
	}
	return strings.Join(blocks, " "), nil
}

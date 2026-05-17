package service

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"math/big"
	"net/http"
	"net/url"
	"strings"

	"kineticgo/internal/model"
	"kineticgo/internal/repository"
)

const (
	probeURL      = "http://neverssl.com"
	loginURL      = "http://10.23.2.4/eportal/InterFace.do?method=login"
	portalPageURL = "http://10.23.2.4/eportal/index.jsp"
	rsaExponent   = 65537
	rsaModulesHex = "94dd2a8675fb779e6b9f7103698634cd400f27a154afa67af6166a43fc26417222a79506d34cacc7641946abda1785b7acf9910ad6a0978c91ec84d40b71d2891379af19ffb333e7517e390bd26ac312fe940c340466b4a5d4af1d65c3b5944078f96a1a51a5a53e4bc302818b7c9f63c4a1b07bd7d874cef1c3d4b2f5eb7871"
	userAgent     = "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36"
)

type SuseWifiService struct {
	TaskRepo *repository.TaskRepository
}

type suseWifiConfig struct {
	Address  string `json:"address"`
	Service  string `json:"service"`
	Username string `json:"username"`
	Password string `json:"password"`
}

func NewSuseWifiService(taskRepo *repository.TaskRepository) model.TaskInstance {
	return SuseWifiService{TaskRepo: taskRepo}
}

func (s SuseWifiService) Run(ctx context.Context, scheduleId uint) error {
	info := func(msg string) { TaskLog(ctx, LogInfo, msg) }
	warn := func(msg string) { TaskLog(ctx, LogWarn, msg) }

	cfg, err := s.loadConfig(scheduleId)
	if err != nil {
		return err
	}

	info("开始检测校园网状态")
	queryString, mac, err := s.probeGateway(ctx)
	if err != nil {
		return errors.New("探测失败: " + err.Error())
	}
	if queryString == "" {
		info("已联网，跳过认证")
		return nil
	}
	info("抓到 MAC: " + mac)

	body, err := s.login(ctx, cfg, queryString, mac)
	if err != nil {
		return errors.New("登录失败: " + err.Error())
	}
	if strings.Contains(body, "\"result\":\"success\"") {
		info("认证成功")
		return nil
	}
	warn("登录返回: " + body)
	return errors.New("认证失败: " + body)
}

func (s SuseWifiService) Stop(scheduleId uint) error {
	return nil
}

func (s SuseWifiService) loadConfig(scheduleId uint) (cfg suseWifiConfig, err error) {
	schedule, err := s.TaskRepo.GetTaskScheduleById(scheduleId)
	if err != nil {
		return cfg, errors.New("获取任务配置失败")
	}
	if err = json.Unmarshal([]byte(schedule.Config), &cfg); err != nil {
		return cfg, errors.New("解析任务配置失败")
	}
	// 防御性 trim：复制粘贴可能带进 \r\n、首尾空格、全角空格
	cfg.Address = strings.TrimSpace(cfg.Address)
	cfg.Service = strings.TrimSpace(cfg.Service)
	cfg.Username = strings.TrimSpace(cfg.Username)
	cfg.Password = strings.TrimSpace(cfg.Password)
	return cfg, nil
}

func (s SuseWifiService) probeGateway(ctx context.Context) (queryString, mac string, err error) {
	client := &http.Client{
		CheckRedirect: func(*http.Request, []*http.Request) error { return http.ErrUseLastResponse },
	}
	req, _ := http.NewRequestWithContext(ctx, http.MethodGet, probeURL, nil)
	resp, err := client.Do(req)
	if err != nil {
		return "", "", err
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	i := strings.Index(string(body), "?")
	if i == -1 {
		return "", "", nil
	}
	queryString = string(body)[i+1:]
	if j := strings.Index(queryString, "'"); j != -1 {
		queryString = queryString[:j]
	}

	mac = "11111111"
	if params, err := url.ParseQuery(queryString); err == nil {
		if m := params.Get("mac"); m != "" {
			mac = m
		}
	}
	return queryString, mac, nil
}

func (s SuseWifiService) login(ctx context.Context, cfg suseWifiConfig, queryString, mac string) (string, error) {
	client := &http.Client{}

	portalReq, _ := http.NewRequestWithContext(ctx, http.MethodGet, portalPageURL+"?"+queryString, nil)
	if resp, err := client.Do(portalReq); err != nil {
		return "", err
	} else {
		resp.Body.Close()
	}

	data := url.Values{}
	data.Set("userId", cfg.Username)
	data.Set("password", rsaEncrypt(cfg.Password+">"+mac))
	data.Set("service", url.QueryEscape(cfg.Address+cfg.Service))
	data.Set("queryString", queryString)
	data.Set("operatorPwd", "")
	data.Set("operatorUserId", "")
	data.Set("validcode", "")
	data.Set("passwordEncrypt", "true")

	req, _ := http.NewRequestWithContext(ctx, http.MethodPost, loginURL, strings.NewReader(data.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("User-Agent", userAgent)
	req.Header.Set("Referer", portalPageURL+"?"+queryString)

	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	out, _ := io.ReadAll(resp.Body)
	return string(out), nil
}

func rsaEncrypt(plaintext string) string {
	m := new(big.Int).SetBytes([]byte(plaintext))
	n := new(big.Int)
	n.SetString(rsaModulesHex, 16)
	c := new(big.Int).Exp(m, big.NewInt(rsaExponent), n)
	return fmt.Sprintf("%0256x", c)
}
